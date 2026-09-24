# Food writes: atomicity and retry handoff

## Goal

Prevent partial meals and duplicate foods when the user retries recipe logging or AI food confirmation after an uncertain network response.

## Changes

- Both endpoints prevalidate every ingredient, then call `Store.CreateFoodEntries` once. Memory writes under one lock; PostgreSQL writes entries and the user-scoped operation marker in one transaction (`000020_food_operations`).
- An optional `idempotency_key` preserves compatibility with existing callers. The mobile app supplies a stable key per user and draft, including after app restart and an arbitrarily long timeout; a confirmed success clears it. Changing the meal starts a new action, and concurrent identical taps share one request.
- PostgreSQL refuses to reconnect inside an open transaction. Otherwise a connection loss could move later statements into autocommit on another session.
- A changed payload with the same key returns a conflict. The API keeps calculating calories and macros from catalog foods.
- Logout clears pending keys for that owner only. IANA `time_zone` determines the returned day; omitted parameter retains UTC behavior.

## Evidence and unresolved items

- Local `npm run typecheck`, `test-mobile-food-idempotency.mjs`, `test-mobile-nutrition-regressions.mjs` and `git diff --check` passed.
- New Go tests cover prevalidation, rollback, replay, concurrent PostgreSQL connections, connection loss, and HTTP endpoints. Go and PostgreSQL are unavailable in this workspace; these tests must pass in the GitHub CI backend job before accepting the change.
- Offline food writes still await a network connection; idempotency preserves safe retries but does not implement a durable offline queue. An ambiguous failed request intentionally retains its key until a retry confirms success or the user changes the meal.

## Handoff

Lead engineer: review CI backend and PostgreSQL integration results before merging into `main`; then verify a physical Android reconnect scenario.
