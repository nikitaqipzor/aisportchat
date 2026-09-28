# Nutrition flow repair — engineering handoff

## Goal
Keep diary writes safe across lost HTTP responses and make recipe and barcode search recoverable.

## Changes
- Manual food, repeat, and undo now use owner-scoped persistent operation keys. The API routes those writes through the existing transactional food batch store; a retry reuses the previous operation result.
- Undo remains available after a failed restore, including after its usual dismissal timer would have elapsed; the button is disabled while a restore is in flight.
- Barcode errors retry barcode lookup. Editing ingredient search invalidates older results, and a successful recipe creation leaves its draft visible while list refresh is unavailable without allowing another creation.
- OpenAPI describes the new optional keys. Existing clients can omit them.

## Evidence
- `apps/mobile/node_modules/.bin/tsc --noEmit -p apps/mobile/tsconfig.json` passed.
- `node scripts/test-mobile-food-idempotency.mjs` and `node scripts/test-mobile-nutrition-regressions.mjs` passed.
- `python3 scripts/verify_android_native.py` passed 125 checks.
- `git diff --check` passed.
- `services/api/internal/nutrition/food_batch_test.go` now verifies manual and repeat retries; Go is not installed locally, so backend test execution is pending CI.

## Risks and unresolved items
- Recipe *creation* still has no server operation key. If its successful response is lost, retrying creation may duplicate the saved recipe. It needs its own backend operation contract.
- Authentication, date changes, and real network interruptions still need on-device integration testing.

## Handoff
QA: run the new Go test and end-to-end network-loss scenarios; lead engineer: review route changes and release readiness.
