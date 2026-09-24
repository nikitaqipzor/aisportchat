import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {createRequire} from 'node:module';
import {loadTypeScript} from './load-typescript.mjs';

const mod = await loadTypeScript();
const ts = mod.default ?? mod;
const temporary = fs.mkdtempSync(path.join(os.tmpdir(), 'fitness-devices-status-'));
try {
  const source = fs.readFileSync('apps/mobile/src/health/status.ts', 'utf8');
  const javascript = ts.transpileModule(source, {
    compilerOptions: {target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS},
  }).outputText;
  fs.writeFileSync(path.join(temporary, 'status.js'), javascript);
  const {formatHealthDate, selectHealthSnapshot, healthConnectionState} = createRequire(path.join(temporary, 'entry.js'))('./status.js');
  const today = '2026-09-24';
  const old = {date: '2026-09-23', imported_at: '2026-09-23T18:00:00Z'};
  const current = {date: today, imported_at: '2026-09-24T08:00:00Z'};
  const future = {date: '2026-09-25', imported_at: '2026-09-25T08:00:00Z'};
  const granted = {sdk_status: 'available', permissions_granted: true, mi_fitness_installed: true};

  assert.equal(formatHealthDate(old.date), '23.09.2026');
  assert.equal(selectHealthSnapshot(today, old, future, current), current, 'prefer the actual current day');
  assert.equal(selectHealthSnapshot(today, old, future), old, 'yesterday remains visibly dated; ignore future records');
  assert.equal(selectHealthSnapshot(today, future), null);
  assert.equal(selectHealthSnapshot(today, old, {...old, imported_at: '2026-09-23T20:00:00Z'}).imported_at, '2026-09-23T20:00:00Z');

  const historical = healthConnectionState(granted, old, today);
  assert.equal(historical.step, 5);
  assert.equal(historical.isToday, false);
  assert.equal(historical.badge, 'ОБНОВИТЬ');
  assert.match(historical.title, /23\.09\.2026/);
  const currentState = healthConnectionState(granted, current, today);
  assert.equal(currentState.isToday, true);
  assert.equal(currentState.badge, 'ГОТОВО');

  const revoked = healthConnectionState({...granted, permissions_granted: false}, current, today);
  assert.equal(revoked.step, 3, 'past imports cannot override revoked permissions');
  assert.equal(revoked.granted, false);
  assert.equal(revoked.badge, 'НАСТРОЙКА');
  assert.equal(healthConnectionState(null, current, today).step, 1, 'unverified native status cannot show completion');
  assert.equal(healthConnectionState(granted, null, today).step, 4);

  const screen = fs.readFileSync('apps/mobile/src/screens/ConnectedDevicesScreen.tsx', 'utf8');
  assert.match(screen, /AppState\.addEventListener\('change',[\s\S]*?next==='active'&&[^)]*\)void load\(\)/);
  assert.match(screen, /const before=await healthConnect\.status\(\)/, 'check actual permissions immediately before syncing');
  assert.match(screen, /setSnapshot\(selectHealthSnapshot\(day,today,result\.latest,drained\.latest,snapshot\)\)/);
  assert.match(screen, /snapshotIsToday\?`Сегодня ·/, 'only current-day data gets Today heading');
  console.log('mobile devices status date and permission scenarios: PASS');
} finally {
  fs.rmSync(temporary, {recursive: true, force: true});
}
