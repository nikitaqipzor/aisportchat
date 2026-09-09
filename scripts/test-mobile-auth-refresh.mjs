import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {execFileSync} from 'node:child_process';
import {pathToFileURL} from 'node:url';
import {createRequire} from 'node:module';

async function loadTypeScript() {
  try {
    return await import('typescript');
  } catch {
    const root = execFileSync('npm', ['root', '-g'], {encoding: 'utf8'}).trim();
    return import(pathToFileURL(path.join(root, 'typescript/lib/typescript.js')).href);
  }
}

const tsModule = await loadTypeScript();
const ts = tsModule.default ?? tsModule;
const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'fitness-auth-test-'));
const srcRoot = path.resolve('apps/mobile/src');

function compile(relative) {
  const source = fs.readFileSync(path.join(srcRoot, relative), 'utf8');
  const output = ts.transpileModule(source, {
    fileName: relative,
    compilerOptions: {
      target: ts.ScriptTarget.ES2022,
      module: ts.ModuleKind.CommonJS,
      esModuleInterop: true,
    },
  }).outputText;
  const target = path.join(tmp, 'src', relative.replace(/\.ts$/, '.js'));
  fs.mkdirSync(path.dirname(target), {recursive: true});
  fs.writeFileSync(target, output);
}

compile('config/api.ts');
compile('api/client.ts');
const nativeSpecSource = fs.readFileSync(path.resolve('apps/mobile/specs/NativeAIFitnessConfig.ts'), 'utf8');
const nativeSpecOutput = ts.transpileModule(nativeSpecSource, {compilerOptions: {target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS, esModuleInterop: true}}).outputText;
fs.mkdirSync(path.join(tmp, 'specs'), {recursive: true});
fs.writeFileSync(path.join(tmp, 'specs/NativeAIFitnessConfig.js'), nativeSpecOutput);
fs.mkdirSync(path.join(tmp, 'node_modules/react-native'), {recursive: true});
fs.writeFileSync(path.join(tmp, 'node_modules/react-native/index.js'), `
exports.NativeModules = {SourceCode: {scriptURL: 'http://10.0.2.2:8081/index.bundle'}};
exports.TurboModuleRegistry = {get(){ return null; }};
`);

global.__DEV__ = true;
global.__AI_FITNESS_API_BASE_URL__ = 'http://test.local/api/v1';

const requireFromTmp = createRequire(path.join(tmp, 'entry.js'));
const {api} = requireFromTmp('./src/api/client.js');

const expired = {access_token: 'expired-access', refresh_token: 'refresh-1', access_expires_at: new Date(0).toISOString()};
const fresh = {access_token: 'fresh-access', refresh_token: 'refresh-2', access_expires_at: new Date(Date.now() + 60_000).toISOString()};
let stored = expired;
let changed = null;
let refreshCalls = 0;
let protectedCalls = 0;

global.fetch = async (url, init = {}) => {
  const value = String(url);
  if (value.endsWith('/auth/refresh')) {
    refreshCalls += 1;
    await new Promise(resolve => setTimeout(resolve, 15));
    return new Response(JSON.stringify({tokens: fresh}), {status: 200, headers: {'Content-Type': 'application/json'}});
  }
  protectedCalls += 1;
  const authorization = new Headers(init.headers).get('Authorization');
  if (authorization !== 'Bearer fresh-access') {
    return new Response(JSON.stringify({error: 'expired'}), {status: 401, headers: {'Content-Type': 'application/json'}});
  }
  if (value.endsWith('/profile')) {
    return new Response(JSON.stringify({profile: {}}), {status: 200, headers: {'Content-Type': 'application/json'}});
  }
  return new Response(JSON.stringify({profile_completed: true, goal_completed: true, training_completed: true, completed: true}), {status: 200, headers: {'Content-Type': 'application/json'}});
};

api.configureAuthSession({
  loadTokens: async () => stored,
  saveTokens: async tokens => { stored = tokens; },
  clearTokens: async () => { stored = null; },
  onTokensChanged: tokens => { changed = tokens; },
});
api.setCurrentTokens(expired);

const [status, profile] = await Promise.all([
  api.onboardingStatus(expired.access_token),
  api.getProfile(expired.access_token),
]);

assert.equal(status.completed, true);
assert.ok(profile.profile);
assert.equal(refreshCalls, 1, 'concurrent 401s must share one refresh request');
assert.equal(stored.access_token, 'fresh-access');
assert.equal(changed.access_token, 'fresh-access');
assert.ok(protectedCalls >= 4, 'both requests should be retried after refresh');
console.log(`mobile auth refresh: PASS (refreshCalls=${refreshCalls}, protectedCalls=${protectedCalls})`);
