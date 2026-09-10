# Staging deployment and connected Android preview

This runbook turns the source-only preview into a phone-testable staging system. It keeps PostgreSQL and media on persistent Docker volumes, runs every migration before the API starts, rejects development secrets in production mode, and terminates HTTPS through Caddy.

## Server prerequisites

- a Linux VPS with Docker Engine and the Compose plugin;
- ports 80 and 443 reachable from the internet;
- a DNS A/AAAA record such as `api.example.com` pointing to the VPS;
- at least 4 GB RAM and 20 GB persistent disk for an early private beta.

## First deployment

```bash
git clone https://github.com/nikitaqipzor/aisportchat.git
cd aisportchat
cp infra/staging/.env.example infra/staging/.env
```

Edit `.env`. Generate URL-safe secrets with `openssl rand -hex 24` for PostgreSQL and `openssl rand -hex 32` for tokens. Never commit `.env`.

```bash
docker compose --env-file infra/staging/.env -f infra/staging/docker-compose.yml up -d --build
docker compose --env-file infra/staging/.env -f infra/staging/docker-compose.yml ps
curl --fail https://api.example.com/healthz
```

Caddy requests and renews the TLS certificate automatically after DNS resolves and ports 80/443 are open. The public mobile API URL is `https://api.example.com/api/v1`.

## Updating staging

Back up PostgreSQL and uploaded media before deployment. Then:

```bash
git pull --ff-only origin main
docker compose --env-file infra/staging/.env -f infra/staging/docker-compose.yml build api
docker compose --env-file infra/staging/.env -f infra/staging/docker-compose.yml up -d
curl --fail https://api.example.com/healthz
```

The migration container is idempotent and must finish successfully before the API starts. Do not manually mark a migration as applied.

## Building an APK connected to staging

In GitHub open **Actions → connected-preview-apk → Run workflow** and enter the full URL ending in `/api/v1`, for example:

```text
https://api.example.com/api/v1
```

The workflow rejects plain HTTP, probes `/healthz`, type-checks the app, builds a standalone APK with the URL embedded in `BuildConfig`, and uploads it as `athletica-ai-connected-preview-apk` for 14 days.

## Recovery and data safety

- Database data is in the `postgres_data` volume.
- Body/media files are in the `media_data` volume.
- Caddy certificates are in `caddy_data`.
- `docker compose down` preserves volumes; never add `--volumes` during routine updates.
- Rotate `AUTH_TOKEN_SECRET` only with an explicit session invalidation plan because existing refresh tokens will stop working.

This is a private-beta staging topology, not a claim of production readiness. Device smoke testing and backups remain release gates.
