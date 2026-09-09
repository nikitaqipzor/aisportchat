import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {execFileSync} from 'node:child_process';
import {pathToFileURL} from 'node:url';
import {createRequire} from 'node:module';

async function loadTypeScript() {
  try { return await import('typescript'); }
  catch {
    const root = execFileSync('npm', ['root', '-g'], {encoding: 'utf8'}).trim();
    return import(pathToFileURL(path.join(root, 'typescript/lib/typescript.js')).href);
  }
}
const tsModule = await loadTypeScript();
const ts = tsModule.default ?? tsModule;
const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'fitness-technique-map-'));
const source = fs.readFileSync('apps/mobile/src/domain/technique.ts', 'utf8');
const output = ts.transpileModule(source, {compilerOptions: {target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS}}).outputText;
fs.writeFileSync(path.join(tmp, 'technique.js'), output);
const requireFromTmp = createRequire(path.join(tmp, 'entry.js'));
const {techniqueKeyForExercise} = requireFromTmp('./technique.js');
assert.equal(techniqueKeyForExercise('barbell_squat'), 'squat');
assert.equal(techniqueKeyForExercise('split_squat'), 'lunge');
assert.equal(techniqueKeyForExercise('dumbbell_curl'), 'biceps_curl');
assert.equal(techniqueKeyForExercise('pushup'), 'push_up');
assert.equal(techniqueKeyForExercise('overhead_press'), 'shoulder_press');
assert.equal(techniqueKeyForExercise('bench_press'), null);
console.log('mobile technique mapping: PASS');
