# Sprint 2D — Program Engine / Fitness 2.0 closeout

## Goal
Close Fitness 2.0 with a deterministic multi-week training-plan layer that sits above individual workout generation.

## Implemented

### Program generation
`POST /api/v1/programs/generate` creates a 4, 8 or 12 week plan from:
- persisted user goal;
- training preferences;
- selected environment;
- requested workouts/week (1–7).

Programs start on the nearest Monday on/after the requested date. This keeps every program week a complete Monday–Sunday unit and prevents shifted first-week sessions from overlapping week 2.

### Split + calendar
A deterministic goal-aware split assigns primary/secondary muscle focuses. Calendar spacing is stable for 1–7 sessions/week.

### Deload
Weeks 4, 8 and 12 are planned deload weeks.
- volume multiplier: `0.60`
- intensity multiplier: `0.90`

These values are not visual-only. Creating a workout from a program session uses `WorkoutService.GenerateAdapted`, which reduces actual target sets and target weights. Arbitrary multipliers are not exposed on the public standalone workout-generation endpoint.

### Session lifecycle
Statuses:
- `planned`
- `rescheduled`
- `missed`
- `completed`
- `skipped`

When a planned/rescheduled date passes without a completed workout, reading the program refreshes it to `missed`. A pre-generated but never-started planned workout does not suppress the missed state.

### Rescheduling
- manual reschedule to a future date;
- automatic reschedule of a missed session to the nearest unoccupied future date inside the current program horizon.

An already-created planned workout can remain linked while its program session is rescheduled. Active/completed workouts cannot be auto-rescheduled.

### Workout linkage
`POST /programs/sessions/{session_id}/workout`:
1. gets the deterministic program session;
2. applies deload/adaptation multipliers;
3. creates the normal persisted workout;
4. links `workout_id` back to the program session.

The endpoint is idempotent for an already linked workout.

Finishing a linked workout automatically marks the calendar session completed. If every session is completed/skipped, the program status becomes `completed`.

### Analytics
`GET /programs/{program_id}/analytics` returns:
- planned/completed/missed/rescheduled/upcoming sessions;
- adherence percent;
- current week;
- total weeks;
- deload session count;
- planned sets by muscle;
- completed sets by muscle;
- next session.

Adherence is computed from sessions that have reached an outcome (`completed / (completed + missed)`). Future scheduled sessions do not lower the score.

### AI explanation
`POST /programs/sessions/{session_id}/explain` sends a bounded description of the deterministic decision to the existing AI Coach provider. The model explains *why* the system used deload/reschedule values, but it does not choose or mutate those values.

Without `OPENAI_API_KEY`, the deterministic local provider still gives a basic explanation for deload/adaptation requests.

## Storage
Migration `000011_program_engine` adds:
- `programs`
- `program_sessions`
- unique active program per user;
- unique session slots;
- unique linked workout per session;
- date/user indexes.

Both MemoryStore and PostgreSQL implement the same Program Store contract.

## Mobile
Added:
- `ProgramsScreen`
- `ProgramSetupScreen`
- `ProgramDetailScreen`

User flow:
```text
Home
 -> Program
 -> create 4/8/12-week program
 -> calendar / adherence / current week
 -> session
 -> create adapted workout
 -> normal workout preview/active flow
 -> finish
 -> calendar session completed
```

Missed sessions expose `Автоперенос`; deload/adapted sessions expose `Почему?` for an AI explanation.

## API additions
- `POST /api/v1/programs/generate`
- `GET /api/v1/programs/active`
- `GET /api/v1/programs/history`
- `GET /api/v1/programs/{program_id}`
- `GET /api/v1/programs/{program_id}/analytics`
- `POST /api/v1/programs/{program_id}/archive`
- `POST /api/v1/programs/sessions/{session_id}/workout`
- `POST /api/v1/programs/sessions/{session_id}/reschedule`
- `POST /api/v1/programs/sessions/{session_id}/auto-reschedule`
- `POST /api/v1/programs/sessions/{session_id}/explain`

OpenAPI: **0.9.0**.

## Tests
Coverage includes:
- 4-week calendar generation;
- deload sessions/multipliers;
- workout linkage and completion;
- completed-set analytics;
- missed-session detection;
- automatic reschedule;
- Friday/midweek program creation starting on next Monday without duplicate dates;
- HTTP program lifecycle;
- actual deload set reduction;
- AI explanation endpoint.

## Fitness 2.0 status
With Sprint 2D, Phase 2 is functionally closed:
- deterministic nutrition;
- custom food/recipes/body progress;
- bounded AI food + Coach;
- weekly AI report;
- deterministic multi-week training programs;
- adherence/deload/rescheduling/program analytics.

Next phase: **Phase 3 / AI Fitness — Body Scan, pose estimation, technique analysis, rep counting and recovery/readiness.**
