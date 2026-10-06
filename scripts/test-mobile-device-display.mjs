import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';

const screen = name => readFileSync(`apps/mobile/src/screens/${name}.tsx`, 'utf8');
const home = screen('HomeScreen');
const devices = screen('ConnectedDevicesScreen');
const report = screen('WeeklyAIReportScreen');

assert.match(home, /healthConnect\.status\(\)/, 'Home must check live Android permissions');
assert.match(home, /deviceStatus\.permissions_granted/, 'Home must gate synced status on permissions');
assert.match(home, /health_insights\?\.freshness\.status === 'fresh'/, 'Home must gate up-to-date status on backend freshness');
assert.match(home, /wearable\?\.date === currentLocalDate\(\)/, 'Home must use the current day');
assert.doesNotMatch(home, /Xiaomi Watch S3/, 'Home must not invent a model');
assert.match(devices, /function distanceKm\(/, 'Distance must have a decimal kilometres formatter');
assert.match(devices, /maximumFractionDigits: 2/, 'Subkilometre distances must remain visible');
assert.match(devices, /distanceKm\(snapshot\.distance_m\)/, 'Snapshot must use kilometre formatter');
assert.match(report, /formatHealthDate\(report\.stats\.from_date\)/, 'Date-only report bounds must not shift across time zones');
assert.doesNotMatch(report, /new Date\(report\.stats\.(?:from_date|to_date)\)/, 'Do not parse date-only bounds as UTC timestamps');
console.log('mobile health device display contracts: PASS');
