# Lead review: pilot reliability increment

## Goal
Make program history usable past 50 records, retain private body-scan media cleanup across failures and concurrent edits, and make offline workout retries safe after an ambiguous network acknowledgement.

## Changes accepted for integration review
- Program history uses a `(updated_at, id)` cursor in the API/store and a load-more UI with deduplication and retry.
- Body-scan upload stages each unique blob key, swaps references and queues previous keys in one transaction, and retries deletion with queue-row synchronization. A worker retries on startup and every minute.
- Mobile sends persisted offline operation IDs; API/store apply each keyed workout operation and save its result in a user-scoped receipt with the mutation. Different payload reuse is rejected.
- New behavioral tests were added to `scripts/verify.sh` and `.github/workflows/ci.yml`.

## Two-stage review
1. Design: rejected offset pagination for keyset ordering; required an outbox-row lock across media deletion to prevent a staged key becoming active during cleanup; required atomic operation receipt and mutation with payload binding.
2. Code: identified and corrected a PostgreSQL log-set versus finish race by locking the workout row before keyed mutation. Inspected shared `store/memory.go`, router, migrations, OpenAPI, mobile UX, and account-deletion interaction. No broad rewrites of shared files were made during parallel work.

## Tests and evidence
- Local PASS: `node apps/mobile/src/screens/ProgramsScreen.test.mjs`, `node scripts/test-mobile-workout-offline-operations.mjs`, `node scripts/test-mobile-program-resume.mjs`, `node scripts/test-mobile-core-screens.mjs`.
- Local PASS: `npm run typecheck --if-present` in `apps/mobile`, `python3 scripts/verify_android_native.py`, `python3 scripts/verify-api-contract.py`, `python3 scripts/verify-migrations.py`, `git diff --check`.
- Local NOT RUN: Go unit/integration/race tests and gofmt, because `go` and `gofmt` are not installed in this workspace. PostgreSQL integration tests additionally require `POSTGRES_TEST_DSN`.
- Remote gate: run CI Go/PostgreSQL suites and the wired mobile checks before release acceptance.

## Risks and unresolved items
- Keyset pagination does not provide a frozen snapshot when unseen programs are updated during paging; a moved row can cross the cursor boundary.
- Abandoned uploads intentionally wait up to one hour before retry cleanup. Failed media deletion remains queued and may delay physical removal; the worker logs failures.
- Durable receipts live for the account lifetime and grow with offline operations. Account deletion cascades receipts; monitor table growth.
- Legacy workout requests without an idempotency key keep previous semantics; only queued operations with IDs receive replay guarantees.
- Local static checks cannot establish Go compilation, PostgreSQL transaction behavior, or Android device behavior.

## Handoff
To root integrator / QA: review the three feature handoffs, run remote Go and PostgreSQL CI gates, confirm no shared-file loss in final diff, then decide whether this increment can be accepted for the pilot. No commit or push performed by this lead.
