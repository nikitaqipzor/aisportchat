# Android CI toolchain and native-dependency fix

## Goal

Restore the Android debug CI job with an internally consistent stable toolchain and React Native 0.87-compatible native dependencies.

## Changes

- CI now installs Android SDK Platform 36 and Build Tools 36.0.0.
- The Android root Gradle configuration now compiles with SDK 36 and Build Tools 36.0.0.
- `targetSdkVersion` remains 36, so compile and target SDK versions are aligned.
- The native Android verifier and setup documentation now assert/document the same versions.
- CI uses JDK 21, matching the committed Gradle wrapper bytecode level.
- `react-native-safe-area-context` is pinned to 5.9.1; unlike 5.5.2, this release no longer calls the removed `UIManagerModule.uiImplementation` API and therefore compiles with React Native 0.87.
- Native modules now obtain the foreground activity through `reactApplicationContext.currentActivity`, the React Native 0.87-compatible accessor.
- Technique framing converts its primitive landmark index array with `map(...).filterNotNull()`, avoiding the unavailable `IntArray.mapNotNull` overload while preserving visibility filtering.
- Food-photo decoding uses the non-deprecated `InputStream.readBytes()` overload required by the Kotlin 2.2 warnings-as-errors policy; the existing 5 MB validation remains unchanged.

## Rationale

- The repository is pinned to React Native 0.87.0. Its minimum Android compile SDK is 34; SDK 36 therefore satisfies the library minimum.
- React Native 0.87's template was bumped to compile/build tools 37, but the failed CI run showed those exact packages were unavailable from the stable `sdkmanager` channel used by GitHub Actions.
- Android SDK Platform 36 is a stable platform release, and Android's official Build Tools documentation lists 36.0.0. SDK 36 also matches the project's existing target SDK.

Primary references:

- React Native 0.87 release: <https://reactnative.dev/blog/2026/08/11/react-native-0.87>
- Android SDK Platform releases: <https://developer.android.com/tools/releases/platforms>
- Android SDK Build Tools releases: <https://developer.android.com/tools/releases/build-tools>

## Tests and evidence

- `python3 scripts/verify_android_native.py` — 121 native checks passed; the process returned non-zero only because the environment does not provide the JDK `jar` executable used for wrapper inspection.
- `unzip -l apps/mobile/android/gradle/wrapper/gradle-wrapper.jar` confirmed `org/gradle/wrapper/GradleWrapperMain.class` is present, covering the one environment-blocked verifier assertion.
- Static cross-file version consistency check — passed for CI, Gradle, verifier, and setup documentation; no stale Android 37 references remain in those Android-CI files.
- `git diff --check` — passed.
- `npm ci` and `npm run typecheck` — passed with the updated dependency lock.
- `node scripts/verify-mobile-syntax.mjs` plus all four mobile runtime-regression scripts — passed.
- `python3 scripts/verify_android_native.py` — 122 native checks passed after the Kotlin and minimum-SDK repairs.
- Full Gradle APK assembly is delegated to the GitHub-hosted Android gate because this workspace has no Android SDK.

## Risks

- This intentionally differs from the React Native 0.87 default template value of 37. The RN release states a minimum compile SDK of 34, so compile SDK 36 remains above the minimum, but an upstream dependency could independently require 37 in a future update.
- The CI job must prove package installation and full APK assembly in the hosted runner environment before merge.

## Unresolved items

- No Android 16 emulator or physical-device behavior was tested in this environment.
- No release-signing build was attempted; this change targets the debug release gate only.

## Handoff

QA should run the `android-debug` job, confirm `sdkmanager` installs all four packages, and verify that `npm run android:assemble:debug` produces `app-debug.apk`.
