# Changelog — Sprint 4C R2 QA Hardening

Date: 2026-09-02

## QA findings fixed

- Fixed backend portability failure when `CGO_ENABLED=0`:
  - `cmd/api` no longer references the cgo-only `store.NewPostgres` symbol directly;
  - added build-specific `OpenPostgresStore` factories;
  - non-cgo builds now compile and return a clear runtime error only if PostgreSQL is selected without cgo/libpq.
- Added `CGO_ENABLED=0 go test ./...` to the repository verifier and GitHub backend gate.
- Stabilized race-detector gates with `-p=1` after repeated evidence that concurrent package race runs can stall in constrained environments while every package passes individually.
- Added router↔OpenAPI drift verification (78 public operations; `/healthz` intentionally excluded from the `/api/v1` contract).
- Added migration structure verification for all 18 up/down migration pairs and contiguous numbering.

## QA evidence

Passed locally:
- `go test ./...`
- `go vet ./...`
- `CGO_ENABLED=0 go test ./...`
- critical `go test -race -p=1 ...`
- 20x repeated deterministic domain regression for health/recovery/nutrition/workouts/programs/technique
- mobile TypeScript syntax verifier
- mobile auth refresh/session scope/technique mapping/passive-health owner isolation regressions
- Android native static verifier
- production JWT/store guard tests
- auth login rate-limit test

## Release status

Still **NOT RC** in this environment. Strict blockers are external/evidence gates:
- npm registry unavailable, so a trustworthy `package-lock.json` cannot be generated and `npm ci` cannot be proven;
- PostgreSQL server/client runtime unavailable locally;
- Android SDK and Gradle distribution unavailable locally;
- physical Android/Xiaomi device evidence is not available.

Do not fake release evidence. `STRICT_RELEASE=1 ./scripts/release-preflight.sh` must exit 0 in a suitable build/device environment before RC.
