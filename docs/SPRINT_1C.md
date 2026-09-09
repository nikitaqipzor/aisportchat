# Sprint 1C — Persistence, Resume & Offline

## Goal
Make the MVP resilient enough for real daily workout logging: persistent backend state, protected session storage on device, offline set capture, automatic synchronization, workout resume, and exercise technique guidance.

## Delivered

### Backend persistence
- real PostgreSQL `Store` adapter using system `libpq` through cgo;
- runtime chooses PostgreSQL when `DATABASE_URL` is present or `STORE_BACKEND=postgres`;
- `MemoryStore` remains available for unit/integration tests;
- refresh sessions, profile, goals, preferences, workouts, exercises and sets persist through the same `Store` contract;
- migration `000006_runtime_persistence` adds workout metadata required by Sprint 1B runtime;
- migration `000007_exercise_guidance` prepares DB fields for technique content.

### Runtime recovery
- `GET /api/v1/workouts/active`;
- API returns the latest active workout or HTTP 204;
- active workout is covered by the HTTP lifecycle test;
- after finishing, the endpoint returns no active workout.

### Mobile session persistence
- access/refresh token pair stored through `react-native-keychain`;
- tokens are separated from normal app storage;
- bootstrap restores the session and rotates access credentials through refresh when needed;
- local logout clears protected token storage.

### Offline workout queue
- completed sets are optimistically written to the UI;
- failed network writes are persisted in an AsyncStorage queue;
- queue is ordered and idempotent by workout exercise + set number;
- app foreground triggers synchronization;
- queued finish operation is supported;
- local active workout snapshot survives app restart.

### Exercise technique cards
Every catalog exercise now carries:
- short description;
- ordered execution steps;
- common mistakes;
- technique tips.

Preview and active workout screens can open the guide without leaving the workout.

### Infrastructure
`docker compose -f infra/docker-compose.yml up --build` now defines:
1. PostgreSQL;
2. Redis;
3. migration job;
4. API container.

The API starts only after PostgreSQL is healthy and migrations complete successfully.

## Verification in current environment
- `go test ./...` ✅
- `go vet ./...` ✅
- active-workout HTTP lifecycle test ✅
- exercise-guidance payload test ✅
- all 15 TS/TSX source files pass TypeScript syntax transpilation ✅
- Docker/PostgreSQL end-to-end execution was not possible because Docker daemon/server binaries are not available in this execution environment.
- native Android/iOS compilation remains a device/developer-machine verification step because node_modules and native scaffold cannot be fetched here.

## Next vertical slice
Sprint 1D should finish the MVP product surface:
- body-map muscle picker;
- workout templates and favorites;
- personal-record detection;
- richer history filters and workout detail;
- local notification/rest timer integration;
- initial native Android/iOS scaffold verification;
- first deployable internal build.

## Migration safety

The Docker migration runner creates `schema_migrations` and applies each `*.up.sql` migration once inside a transaction. Repeated environment starts skip previously applied files.
