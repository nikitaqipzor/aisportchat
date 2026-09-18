# Food photo production hardening

## Goal

Make food-photo input safe and recoverable on a physical Android device without treating AI estimates as nutrition facts.

## Changes

- Added an in-memory image preview and explicit privacy/accuracy disclosure.
- System-picker cancellation is no longer shown as an application failure.
- Retry reuses the already selected photo; choosing another photo remains separate.
- Native Android input rejects empty, unsupported, and over-5-MiB files with stable error codes and bounded streaming reads.
- API validates MIME, base64 integrity, decoded size, and empty payloads before invoking the provider.
- Recognized catalog matches can be excluded before confirmation; unmatched or incorrect results can return to text correction.

## Risks

- Recognition quality still depends on the configured vision provider and the food catalog.
- The photo exists transiently in React Native memory for preview/retry until the screen is left; it is not persisted by the app.
- Camera/gallery behavior still needs a physical-device pass across multiple Android vendors.

## Tests/evidence

- `go test ./internal/aifitness ./internal/httpapi`
- `node scripts/test-mobile-nutrition-regressions.mjs`
- `npm run typecheck` in `apps/mobile`
- `python3 scripts/verify_android_native.py`

## Unresolved items

- Run a release/internal APK test with JPEG, PNG, WebP, cancellation, corrupted image, oversize image, offline retry, and a provider timeout.

## Handoff

QA should verify that cancellation is silent, retry does not reopen the gallery, excluded foods are not written, and raw photos never appear in diary/API persistence.
