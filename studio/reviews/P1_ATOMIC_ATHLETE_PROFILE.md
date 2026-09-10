# P1 — atomic athlete profile handoff

## Goal

Remove partial athlete-profile saves by replacing the mobile screen's three sequential writes with one authenticated aggregate transaction.

## Changes

- Added `PUT /api/v1/profile/athlete` for body measurements/details, goal and training preferences.
- The handler validates the complete aggregate before calling storage and derives ownership only from the authenticated context. Nested `user_id` values are ignored.
- Memory storage replaces all three records under one mutex.
- PostgreSQL stores all profile records and training join rows in one transaction; a failure at any step rolls the complete update back.
- Mobile exposes `api.setAthleteProfile` and the profile screen now saves through that single request.
- OpenAPI documents the aggregate request/response.

## Tests / evidence

- Router coverage saves the aggregate, verifies client-supplied ownership cannot escape auth scope, rejects an invalid equipment id and confirms the earlier profile/goal/training records remain unchanged.
- PostgreSQL integration coverage forces a late equipment foreign-key failure and verifies profile and goal rollback.
- `python3 scripts/verify-api-contract.py` — passed, 82 OpenAPI operations aligned.
- `cd apps/mobile && npm ci && npm run typecheck` — passed.
- `python3 scripts/verify_android_native.py` — passed, 125 checks.
- Go tests and `gofmt` were not executed locally because this environment has no Go toolchain; CI must run them with PostgreSQL.

## Risks

- The legacy individual profile/goal/training endpoints remain for onboarding and backward compatibility. Only the post-onboarding athlete profile screen uses the aggregate write.
- PostgreSQL rollback coverage requires `POSTGRES_TEST_DSN` and CGO/libpq in CI.

## Unresolved items

- Physical-device save/retry testing remains part of the connected APK pass.

## Handoff

QA: run `go test ./...` plus the PostgreSQL integration suite, then verify one successful save and a forced offline retry on-device. Lead: merge this package before navigation/session integrations; it does not modify `App.tsx`, session handling or the shared request transport.
