import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {createRequire} from 'node:module';
import {loadTypeScript} from './load-typescript.mjs';
const mod = await loadTypeScript();
const ts = mod.default ?? mod;
const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'fitness-session-test-'));
const source = fs.readFileSync('apps/mobile/src/storage/session.ts', 'utf8');
const output = ts.transpileModule(source, {compilerOptions: {target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS, esModuleInterop: true}}).outputText;
fs.mkdirSync(path.join(tmp, 'storage'), {recursive: true});
fs.writeFileSync(path.join(tmp, 'storage/session.js'), output);

fs.mkdirSync(path.join(tmp, 'node_modules/@react-native-async-storage/async-storage'), {recursive: true});
fs.writeFileSync(path.join(tmp, 'node_modules/@react-native-async-storage/async-storage/index.js'), `
const data = new Map();
module.exports = {
  __data: data,
  async getItem(k){ return data.has(k) ? data.get(k) : null; },
  async setItem(k,v){ data.set(k,v); },
  async removeItem(k){ data.delete(k); },
  async multiRemove(keys){ for (const k of keys) data.delete(k); },
};
`);
fs.mkdirSync(path.join(tmp, 'node_modules/react-native-keychain'), {recursive: true});
fs.writeFileSync(path.join(tmp, 'node_modules/react-native-keychain/index.js'), `
let value = null;
exports.ACCESSIBLE = {WHEN_UNLOCKED_THIS_DEVICE_ONLY: 'secure'};
exports.setGenericPassword = async (username,password) => { value = {username,password}; return true; };
exports.getGenericPassword = async () => value || false;
exports.resetGenericPassword = async () => { value = null; return true; };
`);

const requireFromTmp = createRequire(path.join(tmp, 'entry.js'));
const {sessionStorage} = requireFromTmp('./storage/session.js');
const token = user => ({access_token: `access-${user}`, refresh_token: `refresh-${user}`, access_expires_at: new Date(Date.now()+60_000).toISOString()});

await sessionStorage.saveTokens(token('A'), 'user-A');
await sessionStorage.enqueueSet('workout-A', {workout_exercise_id: 'exercise-A', set_number: 1, repetitions: 10});
assert.equal((await sessionStorage.loadQueue()).length, 1);

// Even without an explicit cleanup, changing account ownership hides and deletes A's envelope.
await sessionStorage.saveTokens(token('B'), 'user-B');
assert.equal((await sessionStorage.loadQueue()).length, 0, 'user B must never see user A queue');

await sessionStorage.enqueueSet('workout-B', {workout_exercise_id: 'exercise-B', set_number: 1, repetitions: 12});
assert.equal((await sessionStorage.loadQueue()).length, 1);
await sessionStorage.clearTransientData();
assert.equal((await sessionStorage.loadQueue()).length, 0, 'logout cleanup must clear queue');
console.log('mobile session scope: PASS');
