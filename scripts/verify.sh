#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "== Go format/test/vet =="
cd "$ROOT/services/api"
gofmt -w .
go test ./...
go vet ./...

echo "== Go non-cgo compile/test =="
CGO_ENABLED=0 go test ./...

echo "== YAML parse =="
cd "$ROOT"
python3 - <<'PY'
from pathlib import Path
import yaml
for path in [Path('infra/docker-compose.yml'), Path('services/api/openapi/openapi.yaml')]:
    yaml.safe_load(path.read_text())
    print(f'OK {path}')
PY

echo "== API contract + migration structure =="
python3 "$ROOT/scripts/verify-api-contract.py"
python3 "$ROOT/scripts/verify-migrations.py"

echo "== Mobile TypeScript syntax =="
GLOBAL_NODE_MODULES="$(npm root -g 2>/dev/null || true)"
if [ -n "$GLOBAL_NODE_MODULES" ] && [ -d "$GLOBAL_NODE_MODULES/typescript" ]; then
  NODE_PATH="$GLOBAL_NODE_MODULES" node <<'JS'
const fs = require('fs');
const path = require('path');
const ts = require('typescript');
const root = path.resolve('apps/mobile');
const files = [];
function walk(dir) {
  for (const entry of fs.readdirSync(dir, {withFileTypes:true})) {
    if (entry.name === 'node_modules') continue;
    const p = path.join(dir, entry.name);
    if (entry.isDirectory()) walk(p);
    else if (/\.(ts|tsx)$/.test(entry.name)) files.push(p);
  }
}
walk(root);
let failures = 0;
for (const file of files) {
  const source = fs.readFileSync(file, 'utf8');
  const result = ts.transpileModule(source, {
    compilerOptions: {
      target: ts.ScriptTarget.ES2022,
      module: ts.ModuleKind.ESNext,
      jsx: ts.JsxEmit.ReactJSX,
      strict: true,
    },
    fileName: file,
    reportDiagnostics: true,
  });
  const diagnostics = (result.diagnostics || []).filter(d => d.category === ts.DiagnosticCategory.Error);
  if (diagnostics.length) {
    failures += diagnostics.length;
    console.error(`FAIL ${file}`);
    for (const d of diagnostics) console.error(ts.flattenDiagnosticMessageText(d.messageText, '\n'));
  }
}
console.log(`Checked ${files.length} TS/TSX files; syntax failures=${failures}`);
if (failures) process.exit(1);
JS
else
  echo "SKIP mobile syntax: global TypeScript module not installed"
fi


echo "== Android native static verification =="
python3 "$ROOT/scripts/verify_android_native.py"

echo "== Mobile runtime regression tests =="
node "$ROOT/scripts/test-mobile-auth-refresh.mjs"
node "$ROOT/scripts/test-mobile-session-scope.mjs"
node "$ROOT/scripts/test-mobile-technique-mapping.mjs"
node "$ROOT/scripts/test-mobile-health-passive-sync.mjs"
node "$ROOT/scripts/test-mobile-rest-timer.mjs"
node "$ROOT/scripts/test-mobile-analytics-races.mjs"
node "$ROOT/scripts/test-mobile-home-reliability.mjs"
node "$ROOT/scripts/test-mobile-nutrition-regressions.mjs"
node "$ROOT/scripts/test-mobile-core-screens.mjs"
