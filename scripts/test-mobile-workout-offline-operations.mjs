import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {createRequire} from 'node:module';
import {loadTypeScript} from './load-typescript.mjs';

const mod = await loadTypeScript();
const ts = mod.default ?? mod;
const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'fitness-offline-operations-'));
for (const name of ['session', 'sync']) {
  const source = fs.readFileSync(`apps/mobile/src/storage/${name}.ts`, 'utf8');
  fs.mkdirSync(path.join(tmp, 'storage'), {recursive: true});
  fs.writeFileSync(path.join(tmp, 'storage', `${name}.js`), ts.transpileModule(source, {compilerOptions: {target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS, esModuleInterop: true}}).outputText);
}
const client = fs.readFileSync('apps/mobile/src/api/client.ts', 'utf8');
assert.match(client, /'Idempotency-Key': operationId/g);
fs.mkdirSync(path.join(tmp, 'node_modules/@react-native-async-storage/async-storage'), {recursive: true});
fs.writeFileSync(path.join(tmp, 'node_modules/@react-native-async-storage/async-storage/index.js'), `
const data = new Map(); module.exports = {
  async getItem(k){ return data.get(k) ?? null; }, async setItem(k,v){ data.set(k,v); },
  async removeItem(k){ data.delete(k); }, async getAllKeys(){ return [...data.keys()]; }
};`);
fs.mkdirSync(path.join(tmp, 'node_modules/react-native-keychain'), {recursive: true});
fs.writeFileSync(path.join(tmp, 'node_modules/react-native-keychain/index.js'), `
let value; exports.ACCESSIBLE = {WHEN_UNLOCKED_THIS_DEVICE_ONLY:'secure'};
exports.setGenericPassword = async (username,password) => (value={username,password});
exports.getGenericPassword = async () => value ?? false;
exports.resetGenericPassword = async () => (value=undefined);`);
fs.mkdirSync(path.join(tmp, 'api'), {recursive: true});
fs.writeFileSync(path.join(tmp, 'api/client.js'), `module.exports = {api: global.__offlineApi};`);
const requireTmp = createRequire(path.join(tmp, 'entry.js'));
const {sessionStorage} = requireTmp('./storage/session.js');
const tokens = {access_token:'access', refresh_token:'refresh', access_expires_at:'2099-01-01T00:00:00Z'};
await sessionStorage.saveTokens(tokens, 'user-A');
await sessionStorage.enqueueSet('workout-1', {workout_exercise_id:'exercise-1', set_number:1, repetitions:8});
const [setOperation] = await sessionStorage.loadQueue();
const calls = [];
let ambiguous = true;
global.__offlineApi = {
  logSet: async (_token, _workout, payload, id) => {
    calls.push(id);
    assert.equal(id, setOperation.id);
    if (ambiguous) { ambiguous = false; throw new Error('response lost after commit'); }
    return {workout:{id:'workout-1',status:'active'}, exercises:[], payload};
  },
  finishWorkout: async (_token, _workout, id) => { calls.push(id); return {workout:{workout:{id:'workout-1',status:'completed'},exercises:[]}}; },
  cancelWorkout: async (_token, _workout, id) => { calls.push(id); return {workout:{id:'workout-1',status:'cancelled'},exercises:[]}; },
};
const {flushOfflineQueue} = requireTmp('./storage/sync.js');
assert.equal((await flushOfflineQueue(tokens)).remaining, 1, 'ambiguous response retains operation');
assert.equal((await sessionStorage.loadQueue())[0].id, setOperation.id, 'retry keeps original key');
assert.equal((await flushOfflineQueue(tokens)).remaining, 0);
assert.deepEqual(calls, [setOperation.id, setOperation.id]);
await sessionStorage.enqueueFinish('workout-1');
const finishId = (await sessionStorage.loadQueue())[0].id;
assert.equal((await flushOfflineQueue(tokens)).finish.workout.workout.status, 'completed');
assert.equal(calls.at(-1), finishId);
await sessionStorage.enqueueCancel('workout-2');
const cancelId = (await sessionStorage.loadQueue())[0].id;
await flushOfflineQueue(tokens);
assert.equal(calls.at(-1), cancelId);
console.log('mobile workout offline operation retries: PASS');
