#!/usr/bin/env python3
from __future__ import annotations
import json
import re
import subprocess
import sys
import zipfile
import xml.etree.ElementTree as ET
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
MOBILE = ROOT / 'apps' / 'mobile'
ANDROID = MOBILE / 'android'
errors: list[str] = []
checks: list[str] = []

def ok(msg: str):
    checks.append(msg)

def require(cond: bool, msg: str):
    if not cond:
        errors.append(msg)
    else:
        ok(msg)

def text(path: Path) -> str:
    if not path.exists():
        errors.append(f'missing file: {path.relative_to(ROOT)}')
        return ''
    return path.read_text(encoding='utf-8')

# Required native project shape.
required = [
    ANDROID / 'settings.gradle',
    ANDROID / 'build.gradle',
    ANDROID / 'gradle.properties',
    ANDROID / 'gradlew',
    ANDROID / 'gradlew.bat',
    ANDROID / 'gradle/wrapper/gradle-wrapper.jar',
    ANDROID / 'gradle/wrapper/gradle-wrapper.properties',
    ANDROID / 'app/build.gradle',
    ANDROID / 'app/debug.keystore',
    ANDROID / 'app/src/main/AndroidManifest.xml',
    ANDROID / 'app/src/main/java/com/aifitnessos/MainActivity.kt',
    ANDROID / 'app/src/main/java/com/aifitnessos/MainApplication.kt',
    ANDROID / 'app/src/main/java/com/aifitnessos/technique/TechniqueLiveActivity.kt',
    ANDROID / 'app/src/main/java/com/aifitnessos/technique/TechniqueLiveModule.kt',
    ANDROID / 'app/src/main/java/com/aifitnessos/technique/TechniquePoseOverlayView.kt',
    ANDROID / 'app/src/main/java/com/aifitnessos/technique/LiveRepCounter.kt',
    ANDROID / 'app/src/main/java/com/aifitnessos/healthconnect/HealthConnectModule.kt',
    ANDROID / 'app/src/main/java/com/aifitnessos/healthconnect/PermissionsRationaleActivity.kt',
    ANDROID / 'app/src/main/java/com/aifitnessos/healthconnect/HealthConnectReader.kt',
    ANDROID / 'app/src/main/java/com/aifitnessos/healthconnect/HealthPassiveSyncWorker.kt',
    ANDROID / 'app/src/main/java/com/aifitnessos/healthconnect/HealthPassiveSyncStore.kt',
    ANDROID / 'app/src/main/java/com/aifitnessos/healthconnect/HealthPassiveSyncScheduler.kt',
]
for p in required:
    require(p.exists(), f'present: {p.relative_to(ROOT)}')

# JSON + app/component contract.
package = json.loads(text(MOBILE / 'package.json') or '{}')
appjson = json.loads(text(MOBILE / 'app.json') or '{}')
require(package.get('dependencies', {}).get('react-native') == '0.87.0', 'React Native pinned to 0.87.0')
require(package.get('codegenConfig', {}).get('android', {}).get('javaPackageName') == 'com.aifitnessos.specs', 'Codegen Java package is com.aifitnessos.specs')
component = appjson.get('name')
activity = text(ANDROID / 'app/src/main/java/com/aifitnessos/MainActivity.kt')
require(component and f'getMainComponentName(): String = "{component}"' in activity, 'MainActivity component matches app.json')

# Gradle contract.
root_build = text(ANDROID / 'build.gradle')
app_build = text(ANDROID / 'app/build.gradle')
settings = text(ANDROID / 'settings.gradle')
for value, needle in [
    ('compileSdk 36', 'compileSdkVersion = 36'),
    ('targetSdk 36', 'targetSdkVersion = 36'),
    ('minSdk 24', 'minSdkVersion = 24'),
    ('Kotlin 2.2.0', 'kotlinVersion = "2.2.0"'),
]:
    require(needle in root_build, value)
require('namespace "com.aifitnessos"' in app_build, 'Android namespace is com.aifitnessos')
require('applicationId "com.aifitnessos"' in app_build, 'Android applicationId is com.aifitnessos')
require('buildConfigField "String", "API_BASE_URL"' in app_build, 'BuildConfig API_BASE_URL is defined')
require('autolinkLibrariesWithApp()' in app_build, 'React Native autolinking enabled')
require('com.facebook.react.settings' in settings, 'React Native settings plugin enabled')
require('com.google.mediapipe:tasks-vision:0.10.29' in app_build, 'MediaPipe Tasks Vision pinned to 0.10.29')
require('androidx.camera:camera-core:1.5.3' in app_build, 'CameraX core pinned to stable 1.5.3')
require('androidx.camera:camera-camera2:1.5.3' in app_build, 'CameraX camera2 pinned to stable 1.5.3')
require('androidx.camera:camera-lifecycle:1.5.3' in app_build, 'CameraX lifecycle pinned to stable 1.5.3')
require('androidx.camera:camera-view:1.5.3' in app_build, 'CameraX PreviewView pinned to stable 1.5.3')
require('androidx.health.connect:connect-client:1.1.0' in app_build, 'Health Connect pinned to stable 1.1.0')
require('androidx.work:work-runtime-ktx:2.11.2' in app_build, 'WorkManager pinned to stable 2.11.2')
require('fetchPoseLandmarkerModel' in app_build, 'Pose Landmarker model auto-download is wired to Android build')
require('noCompress "task", "tflite"' in app_build, 'MediaPipe model assets are not compressed')

# XML validity.
xml_files = list((ANDROID / 'app/src/main/res').rglob('*.xml')) + [ANDROID / 'app/src/main/AndroidManifest.xml']
for p in xml_files:
    try:
        ET.parse(p)
        ok(f'valid XML: {p.relative_to(ROOT)}')
    except Exception as exc:
        errors.append(f'invalid XML {p.relative_to(ROOT)}: {exc}')

manifest = text(ANDROID / 'app/src/main/AndroidManifest.xml')
require('android.permission.INTERNET' in manifest, 'Manifest has INTERNET permission')
require('android.permission.CAMERA' in manifest, 'Manifest has CAMERA permission')
require('.technique.TechniqueLiveActivity' in manifest, 'Technique Live Activity registered')
require('android.permission.POST_NOTIFICATIONS' in manifest, 'Manifest has POST_NOTIFICATIONS permission')
for perm in ['READ_STEPS','READ_DISTANCE','READ_ACTIVE_CALORIES_BURNED','READ_SLEEP','READ_EXERCISE','READ_HEART_RATE','READ_HEALTH_DATA_IN_BACKGROUND']:
    require(f'android.permission.health.{perm}' in manifest, f'Manifest has Health Connect {perm} permission')
require('com.google.android.apps.healthdata' in manifest, 'Health Connect provider is query-visible')
require('com.xiaomi.wearable' in manifest, 'Mi Fitness package is query-visible')
require('ACTION_SHOW_PERMISSIONS_RATIONALE' in manifest, 'Health Connect privacy rationale activity registered')
require('android.intent.category.HEALTH_PERMISSIONS' in manifest, 'Android 14 Health permissions rationale alias registered')
require('.resttimer.RestTimerReceiver' in manifest, 'Rest timer receiver registered')
require('${applicationId}.fileprovider' in manifest, 'Body Scan FileProvider registered')
require((ANDROID / 'app/src/main/res/xml/file_paths.xml').exists(), 'Body Scan FileProvider paths present')
require('android:exported="false"' in manifest, 'Rest timer receiver is not exported')

# TurboModule spec/implementation mapping.
specs = {
    'NativeRestTimerNotifications': ('RestTimerNotifications', 'resttimer/RestTimerNotificationsModule.kt'),
    'NativeFoodPhotoPicker': ('FoodPhotoPicker', 'foodphoto/FoodPhotoPickerModule.kt'),
    'NativeAIFitnessConfig': ('AIFitnessConfig', 'nativebridge/AIFitnessConfigModule.kt'),
    'NativeBodyPhotoCapture': ('BodyPhotoCapture', 'bodyphoto/BodyPhotoCaptureModule.kt'),
    'NativeTechniqueVideo': ('TechniqueVideo', 'technique/TechniqueVideoModule.kt'),
    'NativeTechniqueLive': ('TechniqueLive', 'technique/TechniqueLiveModule.kt'),
    'NativeHealthConnect': ('HealthConnect', 'healthconnect/HealthConnectModule.kt'),
}
package_file = text(ANDROID / 'app/src/main/java/com/aifitnessos/nativebridge/AIFitnessNativePackage.kt')
main_app = text(ANDROID / 'app/src/main/java/com/aifitnessos/MainApplication.kt')
require('add(AIFitnessNativePackage())' in main_app, 'AIFitnessNativePackage registered in MainApplication')
for spec_name, (module_name, rel_impl) in specs.items():
    spec = text(MOBILE / 'specs' / f'{spec_name}.ts')
    impl = text(ANDROID / 'app/src/main/java/com/aifitnessos' / rel_impl)
    require(f"TurboModuleRegistry.get<Spec>('{module_name}')" in spec, f'{spec_name} registers JS module name {module_name}')
    require(f': {spec_name}Spec(' in impl, f'{rel_impl} extends generated {spec_name}Spec')
    require(f'const val NAME = "{module_name}"' in impl, f'{rel_impl} exposes module name {module_name}')
    require(f'{Path(rel_impl).stem}.NAME' in package_file, f'{rel_impl} registered in native package')

# Live Technique contracts.
live_activity = text(ANDROID / 'app/src/main/java/com/aifitnessos/technique/TechniqueLiveActivity.kt')
require('ProcessCameraProvider' in live_activity and 'ImageAnalysis' in live_activity and 'PreviewView' in live_activity, 'Technique Live Activity uses CameraX preview + image analysis')
require('PoseLandmarker' in live_activity and 'RunningMode.IMAGE' in live_activity, 'Technique Live Activity performs on-device MediaPipe inference')
require('live_rep_count' in live_activity and 'frames' in live_activity, 'Technique Live Activity exports rep count and sampled landmarks')
require('MAX_SAMPLES = 320' in live_activity, 'Technique Live pose payload is bounded')
health_module = text(ANDROID / 'app/src/main/java/com/aifitnessos/healthconnect/HealthConnectModule.kt')
health_reader = text(ANDROID / 'app/src/main/java/com/aifitnessos/healthconnect/HealthConnectReader.kt')
health_worker = text(ANDROID / 'app/src/main/java/com/aifitnessos/healthconnect/HealthPassiveSyncWorker.kt')
health_store = text(ANDROID / 'app/src/main/java/com/aifitnessos/healthconnect/HealthPassiveSyncStore.kt')
require('com.xiaomi.wearable' in health_reader, 'Health Connect prioritizes Mi Fitness source package')
require('AggregateRequest' in health_reader and 'dataOriginFilter = origins' in health_reader, 'Health cumulative metrics are aggregated and origin-filtered')
require('SleepSessionRecord' in health_reader and 'HeartRateRecord' in health_reader, 'Health Connect reads sleep and exercise heart rate')
require('FEATURE_READ_HEALTH_DATA_IN_BACKGROUND' in health_module, 'Health Connect checks background-read feature availability')
require('CoroutineWorker' in health_worker and 'HealthConnectReader.readSnapshot' in health_worker, 'Passive sync uses WorkManager + Health Connect reader')
require('AndroidKeyStore' in health_store and 'AES/GCM/NoPadding' in health_store, 'Passive health cache is encrypted with Android Keystore AES-GCM')
require('ackPendingSnapshot' in health_module, 'Passive cache requires explicit acknowledgement after upload')
require('ownerUserId' in health_store and 'KEY_ACTIVE_OWNER' in health_store, 'Passive health cache is bound to an authenticated app user')
require('clearPassiveSync' in health_module and 'disableAndClear' in text(ANDROID / 'app/src/main/java/com/aifitnessos/healthconnect/HealthPassiveSyncScheduler.kt'), 'Passive health sync can be cancelled and cleared on logout')
require('pending(reactApplicationContext, ownerUserId.trim())' in health_module, 'Native pending health reads are owner-scoped')
require('bindSessionOwner' in health_module and 'bindAuthenticatedOwner' in text(ANDROID / 'app/src/main/java/com/aifitnessos/healthconnect/HealthPassiveSyncScheduler.kt'), 'Authenticated login suspends passive worker owned by another app user')
require('READ_HEART_RATE_VARIABILITY' not in manifest and 'READ_OXYGEN_SATURATION' not in manifest, 'Health Connect does not over-request unsupported Xiaomi HRV/SpO2')
require('android:screenOrientation="portrait"' in manifest, 'Technique Live Activity is portrait locked for calibration consistency')

# Make sure stale legacy bridge tree is gone.
require(not (MOBILE / 'native/android').exists(), 'legacy apps/mobile/native/android bridge tree removed')

# Gradle wrapper metadata and executable bit.
wrapper_props = text(ANDROID / 'gradle/wrapper/gradle-wrapper.properties')
require('gradle-9.4.1-bin.zip' in wrapper_props, 'Gradle wrapper targets 9.4.1')
require((ANDROID / 'gradlew').stat().st_mode & 0o111 != 0, 'gradlew is executable')

# Wrapper JAR has entry point.
jar = ANDROID / 'gradle/wrapper/gradle-wrapper.jar'
if jar.exists():
    try:
        with zipfile.ZipFile(jar) as archive:
            require('org/gradle/wrapper/GradleWrapperMain.class' in archive.namelist(), 'Gradle wrapper JAR has GradleWrapperMain')
    except Exception as exc:
        errors.append(f'cannot inspect Gradle wrapper jar: {exc}')

# Debug signing key.
keystore = ANDROID / 'app/debug.keystore'
if keystore.exists():
    proc = subprocess.run([
        'keytool', '-list', '-keystore', str(keystore), '-storepass', 'android', '-alias', 'androiddebugkey'
    ], text=True, stdout=subprocess.PIPE, stderr=subprocess.STDOUT)
    require(proc.returncode == 0, 'Android debug keystore is readable and has androiddebugkey')

# Production safety.
api_ts = text(MOBILE / 'src/config/api.ts')
require("!normalized.startsWith('https://')" in api_ts, 'production API URL enforces HTTPS')
require('AI_FITNESS_API_BASE_URL' in app_build, 'Android build accepts AI_FITNESS_API_BASE_URL')
require('release {\n            signingConfig signingConfigs.debug' not in app_build, 'release build does not blindly use debug signing')

print(f'Android native checks passed: {len(checks)}')
if errors:
    print(f'Android native checks failed: {len(errors)}', file=sys.stderr)
    for e in errors:
        print(f'  - {e}', file=sys.stderr)
    sys.exit(1)
for msg in checks:
    print(f'  ✓ {msg}')
