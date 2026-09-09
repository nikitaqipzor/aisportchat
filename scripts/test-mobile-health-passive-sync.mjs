import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import {execFileSync} from 'node:child_process';
import {pathToFileURL} from 'node:url';
import {createRequire} from 'node:module';

async function loadTypeScript(){try{return await import('typescript')}catch{const root=execFileSync('npm',['root','-g'],{encoding:'utf8'}).trim();return import(pathToFileURL(path.join(root,'typescript/lib/typescript.js')).href)}}
const mod=await loadTypeScript();const ts=mod.default??mod;const tmp=fs.mkdtempSync(path.join(os.tmpdir(),'fitness-health-sync-'));
const source=fs.readFileSync('apps/mobile/src/health/sync.ts','utf8');
const output=ts.transpileModule(source,{compilerOptions:{target:ts.ScriptTarget.ES2022,module:ts.ModuleKind.CommonJS,esModuleInterop:true}}).outputText;
fs.mkdirSync(path.join(tmp,'health'),{recursive:true});fs.writeFileSync(path.join(tmp,'health/sync.js'),output);
fs.mkdirSync(path.join(tmp,'api'),{recursive:true});fs.mkdirSync(path.join(tmp,'domain'),{recursive:true});fs.mkdirSync(path.join(tmp,'native'),{recursive:true});fs.mkdirSync(path.join(tmp,'storage'),{recursive:true});
fs.writeFileSync(path.join(tmp,'domain/date.js'),`exports.currentLocalDate=()=> '2026-09-02'; exports.localDateOffset=o=>o===0?'2026-09-02':'2026-09-01';`);
fs.writeFileSync(path.join(tmp,'api/client.js'),`
let failDate=null; const imported=[];
exports.__state={imported,setFailDate:v=>failDate=v};
exports.api={
 async importHealthSnapshot(token,payload){if(payload.date===failDate) throw new Error('network'); imported.push({token,date:payload.date}); return {...payload,id:'id-'+payload.date,imported_at:new Date().toISOString()};},
 async healthToday(){return undefined;}
};
`);
fs.writeFileSync(path.join(tmp,'storage/session.js'),`
let owner='user-a';
exports.__state={setOwner:v=>owner=v};
exports.sessionStorage={async currentUserId(){return owner;}};
`);
fs.writeFileSync(path.join(tmp,'native/healthConnect.js'),`
const snapshots={
 'user-a':[
  {date:'2026-09-01',provider:'health_connect',source_package:'com.xiaomi.wearable',source_label:'Mi Fitness',steps:1,distance_m:0,active_calories_kcal:0,sleep_minutes:0,deep_sleep_minutes:0,light_sleep_minutes:0,rem_sleep_minutes:0,awake_minutes:0,exercise_minutes:0,exercise_sessions:0,data_types:['steps'],captured_at:new Date().toISOString()},
  {date:'2026-09-02',provider:'health_connect',source_package:'com.xiaomi.wearable',source_label:'Mi Fitness',steps:2,distance_m:0,active_calories_kcal:0,sleep_minutes:0,deep_sleep_minutes:0,light_sleep_minutes:0,rem_sleep_minutes:0,awake_minutes:0,exercise_minutes:0,exercise_sessions:0,data_types:['steps'],captured_at:new Date().toISOString()}
 ],
 'user-b':[
  {date:'2026-09-02',provider:'health_connect',source_package:'com.xiaomi.wearable',source_label:'Mi Fitness',steps:999,distance_m:0,active_calories_kcal:0,sleep_minutes:0,deep_sleep_minutes:0,light_sleep_minutes:0,rem_sleep_minutes:0,awake_minutes:0,exercise_minutes:0,exercise_sessions:0,data_types:['steps'],captured_at:new Date().toISOString()}
 ]
}; const acked=[];
exports.MI_FITNESS_PACKAGE='com.xiaomi.wearable';
exports.__state={snapshots,acked};
exports.healthConnect={
 async pendingSnapshots(owner){return (snapshots[owner]??[]).filter(x=>!acked.some(a=>a.owner===owner&&a.date===x.date));},
 async ackPendingSnapshot(owner,date){acked.push({owner,date});return true;},
 async status(){return {permissions_granted:true};},
 async readDay(date){return snapshots['user-a'].find(x=>x.date===date)??{...snapshots['user-a'][0],date,data_types:[]};}
};
`);
const requireFromTmp=createRequire(path.join(tmp,'entry.js'));const sync=requireFromTmp('./health/sync.js');const apiState=requireFromTmp('./api/client.js').__state;const nativeState=requireFromTmp('./native/healthConnect.js').__state;const sessionState=requireFromTmp('./storage/session.js').__state;

apiState.setFailDate('2026-09-02');
let failed=false;try{await sync.flushPassiveHealthSnapshots('token-a')}catch{failed=true}
assert.equal(failed,true);assert.deepEqual(nativeState.acked,[{owner:'user-a',date:'2026-09-01'}],'snapshot must be acknowledged only for the authenticated owner after server accepts it');
apiState.setFailDate(null);const result=await sync.flushPassiveHealthSnapshots('token-a');assert.equal(result.imported,1);assert.deepEqual(nativeState.acked,[{owner:'user-a',date:'2026-09-01'},{owner:'user-a',date:'2026-09-02'}]);

// Cross-account privacy: B can only drain B's envelope, never leftovers from A.
sessionState.setOwner('user-b');
const b=await sync.flushPassiveHealthSnapshots('token-b');
assert.equal(b.imported,1);
assert.deepEqual(apiState.imported.map(x=>x.token),['token-a','token-a','token-b']);
assert.deepEqual(nativeState.acked.at(-1),{owner:'user-b',date:'2026-09-02'});

sessionState.setOwner(undefined);
const noOwner=await sync.flushPassiveHealthSnapshots('token-x');
assert.deepEqual(noOwner,{processed:0,imported:0,empty:0});
console.log('mobile passive health sync owner isolation: PASS');
