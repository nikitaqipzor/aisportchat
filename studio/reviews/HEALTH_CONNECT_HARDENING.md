# Health Connect production UX hardening

## Goal

Make Connected Devices recoverable when Health Connect is missing, outdated, partially permitted, empty, offline, or opened from different parent screens.

## Changes

- The back control now names and returns to the actual origin: Home or Analytics.
- Returning from Android Health Connect settings automatically refreshes native permission/provider state and server data.
- Missing permission categories are shown explicitly instead of as a generic denial.
- The native bridge opens the Health Connect store page when the provider needs installation/update, with a browser fallback.
- Opening settings now reports a failure instead of silently succeeding when Android cannot resolve the intent.
- Sync errors are translated into permission, provider, or network recovery guidance.
- A zero-record sync is an actionable non-error state; a successful sync reports imported-day count and completion time.
- Static regression checks cover origin-aware navigation, settings recovery, and sync recovery feedback.

## Risks

- OEM Android builds can customize Health Connect and store intent handling. The native flow therefore has both Play Store app and HTTPS fallbacks, but still needs physical-device validation.
- Source provenance is truthful only to the package/source labels returned by Health Connect and the API. The UI intentionally does not infer a watch model.

## Tests / evidence

- `cd apps/mobile && npm run typecheck` — PASS.
- `node scripts/test-mobile-core-screens.mjs` — PASS.
- `python3 scripts/verify_android_native.py` — PASS, 125 checks.

## Unresolved

- Validate permission denial/revocation, Health Connect update recovery, and Mi Fitness data origin on a physical Xiaomi/Android device.
- Confirm OEM-specific background execution timing; Android periodic work remains best-effort rather than an exact hourly guarantee.

## Handoff

Lead/QA should run the physical-device sequence: open from Home and Analytics, revoke one permission, return from settings, sync with and without Mi Fitness records, disconnect the network during upload, then validate provenance against Mi Fitness.
