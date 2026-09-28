// Run with: node src/screens/ProgramsScreen.test.mjs
import assert from 'node:assert/strict';
import fs from 'node:fs';
import vm from 'node:vm';
import {loadTypeScript} from '../../../../scripts/load-typescript.mjs';

const mod = await loadTypeScript();
const ts = mod.default ?? mod;

const hooks = [];
let position = 0;
let effect;
let requests = [];
let failMore = false;
const make = (id, status = 'archived') => ({program: {id, title: id, status, environment: 'gym'}, sessions: []});
const page = (ids, hasMore, nextCursor) => ({items: ids.map(id => make(id)), has_more: hasMore, next_cursor: nextCursor});
const api = {
  activeProgram: async () => undefined,
  programHistory: async (_token, limit, cursor) => {
    requests.push([limit, cursor]);
    if (cursor === 'page-2' && failMore) throw Error('temporary failure');
    if (!cursor) return {items: [make('active', 'active'), ...Array.from({length: 49}, (_, i) => make(`p${i}`))], has_more: true, next_cursor: 'page-2'};
    if (cursor === 'page-2') return page(['p48', 'p49', 'p50'], false, '');
    throw Error(`unexpected cursor ${cursor}`);
  },
};
const react = {
  createElement: (type, props, ...children) => ({type, props: {...props, children}}),
  useState(initial) {
    const index = position++;
    if (!(index in hooks)) hooks[index] = initial;
    return [hooks[index], value => { hooks[index] = typeof value === 'function' ? value(hooks[index]) : value; }];
  },
  useRef(initial) { const index = position++; return hooks[index] ??= {current: initial}; },
  useCallback: fn => fn,
  useEffect: fn => { effect ??= fn; },
};
const native = {ActivityIndicator: 'ActivityIndicator', Pressable: 'Pressable', RefreshControl: 'RefreshControl', ScrollView: 'ScrollView', StyleSheet: {create: x => x}, Text: 'Text', View: 'View'};
const modules = {
  react: {...react, default: react}, 'react-native': native,
  '../api/client': {api}, '../components/AppButton': {AppButton: 'AppButton'},
  '../theme/tokens': {colors: new Proxy({}, {get: () => 'black'}), control: {minTouch: 44}, radius: {md: 8, lg: 12}, spacing: new Proxy({}, {get: () => 12})},
  '../domain/date': {formatCalendarDate: x => x},
};
const source = fs.readFileSync(new URL('./ProgramsScreen.tsx', import.meta.url), 'utf8');
const code = ts.transpileModule(source, {compilerOptions: {module: ts.ModuleKind.CommonJS, jsx: ts.JsxEmit.React, esModuleInterop: true}}).outputText;
const sandbox = {exports: {}, require: id => { assert.ok(id in modules, `unexpected import ${id}`); return modules[id]; }, console, Set};
vm.runInNewContext(code, sandbox, {filename: 'ProgramsScreen.js'});
const Screen = sandbox.exports.ProgramsScreen;
function render() { position = 0; return Screen({accessToken: 'token', onBack() {}, onCreate() {}, onOpen() {}}); }
function find(node, predicate) {
  if (Array.isArray(node)) return node.map(child => find(child, predicate)).find(Boolean);
  if (!node || typeof node !== 'object') return undefined;
  if (predicate(node)) return node;
  return node.props?.children?.map(child => find(child, predicate)).find(Boolean);
}
function countCards(tree) {
  let count = 0;
  function walk(node) {
    if (Array.isArray(node)) { node.forEach(walk); return; }
    if (!node || typeof node !== 'object') return;
    if (node.props?.accessibilityLabel?.startsWith('Открыть программу ')) count++;
    node.props?.children?.forEach(walk);
  }
  walk(tree);
  return count;
}
render();
await effect(); // useEffect calls the initial async load without returning its promise.
await new Promise(resolve => setImmediate(resolve));
let tree = render();
assert.equal(countCards(tree), 49);
assert.deepEqual(requests[0], [50, undefined]);
failMore = true;
await find(tree, node => node.props?.testID === 'programs-load-more').props.onPress();
await new Promise(resolve => setImmediate(resolve));
tree = render();
assert.equal(countCards(tree), 49, 'failed page preserves current programs');
assert.equal(find(tree, node => node.props?.testID === 'programs-load-more').props.label, 'Повторить загрузку');
failMore = false;
await find(tree, node => node.props?.testID === 'programs-load-more').props.onPress();
await new Promise(resolve => setImmediate(resolve));
tree = render();
assert.equal(countCards(tree), 51, 'repeated items and active item are not duplicated');
assert.equal(find(tree, node => node.props?.testID === 'programs-load-more'), undefined, 'last page removes the control');
assert.deepEqual(requests.map(([, cursor]) => cursor), [undefined, 'page-2', 'page-2']);
console.log('ProgramsScreen paginated history: PASS');
