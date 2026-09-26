# Railway staging deployment

This runbook deploys the Go API for phone testing. The Android application is
not hosted on Railway; a connected APK is built after the public API is healthy.

## 1. Create the Railway project

1. Create a project from `nikitaqipzor/aisportchat` and select branch `main`.
2. Add Railway PostgreSQL to the project.
3. Create the API service from the same GitHub repository.
4. Set `RAILWAY_DOCKERFILE_PATH=services/api/Dockerfile` on the API service.

The Docker build context must remain the repository root because the Dockerfile
copies `services/api` from the monorepo.

## 2. Configure API variables

Set these variables on the API service:

```text
APP_ENV=production
STORE_BACKEND=postgres
DATABASE_URL=${{Postgres.DATABASE_URL}}
AUTH_TOKEN_SECRET=<at-least-32-random-bytes>
MEDIA_ROOT=/data/media
OPENAI_API_KEY=<optional-for-internal-testing>
OPENAI_MODEL=gpt-5.6-terra
```

Generate the token secret locally with `openssl rand -hex 32`. Do not commit it.
Railway injects `PORT`; the API prefers it over the local `API_PORT` fallback.

The authentication limiter uses the direct peer address unless the optional
`AUTH_TRUSTED_PROXY_CIDRS` variable explicitly trusts the ingress proxy. Determine and
verify the provider's actual source range before setting it; never trust every
private address or accept client-supplied `X-Forwarded-For` directly. Test two
separate phones behind the published URL and verify that one client's attempts
do not consume the other's IP bucket. The Docker Compose staging topology pins
Caddy to `172.30.239.10` and trusts that address only. Caddy explicitly replaces
incoming `X-Forwarded-For` with its direct client peer before proxying.

Redis is not required by the current API implementation. Add it only when a
runtime feature starts consuming `REDIS_ADDR`.

## 3. Configure migrations and media

In API service settings, set the pre-deploy command to:

```text
/migrations/run.sh
```

Set a finite pre-deploy timeout such as 300 seconds. The image includes both
the migration files and `psql`; the runner uses `DATABASE_URL` and records every
applied migration in `schema_migrations`.

Attach one persistent Railway volume to the API service at:

```text
/data/media
```

Without this volume, uploaded food photos and body-scan media disappear on a
redeploy. For a larger beta, move private media to S3-compatible object storage.

## 4. Public networking and healthcheck

Generate a Railway public domain for the API service and configure:

```text
Healthcheck path: /healthz
Healthcheck timeout: 300 seconds
Restart policy: ON_FAILURE
```

Verify the deployment using the generated HTTPS domain:

```bash
curl --fail https://<service>.up.railway.app/healthz
```

The mobile API base URL is:

```text
https://<service>.up.railway.app/api/v1
```

## 5. Build the connected APK

Open GitHub Actions, run `connected-preview-apk`, and provide the full mobile
API base URL ending in `/api/v1`. The workflow probes `/healthz`, embeds the
HTTPS URL into the Android build, and publishes the connected APK artifact.

Install that new APK for device testing. The generic internal APK uses
`https://preview.invalid/api/v1` and cannot test authentication or sync.

## Release gates

Before sharing outside the owner-only test group:

- enable PostgreSQL and volume backups;
- verify registration, login, refresh, deletion, photo upload and body scans;
- verify account isolation with two different users;
- review Railway logs without exposing tokens or health data;
- keep the service at one replica while local media uses a single volume.
