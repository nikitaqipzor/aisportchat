# Closed pilot backend/security/operations lead review

## Goal

Review the delegated backend, security and operations fixes together without contacting staging/main or making a release claim.

## Changes accepted for candidate review

- Atomic refresh-session consumption in memory and PostgreSQL blocks concurrent reuse and never issues replacement tokens when consumption fails.
- File media storage rejects unsafe keys and symlinked paths and uses independent temporary files for concurrent writes.
- Workout service and direct store paths reject NaN/infinite set measurements; memory account deletion removes serialized workout-operation receipts under the operation lock.
- Migration setup and each migration use the same advisory lock, with the ledger rechecked inside the locked transaction. The destructive down/up smoke requires explicit opt-in and CI exercises concurrent runners.
- `/readyz` checks the backing PostgreSQL connection where the store supports Ping; staging Compose probes it. `/healthz` remains liveness and release identity. The API contract includes `/readyz`.
- Pilot backup/recovery boundaries are documented in `docs/PILOT_OPERATIONS.md`.

## Independent review and evidence

- Reviewed the combined diff for owner isolation, lock ordering, conditional refresh update, migration lock scope, and staging probe dependency. Requested and verified two second-stage corrections: ledger creation now shares the advisory lock, and direct store workout writes now reject nonfinite values.
- `git diff --check`: pass.
- `python3 scripts/verify-api-contract.py`: pass, 85 operations and 85 routes.
- `python3 scripts/verify-migrations.py`: pass, 23 contiguous up/down pairs.
- `bash -n` on modified Bash scripts and `sh -n services/api/migrations/run.sh`: pass.
- Destructive migration smoke without opt-in: exits 2 before database access.
- Go tests, gofmt, PostgreSQL migration smoke/integration, Docker build/Compose rendering: not executed because Go, psql and Docker are absent here. New tests are present but remain unverified at runtime.

## Risks and release boundary

- Do not accept a pilot release until CI executes Go race tests, PostgreSQL migration/refresh tests and Docker build/Compose checks on this exact candidate, and operators test an encrypted, consistent database-plus-media backup and restore. Backups and retention are currently manual.
- File path checks assume an API-owned media directory; a same-UID local attacker changing directories concurrently could race checks. PostgreSQL Ping uses synchronous libpq calls, so the `/readyz` request context alone cannot guarantee a two-second bound under a hung connection.
- No commit, push, staging/main mutation or deployment was performed.

## Handoff

QA/release operator: run the unexecuted gates in a Go/PostgreSQL/Docker environment, capture the backup/restore result, then return evidence to the Lead Engineer for release acceptance.
