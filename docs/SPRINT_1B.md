# Sprint 1B — Workout Core

## Goal
Turn the MVP from onboarding screens into a usable workout tracker with a persisted workout lifecycle.

## Delivered

### Backend
- authenticated workout generation;
- onboarding-aware defaults for level, session duration and equipment;
- deterministic exercise selection with equipment filtering;
- ~65% previous-session core retention plus exercise rotation;
- persisted workout status: `planned -> active -> completed`;
- set logging with weight, reps, RPE and RIR;
- rest time attached to each exercise;
- compatible exercise replacement before completed sets exist;
- workout history;
- total training volume calculation;
- next-session progressive overload recommendations;
- expanded in-code catalog to 60+ exercises;
- integration tests for workout lifecycle, history, rotation, replacement and progression.

### Mobile source
- real Home -> Generate flow;
- workout preview;
- exercise replacement from preview;
- active workout screen;
- weight / repetitions / RIR input;
- saved-set list;
- rest timer;
- exercise navigation;
- workout summary;
- next-progression recommendations;
- workout history;
- login resumes directly to Home when onboarding is already complete.

### Contracts / database
- OpenAPI `0.3.0`;
- migration `000004_workout_tracking`;
- migration `000005_expand_catalog`;
- indexes for latest completed workouts by muscle/environment.

## API lifecycle

```text
POST /workouts/generate
        ↓
planned workout
        ↓
POST /workouts/{id}/start
        ↓
active workout
        ↓
PUT /workouts/{id}/sets
        ↓
POST /workouts/{id}/finish
        ↓
completed workout + progression recommendations
```

## Progressive overload v1

The engine intentionally stays deterministic.

1. Find the most recent completed workout containing an exercise.
2. Select the strongest working set (highest weight, then highest reps).
3. If the user reached the top of the target rep range without a near-maximal effort signal, suggest a small load increase.
4. If reps fell below the minimum while RPE/RIR indicates maximal effort, suggest a small reduction.
5. Otherwise keep load stable and ask the user to progress repetitions first.

This is not an AI/LLM decision. Later AI Coach will explain these deterministic recommendations.

## Known technical debt

### PostgreSQL runtime adapter
The domain storage interface and SQL schema exist, but the executable still uses `MemoryStore` because this execution environment cannot fetch `pgx` and has no running Docker/PostgreSQL instance.

Next infrastructure task on a normal developer machine:
- add `github.com/jackc/pgx/v5`;
- implement `store/postgres`;
- run integration tests against PostgreSQL 18;
- select adapter from `DATABASE_URL`.

### Native mobile scaffold
`apps/mobile` contains the application source. The bare React Native iOS/Android scaffold still needs to be generated on a developer workstation and the source copied into it.

### Offline queue
Workout logging is API-backed but not yet offline-first. This remains the last major MVP infrastructure item.

## Definition of Done

Sprint 1B is considered complete when:
- Go tests pass;
- Go vet passes;
- TS/TSX source passes syntax transpilation;
- registration/onboarding tests remain green;
- workout lifecycle integration test passes;
- a second generated workout can contain an increased target weight based on prior performance.
