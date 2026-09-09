import fs from 'node:fs';
import path from 'node:path';
import {loadTypeScript} from './load-typescript.mjs';
const tsModule = await loadTypeScript();
const ts = tsModule.default ?? tsModule;

const root = path.resolve('apps/mobile');
const files = [];
function walk(dir) {
  for (const entry of fs.readdirSync(dir, {withFileTypes: true})) {
    if (entry.name === 'node_modules') continue;
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) walk(full);
    else if (/\.(ts|tsx)$/.test(entry.name)) files.push(full);
  }
}
walk(root);
let failures = 0;
for (const file of files) {
  const source = fs.readFileSync(file, 'utf8');
  const result = ts.transpileModule(source, {
    fileName: file,
    reportDiagnostics: true,
    compilerOptions: {
      jsx: ts.JsxEmit.ReactJSX,
      target: ts.ScriptTarget.ES2022,
      module: ts.ModuleKind.ESNext,
    },
  });
  for (const diagnostic of result.diagnostics ?? []) {
    if (diagnostic.category !== ts.DiagnosticCategory.Error) continue;
    failures += 1;
    const message = ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n');
    const pos = diagnostic.file && diagnostic.start != null ? diagnostic.file.getLineAndCharacterOfPosition(diagnostic.start) : null;
    console.error(`${file}${pos ? `:${pos.line + 1}:${pos.character + 1}` : ''}: ${message}`);
  }
}
console.log(`mobile syntax: ${files.length} files, ${failures} error(s)`);
process.exitCode = failures ? 1 : 0;
