# Backend QA — stages 2, 4 and 6

## Goal

Independently review the athlete-profile, manual-workout and training-progress backend slices for persistence parity, owner isolation, deterministic validation, router safety and OpenAPI consistency.

## Findings and changes

- Fixed PostgreSQL profile PATCH semantics. Omitted `unit_system`, `injuries` and `limitations` now preserve stored values, matching the memory store; explicit empty arrays still clear the private notes.
- Added a PostgreSQL integration regression covering the preserve/clear round trip.
- Stopped `GET /profile` from silently returning an empty profile when profile storage fails; it now returns a generic server error without leaking the database error.
- Confirmed manual workout creation derives ownership exclusively from the authenticated context and all subsequent reads are owner-scoped. Added a cross-user service regression.
- Confirmed manual workout bounds, duplicate IDs, catalog IDs and enabled training environment are validated before persistence. AI is not involved in this facts path.
- Fixed progress windows to include the complete first UTC calendar day instead of using a rolling time-of-day cutoff.
- Removed the public 100-workout history cap from progress aggregation. The internal store path now reads up to 500 workouts, sufficient for the supported 90-day window at the profile maximum of 14 workouts per week.
- Added the missing OpenAPI response schema for `GET /workouts/progress-summary`.
- Reviewed route patterns: static `/workouts/manual`, `/workouts/history`, `/workouts/progress-summary` and `/workouts/active` do not conflict with the method-aware `{workout_id}` route in Go 1.22 `ServeMux`.
- Reviewed migration `000019`: ordering is append-only, the down migration reverses only its three columns, and the database age constraint matches HTTP/OpenAPI bounds.

## Risks

- Progress calendar grouping is UTC until the profile stores an IANA timezone.
- Progress personal-record counting still reads the most recent 200 events. This is adequate for the MVP but should become a date-range/count query before high-volume production use.
- Manual creation validates environment but does not reject exercises requiring equipment absent from profile preferences. The current mobile catalog also displays those exercises, so enforcing this safely requires a coordinated catalog/UI contract change.
- The PostgreSQL store currently loads workout details with one query per workout. The 500-row correctness cap is safe for the MVP, but a range aggregation query is needed before scale testing.

## Tests and evidence

- PASS: `python3 scripts/verify-api-contract.py` — 81 OpenAPI operations aligned with 82 router routes plus `/healthz`.
- PASS: `python3 scripts/verify-migrations.py` — 19 ordered up/down migration pairs.
- PASS: `node scripts/verify-mobile-syntax.mjs` — 60 files, zero syntax errors.
- PASS: mobile `npm run typecheck`.
- PASS: OpenAPI YAML parsed with Python/PyYAML.
- PASS: `git diff --check`.
- NOT RUN: Go unit and PostgreSQL integration tests; Go/gofmt and a PostgreSQL test service are unavailable in this workspace.

## Unresolved items

- Run `go test ./...` and `POSTGRES_TEST_DSN=... go test ./internal/httpapi -run TestPostgresCriticalReleaseFlow` in CI.
- Add user timezone before presenting streak/day boundaries as local-calendar metrics.
- Coordinate equipment-aware exercise filtering between `GET /exercises`, manual-workout validation and mobile UX.

## Handoff

Lead Engineer / CI: run Go and PostgreSQL gates, then accept only if the new PostgreSQL profile regression and workout tests pass. Product engineering should schedule the timezone and equipment-contract follow-ups before production readiness.
