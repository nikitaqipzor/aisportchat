# Four native flows — Lead acceptance

## Scope

Production hardening for Body Scan, Mi Fitness through Health Connect,
Technique CV, and Food Photo.

## Accepted changes

- Body Scan exposes deterministic capture-comparability metrics only, remains
  compatible with the previous API response, and keeps metadata when private
  blob deletion fails so deletion can be retried safely.
- Connected Devices explicitly describes the supported Mi Fitness source,
  improves permission/settings recovery, refreshes status after returning to
  the app, and preserves source provenance.
- Technique adds capture preflight and structured cancellation/permission
  errors, expands only safe aliases of the five calibrated motion patterns,
  and cannot attach an unsupported movement analysis to a workout set.
- Food Photo previews the selected image, treats picker cancellation as a
  neutral action, validates supported image formats and size, and requires the
  user to confirm or exclude each matched catalog item.

## Evidence

- mobile TypeScript: PASS;
- all 11 `scripts/test-mobile-*.mjs` suites: PASS;
- mobile syntax: 64 files, 0 errors;
- Android native static checks: 125 PASS;
- OpenAPI/router alignment: 82 operations / 83 routes + health check: PASS;
- migrations: 19 up/down pairs: PASS;
- independent QA: CONDITIONAL PASS, no code-level Critical/High issues;
- independent UX review: PASS for the four remediated flows.

## Conditional items

- Go is unavailable in the current workspace, so the new Go unit tests must be
  executed by CI.
- The local JDK is 17 while the Gradle wrapper requires JDK 21, so the internal
  APK build must be executed by CI.
- Before a public beta, run the internal APK on a physical Android/Xiaomi phone
  for camera, gallery picker, CameraX/MediaPipe, Mi Fitness, Health Connect
  permission denial, settings return, background sync, and process recreation.

## Decision

Accepted for CI and physical-device alpha testing. Not accepted as a public
release until the conditional items above pass.
