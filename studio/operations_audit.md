# Closed pilot operations audit

## Goal

Reduce migration race and unsafe smoke-test risk, verify Compose readiness wiring,
and document the pilot backup/recovery boundary before a closed pilot.

## Changes

- `services/api/migrations/run.sh` serializes migration-ledger initialization
  and each migration transaction with a shared PostgreSQL transaction advisory
  lock. Each runner rechecks the migration ledger after acquiring the lock, so
  a concurrent runner skips already committed work.
- The PostgreSQL migration smoke now exercises two concurrent copies of the
  production runner. It requires `ALLOW_DESTRUCTIVE_MIGRATION_SMOKE=1` because
  it intentionally applies all down migrations and removes the migration
  ledger. CI sets this only for its disposable PostgreSQL service.
- The staging API image installs `curl`, and Compose uses it to probe
  `/readyz`. The staging contract verifier checks the Compose probe, image
  dependency and route registration.
- Added `docs/PILOT_OPERATIONS.md` with backup protection, consistency, restore
  and release identity guidance.

## Risks

- The repository does not automate backups, encryption, retention, or a
  coordinated database/media restore. A PostgreSQL dump and `/data/media`
  copy must be protected together and restored in isolation before operators
  switch staging over.
- Smoke-test migrations are destructive by design. Operators must run them
  only against a disposable database and explicitly set the opt-in variable.
- `/readyz` checks HTTP readiness and PostgreSQL reachability; it does not
  establish media availability, backup recoverability or end-to-end behavior.

## Tests and evidence

- PASS: `python3 scripts/verify-migrations.py` (23 up/down pairs plus runner
  lock, under-lock recheck, smoke opt-in and concurrent-runner assertions).
- PASS: `bash -n` on the migration runner and changed shell scripts.
- PASS: staging readiness YAML/static wiring inspection, including `/readyz`
  route registration and installed `curl`.
- PASS: `git diff --check`.
- PASS: safety check that migration smoke exits before database access without
  the destructive opt-in.
- NOT RUN: real PostgreSQL concurrent/up/down/up smoke; neither `psql` nor
  Docker is installed in this workspace.
- NOT RUN: `scripts/verify-staging-deploy.sh` Compose rendering; Docker is not
  installed. No staging URL or credentials were supplied, and no staging/main
  system was contacted.

## Unresolved items

- Provision and test protected backups and a database-plus-media restore under
  the pilot's approved retention policy.
- Run CI PostgreSQL smoke and Docker Compose contract/build checks on the
  reviewed candidate SHA.
- Verify actual deployed Git SHA independently of `PILOT_RELEASE_ID`, then
  exercise `/readyz` and connected API/mobile flows against separate staging.
- Physical device smoke, pilot privacy notice/consent, and external deletion
  retention evidence remain outside this operations change.

## Handoff

Lead Engineer / independent QA: review migration lock semantics and the
destructive smoke guard, then require PostgreSQL and Compose CI evidence before
release acceptance. Operations: establish and test encrypted backups and
restore before inviting pilot users. No commit, push or deployment was made.
