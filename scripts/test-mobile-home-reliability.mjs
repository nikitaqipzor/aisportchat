import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {createRequire} from 'node:module';
import {loadTypeScript} from './load-typescript.mjs';

const home = fs.readFileSync('apps/mobile/src/screens/HomeScreen.tsx', 'utf8');
const manual = fs.readFileSync('apps/mobile/src/screens/ManualWorkoutScreen.tsx', 'utf8');
assert.match(home, /useState<Environment \| null>\(null\)/, 'Home must not assume a default gym');
assert.match(home, /disabled=\{!environmentReady\}/, 'manual CTA must be blocked without verified environment');
assert.match(home, /home-body-map-blocked/, 'body map must be blocked without verified environment');
assert.match(home, /cached-stale/, 'Home must expose stale cached profile state');
assert.match(home, /readinessGuard/, 'readiness responses must be latest-request guarded');
assert.match(manual, /useState<Environment\|null>\(null\)/, 'manual workout must not assume a default gym');
assert.match(manual, /if\(!environment\|\|!allowedEnvironments\.includes\(environment\)\)/, 'manual creation must reject environment bypass');

const mod = await loadTypeScript();
const ts = mod.default ?? mod;
const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'fitness-home-reliability-'));
const source = fs.readFileSync('apps/mobile/src/storage/session.ts', 'utf8');
const output = ts.transpileModule(source, {compilerOptions: {target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS, esModuleInterop: true}}).outputText;
fs.mkdirSync(path.join(tmp, 'storage'), {recursive: true});
fs.writeFileSync(path.join(tmp, 'storage/session.js'), output);
fs.mkdirSync(path.join(tmp, 'node_modules/@react-native-async-storage/async-storage'), {recursive: true});
fs.writeFileSync(path.join(tmp, 'node_modules/@react-native-async-storage/async-storage/index.js'), `
const data = new Map(); module.exports = {__data:data, async getItem(k){return data.get(k)??null}, async setItem(k,v){data.set(k,v)}, async removeItem(k){data.delete(k)}};
`);
fs.mkdirSync(path.join(tmp, 'node_modules/react-native-keychain'), {recursive: true});
fs.writeFileSync(path.join(tmp, 'node_modules/react-native-keychain/index.js'), `
let value=null; exports.ACCESSIBLE={WHEN_UNLOCKED_THIS_DEVICE_ONLY:'secure'}; exports.setGenericPassword=async(username,password)=>{value={username,password};return true}; exports.getGenericPassword=async()=>value||false; exports.resetGenericPassword=async()=>{value=null;return true};
`);
const requireFromTmp = createRequire(path.join(tmp, 'entry.js'));
const {sessionStorage} = requireFromTmp('./storage/session.js');
const asyncStorage = requireFromTmp('@react-native-async-storage/async-storage');
const token = user => ({access_token:`access-${user}`,refresh_token:`refresh-${user}`,access_expires_at:new Date(Date.now()+60_000).toISOString()});

await sessionStorage.saveTokens(token('A'),'user-A');
await sessionStorage.saveTrainingEnvironments(['gym'], 'user-A');
const raw = [...asyncStorage.__data.values()].find(value => value.includes('cachedAt'));
const cachedEnvironment = JSON.parse(raw);
assert.deepEqual(Object.keys(cachedEnvironment).sort(), ['cachedAt','environments'], 'cache schema must remain minimal');
assert.equal(raw.includes('access-A'), false, 'cache must not contain tokens or PII');
await sessionStorage.saveTokens(token('B'),'user-B');
assert.equal(await sessionStorage.saveTrainingEnvironments(['home'],'user-A'), false, 'deferred A response must not write under B');
assert.equal(await sessionStorage.loadTrainingEnvironments(), null, 'B must never read A cache');
await sessionStorage.saveTrainingEnvironments(['band'],'user-B');
await sessionStorage.clearTokens();
assert.ok([...asyncStorage.__data.values()].some(value=>value.includes('band')), 'automatic token expiry must preserve environment cache');

console.log('mobile home reliability: PASS');
