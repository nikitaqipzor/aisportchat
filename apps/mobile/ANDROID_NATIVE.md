# AI Fitness OS — Native Android

This folder is now a real React Native 0.87 Android application, not a source fragment.

## Toolchain

- React Native 0.87.0
- React 19.2.3
- Node >= 22.11
- JDK 17+ (JDK 21 recommended)
- compileSdk 37
- targetSdk 36
- minSdk 24
- Kotlin 2.2.0
- Gradle 9.4.1
- New Architecture + Hermes enabled

## First build

1. Install Android Studio and SDK Platform 37 / Build Tools 37.0.0.
2. Copy `android/local.properties.example` to `android/local.properties` and set `sdk.dir`.
3. Run `npm install` from `apps/mobile`.
4. Start Metro: `npm start`.
5. In another terminal: `npm run android`.

For a physical Android phone, enable USB debugging and verify `adb devices`. Metro's host is used automatically for the dev API URL; the backend must be reachable from the phone on port 8080.

## Native modules

The app contains first-party Turbo Native Modules generated with React Native Codegen:

- `RestTimerNotifications` — schedules/cancels rest timer notifications.
- `FoodPhotoPicker` — opens Android image picker and returns an image data URL (max 5 MiB).
- `AIFitnessConfig` — exposes build-time API base URL to JS.

Run `npm run android:codegen` after changing files under `specs/`.

## Release API URL

Pass an HTTPS URL via Gradle property or environment variable:

```bash
AI_FITNESS_API_BASE_URL=https://api.example.com/api/v1 npm run android:bundle:release
```

Release builds disable cleartext HTTP at manifest level.

## Release signing

Set:

- `AI_FITNESS_UPLOAD_STORE_FILE`
- `AI_FITNESS_UPLOAD_STORE_PASSWORD`
- `AI_FITNESS_UPLOAD_KEY_ALIAS`
- `AI_FITNESS_UPLOAD_KEY_PASSWORD`

Never commit a production keystore or its passwords.


## Technique CV
`TechniqueVideo` is a TurboModule backed by MediaPipe Pose Landmarker. Android `preBuild` downloads the lite pose model automatically when it is absent. To prefetch it explicitly:

```bash
./scripts/fetch-mediapipe-model.sh
```

The raw MP4 is temporary app-cache data. Pose landmarks are extracted locally and the video is deleted after analysis.
