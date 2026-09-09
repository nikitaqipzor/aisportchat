# AI Fitness OS Mobile — Native Android checkpoint

Target: bare React Native 0.87 + TypeScript with a committed native Android application.

The Android project is now part of the repository under `apps/mobile/android/`; do **not** regenerate or replace it with a temporary template. iOS is still pending.

See `ANDROID_NATIVE.md` and `../../docs/ANDROID_INTERNAL_BUILD.md`.

## Current user path

```text
Register/Login -> Onboarding -> Home
  -> Body map -> muscle -> workout -> sets -> summary/history/PR
  -> Nutrition -> calories/macros -> foods/recipes -> progress
  -> AI food text -> confirm -> diary
  -> AI food photo -> confirm -> diary
  -> AI Coach -> read-only fitness tools
  -> Weekly AI report
  -> Program -> 4/8/12 weeks -> calendar/adherence/deload
     -> program session -> adapted workout -> completion sync
```

## Storage design
- `react-native-keychain`: access + refresh tokens.
- `@react-native-async-storage/async-storage`: user-scoped active workout snapshot, pending workout mutation queue and last 30 non-sensitive AI Coach messages.

Do not move auth tokens into AsyncStorage.

## Android native modules
Native Android modules are TurboModules backed by Codegen specs in `specs/`:

- `NativeRestTimerNotifications` — background rest timer notification.
- `NativeFoodPhotoPicker` — Android system image picker for AI food analysis.
- `NativeAIFitnessConfig` — exposes build-time `API_BASE_URL` to JavaScript.

The Kotlin implementations are under `android/app/src/main/java/com/aifitnessos/` and are registered through `AIFitnessNativePackage` in `MainApplication`.

## Build

```bash
npm install
npm start
npm run android
```

or build the APK directly:

```bash
cd android
./gradlew assembleDebug
```

Expected artifact:

```text
android/app/build/outputs/apk/debug/app-debug.apk
```

## Navigation note
The MVP still uses an explicit screen state-machine plus Android `BackHandler`. It can be migrated to React Navigation after device smoke-tests without changing backend contracts.

## Platform status

The current release target is **Android-first**. `apps/mobile/android/` is the supported native project.
An iOS native project has not been created yet, so the repository intentionally does not expose an `npm run ios` command. iOS/HealthKit is a separate post-Android release milestone.
