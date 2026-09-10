# Home / manual profile reliability handoff

## Goal

Remove the implicit gym fallback and keep Home and manual workout safe during
network failure, account changes, and out-of-order requests.

## Changes

- Home profile state is explicit: `loading`, `loaded`, `cached-stale`, `error`.
- Home manual CTA, body map, and muscle cards stay blocked until an environment
  is verified remotely or loaded from the same-owner cache.
- Manual workout independently verifies environments and never loads its catalog
  or creates a workout using an assumed environment.
- Added minimal owner-scoped cache `{environments, cachedAt}`. It captures the
  initiating owner and conditionally writes only if that owner is still active.
- Explicit logout clears the cache; automatic token expiry preserves it.
- Home readiness, Home profile, manual profile/catalog, and athlete profile loads
  reject stale responses and updates after unmount.

## Privacy and risks

- Cache contains no profile fields, tokens, injury data, or other PII.
- Cached environments are visibly marked stale and offer retry.
- The shared `LatestRequestGuard` may overlap an analytics implementation during
  integration. Keep one compatible implementation at
  `src/domain/latestRequest.ts`.

## Evidence

- `node scripts/test-mobile-home-reliability.mjs` — PASS
- `node scripts/test-mobile-session-scope.mjs` — PASS
- `node scripts/verify-mobile-syntax.mjs` — PASS (61 files)
- `npm --prefix apps/mobile run typecheck` — PASS

## Unresolved

- Physical-device interaction and offline visual QA were not executed here.

## Handoff

Integrator: cherry-pick this commit, resolve any duplicate latest-request helper
with analytics, then run the combined mobile gate and Android build.
