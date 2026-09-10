import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {createRequire} from 'node:module';
import {loadTypeScript} from './load-typescript.mjs';

const tsModule = await loadTypeScript();
const ts = tsModule.default ?? tsModule;
const sourcePath = path.resolve('apps/mobile/src/analytics/latestRequestGuard.ts');
const output = ts.transpileModule(fs.readFileSync(sourcePath, 'utf8'), {
  compilerOptions: {target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS},
}).outputText;
const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'fitness-analytics-races-'));
fs.writeFileSync(path.join(tmp, 'guard.js'), output);
const {beginLoad, failLoad, LatestRequestGuard, LatestRefreshOwner} = createRequire(path.join(tmp, 'entry.js'))('./guard.js');

const guard = new LatestRequestGuard();
guard.mount();

// Inverted completion: the later response owns the section.
const slowA = guard.begin('summary');
const fastB = guard.begin('summary');
assert.equal(guard.isLatest(fastB), true);
assert.equal(guard.isLatest(slowA), false);

// Different keys remain independent.
const history = guard.begin('history');
assert.equal(guard.isLatest(history), true);
assert.equal(guard.isLatest(fastB), true);

// A-B-A on one key still accepts only the final A generation.
const a1 = guard.begin('health');
guard.begin('health');
const a2 = guard.begin('health');
assert.equal(guard.isLatest(a1), false);
assert.equal(guard.isLatest(a2), true);

// Unmount invalidates every in-flight completion.
const beforeUnmount = guard.begin('nutrition');
guard.unmount();
assert.equal(guard.isLatest(beforeUnmount), false);
guard.mount();

// Token rotation invalidates old-token responses, including same-key requests.
const oldToken = guard.begin('workouts');
guard.rotateAuth();
const newToken = guard.begin('workouts');
assert.equal(guard.isLatest(oldToken), false);
assert.equal(guard.isLatest(newToken), true);

// Two overlapping refreshes: only the latest may release the spinner.
const refreshes = new LatestRefreshOwner();
const firstRefresh = refreshes.begin();
const secondRefresh = refreshes.begin();
assert.equal(refreshes.release(firstRefresh), false);
assert.equal(refreshes.release(secondRefresh), true);

// Token rotation owns cancellation and immediately releases the old spinner.
const activeRefresh = refreshes.begin();
refreshes.invalidate();
assert.equal(refreshes.release(activeRefresh), false);

// A refresh keeps last-known data visible and marks it stale on failure.
const cached = {phase: 'fresh', data: {points: 3}, error: ''};
assert.equal(beginLoad(cached).phase, 'stale');
const stale = failLoad(beginLoad(cached), 'offline');
assert.deepEqual(stale, {phase: 'stale', data: {points: 3}, error: 'offline'});
assert.equal(failLoad({phase: 'loading', data: null, error: ''}, 'offline').phase, 'error');

console.log('mobile analytics request races: PASS (latest-wins, stale, unmount, A-B-A, overlap, token rotation)');
