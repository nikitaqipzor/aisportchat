import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {createRequire} from 'node:module';
import {loadTypeScript} from './load-typescript.mjs';

const tsModule = await loadTypeScript();
const ts = tsModule.default ?? tsModule;
const target = fs.mkdtempSync(path.join(os.tmpdir(), 'nutrition-calendar-'));
try {
  const source = fs.readFileSync('apps/mobile/src/domain/date.ts', 'utf8');
  const output = ts.transpileModule(source, {compilerOptions: {target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS}}).outputText;
  fs.writeFileSync(path.join(target, 'date.js'), output);
  const load = createRequire(path.join(target, 'entry.js'));
  const {currentLocalDate, shiftLocalDate, localTimeZone, nextNutritionDateOnRollover, formatCalendarDate} = load('./date.js');
  const oldZone = process.env.TZ;
  try {
    process.env.TZ = 'America/New_York';
    assert.equal(localTimeZone(), 'America/New_York');
    assert.equal(formatCalendarDate('2026-09-02T00:00:00Z'), '02.09.2026', 'program calendar dates must not shift to previous day west of UTC');
    assert.equal(shiftLocalDate('2026-03-08', 1), '2026-03-09');
    assert.equal(shiftLocalDate('2026-11-01', -1), '2026-10-31');
    process.env.TZ = 'Europe/Moscow';
    assert.equal(localTimeZone(), 'Europe/Moscow');
    const RealDate = global.Date;
    try {
      let instant = '2026-09-01T20:59:00Z';
      global.Date = class extends RealDate {
        constructor(...args) { super(...(args.length ? args : [instant])); }
      };
      const before = currentLocalDate();
      instant = '2026-09-01T21:01:00Z';
      assert.equal(currentLocalDate(), '2026-09-02', 'a local September 2 remains September 2 while UTC is September 1');
      const after = currentLocalDate();
      assert.equal(nextNutritionDateOnRollover(before, after, before, true), '2026-09-02', 'following today advances at local midnight');
      assert.equal(nextNutritionDateOnRollover(before, after, before, false), null, 'manually selected archive remains on its date');
      assert.equal(nextNutritionDateOnRollover(before, after, '2026-08-30', false), null);
      assert.equal(nextNutritionDateOnRollover(after, after, after, true), null, 'foreground without a date change does not navigate');
    } finally {
      global.Date = RealDate;
    }
  } finally {
    if (oldZone === undefined) delete process.env.TZ;
    else process.env.TZ = oldZone;
  }
  const screen = fs.readFileSync('apps/mobile/src/screens/NutritionScreen.tsx', 'utf8');
  assert.match(screen, /AppState\.addEventListener\('change'/, 'foreground date reconciliation must be wired');
  assert.match(screen, /setTimeout\(\(\) => \{ reconcileToday\(\)/, 'open foreground must refresh at midnight');
  console.log('mobile nutrition calendar regression: PASS');
} finally {
  fs.rmSync(target, {recursive: true, force: true});
}
