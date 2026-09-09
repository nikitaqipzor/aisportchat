# Android internal build — Native React Native 0.87

`apps/mobile/android/` is now a committed native Android application. Do not run `react-native init` over it.

## Toolchain

- Node.js >= 22.11
- JDK 21 (required by the committed Gradle wrapper)
- Android Studio
- Android SDK Platform 36
- Android Build Tools 36.0.0
- Android SDK Platform Tools

The project uses minSdk 26, compileSdk 36, targetSdk 36, Kotlin 2.2.0 and Gradle 9.4.1. Android 8.0 is the oldest supported release because the pinned stable `androidx.health.connect:connect-client:1.1.0` dependency requires API 26; do not bypass this contract with a manifest override.

## 1. Install dependencies

```bash
cd apps/mobile
npm install
```

## 2. Configure Android SDK

Create `android/local.properties` from the example:

```properties
sdk.dir=/absolute/path/to/Android/Sdk
```

Android Studio can create this automatically.

## 3. Development API URL

During Metro development the JS layer derives the LAN/Metro host automatically. You can also pass an explicit native URL at build time:

```bash
cd android
./gradlew installDebug -PAI_FITNESS_API_BASE_URL=http://192.168.1.50:8080/api/v1
```

For production/release use HTTPS:

```bash
./gradlew bundleRelease -PAI_FITNESS_API_BASE_URL=https://api.example.com/api/v1
```

## 4. Start Metro

Terminal 1:

```bash
cd apps/mobile
npm start
```

## 5. Build/install debug

Terminal 2:

```bash
cd apps/mobile
npm run android
```

or:

```bash
cd apps/mobile/android
./gradlew assembleDebug
```

Expected APK:

```text
apps/mobile/android/app/build/outputs/apk/debug/app-debug.apk
```

The debug APK expects Metro. For a standalone visual-test APK with an embedded
Hermes bundle, run:

```bash
cd apps/mobile
npm run android:assemble:internal
```

Expected standalone APK:

```text
apps/mobile/android/app/build/outputs/apk/internal/app-internal.apk
```

The internal build uses debug signing, installs as the separate
`com.aifitnessos.preview` application, and does not require Metro. When no API
URL is supplied, it uses the reserved unreachable preview endpoint so the UI
can open safely while server-backed actions report a network error. Pass a real
HTTPS `AI_FITNESS_API_BASE_URL` to test authentication and synchronized flows.

## 6. Physical Android device

Enable Developer options + USB debugging, then:

```bash
adb devices
cd apps/mobile
npm run android
```

If API runs on the development computer, prefer an explicit LAN URL or `adb reverse` where appropriate.

## Native modules

The app is on React Native New Architecture and uses Codegen/TurboModules.

Specs:

```text
apps/mobile/specs/NativeRestTimerNotifications.ts
apps/mobile/specs/NativeFoodPhotoPicker.ts
apps/mobile/specs/NativeAIFitnessConfig.ts
```

Implementations:

```text
android/app/src/main/java/com/aifitnessos/resttimer/
android/app/src/main/java/com/aifitnessos/foodphoto/
android/app/src/main/java/com/aifitnessos/nativebridge/
```

Manual package registration lives in `MainApplication.kt` through `AIFitnessNativePackage`.

## Release signing

Never commit production signing credentials. Configure the following Gradle properties or environment variables:

```text
AI_FITNESS_UPLOAD_STORE_FILE
AI_FITNESS_UPLOAD_STORE_PASSWORD
AI_FITNESS_UPLOAD_KEY_ALIAS
AI_FITNESS_UPLOAD_KEY_PASSWORD
```

A release build without these credentials remains unsigned rather than falling back to the debug key.

## Device smoke checklist

1. install clean build;
2. register/login and restart app;
3. verify token/session recovery;
4. complete onboarding;
5. generate and start workout;
6. log a set and background the app during rest;
7. verify native notification;
8. reopen app and restore active workout;
9. cancel one workout and verify it is not restored;
10. finish a workout and view PR/history;
11. log nutrition and test Undo;
12. open AI food photo and select an image;
13. open AI Coach;
14. create a training program and launch a session from its calendar.
