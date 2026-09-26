# Mobile account deletion handoff — 2026-09-26

## Goal
Give athletes an explicit, reachable account deletion action, avoid deleting local data before server confirmation, and clear private on-device state after committed deletion.

## Changes
- Profile offers a data and account section, with a separate destructive confirmation that calls out unsynced operations and active workouts. It remains reachable if profile loading fails.
- `DELETE /api/v1/auth/account` uses bearer authentication. A single access-token refresh retry applies to 401; network timeouts are not retried automatically. Both 204 and 202 mean committed deletion. 202 is explained as server photo cleanup pending.
- After 204/202, capture the local owner beforehand, invalidate API refresh, clear Keychain, remove owner-scoped AsyncStorage (or all app-owned keys if a legacy session has no owner), cancel rest timer, clear native health passive cache, and remove camera/technique temporary files. A failed local cleanup prompts the user to clear Android app data.
- No fake privacy consent control or unpublished policy link was added. Formal notice and consent remain a release gate.

## Tests/evidence
- `npm run typecheck`: pass.
- `node scripts/test-mobile-auth-refresh.mjs`: pass, including 401 refresh and 202 handling.
- `node apps/mobile/src/storage/offlineQueue.test.mjs`: pass, including owner-scoped and legacy cache removal.
- `node scripts/verify-mobile-syntax.mjs`: pass (67 files, 0 errors).
- `python3 scripts/verify_android_native.py`: pass (125 checks). Actual APK build/device verification remains pending.

## Risks/unresolved
- A network timeout after a server commit is ambiguous; UI says deletion may have completed and leaves local auth in place until a confirmed response. A later login/retry can resolve this. A subsequent 401 may also mean already deleted.
- Native temporary file deletion only targets private app cache directories `body_scan_camera`, `technique_video`, `technique_live`. Android build must regenerate TurboModule codegen for new `clearPrivateTempFiles()` method.
- Server backup/external processor retention and publication of a privacy notice still need product/legal decisions before a public pilot.

## Handoff
QA: exercise 204, 202, 401 refresh, server 500, network timeout, offline queue and Body Scan cache with a real Android internal build. Lead engineer: accept the complete server/mobile contract and copy against the privacy notice before pilot.
