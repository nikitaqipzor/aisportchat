# Stage 4 workout MVP handoff

## Goal

Complete the missing MVP path for a user-authored workout while preserving the existing deterministic workout lifecycle, owner-scoped offline queue, rest timer, confirmed completion and history.

## Changes

- Added authenticated catalog filtering by muscle, environment and search query.
- Added `POST /workouts/manual` with deterministic catalog validation, environment/profile enforcement, bounded sets/repetitions/weight/rest/duration, duplicate rejection and owner-scoped persistence.
- Added the mobile manual workout builder with profile-allowed environments, muscle and exercise selection, editable plan fields, validation and loading/error/empty states.
- Connected the builder to Home and the existing preview → start → set logging → rest timer → confirmed finish → summary/history flow.
- Updated mobile types and OpenAPI contracts.
- Added service tests covering successful persistence and invalid/duplicate/environment-incompatible exercise rejection.

## Risks

- Manual creation currently requires network access. Once the planned workout has started, existing active-workout snapshots and the user-scoped queue preserve set logging and finish/cancel mutations offline.
- The current catalog is intentionally limited; importing the 918-exercise content contract remains a separate content milestone.
- Physical-device keyboard, dynamic type, TalkBack and notification behavior still require QA.

## Tests / evidence

- `cd apps/mobile && npm run typecheck` — passed.
- `node scripts/verify-mobile-syntax.mjs` — passed, 59 files / 0 errors.
- `python3 scripts/verify_android_native.py` — passed, 125 checks.
- OpenAPI YAML parsed with PyYAML — passed.
- `git diff --check` — passed.
- Go tests were not executed because this environment has no `go` or `gofmt` binary. New service tests are included for CI.

## Unresolved items

- Run `gofmt` and `go test ./...` in CI/Go-enabled development environment.
- Run end-to-end device QA against a deployed backend and verify the finished workout appears in History after offline queue synchronization.

## Handoff

QA: test the manual builder with home-only, gym-only and band-only accounts; reject invalid numeric bounds; start and finish both complete and early workouts; repeat with connectivity removed after start; relaunch and verify restoration/synchronization/history. Lead Engineer should accept only after Go CI and device QA are green.
