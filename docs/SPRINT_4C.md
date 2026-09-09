# Sprint 4C — Passive Health Sync, Readiness Explanation Timeline & Release Hardening

Status: **SOURCE ACCEPTED / RELEASE CANDIDATE NOT YET ELIGIBLE**.

## Product work

### Passive Health Sync
- Android WorkManager can read permitted Health Connect data in the background.
- Background code never stores API access/refresh tokens and never calls the backend directly.
- Pending daily snapshots are encrypted with Android Keystore and uploaded by the authenticated JS layer on the next app session.
- Empty Health Connect days are not persisted as fake coverage.

### Cross-account privacy hardening
- Passive snapshots are partitioned by AI Fitness OS `user_id`.
- JS can only read/acknowledge the current authenticated user's queue.
- Logout cancels that user's worker and clears that user's pending health queue.
- Legacy unscoped v1 pending data is deleted because its owner cannot be proven.
- On bootstrap/login, a worker bound to a different owner is suspended while its encrypted queue stays isolated.

### Readiness Explanation Timeline
- Recovery exposes a deterministic factor timeline instead of an opaque AI score.
- Sleep, energy, stress, soreness and recent load show their actual contribution to readiness.
- Health trends such as steps/calories can appear as context without falsely claiming causal score impact.
- Workout adaptation explains the effective volume/intensity multipliers after Program + Recovery rules.

## Release hardening

### Authentication
- IP + account rate limiting for register/login.
- IP + refresh-token fingerprint throttling for refresh.
- `429 Too Many Requests` + `Retry-After` is part of OpenAPI.
- Wrong-password responses remain generic to avoid account enumeration.

### Production runtime
- `APP_ENV=production` rejects known/default/dev JWT secrets and secrets shorter than 32 bytes.
- Production refuses the in-memory store; PostgreSQL is mandatory.

### PostgreSQL release gate
- Real PostgreSQL HTTP integration test is available when `POSTGRES_TEST_DSN` is supplied.
- Critical flow: auth/onboarding → workout → nutrition → Xiaomi health → recovery → cross-user isolation.
- Migration smoke executes `up → down(reverse) → up` on a disposable database.
- GitHub Actions config includes a PostgreSQL 18 service.

### Android build gate
- GitHub Actions config provisions JDK + Android SDK/Build Tools/NDK and runs TypeScript/Codegen/`assembleDebug`.
- Actual Android build is not claimed in the current container because Android SDK and npm registry access are unavailable.

### Mobile dependency reproducibility
- direct dependencies are exact-pinned;
- `.npmrc` requires lockfiles/exact saves;
- CI install helper prefers `npm ci`;
- **`package-lock.json` is still missing** because this execution environment cannot reach the npm registry. This remains a strict release blocker.

### Test quality
- Nutrition service coverage: **85.5%**.
- Workout service/package coverage: **80.3%**.
- Whole Go backend coverage in this environment: **56.2%**.
- Auth/health/recovery/nutrition/workout critical packages pass the Go race detector.

### Accessibility pass
- shared `AppButton` now provides fallback `testID`, button role/label/state and minimum touch target;
- Auth, Goal, Profile Setup, Training Setup, Workout Preview, Recovery and AI Coach received stable selectors/semantics/error states;
- AI Coach supports cancel/retry and abortable requests;
- accessibility is improved but **not considered complete** across the whole app.

## OpenAPI

Version: **1.6.1**.

## Release preflight

Run normal diagnostics:

```bash
./scripts/release-preflight.sh
```

Run the actual release gate:

```bash
STRICT_RELEASE=1 ./scripts/release-preflight.sh
```

Strict mode must not be bypassed. At the time of this checkpoint it intentionally fails on external evidence that is unavailable in this container.

## Open release gates

1. generate and commit `apps/mobile/package-lock.json`, then prove clean `npm ci`;
2. run PostgreSQL 18 migration + HTTP integration jobs successfully;
3. run Android native CI build successfully;
4. install/run on a physical Android device;
5. validate Xiaomi Watch S3 → Mi Fitness → Health Connect values against Mi Fitness UI;
6. record cross-account passive-sync device smoke evidence.

No production Release Candidate should be labeled ready until all six gates are green.
