# Sprint 3C — Recovery & Readiness

## Goal
Turn recovery from a static label into a deterministic input to the workout decision engine.

## Daily check-in
The user logs a user-local date plus:
- sleep duration (0–16 h);
- sleep quality 1–5;
- energy 1–5;
- stress 1–5;
- optional soreness 1–5 for each supported muscle group;
- optional note.

The check-in is subjective fitness data. The UI explicitly avoids presenting it as medical diagnosis.

## Deterministic Readiness Engine
Five normalized factor scores are produced:
- sleep — 30%;
- energy — 20%;
- stress — 15%;
- soreness — 20%;
- completed-set load during the prior 48 h — 15%.

The weighted result is clamped to 0–100 and mapped to `ready`, `good`, `moderate`, or `low`.

If today's check-in is absent, workout generation remains unchanged. Training-history-only data can still be shown, but is not allowed to silently reduce the workout.

## Per-muscle recovery
For all 11 muscle groups the engine derives:
- subjective soreness;
- equivalent sets over 48 h and 7 d;
- hours since the last completed workout affecting the muscle;
- primary sets at weight 1.0 and secondary-muscle sets at weight 0.5;
- recovery score 0–100;
- `ready / moderate / fatigued` status.

This is a training-readiness heuristic, not a physiological measurement of muscle recovery.

## Adaptive workout generation
Normal workout generation now accepts a user-local `local_date`.

When a completed check-in exists:
- readiness >= 80: volume 1.00 / intensity 1.00;
- 65–79: 0.90 / 0.95;
- 50–64: 0.80 / 0.90;
- <50: 0.65 / 0.85.

A low score for the selected muscle can reduce the recommendation further. The public client cannot post arbitrary multipliers; only the backend Recovery Engine can select them.

Program sessions compose both deterministic layers:
`program_multiplier × readiness_multiplier`.

`GenerateAdapted` enforces absolute safety bounds already used by Program Engine.

## UI
Added:
- Home readiness card;
- Recovery screen;
- daily check-in;
- recovery body map front/back;
- sorted per-muscle recovery list;
- readiness factor breakdown;
- muscle detail readiness card;
- adaptation note in workout preview through existing progression-note UI.

## AI boundary
AI Coach receives readiness as read-only context and gets a `get_readiness` read-only tool. It can explain why volume changed but cannot modify check-in data, score, or multipliers.

Weekly AI stats now expose recovery check-in days and average readiness for the week.

## Persistence
Migration `000016_recovery_readiness` adds `recovery_checkins` with a unique `(user_id, local_date)` key and JSONB muscle-soreness data.

## API
- `GET /api/v1/recovery/today?date=YYYY-MM-DD`
- `PUT /api/v1/recovery/check-in`
- `GET /api/v1/recovery/history?limit=30`
- `POST /api/v1/workouts/generate` accepts `local_date`
- program-session workout endpoint accepts a user-local `date` query.

OpenAPI: 1.3.0.

## Tests
Coverage includes:
- high/low readiness scoring;
- invalid date/muscle validation;
- muscle recovery from completed workout sets;
- muscle-specific volume reduction;
- HTTP check-in/readiness/history flow;
- HTTP proof that low recovery changes actual workout target sets;
- existing full backend/mobile/native regression gate.

## Not included yet
Wearable-derived sleep, HRV and resting heart rate are intentionally deferred to Phase 4. Sprint 3C does not pretend subjective data is a wearable or medical measurement.
