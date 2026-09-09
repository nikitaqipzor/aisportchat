# Standalone Android test APK

## Goal

Produce an installable phone-preview APK that embeds the React Native Hermes bundle and does not depend on Metro.

## Changes

- Added an `internal` Android build type derived from debug but marked non-debuggable so the React Native Gradle plugin packages `index.android.bundle`.
- Kept debug signing for test installation and disabled cleartext traffic.
- Added a reserved HTTPS fallback API endpoint only for builds where no server URL is supplied; this avoids a startup configuration crash without pretending that server flows are available.
- CI now builds and uploads `app-internal.apk` as `athletica-ai-internal-apk`.

## Tests and evidence

- `npm run typecheck` passes.
- Mobile syntax and runtime regression scripts pass.
- Android native static verifier passes.
- GitHub Actions must prove `assembleInternal`, bundle packaging and artifact upload before delivery.

## Risks and unresolved items

- The internal APK is test-only and signed with the debug key.
- Without a deployed HTTPS API, authentication and synchronized data flows will return network errors.
- Physical-device visual, permission and accessibility QA remains required.

## Handoff

QA should install the generated internal APK without Metro, verify that the first screen renders, and record device/model/Android version plus screenshots of any issue.
