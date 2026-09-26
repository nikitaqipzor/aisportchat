// Run with: node src/screens/ActiveWorkoutScreen.test.mjs
import assert from 'node:assert/strict';
import fs from 'node:fs';
import vm from 'node:vm';
import ts from 'typescript';

const hooks = [];
let cursor = 0;
let alert;
let pendingLog;
let pendingFinish;
let logCalls = 0;
let finishCalls = 0;
let cancelCalls = 0;
let queueCalls = 0;
let queueFails = false;
let cacheFails = false;
let networkError = false;
let currentWorkout;
const defer = () => {
  let resolve;
  let reject;
  const promise = new Promise((yes, no) => { resolve = yes; reject = no; });
  return {promise, resolve, reject};
};
const react = {
  createElement: (type, props, ...children) => ({type, props: {...props, children}}),
  useState(initial) {
    const index = cursor++;
    if (!(index in hooks)) hooks[index] = typeof initial === 'function' ? initial() : initial;
    return [hooks[index], value => { hooks[index] = typeof value === 'function' ? value(hooks[index]) : value; }];
  },
  useRef(initial) {
    const index = cursor++;
    return hooks[index] ??= {current: initial};
  },
  useEffect: () => {},
  useMemo: factory => factory(),
};
const styles = {create: value => value};
const native = {
  ActivityIndicator: 'ActivityIndicator', Alert: {alert: (...args) => { alert = args; }},
  AppState: {addEventListener: () => ({remove: () => {}})},
  Pressable: 'Pressable', ScrollView: 'ScrollView', StyleSheet: styles,
  Text: 'Text', TextInput: 'TextInput', View: 'View',
};
const modules = {
  react: {...react, default: react}, 'react-native': native,
  '../api/client': {api: {logSet: () => { logCalls++; return pendingLog.promise; }, isNetworkError: () => networkError}},
  '../components/ExerciseGuide': {ExerciseGuide: 'ExerciseGuide'},
  '../components/AppButton': {AppButton: 'AppButton'},
  '../storage/session': {sessionStorage: {
    enqueueSet: async () => { queueCalls++; if (queueFails) throw Error('storage full'); },
    saveActiveWorkout: async () => { if (cacheFails) throw Error('cache full'); },
  }},
  '../domain/restTimer': {restSecondsRemaining: () => 0, restTimerCoordinator: {invalidate: async () => {}, start: async () => {}}},
  '../domain/technique': {techniqueKeyForExercise: () => null},
  '../theme/tokens': {colors: new Proxy({}, {get: () => 'black'}), control: {minTouch: 44, buttonHeight: 48}, radius: {md: 8, lg: 12, pill: 99}, spacing: {lg: 16, md: 8, xl: 24}},
};
const source = fs.readFileSync(new URL('./ActiveWorkoutScreen.tsx', import.meta.url), 'utf8');
const js = ts.transpileModule(source, {compilerOptions: {module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.React, esModuleInterop: true}}).outputText;
const sandbox = {exports: {}, require: id => { assert.ok(id in modules, `unexpected import: ${id}`); return modules[id]; }, console, Date, setInterval, clearInterval};
vm.runInNewContext(js, sandbox, {filename: 'ActiveWorkoutScreen.js'});
const Screen = sandbox.exports.ActiveWorkoutScreen;
const makeWorkout = () => ({workout: {id: 'w1'}, exercises: [{exercise: {id: 'e1', name: 'Squat'}, workout_exercise: {id: 'we1', target_sets: 2, target_reps_min: 5, target_reps_max: 8, rest_seconds: 30}, sets: []}]});
currentWorkout = makeWorkout();
const props = {
  accessToken: 'token',
  get workout() { return currentWorkout; },
  onWorkoutChange: updated => { currentWorkout = updated; },
  onFinish: () => { finishCalls++; return pendingFinish.promise; },
  onCancel: async () => { cancelCalls++; },
};
function render() { cursor = 0; return Screen(props); }
function find(node, id) {
  if (!node || typeof node !== 'object') return undefined;
  if (node.props?.testID === id) return node;
  return node.props?.children?.map(child => find(child, id)).find(Boolean);
}
const tick = () => new Promise(resolve => setImmediate(resolve));

pendingLog = defer();
let tree = render();
hooks[2] = '5'; // Simulate the default reps filled by the mount effect.
tree = render();
find(tree, 'workout-save-set').props.onPress();
find(tree, 'workout-save-set').props.onPress();
find(tree, 'workout-finish').props.onPress();
find(tree, 'workout-cancel').props.onPress();
assert.equal(logCalls, 1, 'repeat save must not send a second log');
assert.equal(alert, undefined, 'finish/cancel cannot start during pending log');
tree = render();
assert.equal(find(tree, 'workout-finish').props.disabled, true);
assert.equal(find(tree, 'workout-cancel').props.disabled, true);
pendingLog.resolve({...currentWorkout, exercises: [{...currentWorkout.exercises[0], sets: [{id: 's1', set_number: 1, repetitions: 5}]}]});
await tick();
tree = render();
assert.equal(currentWorkout.exercises[0].sets.length, 1);
assert.equal(find(tree, 'workout-finish').props.disabled, false);
find(tree, 'workout-finish').props.onPress();
find(tree, 'workout-cancel').props.onPress();
assert.match(alert[0], /Завершить/);
alert[3].onDismiss();
find(tree, 'workout-finish').props.onPress();
pendingFinish = defer();
alert[2][1].onPress();
alert[2][1].onPress();
await tick();
assert.equal(finishCalls, 1, 'repeat confirmation must not finish twice');
find(tree, 'workout-save-set').props.onPress();
assert.equal(logCalls, 1, 'save cannot overlap finish');
pendingFinish.reject(Error('server failed'));
await tick();
tree = render();
assert.equal(find(tree, 'workout-finish').props.disabled, false, 'failed finish unlocks actions');
find(tree, 'workout-cancel').props.onPress();
alert[3].onDismiss();
find(tree, 'workout-cancel').props.onPress();
alert[2][1].onPress();
alert[2][1].onPress();
await tick();
assert.equal(cancelCalls, 1, 'repeat cancellation must not cancel twice');

// Local queue failure keeps the set visible as unsaved and allows retry.
networkError = true;
queueFails = true;
pendingLog = defer();
tree = render();
find(tree, 'workout-save-set').props.onPress();
pendingLog.reject(Error('offline'));
await tick();
tree = render();
assert.equal(queueCalls, 1);
assert.equal(currentWorkout.exercises[0].sets.length, 1);
assert.equal(find(tree, 'workout-save-set').props.disabled, false);
queueFails = false;
cacheFails = true;
pendingLog = defer();
find(tree, 'workout-save-set').props.onPress();
pendingLog.reject(Error('offline'));
await tick();
assert.equal(queueCalls, 2);
assert.equal(currentWorkout.exercises[0].sets.length, 2);
assert.ok(JSON.stringify(render()).includes('не удалось обновить локальную тренировку'));
console.log('ActiveWorkoutScreen async action guards: PASS');
