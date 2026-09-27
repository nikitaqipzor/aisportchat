# Railway staging deployment

This runbook deploys the Go API for phone testing. The Android application is
not hosted on Railway; a connected APK is built after the public API is healthy.

## 1. Create the Railway project

1. Create a separate staging project from `nikitaqipzor/aisportchat`. For the
   owner phone pilot, select the PR branch containing the candidate changes,
   rather than `main`. Keep production services and variables separate.
2. Add Railway PostgreSQL to the project.
3. Create the API service from the same GitHub repository.
4. Set `RAILWAY_DOCKERFILE_PATH=services/api/Dockerfile` on the API service.

Record the full 40-character Git commit SHA of the deployed candidate. The
branch name alone is insufficient because a later push can move it. Deploy
the exact candidate commit, wait for its migration and healthcheck to finish,
and compare Railway's deployed commit SHA with the recorded SHA before making
the APK. A green CI run on an earlier commit does not cover later changes.

The Docker build context must remain the repository root because the Dockerfile
copies `services/api` from the monorepo.

## 2. Configure API variables

Set these variables on the API service:

```text
APP_ENV=production
STORE_BACKEND=postgres
DATABASE_URL=${{Postgres.DATABASE_URL}}
AUTH_TOKEN_SECRET=<at-least-32-random-bytes>
PILOT_RELEASE_ID=<candidate-commit-sha>
MEDIA_ROOT=/data/media
OPENAI_API_KEY=<optional-for-internal-testing>
OPENAI_MODEL=gpt-5.6-terra
```

Set `PILOT_RELEASE_ID` to the deployed candidate's full commit SHA. It is a
public build identifier, not a secret; keep tokens and database URLs out of
this field. Generate the token secret locally with `openssl rand -hex 32`. Do
not commit it.
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
curl --fail --silent --show-error https://<service>.up.railway.app/healthz
```

Check that the JSON has `status: "ok"` and `release_id` exactly equal to the
deployed commit SHA. A 200 response alone can come from the wrong deployment.
Also probe `https://<service>.up.railway.app/api/v1/muscles` to confirm the
public domain reaches this API through the expected base path.

The mobile API base URL is:

```text
https://<service>.up.railway.app/api/v1
```

## 5. Build the connected APK

Once the candidate commit's CI is green, select that candidate branch in
GitHub Actions and run `connected-preview-apk`. Provide the full
mobile API base URL ending in `/api/v1` and its full expected release ID
(the deployed candidate commit SHA). Confirm that the workflow run's checkout
SHA is the same SHA. The preflight checks the server identity and API path,
then embeds the HTTPS URL into the Android build and publishes the artifact.
If the branch advances while building, stop and build again from the new
candidate only after redeploying that exact SHA.

Install that new APK for device testing. The generic internal APK uses
`https://preview.invalid/api/v1` and cannot test authentication or sync.

## Release gates

Before sharing outside the owner-only test group:

- enable PostgreSQL and volume backups;
- verify registration, login, refresh, deletion, photo upload and body scans;
- verify account isolation with two different users;
- review Railway logs without exposing tokens or health data;
- keep the service at one replica while local media uses a single volume.

After the owner pilot is accepted, merge through the normal review and CI
process. Redeploy staging/main from the merged commit and build a new APK
against that commit's release ID. Never repoint this pilot APK or staging
database at an existing production service during the experiment.
