# Lead integration and QA handoff — 2026-09-25

## Goal
Repair confirmed cross-screen bugs from the full route audit, including the three P1 data integrity issues, and bring the changes to a reviewable PR state.

## Changes and ownership
- `archived_program` (gpt-6-sol): inactive program returns HTTP 409; workout insertion, exercise insertion, and session link are one PostgreSQL transaction under parent program and session row locks, or one memory store critical section. Archive and concurrent creation serialize. OpenAPI and memory/PostgreSQL regression cases updated.
- `bodyscan` (gpt-6-sol): unique media keys for every reshoot; the original photo survives failed metadata writes; new media cleanup verifies whether metadata committed. PostgreSQL locks the owned draft scan before upsert; regression tests cover failure, replacement and ownership.
- `active_workout` (gpt-6-sol): synchronous action guard blocks Finish/Cancel while saving; Android dialog dismiss unlocks it. Offline queue mutations and flushes are serialized; acknowledged IDs are removed without dropping concurrent writes. Behavioral screen and queue tests added.
- `nutrition_p2` (gpt-5.6-sol): custom product validation matches API; navigation retains food search and AI text drafts; hardware Back matches the visible Back path. Nutrition regression updated.
- `other_mobile_p2` (gpt-5.6-sol): profile accepts empty equipment without changing the persisted choice; AI Coach prevents duplicate requests and serializes local chat saves; Technique rejects zero repetitions; program detail drops stale responses, and program list requests the server's maximum valid page size (50). Program and technique regressions updated.
- Lead: corrected repeat-workout Back to retain the original detail ID; cleared nutrition drafts on logout or session change; Finish/Cancel report non-network failures to the action screen and queue terminal actions after pending sets. Added device display and new async tests to CI and `scripts/verify.sh`.

## Two-stage review
1. **Implementation review:** examined transaction order and user ownership in Go, private image cleanup and committed-write ambiguity, synchronous mobile action guards, queue concurrency, API field limits, and screen navigation. Returned four issues to owners: Android Alert dismiss, profile's real empty equipment choice, AI chat persistence ordering, and stale source assertions; all were corrected. App's repeat target and terminal offline ordering were integrated by lead.
2. **Regression review:** all 17 `scripts/test-mobile-*.mjs` scenarios passed; `ActiveWorkoutScreen.test.mjs` and `offlineQueue.test.mjs` passed; mobile syntax 66/0, TypeScript typecheck, Android native static checks 125, and `git diff --check` passed. The new behavior tests run in CI's Android job after mobile dependencies install; `test-mobile-device-display.mjs` runs in static mobile CI. PostgreSQL integration uses `POSTGRES_TEST_DSN`, configured by CI. Go/gofmt are absent here, so the Go tests, race tests and migration integration **were not run locally**. CI must pass before merge.

## Risks and release gate
- A failed deletion of a replaced BodyScan blob, concurrent reshoots, or an ambiguous metadata read after commit can leave private orphaned media. A durable cleanup/reconciliation job is needed to guarantee eventual deletion.
- Program history still stops at 50 items: the existing API has no cursor or offset. Showing all programs beyond 50 requires a paginated API/store/mobile flow. The previous 10-item cutoff is fixed for histories of at most 50.
- Ambiguous server success before offline queue acknowledgment can replay an operation. Confirm terminal operation idempotency before considering the offline flow fully reliable.
- Device taps, camera, Health Connect, loss and restoration of connectivity, and actual APK functionality are not proven by static tests. Run them against a deployed API on a device. CI Go and Android build are mandatory gates before PR acceptance.

## Handoff
Root/remote PR owner: freeze local edits, run remote CI including Go race + PostgreSQL concurrency, check Android APK build; if green, review on-device journeys. Keep PR as draft until the release gates and user device pass are complete.
