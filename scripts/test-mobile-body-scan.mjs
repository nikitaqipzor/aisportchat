import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';

const screen = readFileSync(new URL('../apps/mobile/src/screens/BodyScanScreen.tsx', import.meta.url), 'utf8');
const client = readFileSync(new URL('../apps/mobile/src/api/client.ts', import.meta.url), 'utf8');

assert.match(screen, /!loading && current/,
  'capture controls must not render before the initial request resolves');
assert.match(screen, /capture_only_no_body_inference|не анализирует фигуру/,
  'the UI must explicitly avoid body or medical inference');
assert.match(screen, /\(comparison\.view_metrics \?\? \[\]\)\.map/,
  'comparison must explain consistency for every view');
assert.match(screen, /api\.deleteBodyScan\(accessToken, item\.scan\.id\)/,
  'completed scan history must expose private-data deletion');
assert.match(client, /capture_grade\?: 'excellent' \| 'good' \| 'retake_recommended'/,
  'the mobile API contract must include the bounded capture grade');
assert.match(client, /resolution_delta_percent\?: number/,
  'the mobile API contract must expose resolution consistency');
assert.match(screen, /contrast_delta !== undefined/,
  'new comparison fields must remain compatible with an older backend');

console.log('mobile body scan contract: PASS');
