# Body scan media reliability — coder handoff

## Goal

Recover private body scan files left by concurrent reshoots, failed deletion, interrupted uploads, and scan deletion without deleting the active photo.

## Changes

- Migration `000022` adds a per-key cleanup outbox. Upload keys are staged before writing bytes, and staged orphan keys become eligible after one hour.
- PostgreSQL photo replacement locks the scan and staged key, swaps metadata, queues the prior key, and removes the new key's stage in one transaction. Scan deletion queues all referenced keys and removes metadata in one transaction.
- A cleanup worker runs at startup and each minute. It holds a queue row lock across the reference check and file deletion. Activation of a staged key requires that lock, so cleanup cannot race an active photo into deletion. Failed deletion keeps the job for retry.
- The memory store mirrors this behavior. Account removal makes staged jobs immediately eligible; the worker also expedites jobs for deleted PostgreSQL users.
- The service immediately retries eligible cleanup after reshoot or scan deletion. A failed immediate cleanup returns an error, while the durable job remains for the worker. A failed upload or crashed request can leave a staged file for up to an hour.

## Tests and evidence

- Added concurrent reshoot, failed deletion retry, active-reference protection, and deleted-account queue tests; updated the PostgreSQL body scan owner/draft test to stage keys.
- `git diff --check` passed.
- Go tests and `gofmt` could not run: neither binary is installed in this execution environment. The lead/QA stage must run `go test ./internal/bodyscan ./internal/store ./cmd/api` from `services/api`, including PostgreSQL integration with `POSTGRES_TEST_DSN` when available, and format touched Go files.

## Risks and unresolved items

- An interrupted upload can remain private on disk for up to one hour before its staged job is eligible. If the account is deleted, the job becomes eligible immediately.
- The worker handles at most 100 eligible keys per minute. A large backlog drains over several runs.
- PostgreSQL file deletion occurs while holding a database transaction and queue row lock; the 45-second worker attempt and 10-second request cleanup limit stall time. This should be load tested with remote media storage if one is added.
- The API response shape is unchanged. A scan delete may return a cleanup error after metadata is removed; the worker retries that queued file deletion automatically.

## Handoff

QA: run Go tests, PostgreSQL integration, and formatting. Lead Engineer: review deletion lifecycle and accept once tests pass.
