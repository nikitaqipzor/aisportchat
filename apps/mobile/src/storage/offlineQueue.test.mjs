// Run with: node src/storage/offlineQueue.test.mjs
import assert from 'node:assert/strict';
import fs from 'node:fs';
import vm from 'node:vm';
import ts from 'typescript';

const db = new Map();
let credential = false;
const storage = {
  getItem: async key => db.get(key) ?? null,
  setItem: async (key, value) => { db.set(key, value); },
  removeItem: async key => { db.delete(key); },
  getAllKeys: async () => [...db.keys()],
};
const keychain = {
  ACCESSIBLE: {WHEN_UNLOCKED_THIS_DEVICE_ONLY: 'secure'},
  getGenericPassword: async () => credential,
  setGenericPassword: async (username, password) => { credential = {username, password}; return true; },
  resetGenericPassword: async () => { credential = false; },
};
const defer = () => {
  let resolve;
  let reject;
  const promise = new Promise((yes, no) => { resolve = yes; reject = no; });
  return {promise, resolve, reject};
};
let network;
let logCalls = 0;
const api = {
  logSet: async (_token, workoutId) => { logCalls++; await network.promise; return {workout: {id: workoutId, status: 'active'}, exercises: []}; },
  finishWorkout: async (_token, workoutId) => ({workout: {workout: {id: workoutId, status: 'completed'}, exercises: []}}),
  cancelWorkout: async (_token, workoutId) => ({workout: {id: workoutId, status: 'cancelled'}, exercises: []}),
};
function compile(file, dependencies) {
  const source = fs.readFileSync(new URL(file, import.meta.url), 'utf8');
  const output = ts.transpileModule(source, {compilerOptions: {target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS, esModuleInterop: true}}).outputText;
  const sandbox = {exports: {}, require: name => { assert.ok(name in dependencies, `unknown module ${name}`); return dependencies[name]; }, Date, Math};
  vm.runInNewContext(output, sandbox, {filename: file});
  return sandbox.exports;
}
const {sessionStorage} = compile('./session.ts', {
  '@react-native-async-storage/async-storage': storage, 'react-native-keychain': keychain,
});
const {flushOfflineQueue} = compile('./sync.ts', {'../api/client': {api}, './session': {sessionStorage}});
const tokens = {access_token: 'A', refresh_token: 'R', access_expires_at: ''};
const payload = (number, repetitions = 6) => ({workout_exercise_id: 'exercise', set_number: number, repetitions});
await sessionStorage.saveTokens(tokens, 'owner-A');
await sessionStorage.enqueueSet('w1', payload(1));
network = defer();
const first = flushOfflineQueue(tokens);
while (logCalls === 0) await new Promise(resolve => setImmediate(resolve));
await sessionStorage.enqueueSet('w1', payload(2));
network.resolve();
const result = await first;
assert.equal(result.remaining, 1, 'a new set must survive an old flush snapshot');
assert.equal(result.workout, undefined, 'stale server snapshot must not overwrite optimistic state');
assert.equal((await sessionStorage.loadQueue())[0].payload.set_number, 2);

// Replacing a set while its earlier version is in flight retains the new ID.
network = defer();
const second = flushOfflineQueue(tokens);
while (logCalls < 2) await new Promise(resolve => setImmediate(resolve));
await sessionStorage.enqueueSet('w1', payload(2, 8));
network.resolve();
assert.equal((await second).remaining, 1);
assert.equal((await sessionStorage.loadQueue())[0].payload.repetitions, 8);

// Two concurrent flushes serialize, so an acknowledged ID is not replayed.
network = defer();
const before = logCalls;
const a = flushOfflineQueue(tokens);
const b = flushOfflineQueue(tokens);
while (logCalls === before) await new Promise(resolve => setImmediate(resolve));
network.resolve();
await Promise.all([a, b]);
assert.equal(logCalls, before + 1, 'concurrent flush must not replay the same operation');
assert.equal((await sessionStorage.loadQueue()).length, 0);

// A failed request leaves its operation and all following operations pending.
await sessionStorage.enqueueSet('w1', payload(3));
await sessionStorage.enqueueFinish('w1');
network = defer();
const failed = flushOfflineQueue(tokens);
while (logCalls === before + 1) await new Promise(resolve => setImmediate(resolve));
network.reject(Error('offline'));
assert.equal((await failed).remaining, 2);
assert.deepEqual(Array.from(await sessionStorage.loadQueue(), item => item.type), ['log_set', 'finish_workout']);

// Owner switch while the request is outstanding must never mutate B's queue.
network = defer();
const switched = flushOfflineQueue(tokens);
while (logCalls === before + 2) await new Promise(resolve => setImmediate(resolve));
await sessionStorage.saveTokens({...tokens, access_token: 'B'}, 'owner-B');
await sessionStorage.enqueueSet('w2', payload(1));
network.resolve();
await switched;
assert.equal((await sessionStorage.loadQueue())[0].workoutId, 'w2');
await sessionStorage.saveTokens(tokens, 'owner-A');
assert.equal((await sessionStorage.loadQueue()).length, 2, 'A keeps operations for recovery');
// Concurrent local writes retain all sets; cancellation stays after sets.
await Promise.all([
  sessionStorage.enqueueSet('w3', payload(1)),
  sessionStorage.enqueueSet('w3', payload(2)),
]);
await sessionStorage.enqueueFinish('w3');
await sessionStorage.enqueueCancel('w3');
const w3 = (await sessionStorage.loadQueue()).filter(item => item.workoutId === 'w3');
assert.deepEqual(Array.from(w3, item => item.type), ['log_set', 'log_set', 'cancel_workout']);
await sessionStorage.removeWorkoutOperations('w3');
assert.equal((await sessionStorage.loadQueue()).filter(item => item.workoutId === 'w3').length, 0);
db.set('fitness.ai-chat.v3.owner-A', 'private chat');
db.set('fitness.training-environments.v3.owner-A', 'private settings');
db.set('fitness.food-write.v1.owner-A.meal', 'private draft');
db.set('fitness.rest-timer.v1', 'private timer');
db.set('fitness.ai-chat.v3.owner-B', 'other user');
await sessionStorage.clearTokens();
await sessionStorage.clearDeletedAccountData('owner-A');
assert.ok(![...db.keys()].some(key => key.includes('owner-A') || key === 'fitness.rest-timer.v1'));
assert.equal(db.get('fitness.ai-chat.v3.owner-B'), 'other user');
db.set('fitness.ai-chat.v3.owner-B', 'legacy owner unknown');
await sessionStorage.clearDeletedAccountData();
assert.equal(db.size, 0, 'legacy sessions without an owner purge all app-owned caches');
console.log('mobile offline queue concurrency: PASS');
