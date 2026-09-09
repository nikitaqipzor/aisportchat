# Sprint 1A — Auth + Onboarding

## Implemented

- registration and login API;
- password hashing using PBKDF2-HMAC-SHA256 with per-password random salt;
- signed short-lived access tokens;
- opaque refresh tokens stored only as SHA-256 hashes;
- refresh-token rotation;
- logout revocation;
- protected profile routes;
- goal selection;
- height, weight and experience onboarding;
- home/gym/band training preferences;
- equipment selection;
- onboarding completion guard;
- equipment catalog endpoint;
- PostgreSQL migrations for goals, preferences, equipment and refresh sessions;
- catalog seed migration;
- Docker-based migration script;
- React Native onboarding screens wired to the API contract;
- integration test for the full registration -> onboarding -> refresh flow.

## Storage status

The domain and HTTP layers depend on the `store.Store` interface. The repository includes the complete PostgreSQL schema and migration runner, but this checkpoint's executable uses the in-memory adapter because the build environment could not fetch a PostgreSQL Go driver and does not provide Docker/PostgreSQL locally.

Next checkpoint should add the production PostgreSQL adapter (pgx preferred), then switch exercise generation to the DB-backed catalog.

## Next — Sprint 1B

1. Persist generated workouts.
2. Start/finish workout lifecycle.
3. Log sets, repetitions, weight, RPE/RIR.
4. Rest timer state.
5. Workout history.
6. Exercise replacement.
7. Progressive overload engine.
8. First offline queue contract for the mobile app.
