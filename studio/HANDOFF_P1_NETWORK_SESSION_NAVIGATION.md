# P1 Network / Session / Navigation handoff

## Goal

Make authentication, offline ownership and analytics tool navigation safe under timeouts, token rotation, logout/login races and account switching.

## Changes

- Every API request has a 30-second timeout, composes a caller `AbortSignal`, and reports transport failures as `ApiNetworkError`.
- Access-token rotation is centralized in the API client. It is single-flight and guarded by auth generation plus the originating refresh-token identity before and after asynchronous adapter writes.
- Changing or clearing the current auth identity aborts an obsolete refresh. A server 401 revokes credentials through the session adapter without deleting the owner's queue, active workout or AI chat.
- Keychain mutations are serialized. Transient values use owner-specific v3 keys, migrate matching v2 envelopes, and preserve A -> B -> A recovery without exposing A data to B.
- Offline queue flushing no longer refreshes tokens itself and stops without committing if the authenticated owner changes during an API await.
- Offline bootstrap is allowed only for `ApiNetworkError` and only up to seven days after the saved access expiry. Unauthorized responses revoke the session.
- Manual logout starts a best-effort remote logout immediately, warns when a workout or queued operations exist, and preserves owner-scoped recovery data.
- Analytics remains mounted behind Devices, Body Scan and standalone Technique so its selected tab/form state survives. Devices returns to its actual Home/Analytics origin, including Android hardware Back.

## Evidence executed

- `npm --prefix apps/mobile run typecheck` — PASS
- `node scripts/test-mobile-auth-refresh.mjs` — PASS, one refresh for concurrent 401 responses
- `node scripts/test-mobile-session-scope.mjs` — PASS, A -> B -> A isolation/preservation and revoked marker
- `node scripts/test-mobile-core-screens.mjs` — PASS
- `node scripts/test-mobile-health-passive-sync.mjs` — PASS
- `node scripts/verify-mobile-syntax.mjs` — PASS, 60 files
- `python3 scripts/verify_android_native.py` — PASS, 125 checks
- `git diff --check` — PASS

## Risks / unresolved

- No physical-device or live backend test was performed in this worktree.
- The seven-day offline bootstrap grace period is a product/security policy and should be confirmed before public beta.
- This package intentionally does not implement or change the atomic profile-save API; its App integration point is preserved for the profile package.

## Handoff

Integration owner should merge this commit before UI packages that also touch `App.tsx`, resolve those changes around the centralized auth adapter and persistent analytics mount, then rerun all mobile gates and build the connected APK.
