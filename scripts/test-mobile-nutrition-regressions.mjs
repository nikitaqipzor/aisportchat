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
assert.match(input, /setDraft\(null\);onTextChange\?\.\(value\)/);
assert.match(input, /version===parseVersion\.current/);

const photo = fs.readFileSync('apps/mobile/src/screens/FoodPhotoScreen.tsx', 'utf8');
assert.match(photo, /isFoodPhotoPickCancelled/);
assert.match(photo, /Повторить анализ/);
assert.match(photo, /source=\{\{uri:selectedImage\}\}/);
assert.match(photo, /onManualEntry=\{onBack\}/);

const draftView = fs.readFileSync('apps/mobile/src/components/AIFoodDraftView.tsx', 'utf8');
assert.match(draftView, /accessibilityRole="checkbox"/);
assert.match(draftView, /included\[x\.i\]/);

const picker = fs.readFileSync('apps/mobile/android/app/src/main/java/com/aifitnessos/foodphoto/FoodPhotoPickerModule.kt', 'utf8');
assert.match(picker, /PHOTO_PICK_TOO_LARGE/);
assert.match(picker, /PHOTO_PICK_UNSUPPORTED/);

const diary = fs.readFileSync('apps/mobile/src/screens/NutritionScreen.tsx', 'utf8');
assert.match(diary, /withFoodIdempotency\(`repeat-\$\{entry\.id\}`/);
assert.match(diary, /withFoodIdempotency\(`undo-\$\{entry\.id\}`/);
assert.match(diary, /setUndoEntry\(null\);\s*await load\(\)/);
const foodSearch = fs.readFileSync('apps/mobile/src/screens/FoodSearchScreen.tsx', 'utf8');
assert.match(foodSearch, /withFoodIdempotency\('manual'/);
assert.match(foodSearch, /lastSearch\.current\.kind==='barcode'\?lookupBarcode\(\)/);
const recipes = fs.readFileSync('apps/mobile/src/screens/RecipesScreen.tsx', 'utf8');
assert.match(recipes, /version===searchVersion\.current/);
assert.match(recipes, /createdPendingRefresh\.current=true;await load\(\)/);
assert.match(diary, /const today = currentLocalDate\(\)/);
assert.match(diary, /if \(today !== requestedDate\) openDate\(today\)/);
assert.match(diary, /else \{ setDay\(null\); await load\(\); \}/);
assert.doesNotMatch(diary, /setDay\(repeatedDay\)/);

console.log('mobile nutrition behavioral/static regressions: PASS');
