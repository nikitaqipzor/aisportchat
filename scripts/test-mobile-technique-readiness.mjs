import assert from 'node:assert/strict';
import fs from 'node:fs';

const screen = fs.readFileSync('apps/mobile/src/screens/TechniqueScreen.tsx', 'utf8');
const activity = fs.readFileSync('apps/mobile/android/app/src/main/java/com/aifitnessos/technique/TechniqueLiveActivity.kt', 'utf8');
const module = fs.readFileSync('apps/mobile/android/app/src/main/java/com/aifitnessos/technique/TechniqueLiveModule.kt', 'utf8');

assert.match(screen, /testID="technique-preflight"/, 'Technique requires an explicit preflight');
assert.match(screen, /accessibilityRole="checkbox"/, 'preflight checklist is accessible');
assert.match(screen, /Object\.values\(preflight\)\.every\(Boolean\)/, 'capture remains disabled until preflight passes');
assert.match(screen, /techniqueCaptureError/, 'capture errors use cancellation/permission-aware messaging');
assert.match(screen, /if \(workoutContext\) setSelected\(linkedKey \? catalog\.items\.find/,
  'an unsupported linked workout must not silently select another movement');
assert.match(screen, /result && result\.rep_count > 0 && workoutContext && linkedKey && onUseLinkedResult/,
  'only a safely mapped movement with a positive rep count can write repetitions back to a workout');
assert.match(screen, /result && result\.rep_count === 0 && workoutContext && linkedKey && !resultFromHistory/,
  'a linked zero-rep result must use the non-transferable warning state');
assert.match(screen, /нулевой результат нельзя перенести в тренировку/,
  'zero-rep result must explain why it cannot be copied');
assert.match(activity, /TECHNIQUE_CAMERA_PERMISSION_DENIED/, 'native activity identifies permission denial');
assert.match(activity, /TECHNIQUE_LIVE_CANCELLED/, 'native activity identifies user cancellation');
assert.match(module, /EXTRA_ERROR_CODE/, 'native module forwards structured error codes');

console.log('mobile technique readiness: PASS');
