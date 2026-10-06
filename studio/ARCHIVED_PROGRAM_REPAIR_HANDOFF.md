# Archived program session workout repair

## Goal
Reject session workout creation for a program that is already archived or otherwise inactive, including sessions with a previously linked workout.

## Changes
- The HTTP handler loads the session's parent program from the authenticated user's store before returning an existing workout or generating a new one. Inactive programs return HTTP 409 without inserting a workout.
- Program session workout creation now uses `CreateProgramSessionWorkout` on the store. PostgreSQL uses one transaction: lock the parent program and session rows, check active status and absent link, insert workout and exercises, update session link, commit. Archive and replacement update the same program row, so they serialize with the transaction. Memory checks and writes under one mutex. A competing POST observes an existing link and returns the linked workout idempotently if the program remains active.
- Regression tests cover explicit archive, automatic archive through program replacement, stale linked workouts, and successful creation/idempotent retrieval on the replacement active program. Rejected requests compare workout history before and after; the unlinked session remains unlinked.
- OpenAPI now documents the 409 response for `POST /programs/sessions/{session_id}/workout` (mounted at `/api/v1`).

## Tests and evidence
- Go tests were not executable in this workspace because `go` and `gofmt` are absent. Run `cd services/api && go test ./internal/httpapi -run 'TestProgramSessionWorkoutRejectsArchived' -count=1` and the full API Go suite in CI.
- Added deterministic archive-during-generation test and memory/PostgreSQL (`POSTGRES_TEST_DSN`, set in CI) concurrent POST/POST/archive integration tests checking there is at most one workout, it is linked to the session, and a subsequent request returns 409. `git diff --check` is the available static whitespace check.

## Remaining risk
The PostgreSQL integration test requires `POSTGRES_TEST_DSN` pointing to a migrated test database; without it the test skips. CI configures this variable. The current handler's preliminary read and after-commit response assembly can overlap an archive, but the create/link decision itself is serialized and cannot produce an orphan. A 201 response can be followed immediately by an archive, which is expected when archive commits after workout creation.

## Handoff
Backend owner: review the transaction lock order and run Go tests, including PostgreSQL integration with `POSTGRES_TEST_DSN`. QA: verify archive and replacement flows before release acceptance.
