# Android minimum SDK compatibility fix

## Goal

Resolve the release-gate manifest merge failure caused by the app declaring API 24 while the stable Health Connect client requires API 26.

## Changes

- Raised the shared Android `minSdkVersion` from 24 to 26.
- Updated the native verifier to enforce API 26.
- Updated Android setup documentation to state Android 8.0 as the oldest supported release and explain the Health Connect constraint.
- Corrected the internal-build document's stale compile SDK / Build Tools values to the repository's configured SDK 36 / Build Tools 36.0.0.

## Rationale

Health Connect is a first-class product boundary for wearable and recovery data, and the app pins `androidx.health.connect:connect-client:1.1.0`. Raising the product minimum to the dependency's supported API level keeps that contract explicit and testable. A manifest `tools:overrideLibrary` would only silence the build-time safety check and could allow unsupported library code onto API 24–25 devices, so it is intentionally not used.

## Tests and evidence

- `python3 scripts/verify_android_native.py` — 122 checks passed, including the API 26 assertion.
- `git diff --check` — passed.
- No activity or native-module Kotlin implementation was changed.

## Risks

- Android 7.0/7.1 devices (API 24–25) can no longer install future builds.
- Full manifest merge and APK assembly remain delegated to the hosted Android CI runner.

## Unresolved items

- Physical-device Health Connect permission and Xiaomi/Mi Fitness data-origin validation remain required before a production-release claim.

## Handoff

QA should confirm `android-debug` completes manifest merge and produces `app-debug.apk` with `minSdkVersion 26`, then verify install/permission flows on a supported physical device.
