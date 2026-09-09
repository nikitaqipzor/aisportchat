# Release readiness

## Current decision

**NOT RC YET.** The source tree is release-hardened, but external build/device/database evidence is still incomplete.

## Resolved Critical / High findings

- [x] Passive Health Sync owner isolation.
- [x] Logout/foreign-owner worker handling.
- [x] Auth brute-force throttling and `429` contract.
- [x] Production JWT secret hard fail.
- [x] Production PostgreSQL-only store guard.
- [x] Real PostgreSQL integration/migration test harness added to CI.
- [x] Android native build job added to CI.
- [x] Nutrition critical coverage raised above 80%.
- [x] Workout critical coverage raised above 80%.
- [x] First accessibility/test selector hardening pass.
- [x] Non-CGO backend portability gate (`CGO_ENABLED=0`).
- [x] Router↔OpenAPI drift gate and migration pair/sequence gate.

## Open release blockers

| Gate | Current state | Required evidence |
|---|---|---|
| npm reproducibility | BLOCKED HERE | committed `package-lock.json` + clean `npm ci` |
| PostgreSQL 18 E2E | CONFIGURED, NOT RUN HERE | green CI migration + HTTP flow |
| Android build | CONFIGURED, NOT RUN HERE | green `assembleDebug`/release build artifact |
| Physical Android smoke | NOT RUN | recorded install/auth/workout/offline/native smoke |
| Xiaomi Health Connect | NOT RUN ON DEVICE | values matched to Mi Fitness for sleep/steps/activity |
| Cross-account passive health | SOURCE TESTED ONLY | device A→logout/crash→B isolation smoke |

## Strict rule

`STRICT_RELEASE=1 ./scripts/release-preflight.sh` must exit successfully before an RC tag is produced.

Do not create placeholder files under `release-evidence/` to force a pass.
