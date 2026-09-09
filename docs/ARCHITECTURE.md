# Architecture

## Principle
LLM is not the source of truth for fitness logic.

- deterministic engines calculate and validate;
- PostgreSQL stores facts;
- ML/CV measures visual signals;
- AI explains, orchestrates and proposes bounded actions.

## Current architecture

```text
React Native mobile source
        |
        +--> Keychain / Keystore: auth tokens
        +--> AsyncStorage: workout snapshot + offline queue
        |
        v
Go HTTP API
        |
        +--> Auth Service
        +--> Profile / Onboarding
        +--> Workout Service
                |
                +--> deterministic Workout Engine
                +--> progression rules
                +--> Store interface
                         |
                         +--> PostgreSQL/libpq adapter (runtime)
                         +--> Memory adapter (tests/dev fallback)
        +--> Nutrition Service
                |
                +--> deterministic target calculator
                +--> daily/history aggregation
                +--> same Store interface
        |
        v
PostgreSQL

Redis is provisioned for the next caching/jobs slice.
```

## Persistence rule
Production/developer Docker runtime uses `STORE_BACKEND=postgres`. The executable automatically selects PostgreSQL whenever `DATABASE_URL` is defined. Memory storage is intentionally retained for isolated tests and quick local API experimentation only.

The first PostgreSQL implementation uses `libpq` through cgo because this isolated build environment cannot fetch external Go modules. All domain code depends only on the `Store` interface, so the transport can later be changed to a pooled `pgx` implementation without rewriting auth/profile/workout services.

## Mobile offline model
Sensitive credentials and normal application state are separated:

- access/refresh tokens -> Keychain/Keystore;
- active workout snapshot -> persistent app storage;
- pending mutations -> ordered offline queue.

The queue currently supports `log_set` and `finish_workout`. Set upserts are naturally idempotent because the backend has a unique `(workout_exercise_id, set_number)` constraint.

## Workout domain boundary
The HTTP layer does not calculate progression. It delegates to `internal/workouts.Service`.

Responsibilities:
- apply user profile defaults;
- enforce onboarding/environment constraints;
- select previous workout context;
- call deterministic generator;
- persist workout lifecycle;
- calculate progression suggestions;
- enrich stored exercise IDs with catalog metadata and technique guidance.

## Modular monolith domains
- auth
- users/profile
- catalog/exercises
- workouts
- programs
- nutrition
- progress
- recovery
- health
- ai
- billing
- notifications

Microservices are intentionally postponed until operational need appears.


## Nutrition domain boundary
Nutrition arithmetic lives in `internal/nutrition`, not in prompts. The service owns target validation, gram-to-macro conversion, daily totals, remaining targets and history aggregation. Food entries store calculated snapshots so future changes to a food item do not rewrite historical consumption.

The 7–90 day history path reads the range in bulk rather than issuing one database query per day. This keeps the path suitable for later charts and AI weekly reports.

## AI domain boundary — Sprint 2C
`internal/aifitness` sits above existing deterministic services. It owns provider orchestration, catalog matching of AI-extracted food names, the read-only Coach tool registry and narrative weekly reports.

Provider modes:
- OpenAI Responses API when `OPENAI_API_KEY` is configured;
- deterministic local provider for network-free tests/development.

Write boundary:
- Coach tools are read-only;
- food extraction returns a draft only;
- the user confirms `food_id + quantity_g`;
- `internal/nutrition` performs final macro arithmetic and persistence.

This keeps model swaps independent from domain calculations and makes AI failures non-destructive.

## Program domain boundary — Sprint 2D
`internal/programs` sits above the Workout Engine and owns multi-week planning facts: calendar dates, split, planned sets, deload multipliers, adherence, missed-state detection and rescheduling.

The boundary is intentional:
- Program Engine decides **when / what focus / how much planned load**;
- Workout Engine decides **which concrete exercises / reps / progression targets**;
- `GenerateAdapted` accepts bounded internal volume/intensity multipliers from Program Engine, but the public standalone workout endpoint cannot post arbitrary adaptation multipliers;
- AI can explain `adaptation_reason`, but cannot change the deterministic program state through the Coach.

Program sessions link to normal persisted workouts by `workout_id`. Completing a workout synchronizes the corresponding program session, so program analytics are derived from facts rather than client-side checkboxes.


## Technique CV boundary — Sprint 3B
Technique video follows a privacy-first split pipeline:

```text
Android camera -> temporary private MP4 -> MediaPipe Pose Landmarker (on device)
                                      -> timestamped 33-landmark frames
                                      -> Go Technique Engine
                                      -> rep state machine + ROM/tempo/symmetry/stability
                                      -> PostgreSQL analysis result
```

Raw exercise video is not an API payload and is deleted after extraction. The backend receives normalized landmarks only. This keeps CV inference close to the camera while retaining deterministic, versioned scoring and cross-session history on the server. `algorithm_version` is persisted with every result so future threshold/model changes do not silently rewrite historical scores.

The first five analyzers are heuristic rules, intentionally separate from MediaPipe inference. A future model can replace the rep/quality classifier without replacing capture or pose extraction.


## Live Technique boundary — Sprint 3B.1

```text
CameraX Preview + ImageAnalysis
  -> MediaPipe Pose Landmarker on Android
  -> native overlay / provisional rep counter
  -> bounded sampled landmarks only
  -> authenticated Technique API
  -> deterministic Go state machine
  -> optional active workout/exercise/set link
```

The native live counter is UX feedback only. Persistence and any value returned to the workout set use the server-side Technique Engine result. Raw live frames/video are not an API payload. The backend validates workout ownership, active status, set bounds, and technique-key compatibility before persisting a linked result.

## Recovery boundary — Sprint 3C
`internal/recovery` owns training-readiness heuristics. It reads facts from completed workouts and one subjective daily check-in, then produces a deterministic readiness score plus per-muscle status.

```text
Daily check-in (sleep / energy / stress / soreness)
                    +
Completed workout sets / exercise muscle mapping
                    |
                    v
          deterministic Recovery Engine
                    |
        +-----------+-----------+
        |                       |
 global readiness           muscle recovery
 0–100 + factors            0–100 + recent load
        |                       |
        +-----------+-----------+
                    |
       bounded volume/intensity multipliers
                    |
                    v
             Workout Engine
```

The public workout API accepts a user-local date, not arbitrary adaptation multipliers. The HTTP orchestration layer asks Recovery Engine for bounded multipliers and then calls `GenerateAdapted`. Program sessions compose `program multiplier × recovery multiplier`; `GenerateAdapted` retains absolute bounds.

AI receives recovery only through read-only context/tools. It may explain why a recommendation changed but cannot write a readiness score or directly modify the workload.

Wearable HRV/RHR/sleep data is intentionally not synthesized in Sprint 3C. Phase 4 will add normalized health-provider inputs behind a separate provider boundary.
