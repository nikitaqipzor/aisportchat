# Connected APK pilot increment — lead handoff

## Goal

Bind a phone pilot APK to the intended staging API source revision and prevent
a healthy but stale/wrong public endpoint from passing the build preflight.

## Scope and changes

- `GET /healthz` retains `status`/`time` and reports optional, trimmed,
  non-secret `release_id` from `PILOT_RELEASE_ID`; OpenAPI and contract checker
  cover the root-level health endpoint.
- Connected APK dispatch requires a full 40-character expected commit SHA.
  The preflight checks that it equals the workflow checkout SHA, validates a
  canonical HTTPS `/api/v1` URL, checks exact health release identity, and
  probes the public muscles API. Redirects and unexpected JSON fail the build.
- Offline preflight fixtures run in the deployment CI job and `verify.sh`.
- Railway and Compose staging runbooks now start from the candidate PR branch
  and record/check its immutable SHA. Compose passes `PILOT_RELEASE_ID`.
  Production/main is not touched during the owner pilot.

## Two-stage review

1. **Design:** required a non-secret exact release marker, explicit SHA input,
   checkout SHA binding, expected API route, and separate staging from the
   current main deployment. The ID is operator-configured, so Railway's
   actual deployed commit must also be checked independently.
2. **Code:** inspected Go response, OpenAPI root server override, shell URL
   parser, curl status/redirect behavior, endpoint fixture shape, workflow
   wiring and runbook consistency. Fixed a missing script executable bit;
   tightened the ID to a full lowercase SHA and added bounded retries.

## Tests and evidence

- PASS: `bash scripts/test-preflight-connected-api.sh`.
- PASS: `python3 scripts/verify-api-contract.py` (84 documented/84 routed).
- PASS: `python3 scripts/verify-migrations.py` (23 pairs).
- PASS: `python3 scripts/verify_android_native.py` (125 checks).
- PASS: Bash syntax check on changed shell scripts, YAML parse of workflow,
  OpenAPI and Compose, and `git diff --check`.
- NOT RUN: Go formatting/tests, Docker Compose contract/render, real Railway
  deployment, connected APK build, and physical phone smoke. This workspace
  has no Go/gofmt or Docker, and no staging URL or secrets were supplied.

## Risks and unresolved items

- `PILOT_RELEASE_ID` is a manual staging setting. Matching it to the workflow
  SHA is useful only after an operator verifies Railway's deployed Git commit
  is that same SHA; a mistaken marker can label old code as new.
- `/healthz` is an HTTP process response and does not prove database queries,
  media persistence or account flows. The API catalog probe checks routing,
  not authenticated behavior.
- The prior green PR evidence for `39c1b8c` does not validate this uncommitted
  increment. CI must run on the eventual candidate SHA, including Go,
  PostgreSQL and Android gates, before an APK is shared.
- Owner must supply a separate staging service/domain, database, media volume
  and secrets, then run the connected workflow and two-account phone smoke.

## Acceptance and handoff

**Source review complete; release blocked.** No commit, push, merge or deploy
was performed. Lead/DevOps: commit the reviewed candidate through the normal
PR process, rerun CI, deploy the same SHA to separate staging and verify its
reported deployed commit and `/healthz` ID. QA: use only the resulting
connected APK for registration, profile, workout/history, media, deletion and
cross-account device smoke. Record device model, Android version and failures.
