# P1 Analytics / Devices rebuild handoff

## Goal

Restore the approved analytics/device hardening without touching `App.tsx`, the API client, or session storage.

## Changes

- `ProgressScreen` owns five independent load states: summary, measurement history, nutrition correlation, workout progress, and health insights.
- Each section exposes loading, fresh, stale, error, and explicit retry behavior. A partial failure no longer makes successful sections appear newly refreshed.
- `LatestRequestGuard` gates completion by per-key generation, auth epoch, and mount epoch. Older, old-token, and post-unmount responses cannot mutate the screen.
- `LatestRefreshOwner` gives overlapping pull-to-refresh operations latest-wins spinner ownership; token rotation invalidates that ownership and releases the spinner immediately.
- Weight chart is one accessible image with a consolidated trend/range label; visual bar children are hidden from accessibility traversal.
- Connected Devices no longer claims an unobserved `Xiaomi Watch S3`. Provider/model labels remain unknown unless the native bridge actually reports Mi Fitness presence.

## Tests / evidence

- `node scripts/test-mobile-analytics-races.mjs` — PASS
  - inverted completion;
  - independent keys and stale-cache preservation;
  - unmount invalidation;
  - A-B-A ordering;
  - two overlapping refreshes;
  - token rotation.
- `npm --prefix apps/mobile run typecheck` — PASS
- `node scripts/test-mobile-core-screens.mjs` — PASS
- `node scripts/verify-mobile-syntax.mjs` — PASS (61 files, 0 errors)
- `git diff --check` — PASS

## Risks

- No rendered emulator/TalkBack session was available. Accessibility behavior is verified statically and by React Native contracts, not on a physical device.
- The current native status contract can identify Mi Fitness installation, but does not expose the wearable model. The UI intentionally says the model is unknown.

## Unresolved

- `App.tsx` owns Analytics → Devices back navigation and is outside this package's permitted scope.
- Physical Health Connect provider combinations still need device QA.

## Handoff

Lead integrator should cherry-pick this commit, resolve only integration-level navigation in its own branch, then run the complete mobile and Android gates.
