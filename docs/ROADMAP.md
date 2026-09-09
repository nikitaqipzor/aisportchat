# Delivery roadmap

## Phase 0 — Foundation
- repository and modular architecture ✅
- API contracts ✅
- auth foundation ✅
- migrations / Docker foundation ✅

## Phase 1 — MVP
- auth/onboarding ✅ Sprint 1A
- exercise catalog ✅ Sprint 1B
- workout generator ✅ Sprint 1B
- active workout ✅ Sprint 1B
- set logging / RPE / RIR ✅ Sprint 1B
- exercise replacement ✅ Sprint 1B
- history ✅ Sprint 1B
- progression engine ✅ Sprint 1B v1
- PostgreSQL runtime adapter ✅ Sprint 1C
- active workout restore ✅ Sprint 1C
- secure mobile token storage ✅ Sprint 1C source
- offline workout queue ✅ Sprint 1C source
- exercise technique instructions ✅ Sprint 1C
- body-map muscle selector ✅ Sprint 1D
- muscle history metrics ✅ Sprint 1D
- personal-record engine ✅ Sprint 1D
- history filters/details ✅ Sprint 1D
- workout favorites/repeat ✅ Sprint 1D
- native Android rest-timer bridge ✅ Sprint 1D source
- Android/iOS generated native shell + device build ⏳ developer machine step

## Phase 2 — Fitness 2.0
### Sprint 2A — deterministic nutrition core ✅
- nutrition profile ✅
- automatic + manual calorie and macro targets ✅
- nutrition day ✅
- meal logging by grams ✅
- starter food catalog/search ✅
- daily totals and remaining balance ✅
- training-day linkage ✅
- 7–90 day nutrition history ✅

### Sprint 2B — nutrition expansion + progress ✅
- barcode lookup foundation ✅
- custom foods with per-user privacy ✅
- saved meals/recipes ✅
- quick repeat ✅
- body measurements + weight/waist trend ✅
- nutrition × training-day correlation ✅
- external food provider adapter / camera barcode scanner ⏳ later integration

### Sprint 2C — bounded AI layer ✅
- AI text food input + catalog matching + confirmation ✅
- food photo recognition foundation + Android gallery bridge ✅
- provider abstraction + local deterministic fallback ✅
- OpenAI Responses provider ✅
- AI Coach read-only tool layer ✅
- deterministic seven-day stats + AI weekly report ✅
- persistent AI conversation memory / tracing ⏳ later observability slice

### Sprint 2D — program engine / Fitness 2.0 closeout ✅
- 4/8/12-week deterministic program generation ✅
- calendar + goal-aware split ✅
- planned muscle volume ✅
- adherence analytics ✅
- missed-session detection ✅
- manual + automatic reschedule ✅
- deload weeks 4/8/12 with actual workout target reduction ✅
- program session ↔ workout lifecycle ✅
- AI explanation of deterministic adaptations ✅


## Quality Gate — before Phase 3
### Sprint 2Q1 — runtime blockers ✅
- central single-flight access-token refresh/retry ✅
- runtime/environment API URL ✅
- per-user offline/transient storage ✅
- cancel workout + program detach/re-create ✅
- finish confirmation + completion percent / ended-early metadata ✅
- client/backend set validation ✅
- bottom navigation + shared button foundation ✅
- nutrition delete Undo + destructive confirmations foundation ✅

### Sprint 2Q2 — Android native foundation ✅ source / ⏳ device build
- committed React Native 0.87 Android native scaffold ✅
- TurboModules: rest timer + food photo + native API config ✅
- debug/release API config at native build layer ✅
- debug signing + configurable release signing ✅
- native Android static/contract verifier ✅
- Android emulator build ⏳ (requires Android SDK + npm access)
- physical Android device smoke test ⏳
- iOS native scaffold ⏳

### Sprint 2Q3 / Release hardening — automated release gate 🟡
- mobile runtime regression suite ✅
- PostgreSQL 18 integration + migration CI harness ✅ configured / ⏳ green external run
- Android native build CI gate ✅ configured / ⏳ green external run
- auth rate limiting + production secret/store guards ✅
- accessibility/testID coverage for core flows 🟡 in progress
- mobile package-lock + clean `npm ci` ⏳ external registry required
- physical Android/Xiaomi smoke ⏳
- final design-token/router migration ⏳ post-RC quality work

## Phase 3 — AI Fitness
### Sprint 3A — Body Scan foundation ✅
- front/side/back capture lifecycle ✅
- Android full-resolution camera TurboModule ✅
- deterministic image quality gate ✅
- private authenticated media storage ✅
- delete metadata + original photos ✅
- scan history + resume draft ✅
- capture consistency + weight/waist deltas ✅
- body-shape / posture CV ⏳ Sprint 3B/3C

### Sprint 3B — Technique CV foundation ✅ source / ⏳ device calibration
- on-device MediaPipe pose landmarks ✅
- recorded exercise video TurboModule ✅
- privacy-first raw-video local processing + cleanup ✅
- movement state machine ✅
- rep counting for first 5 exercises ✅
- ROM / tempo / symmetry / stability scoring ✅
- pose skeleton preview + analysis history ✅
- real-device labeled-video calibration ⏳

### Sprint 3B.1 — Live Technique ✅ source / ⏳ device calibration
- CameraX live preview + on-device MediaPipe ✅
- native skeleton overlay + framing guidance ✅
- provisional live rep counter ✅
- authoritative backend recount ✅
- active workout / exercise / set linkage ✅
- camera-derived reps prefill current set ✅
- eccentric/concentric timing ✅
- top-start / bottom-start movement state machines ✅
- real-device calibration dataset ⏳

### Sprint 3C — Recovery/readiness ✅
- daily sleep / energy / stress check-in ✅
- per-muscle soreness input ✅
- deterministic readiness 0–100 ✅
- factor scores: sleep / energy / stress / soreness / 48h load ✅
- per-muscle recovery score + 48h/7d set load ✅
- interactive recovery body map ✅
- readiness-aware normal workout generation ✅
- program deload × readiness multiplier composition ✅
- AI Coach read-only readiness context ✅
- weekly AI report recovery adherence ✅
- wearable HRV/RHR/sleep ingestion ⏳ Phase 4

## Phase 4 — Fitness OS
### Sprint 4A — Xiaomi Watch S3 / Health Connect ✅ source / ⏳ device validation
- stable Android Health Connect provider ✅
- Mi Fitness preferred source + duplicate-safe cumulative aggregation ✅
- steps / distance / active calories ✅
- sleep total + stages ✅
- exercise sessions + exercise heart-rate summary ✅
- normalized daily health snapshots + PostgreSQL ✅
- Connected Devices permissions/sync UX ✅
- Recovery wearable sleep context ✅
- physical Xiaomi Watch S3 / phone permission + data calibration ⏳

### Sprint 4B — Personal baselines / provenance ✅ source / ⏳ device calibration
- multi-source conflict resolver ✅
- per-source health persistence + deterministic Xiaomi-first resolution ✅
- 7/28-day personal sleep/activity baselines ✅
- current-day exclusion / empty-day coverage protection ✅
- personal sleep trend influences deterministic readiness ✅
- stale wearable sleep → manual check-in fallback ✅
- health import freshness / confidence / provenance diagnostics ✅
- AI Coach read-only personal health insights ✅
- physical Xiaomi 28-day baseline comparison ⏳

### Sprint 4C — Readiness explanation timeline + passive sync ✅ source / 🚫 RC blocked
- factor delta timeline vs yesterday / 7d / 28d ✅
- exact explanation of workout volume/intensity changes ✅
- WorkManager background Health Connect refresh + encrypted pending queue ✅
- owner-scoped passive cache / logout / crash-account-switch isolation ✅
- release auth/security hardening ✅
- Nutrition + Workout domain coverage >80% ✅
- strict release preflight ✅
- package-lock / PostgreSQL CI / Android build / physical Xiaomi evidence ⏳
- robust health outlier calibration after real-device dataset ⏳

### Later
- iOS native scaffold + HealthKit
- direct vendor integrations only where aggregator data is insufficient
- specialized AI agents + orchestrator
