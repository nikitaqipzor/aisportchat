# AI Fitness OS — QA Release Audit 4C R2

Date: 2026-09-02
Decision: **SOURCE PASS / STRICT RC BLOCKED BY EXTERNAL EVIDENCE**

## Test layers executed

### Backend correctness
- `go test ./...` — PASS.
- `go vet ./...` — PASS.
- `CGO_ENABLED=0 go test ./...` — PASS after R2 fix.
- critical race detector, sequential package execution — PASS.
- HTTP integration suite shuffled 5 times — PASS.
- auth/body scan/health/nutrition/programs/recovery/technique/workouts/AI domain suites shuffled 10 times — PASS.
- deterministic health/recovery/nutrition/workout/program/technique packages repeated 20 times — PASS.

### Security/privacy targeted tests
- production weak/default JWT secret rejection — PASS.
- production memory-store rejection — PASS.
- login rate limiting / `429` behavior — PASS.
- body-scan authenticated/private flow — PASS.
- passive Health Sync owner isolation mobile regression — PASS.
- mobile auth refresh single-flight regression — PASS.
- mobile session owner scoping regression — PASS.

### API/schema/static integrity
- OpenAPI 1.6.1 YAML parse — PASS.
- Docker Compose YAML parse — PASS.
- router↔OpenAPI drift check — PASS: 78 public operations aligned; `/healthz` intentionally outside `/api/v1` contract.
- migration structure check — PASS: 18 contiguous up/down pairs (`000001`–`000018`).

### Mobile/native source checks
- 58 TS/TSX files syntax-transpiled with zero errors — PASS.
- Android native verifier — 122 checks PASS.
- technique mapping regression — PASS.
- passive-health owner-isolation regression — PASS.

## Defect found and fixed in this QA cycle

### Q1 — non-CGO backend build failure
Before R2, `CGO_ENABLED=0 go test ./...` failed because `cmd/api/main.go` referenced cgo-only `store.NewPostgres` directly.

Fix:
- introduced build-specific `OpenPostgresStore` factory;
- cgo build opens/pings libpq store;
- non-cgo build compiles and returns a clear runtime error only when PostgreSQL is selected;
- non-cgo test is now a permanent verifier and CI gate.

Status: **CLOSED**.

## QA infrastructure hardening
- race release gate now uses `-p=1` to avoid package-level race-run stalls observed in constrained environments while preserving race instrumentation;
- non-cgo portability gate added to local verifier and GitHub CI;
- API route/OpenAPI drift verifier added;
- migration pair/sequence verifier added.

## Coverage snapshot
- total backend statement coverage: ~56.1%.
- nutrition: 85.5%.
- workouts: 80.3%.
- healthdata: 83.2%.
- technique: 78.1%.
- recovery: 71.2%.
- programs: 68.4%.
- httpapi: 61.3%.
- auth: 28.3% (still a future coverage-improvement area; targeted security behaviors are tested).

## Strict release blockers still open

`STRICT_RELEASE=1 ./scripts/release-preflight.sh` returns 1 for four evidence gates:

1. `package-lock.json` missing. npm registry is unreachable in this environment, and offline cache is incomplete, so a trustworthy lockfile cannot be generated here.
2. real PostgreSQL gate not executable locally: no PostgreSQL server/client runtime / `POSTGRES_TEST_DSN`.
3. Android SDK unavailable; Gradle distribution download is also blocked by network access.
4. no physical Android/Xiaomi device smoke evidence.

These are not to be bypassed with placeholder evidence. RC remains forbidden until a suitable build/device environment makes strict preflight exit 0.
