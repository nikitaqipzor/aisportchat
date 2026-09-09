# Sprint 1D — Body Map, PRs, History & Native Rest Timer

## Goal
Finish the visible MVP surface around training selection and post-workout progress: interactive body navigation, muscle-level history, personal records, richer workout history, favorite/repeat actions, and Android rest-timer notifications.

## Delivered

### Interactive muscle navigation
- front/back body-map component without external graphics dependencies;
- selectable muscle hotspots;
- fallback list of all muscle groups;
- selected environment (home/gym/band) is preserved into the muscle screen;
- muscle details show training-history metrics before generation.

### Muscle statistics
New protected endpoint:

```text
GET /api/v1/muscles/{muscle_id}/stats
```

Returns:
- completed workouts;
- total sets;
- total volume;
- recent four-workout volume;
- last workout volume/date;
- days since last workout;
- personal-record event count.

The UI explicitly labels recency as a training-history signal, not a physiological or medical recovery measurement.

### Personal record engine
Workout completion now detects and persists record events per exercise:
- `max_weight`;
- `max_reps`;
- `estimated_1rm` using the Epley estimate for weighted sets.

Each event stores the previous record value when one exists. Records are persisted in migration `000008_mvp_insights` and are returned in workout detail and finish responses.

New endpoint:

```text
GET /api/v1/records
```

### Favorites and repeat
Completed workouts can be marked favorite:

```text
PATCH /api/v1/workouts/{workout_id}/favorite
```

A completed workout can be copied into a new planned workout:

```text
POST /api/v1/workouts/{workout_id}/repeat
```

The repeated workout keeps the source structure but recalculates current progression targets from the latest performance history.

### Richer history
`GET /api/v1/workouts/history` now supports:
- muscle;
- environment;
- status;
- favorite;
- limit.

Mobile history adds filter chips and opens a workout detail screen with:
- exercises and sets;
- volume;
- PR events;
- favorite toggle;
- repeat-with-progression action.

### Workout summary
The finish screen now surfaces new PR cards before the next-session progression recommendations.

### Android rest timer notification
- JS bridge wrapper with safe fallback;
- Android `AlarmManager` implementation;
- notification receiver/channel;
- Android 13+ notification permission request from JS;
- skipping the timer cancels the scheduled native alarm;
- if the native bridge is not registered, the on-screen timer still works.

Native integration seed lives under:

```text
apps/mobile/native/android/
```

### Native build bootstrap
Added:
- `app.json`;
- `index.js`;
- Babel configuration;
- Metro configuration;
- mobile npm scripts;
- `docs/ANDROID_INTERNAL_BUILD.md`.

The actual React Native-generated `android/` and `ios/` directories could not be downloaded in this environment because npm registry DNS is unavailable. The source and bridge integration instructions are prepared for the first developer-machine build.

## Backend verification
- `go test ./...` ✅
- `go vet ./...` ✅
- Sprint 1D HTTP lifecycle test covers PR -> favorite -> history filters -> muscle stats -> records -> repeat ✅
- OpenAPI 0.5.0 parses ✅
- migration `000008` present ✅

## Mobile verification
- 20 TypeScript/TSX files syntax-transpiled successfully before final documentation update ✅
- JS bootstrap/config files pass `node --check` ✅
- native Android compilation remains pending because npm/native SDK dependencies cannot be fetched here.

## Next phase
Sprint 2A starts Fitness 2.0 with the first nutrition vertical slice:

```text
nutrition profile
  -> calorie/macro targets
  -> daily dashboard
  -> meals
  -> food search/manual food
  -> daily totals
  -> history
```

After the deterministic nutrition core is stable, photo food recognition and AI Coach can be added on top without making the LLM the source of nutritional calculations.
