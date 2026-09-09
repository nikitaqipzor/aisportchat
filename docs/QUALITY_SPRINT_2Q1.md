# Quality Sprint 2Q1 — Runtime blockers

This sprint applies the first fixes from `ai-fitness-os-qa-audit.md` before Phase 3.

## Fixed

### QA-001 — token expiry during long sessions
`apps/mobile/src/api/client.ts` now owns authenticated request retry:

1. protected request receives `401`;
2. one shared refresh request rotates tokens;
3. new tokens are persisted in Keychain;
4. App state is notified;
5. all failed protected requests retry once with the fresh token.

Concurrent 401s share `refreshInFlight`, preventing refresh-token rotation races.
`node scripts/test-mobile-auth-refresh.mjs` is the regression test.

### QA-002 — API URL
`src/config/api.ts` resolves configuration in this order:

1. runtime override `__AI_FITNESS_API_BASE_URL__`;
2. native `AIFitnessConfig.apiBaseUrl`;
3. dev Metro host + API port 8080.

Production rejects non-HTTPS configured URLs.

### QA-004 — cancel workout
Added `POST /workouts/{id}/cancel`. Planned/active workouts can be cancelled. Active restore no longer returns them. A linked Program Session is detached so the workout can be generated again. Offline cancellation is queued.

### QA-005 — cross-user offline queue
AsyncStorage fitness data now uses an owner envelope keyed by the immutable authenticated user id stored with Keychain credentials. A different user cannot read another user's active workout, offline queue, or AI-chat cache. Logout clears v2 and legacy v1 transient keys.

### QA-010 / QA-011
- early workout finish shows confirmation with `done / target`;
- workout stores `completion_percent` and `ended_early`;
- reps must be 1..200;
- weight 0..1000;
- RIR integer 0..10;
- rest notification is cancelled on finish/cancel.

## UX foundation fixes
- bottom navigation: Today / Training / Nutrition / Progress / AI;
- profile/logout moved out of the crowded Home top-row;
- shared accessible `AppButton`;
- root uses `react-native-safe-area-context`;
- Nutrition primary AI action + two secondary actions;
- food deletion has Undo;
- program archive has confirmation and pending state.

## Verification

```bash
./scripts/verify.sh
```

Includes Go tests/vet, YAML checks, TypeScript syntax, mobile auth refresh runtime regression, and per-user storage runtime regression.

## Remaining release gate
The largest remaining critical item is QA-003: create the real React Native Android/iOS native scaffold and compile/install the application. This should be completed before Body Scan/Pose work.
