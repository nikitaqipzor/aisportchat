# Railway deployment handoff

## Goal

Make the Go API deployable to Railway for a connected Android staging test.

## Changes

- API now prefers Railway's injected `PORT` and retains `API_PORT` locally.
- Production image includes PostgreSQL client and all migrations.
- Migration runner supports Railway `DATABASE_URL` and existing Compose variables.
- Docker build context excludes the 319 MB Android workspace and unrelated files.
- Railway setup, PostgreSQL, media volume, healthcheck, and connected APK steps
  are documented in `docs/RAILWAY_DEPLOYMENT.md`.
- CI validates the Railway deployment contract.
- CI builds the exact Railway Docker image and exercises the migration runner
  against PostgreSQL through `DATABASE_URL`.

## Risks

- Railway resources and secrets still require an authenticated deployment.
- An attached volume causes a short redeploy interruption and requires one API
  replica while media is stored locally.
- Without `OPENAI_API_KEY`, AI endpoints use the deterministic local provider.

## Tests and evidence

- Railway deployment contract passes.
- Migration shell syntax and `git diff --check` pass.
- API contract: 82 OpenAPI operations / 83 routes plus `/healthz`.
- Migration contract: 19 up/down pairs.
- Android native verifier: 125 checks pass.
- Mobile syntax: 64 files, zero errors.
- Full Go tests and Docker build run in CI because this environment has neither
  Go nor Docker installed.

## Unresolved items

- Deploy and probe the real Railway domain.
- Run the connected APK workflow with the deployed `/api/v1` URL.
- Complete real-device account, media, body scan, nutrition, and sync smoke tests.

## Handoff

Lead/DevOps: deploy from `main`, verify CI and Railway `/healthz`, then build
`athletica-ai-connected-preview-apk`. QA: install the connected artifact; do not
test network flows with the generic `preview.invalid` APK.
