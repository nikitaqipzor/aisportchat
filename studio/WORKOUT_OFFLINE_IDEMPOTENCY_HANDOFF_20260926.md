# Workout offline operation handoff — 2026-09-26

## Goal

Make offline log-set, finish, and cancel retries safe after a response is lost, keyed by the queue's persisted operation ID.

## Changes

- Mobile sync forwards each persisted `operation.id` as `Idempotency-Key`. Failed calls retain the same queued operation. Legacy direct callers omit the header.
- Workout endpoints validate bounded ASCII keys and bind each key to user, action, workout, and canonical set payload. Reuse with a different request returns 409.
- Migration 000023 adds user-scoped receipts that live until account deletion. PostgreSQL claims the receipt, locks the workout, mutates, captures a result snapshot, and completes the receipt in one transaction. Same-ID retry returns its original snapshot. Memory store mirrors the receipt behavior for tests.
- OpenAPI documents the optional header and conflict semantics.

## Risks and limits

- A new distinct operation for the same set arriving out of order across different devices can still overwrite another device's edit. Receipt deduplication guarantees only that a *previously applied* operation cannot overwrite a later edit on replay.
- Existing headerless clients retain prior behavior. They do not acquire retry deduplication until they send stable keys.
- The store snapshot is durable, while progression recommendations and personal record lists in the service response are recomputed during a finish replay.

## Evidence

- `node scripts/test-mobile-workout-offline-operations.mjs` — PASS (ambiguous response retains key; retries and terminal actions pass it).
- `python3 scripts/verify-api-contract.py` — PASS.
- `python3 scripts/verify-migrations.py` — PASS, migration sequence 000001–000023.
- Go tests were not executable in this workspace because `go` and `gofmt` were absent. Added `internal/store/workout_operations_test.go` for concurrent same-key calls, stale replay, payload binding, finish replay, and cancel replay.

## Unresolved items and handoff

Lead Engineer and QA: run Go tests and PostgreSQL migration smoke in a Go-capable environment, review the PostgreSQL transaction and service replay result, and wire `scripts/test-mobile-workout-offline-operations.mjs` into both CI and `scripts/verify.sh` (lead is coordinating those shared files).
