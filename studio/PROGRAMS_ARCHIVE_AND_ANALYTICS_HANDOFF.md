# Programs archive, analytics, and calendar handoff

## Goal
Prevent starting archived program workouts through history, explain replacement of active programs, keep program history accessible when analytics fails, and preserve calendar days in all time zones.

## Changes and findings
- `StartWorkout` rejects workouts linked to programs other than active in memory and PostgreSQL; PostgreSQL checks eligibility in the same UPDATE statement.
- Program setup checks for an active program and asks the user before replacing it; a failed lookup does not silently archive anything.
- Workout details hides the start action when a linked workout belongs to a previously archived program. The backend is authoritative when old history exceeds 50 items or the history request fails.
- Program list loads history independently of analytics; absent statistics display `—` instead of misleading zero values. Program dates render as calendar days independent of UTC offsets.

## Tests and evidence
- `node scripts/test-mobile-program-resume.mjs` PASS.
- `node scripts/test-mobile-nutrition-calendar.mjs` PASS, including New York date regression.
- Added `TestArchivedProgramWorkoutCannotBeStartedFromHistory` for memory backend, including check that unlinked manual workouts remain startable.
- Local Go runtime is unavailable; backend test requires CI.
- Typecheck executed during parallel edits and temporarily failed in unrelated `App.tsx`/`WorkoutSummaryScreen` integration; re-run after integration.

## Risks and unresolved items
- A concurrent program replacement between client confirmation and generation can still archive the newest program: the backend API does not carry an expected active-program ID. Product-safe compare-and-swap would require an API contract change.
- UI history query is limited to 50; PostgreSQL and memory backend still reject archived program starts even when that UI query cannot identify the relationship.
- Real Android dialogs and physical-device timezone rendering remain untested.

## Handoff
QA: run Go tests and typecheck after all parallel branches have been integrated, then test archived planned workout from History on a device. Lead engineer: review replacement race and release acceptance.
