# Body scan photo replacement repair — backend handoff

## Goal

Keep the previous private image usable if a recapture fails while updating database metadata, and remove newly uploaded private media when its database row was not committed.

## Changes

- `AddPhoto` now gives each upload an independent UUID based storage key. A failed replacement therefore cannot overwrite or delete the old file.
- On database failure it reads the current row with a bounded cleanup context. If the new key is absent, it deletes the newly uploaded blob and reports any cleanup failure. If the key was committed but the result read failed, it keeps the referenced blob and removes the replaced file.
- Following successful upsert it deletes the old blob. Deletion failure is returned while the new metadata and blob remain intact.
- Postgres photo upsert locks the owned parent scan within a transaction, checks its draft status, and holds the lock through insert or replacement. Scan completion's update conflicts with the same row lock.

## Tests and evidence

- Added Go tests for failed recapture with existing image, failed first insert and cleanup, successful replacement, committed upsert with failed result read, and old blob deletion failure.
- Added `POSTGRES_TEST_DSN` gated integration test for foreign user writes, completed scan replacement rejection, and preservation of original photo row.
- `gofmt` and `go test` could not be run in this environment: neither Go nor gofmt is installed on PATH or at `/usr/local/go/bin/go`. CI or a Go equipped environment must run `gofmt -w internal/bodyscan/service.go internal/bodyscan/service_test.go internal/store/postgres_bodyscan.go internal/store/postgres_bodyscan_repair_test.go`, followed by `go test ./internal/bodyscan ./internal/store` from `services/api` with optional `POSTGRES_TEST_DSN`.

## Risks and unresolved items

- If deletion of the old file fails after a committed replacement, the obsolete private file is orphaned. The API reports this cleanup error, but an automated retry needs durable cleanup state. No transaction queue was added under the agreed scope.
- If the database cannot verify whether a failed upsert committed, the new blob is retained to avoid breaking a potentially referenced private image. This rare ambiguous case likewise needs reconciliation.
- Concurrent recaptures of the same view across service instances may leave an intermediate orphan, since application level operations are not serialized around blob cleanup.

## Handoff

Lead engineer: review the database lock and failure semantics; run the Go commands above, including the integration test when PostgreSQL is available; route the result to QA.
