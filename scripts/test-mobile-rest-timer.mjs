import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {createRequire} from 'node:module';
import {loadTypeScript} from './load-typescript.mjs';

const mod = await loadTypeScript();
const ts = mod.default ?? mod;
const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'fitness-rest-timer-'));
const source = fs.readFileSync('apps/mobile/src/domain/restTimer.ts', 'utf8');
const output = ts.transpileModule(source, {compilerOptions: {target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS, esModuleInterop: true}}).outputText;
fs.mkdirSync(path.join(tmp, 'domain'), {recursive: true});
fs.writeFileSync(path.join(tmp, 'domain/restTimer.js'), output);
fs.mkdirSync(path.join(tmp, 'native'), {recursive: true});
fs.writeFileSync(path.join(tmp, 'native/restTimer.js'), 'exports.restTimerNotifications={schedule:async()=>true,cancel:async()=>true};');
fs.mkdirSync(path.join(tmp, 'storage'), {recursive: true});
fs.writeFileSync(path.join(tmp, 'storage/session.js'), "exports.currentOwnerUserId=async()=>\'owner\';");
fs.mkdirSync(path.join(tmp, 'node_modules/@react-native-async-storage/async-storage'), {recursive: true});
fs.writeFileSync(path.join(tmp, 'node_modules/@react-native-async-storage/async-storage/index.js'), 'module.exports={getItem:async()=>null,setItem:async()=>{},removeItem:async()=>{}};');

const requireFromTmp = createRequire(path.join(tmp, 'entry.js'));
const {RestTimerCoordinator, restSecondsRemaining} = requireFromTmp('./domain/restTimer.js');
const deferred = () => { let resolve; let reject; const promise = new Promise((ok, no) => { resolve = ok; reject = no; }); return {promise, resolve, reject}; };

function harness(overrides = {}) {
  let now = 10_000;
  let stored = null;
  const calls = [];
  const dependencies = {
    now: () => now,
    owner: async () => 'owner-A',
    read: async () => stored,
    write: async value => { calls.push('save'); stored = value; },
    remove: async () => { calls.push('clear'); stored = null; },
    schedule: async seconds => { calls.push(`schedule:${seconds}`); return true; },
    cancel: async () => { calls.push('cancel'); return true; },
    ...overrides,
  };
  return {timer: new RestTimerCoordinator(dependencies), calls, get stored() { return stored; }, set stored(value) { stored = value; }, setNow(value) { now = value; }};
}

assert.equal(restSecondsRemaining(12_001, 10_000), 3, 'deadline rounds up partial seconds');
assert.equal(restSecondsRemaining(9_999, 10_000), 0, 'expired deadlines clamp to zero');

const normal = harness();
await normal.timer.start('workout-A', 'Squat', 30);
normal.setNow(25_000);
assert.equal(restSecondsRemaining(JSON.parse(normal.stored).deadlineMs, 25_000), 15);
assert.equal((await normal.timer.restore('workout-A')).workoutId, 'workout-A');
normal.setNow(41_000);
assert.equal(await normal.timer.restore('workout-A'), null, 'expired timer is not restored');
assert.equal(normal.stored, null, 'expired snapshot is removed');

for (const blockedStep of ['owner', 'write', 'schedule']) {
  const gate = deferred();
  const overrides = blockedStep === 'owner'
    ? {owner: () => gate.promise}
    : blockedStep === 'write'
      ? {write: () => gate.promise}
      : {schedule: () => gate.promise};
  const item = harness(overrides);
  const starting = item.timer.start('workout-A', 'Row', 20);
  await Promise.resolve();
  if (blockedStep !== 'owner') await Promise.resolve();
  const invalidating = item.timer.invalidate();
  gate.resolve(blockedStep === 'owner' ? 'owner-A' : blockedStep === 'schedule' ? true : undefined);
  await Promise.all([starting, invalidating]);
  assert.equal(item.stored, null, `destructive interruption during ${blockedStep} must compensate`);
  assert.ok(item.calls.includes('cancel'), `destructive interruption during ${blockedStep} must cancel native alarm`);
}

let firstRemove = true;
const resilient = harness({remove: async () => { if (firstRemove) { firstRemove = false; throw new Error('storage down'); } }});
await resilient.timer.invalidate();
await resilient.timer.start('workout-A', 'Press', 12);
assert.ok(resilient.calls.includes('schedule:12'), 'rejected first task must not poison the next FIFO task');

const ordered = harness();
await ordered.timer.start('workout-A', 'Old', 10);
const skip = ordered.timer.invalidate();
const next = ordered.timer.start('workout-A', 'New', 25);
await Promise.all([skip, next]);
assert.equal(JSON.parse(ordered.stored).exerciseName, 'New', 'skip followed immediately by start keeps the new timer');
assert.equal(ordered.calls.at(-1), 'schedule:25');

const screen = fs.readFileSync('apps/mobile/src/screens/ActiveWorkoutScreen.tsx', 'utf8');
assert.ok(screen.includes("AppState.addEventListener('change'"), 'foreground restore is missing');
const cleanup = screen.slice(screen.indexOf('return () => {'), screen.indexOf('}, [workout.workout.id])'));
assert.ok(!cleanup.includes('invalidate'), 'ordinary unmount/Technique navigation must not destroy timer');

const program = fs.readFileSync('apps/mobile/src/screens/ProgramDetailScreen.tsx', 'utf8');
const startHandler = program.slice(program.indexOf('async function start'), program.indexOf('async function autoMove'));
assert.ok(startHandler.includes('catch(e)') && startHandler.includes("finally{setBusy('')}"), 'program start must contain errors and always release busy state');

console.log('mobile rest timer: PASS');
