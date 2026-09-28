# Backend data integrity handoff

## Goal
Close concrete workout data integrity and account deletion privacy gaps before the pilot.

## Changes and findings
- Reject NaN and infinite weight, RPE and RIR in `workouts.Service.logSet` for both keyed and legacy set writes. Range comparisons alone accepted NaN; in memory a completed workout could then have NaN total volume.
- Reject nonfinite measurements at the keyed store boundary before a memory mutation. Previously a memory set could be written and then `json.Marshal` of its operation receipt would fail, leaving an unreceipted mutation. PostgreSQL now rejects it before starting the transaction as well.
- Reject nonfinite measurements at both direct `UpsertWorkoutSet` store implementations. This closes the legacy store path identified during lead review.
- Remove memory workout operation receipts during account deletion. They contain serialized workout details and previously survived deletion. Deletion acquires `workoutOperationMu` before `mu`, matching the keyed operation lock order; it releases the former before invoking the media cleanup callback.
- Reviewed food batch and program history implementations and existing tests. No additional fix was supported strongly enough to change those paths in this pass.

## Tests and evidence
- Added `TestLogSetRejectsNonFiniteMeasurementsBeforeWriting`, `TestInvalidWorkoutOperationDoesNotWriteSetOrClaimReceipt` (also checks direct memory store write), and `TestDeleteAccountRemovesWorkoutOperationReceipts`.
- `git diff --check` passed.
- `go test ./internal/workouts ./internal/store` could not run: neither `go` nor `gofmt` exists in this workspace (`gofmt: command not found`). PostgreSQL integration tests also require `POSTGRES_TEST_DSN`, which is unset here. Changes are statically reviewed but not compiled in this environment.

## Risks and unresolved items
- Independent reviewer should run gofmt and the two Go package tests in a Go toolchain, and rerun PostgreSQL integration tests with migrated test DB if available.
- PostgreSQL direct store guard has static review only; its integration behavior needs the Go/PostgreSQL test environment.
- Account deletion receipt cleanup is memory only; PostgreSQL receipt rows already cascade on user deletion through the `workout_operations.user_id` foreign key.

## Handoff
Lead Engineer / independent QA: review the scoped patch and tests, run the Go verification gates, then decide pilot acceptance. No commit, push, deployment or staging mutation performed.
