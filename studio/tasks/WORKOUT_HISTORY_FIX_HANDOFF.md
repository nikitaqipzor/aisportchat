# Handoff: workout history and offline summary

## Goal
Preserve workout history across rapid filter changes, paginate filtered history beyond 50 results, and avoid implying an offline workout was already saved remotely.

## Changes
- Workout history filters now run in PostgreSQL or memory before `LIMIT/OFFSET`; API responds with `items` and `has_more`.
- Mobile history clears cards when filters change, ignores late responses, and loads additional pages on demand.
- Offline workout summary explicitly says server confirmation is pending; manual workout asks before dropping selected exercises on muscle/place changes.

## Risks
- Offset pagination can shift when a workout is created or deleted while paging. Reload the screen to see a fresh ordering.
- Android device rendering, offline sync and PostgreSQL integration still need execution in CI or on a device.

## Tests and evidence
- `npx tsc --noEmit -p tsconfig.json` in `apps/mobile`: passed.
- Added Go test `TestWorkoutHistoryFiltersBeforePagination` covering 121 workouts, filters, user isolation and the second page.
- Go tests could not run in this workspace because `go` is not installed.

## Unresolved
- Run backend tests, SQL integration and manual tap checks before release.

## Handoff
QA: exercise filters rapidly with slow network responses, load beyond 50 entries and confirm offline summary after airplane mode. Lead engineer: accept only after backend and mobile CI pass.
