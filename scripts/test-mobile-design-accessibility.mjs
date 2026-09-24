import assert from 'node:assert/strict';
import fs from 'node:fs';

const read = path => fs.readFileSync(path, 'utf8');
const home = read('apps/mobile/src/screens/HomeScreen.tsx');
const nav = read('apps/mobile/src/components/BottomNavigation.tsx');
const map = read('apps/mobile/src/components/BodyMap.tsx');
const guide = read('apps/mobile/src/components/ExerciseGuide.tsx');
const muscles = read('apps/mobile/src/domain/muscles.ts');

const positions = ['home-manual-workout', 'home-body-map-blocked', 'Другие возможности', 'home-recovery', 'home-devices', 'home-ai-coach', 'home-program'].map(marker => {
  const position = home.indexOf(marker);
  assert.ok(position >= 0, `Missing Home action: ${marker}`);
  return position;
});
assert.deepEqual(positions, [...positions].sort((a, b) => a - b), 'Workout actions must appear before secondary cards');
assert.match(home, /testID="home-manual-workout" disabled=\{!environmentReady\}/, 'Workout action remains blocked until environment is verified');

const muscleIds = [...muscles.matchAll(/^  (\w+): \{title:/gm)].map(match => match[1]);
const regionMuscles = [...map.matchAll(/muscles: \[([^\]]+)\]/g)].flatMap(match => [...match[1].matchAll(/'([^']+)'/g)].map(item => item[1]));
assert.deepEqual([...new Set(regionMuscles)].sort(), muscleIds.sort(), 'Map must include every catalog muscle group');
const buttonStyle = map.match(/muscleButton: \{[^\n]+/)?.[0] ?? '';
assert.doesNotMatch(buttonStyle, /position:|\btop:|\bleft:|\bright:/, 'Muscle buttons must remain in document flow');
assert.doesNotMatch(map, /styles\.hotspot/, 'Decorative silhouette must not become an overlapping touch hotspot');
assert.match(map, /testID="body-map-illustration"[\s\S]*?accessibilityElementsHidden[\s\S]*?importantForAccessibility="no-hide-descendants"[\s\S]*?pointerEvents="none"/, 'Silhouette must be decorative and skipped by TalkBack');
assert.ok(map.indexOf('testID="body-map-illustration"') < map.indexOf('regions[side].map'), 'Silhouette must appear separately before muscle buttons');
const silhouetteHeight = Number(map.match(/silhouettePanel: \{height: (\d+)/)?.[1]);
assert.ok(silhouetteHeight > 0 && silhouetteHeight <= 160, 'Silhouette must fit a compact viewport');
assert.match(map, /muscleButton: \{[^\n]*minHeight: control\.minTouch/, 'Every muscle choice requires a full-size touch target');
assert.match(map, /segmentItem: \{[^\n]*minHeight: control\.minTouch/, 'Front/back tabs require full-size touch targets');
assert.match(map, /accessibilityState=\{\{selected: side === value\}\}/, 'Screen reader must know selected body side');

assert.match(nav, /label: \{fontSize: 12/, 'Navigation labels must be legible');
assert.match(nav, /accessibilityState=\{\{selected\}\}/, 'Screen reader must know selected tab');
assert.match(guide, /accessibilityLabel="Закрыть инструкцию"/, 'Guide close action needs an accessible name');
assert.match(guide, /stepNumberText/, 'Technique steps need visible numbers');
assert.match(guide, /accessibilityLabel=\{`Шаг \$\{index \+ 1\}\./, 'Screen reader must announce ordered technique steps');

console.log('mobile design accessibility contracts: PASS');
