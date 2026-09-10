import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {createRequire} from 'node:module';
import {loadTypeScript} from './load-typescript.mjs';

const tsModule = await loadTypeScript();
const ts = tsModule.default ?? tsModule;
const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'fitness-nutrition-'));
const source = fs.readFileSync('apps/mobile/src/domain/nutrition.ts', 'utf8');
const output = ts.transpileModule(source, {
  compilerOptions: {target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.CommonJS},
}).outputText;
fs.writeFileSync(path.join(tmp, 'nutrition.js'), output);
const requireFromTmp = createRequire(path.join(tmp, 'entry.js'));
const {catalogMacrosForQuantity, parseFoodQuantity, repeatedNutritionState} = requireFromTmp('./nutrition.js');

for (const invalid of ['', ' ', 'NaN', '0', '-1', '5000.1', 'Infinity']) {
  assert.equal(parseFoodQuantity(invalid).value, null, `${invalid || '<blank>'} must be rejected`);
}
assert.equal(parseFoodQuantity('125,5').value, 125.5);
assert.equal(parseFoodQuantity('5000').value, 5000);

const draftItem = {
  matched_food: {
    kcal_per_100g: 200,
    protein_per_100g: 20,
    fat_per_100g: 10,
    carbs_per_100g: 30,
  },
};
assert.deepEqual(catalogMacrosForQuantity(draftItem, 250), {
  calories: 500,
  proteinG: 50,
  fatG: 25,
  carbsG: 75,
});

const repeatedDay = {date: '2026-09-10', entries: []};
assert.deepEqual(repeatedNutritionState(repeatedDay), {selectedDate: '2026-09-10', day: repeatedDay});

const setup = fs.readFileSync('apps/mobile/src/screens/NutritionSetupScreen.tsx', 'utf8');
assert.match(setup, /api\.nutritionProfile\(accessToken\)/);
assert.match(setup, /if \(loading \|\| loadError\)/);
assert.match(setup, /Значения по умолчанию не показаны/);

const input = fs.readFileSync('apps/mobile/src/screens/AIFoodInputScreen.tsx', 'utf8');
assert.match(input, /parseVersion\.current\+=1;setBusy\(false\);setText\(value\);setDraft\(null\)/);
assert.match(input, /version===parseVersion\.current/);

const diary = fs.readFileSync('apps/mobile/src/screens/NutritionScreen.tsx', 'utf8');
assert.match(diary, /setSelectedDate\(next\.selectedDate\)/);
assert.match(diary, /setDay\(next\.day\)/);

console.log('mobile nutrition behavioral/static regressions: PASS');
