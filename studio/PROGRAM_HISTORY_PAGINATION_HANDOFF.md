# Program history pagination handoff — 2026-09-26

## Goal
Let an athlete find and open past programs beyond the first 50, with deterministic user-scoped pages and a usable mobile load-more path.

## Changes
- `GET /programs/history` now accepts `limit` (1–50) and an opaque `cursor`, and returns the existing `items` plus `has_more` and `next_cursor`. Bad limits and malformed cursors receive 400. OpenAPI describes the parameters and response.
- Memory and PostgreSQL stores select by `(updated_at, id)` descending and page strictly below the prior page's last tuple. The service reads one extra item to determine `has_more`.
- Mobile Programs loads 50 initially, appends later pages, deduplicates program IDs, retries a failed page without dropping loaded data, and resets on refresh. Workout detail searches successive pages for an archived linked workout.
- Go tests exercise 121 programs, user ownership, invalid parameters, all page boundaries, and a new program inserted between pages. A PostgreSQL integration variant uses `POSTGRES_TEST_DSN`. A mobile behavior test covers >50, duplicate merge, and failed-page retry.

## Evidence
- `node apps/mobile/src/screens/ProgramsScreen.test.mjs` — passed.
- `node scripts/test-mobile-program-resume.mjs` — passed.
- `node scripts/test-mobile-core-screens.mjs` — passed.
- `cd apps/mobile && npm run typecheck` — passed.
- `python3 scripts/verify_android_native.py` — passed (125 checks).
- `git diff --check` — passed during integration.
- Go and PostgreSQL tests were not run here because `go` is absent. Run `go test ./internal/store ./internal/programs ./internal/httpapi` from `services/api`; set `POSTGRES_TEST_DSN` after migrations for the PostgreSQL case.

## Risks and unresolved items
- Cursor paging handles new records ahead of the boundary, but a concurrent edit that changes an unseen program's `updated_at` across the boundary can move it out of a later page. A database snapshot or immutable historical order would be required for a strict point-in-time list.
- PostgreSQL query and its 121-program integration case still need execution in CI. The query loads sessions for each selected program, so large pages issue multiple reads.
- The new mobile test needs to be added to `scripts/verify.sh` and `.github/workflows/ci.yml` by the lead alongside other concurrent test wiring.

## Handoff
Lead engineer: run Go and PostgreSQL integration gates, wire `apps/mobile/src/screens/ProgramsScreen.test.mjs` into CI/verify, review the shared router/store/OpenAPI diff, and accept the program history scope after those checks.
