import assert from 'node:assert/strict';
import fs from 'node:fs';
import vm from 'node:vm';
import {loadTypeScript} from './load-typescript.mjs';

const typescript = await loadTypeScript();
const ts = typescript.default ?? typescript;

const memory = new Map();
let owner = 'owner-a';
const source = fs.readFileSync(new URL('../apps/mobile/src/domain/foodIdempotency.ts', import.meta.url), 'utf8');
const compiled = ts.transpileModule(source, {compilerOptions: {module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022}}).outputText;
const module = {exports: {}};
vm.runInNewContext(compiled, {
  module,
  exports: module.exports,
  require: name => {
    if (name === '@react-native-async-storage/async-storage') return {
      getItem: async key => memory.get(key) ?? null,
      setItem: async (key, value) => {memory.set(key, value);},
      removeItem: async key => {memory.delete(key);},
      getAllKeys: async () => [...memory.keys()],
      multiRemove: async keys => {for (const key of keys) memory.delete(key);},
    };
    if (name === '../storage/session') return {currentOwnerUserId: async () => owner};
    throw Error(`Unexpected import ${name}`);
  },
  Date,
  Math,
  Map,
  Error,
  JSON,
  Number,
  encodeURIComponent,
});
const {withFoodIdempotency} = module.exports;

let firstKey;
await assert.rejects(withFoodIdempotency('recipe-1', 'meal-a', async key => {firstKey = key; throw Error('timeout');}), /timeout/);
const storageKey = 'fitness.food-write.v1.owner-a.recipe-1';
assert.ok(memory.has(storageKey), 'failed operation should remain pending');
const pending = JSON.parse(memory.get(storageKey));
pending.createdAt = Date.now() - 48 * 60 * 60 * 1000;
memory.set(storageKey, JSON.stringify(pending));
const retried = await withFoodIdempotency('recipe-1', 'meal-a', async key => key);
assert.equal(retried, firstKey, 'a retry two days later must reuse the key');
assert.equal(memory.has(storageKey), false, 'success clears pending key');
const newAction = await withFoodIdempotency('recipe-1', 'meal-a', async key => key);
assert.notEqual(newAction, firstKey, 'a second completed user action must get a new key');

let resolveWrite;
let sends = 0;
const ongoing = withFoodIdempotency('ai-photo', 'meal-photo', key => {
  sends += 1;
  return new Promise(resolve => {resolveWrite = () => resolve(key);});
});
const simultaneous = withFoodIdempotency('ai-photo', 'meal-photo', () => {sends += 1; return Promise.resolve('bad duplicate');});
await new Promise(resolve => setImmediate(resolve));
assert.equal(sends, 1, 'double tap must send one request');
resolveWrite();
const [result1, result2] = await Promise.all([ongoing, simultaneous]);
assert.equal(result1, result2, 'simultaneous tap shares the original outcome');

owner = 'owner-b';
const other = await withFoodIdempotency('recipe-1', 'meal-a', key => Promise.resolve(key));
assert.notEqual(other, firstKey, 'a different user cannot replay the first user operation');
assert.equal(memory.size, 0);

// Session cleanup must remove only this user's food drafts and other temporary state.
memory.set('fitness.food-write.v1.owner-a.recipe-1', 'private meal A');
memory.set('fitness.food-write.v1.owner-a.ai-photo', 'private meal B');
memory.set('fitness.food-write.v1.owner-b.recipe-1', 'private meal C');
const sessionSource = fs.readFileSync(new URL('../apps/mobile/src/storage/session.ts', import.meta.url), 'utf8');
const sessionCompiled = ts.transpileModule(sessionSource, {compilerOptions: {module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2022}}).outputText;
const sessionModule = {exports: {}};
vm.runInNewContext(sessionCompiled, {
  module: sessionModule,
  exports: sessionModule.exports,
  require: name => {
    if (name === '@react-native-async-storage/async-storage') return {
      getItem: async key => memory.get(key) ?? null,
      setItem: async (key, value) => {memory.set(key, value);},
      removeItem: async key => {memory.delete(key);},
      getAllKeys: async () => [...memory.keys()],
      multiRemove: async keys => {for (const key of keys) memory.delete(key);},
    };
    if (name === 'react-native-keychain') return {getGenericPassword: async () => false};
    throw Error(`Unexpected import ${name}`);
  },
  Date,
  Promise,
  JSON,
  encodeURIComponent,
});
await sessionModule.exports.sessionStorage.clearTransientData('owner-a');
assert.equal(memory.has('fitness.food-write.v1.owner-a.recipe-1'), false);
assert.equal(memory.has('fitness.food-write.v1.owner-a.ai-photo'), false);
assert.equal(memory.get('fitness.food-write.v1.owner-b.recipe-1'), 'private meal C');
console.log('mobile food idempotency retries, concurrency and user scope: PASS');
