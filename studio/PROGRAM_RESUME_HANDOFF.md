# Program workout resume — engineering handoff

## Goal
Allow a workout created from a program to remain discoverable after leaving preview, and let planned and active workouts be reopened through program and history screens.

## Changes
- `ProgramDetailScreen` opens a linked `workout_id` with `GET /workouts/{id}`; only sessions without a linked workout use `POST /programs/sessions/{id}/workout`. Completed linked workouts open read-only details; active linked workouts resume. Archived/completed programs cannot create a new session workout from their detail view; linked planned workouts display a read-only preview.
- `App.tsx` records where preview and workout details came from. Onscreen and hardware Back share the route. Preview after a program workout returns to program details; after a history workout returns to workout details; manual creation returns home.
- Repeating a completed workout updates the selected detail ID to the new planned workout before entering preview, so Back opens that newly created workout rather than the source workout.
- `WorkoutDetailScreen` exposes actions to open a planned workout or resume an active workout; completed workouts retain Repeat. Its back label follows the source.
- `WorkoutPreviewScreen` prevents start and replacement in read-only archived/completed program previews.
- Program details and optional program analytics now load through independent `Promise.allSettled` results. An analytics failure leaves the program session list and its actions available and shows “Нет данных” for unavailable metrics plus a retry message. A program-fetch failure still shows the actual program error.

## Evidence
- `node scripts/test-mobile-program-resume.mjs`: PASS; checks linked workout reopening, start by the same ID, program and history return routes, repeat destination, active resume, status handling and archived preview. It executes the actual program load callback with successful program fetch plus failed analytics, and with failed program fetch plus successful analytics.
- `node scripts/test-mobile-rest-timer.mjs`: PASS.
- `npm run typecheck -- --pretty false`: PASS after all currently visible parallel edits.
- `git diff --check`: PASS.

## Risks and unresolved items
- Archive restriction here applies only to the program UI. `WorkoutView` has no program ID or start eligibility marker. A linked planned workout in an archived program is read-only when reached from the program, but a user can find the same workout in global History and tap Start. The backend `POST /workouts/{id}/start` currently enforces only planned status. A server-side archived-program guard or an explicit eligibility flag is a separate product decision.
- Verify real-device program → preview → Back → program → reopen → start, history planned → preview → Back → history detail, active resume after app restart, and archived program behavior against a deployed backend. Tests here are source-contract checks, not device E2E.

## Handoff
QA: run integrated regressions and Android smoke after parallel branches settle. Lead Engineer: decide whether the server-side archived-program start guard is required before beta.
