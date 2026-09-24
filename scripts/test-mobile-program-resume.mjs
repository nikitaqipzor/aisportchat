import assert from 'node:assert/strict';
import fs from 'node:fs';

const app = fs.readFileSync('apps/mobile/App.tsx', 'utf8');
const program = fs.readFileSync('apps/mobile/src/screens/ProgramDetailScreen.tsx', 'utf8');
const detail = fs.readFileSync('apps/mobile/src/screens/WorkoutDetailScreen.tsx', 'utf8');
const preview = fs.readFileSync('apps/mobile/src/screens/WorkoutPreviewScreen.tsx', 'utf8');

const between = (text, from, to) => {
  const start = text.indexOf(from);
  assert.ok(start >= 0, `${from} is missing`);
  const end = text.indexOf(to, start);
  assert.ok(end > start, `${to} is missing after ${from}`);
  return text.slice(start, end);
};

const start = between(program, 'async function start(s:ProgramSession)', 'async function autoMove');
assert.match(start, /s\.workout_id\s*\?\s*await api\.getWorkout\(accessToken,s\.workout_id\)/, 'linked workout must be reopened by ID');
assert.match(start, /onWorkout\(next,s\.muscle as MuscleId,program\?\.program\.status==='active'\)/, 'reopened linked workout must reach the navigator with program eligibility');
assert.match(start, /api\.createProgramWorkout\(accessToken,s\.id,currentLocalDate\(\)\)/, 'new session still creates a workout');
assert.match(program, /s\.workout_id\|\|\(program\.program\.status==='active'/, 'linked sessions must be openable, archived sessions must not create workouts');
assert.match(program, /testID=\{`\$\{s\.workout_id\?'program-open':'program-start'\}-\$\{s\.id\}`\}/, 'reopen action needs a stable UI identifier');

// Exercise the actual load callback with independent API outcomes: optional metrics cannot hide sessions.
const loadBody = between(program, 'const load=useCallback(async()=>{', '},[accessToken,programId]);').slice('const load=useCallback(async()=>{'.length);
async function loadWith(programResponse, analyticsResponse) {
  const state = {program: undefined, analytics: 'previous', loading: undefined, loadError: '', analyticsError: ''};
  const load = new Function('api', 'accessToken', 'programId', 'setLoading', 'setLoadError', 'setAnalyticsError', 'setAnalytics', 'setProgram', `return async function () {${loadBody}}`)(
    {getProgram: () => programResponse, programAnalytics: () => analyticsResponse},
    'token', 'program', value => {state.loading = value;}, value => {state.loadError = value;},
    value => {state.analyticsError = value;}, value => {state.analytics = value;}, value => {state.program = value;},
  );
  await load();
  return state;
}
const available = {program: {status: 'active'}, sessions: [{id: 'linked', workout_id: 'workout-1'}]};
const metricsFailure = await loadWith(Promise.resolve(available), Promise.reject(new Error('metrics down')));
assert.equal(metricsFailure.program, available, 'program and linked sessions survive analytics failure');
assert.equal(metricsFailure.analytics, undefined, 'stale metrics must be removed');
assert.match(metricsFailure.analyticsError, /Аналитика временно недоступна/);
assert.equal(metricsFailure.loadError, '', 'optional metrics must not become program failure');
assert.equal(metricsFailure.loading, false);
const programFailure = await loadWith(Promise.reject(new Error('program down')), Promise.resolve({adherence_percent: 85}));
assert.equal(programFailure.program, undefined, 'failed program must not be invented from metrics');
assert.equal(programFailure.loadError, 'program down');
const noFalseZero = between(program, '<View style={styles.analytics}>', '</View>');
assert.ok(noFalseZero.includes("'Нет данных'"), 'missing metrics must show no data');
assert.ok(!noFalseZero.includes('??0'), 'missing metrics must not be displayed as 0%');
assert.match(program, /testID="program-analytics-unavailable"/, 'optional failure must be explained on screen');

const hardwareBack = between(app, "if (step === 'preview')", "if (step === 'manualWorkout')");
assert.match(hardwareBack, /previewOrigin === 'manualWorkout' \? 'home' : previewOrigin/, 'system Back must use the real preview origin');
const previewRender = between(app, "{step === 'preview'", "{step === 'active'");
assert.match(previewRender, /canStart=\{previewCanStart\}/, 'read-only archived program must reach the preview');
assert.match(previewRender, /onBack=\{\(\) => setStep\(previewOrigin === 'manualWorkout' \? 'home' : previewOrigin\)\}/, 'onscreen Back must match system Back');
assert.match(app, /setPreviewOrigin\('programDetail'\)/, 'program preview must return to program details');
assert.match(app, /onStart=\{next => \{void updateWorkout\(next\); setStep\('active'\);\}\}/, 'planned linked workout must start from its preview');
assert.match(app, /setPreviewOrigin\('workoutDetail'\)/, 'history preview must return to workout details');
assert.match(app, /onRepeat=\{next => \{setWorkout\(next\);setSelectedWorkoutId\(next\.workout\.id\)/, 'repeated workout must become the detail target before preview Back');
assert.match(app, /next\.workout\.status==='active'&&canStart/, 'existing active session resumes in active workout');
assert.match(detail, /workout\.workout\.status === 'planned' \? <AppButton label="Открыть план и начать"/, 'history planned workout must be actionable');
assert.match(detail, /workout\.workout\.status === 'active' \? <AppButton label="Продолжить тренировку"/, 'history active workout must be resumable');
assert.match(preview, /\{canStart\?<AppButton label="Начать тренировку"/, 'archived preview must not start');
assert.match(preview, /onStart\(await api\.startWorkout\(accessToken,current\.workout\.id\)\)/, 'preview must start the existing linked workout ID');
assert.match(preview, /\{canStart\?<Pressable testID=\{`workout-preview-replace-/, 'archived preview must not mutate exercises');

console.log('mobile program resume navigation: PASS (linked, history, archived, Back, active)');
