# Stage 6 — Training history and progress handoff

## Goal
Turn persisted completed workouts into a useful MVP progress surface without AI-generated or client-invented fitness facts.

## Changes
- Added authenticated `GET /api/v1/workouts/progress-summary?days=28`.
- The workout service deterministically aggregates completed workouts, training days, frequency, sets, total and rolling 7-day volume, consecutive active ISO weeks, PR events, and per-exercise best weight/reps.
- Extended the mobile History screen with a 28-day summary, week-over-week volume comparison, PR count and actual working bests.
- Preserved the existing environment/status/favorite filters and workout-detail navigation.
- Added loading, retry/error and first-workout empty states. The summary is loaded together with history, so partial API failure is not silently presented as zero progress.
- Updated the OpenAPI contract and mobile API types.

## Privacy and factual boundary
- All reads remain bearer-authenticated and owner-scoped through the existing Store API.
- Metrics use only persisted completed workouts, logged sets and saved PR events.
- No LLM output or synthetic placeholder is stored or displayed as progress.
- This slice adds no new retained personal data and no deletion behavior.

## Tests and evidence
- Added `TestCalculateProgressSummaryUsesCompletedFacts` for volume windows, set/workout totals, frequency inputs, weekly streak, PR count and exercise bests.
- `npm run typecheck` — passed.
- `node scripts/verify-mobile-syntax.mjs` — 58 files, 0 errors.
- `git diff --check` — passed.
- Go tests were not executable locally because the environment has no Go toolchain; CI must run `go test ./...` before acceptance.

## Risks / unresolved
- The summary currently considers the most recent 100 completed workouts, matching the existing Store boundary. Very high-frequency users may need a date-range SQL query later.
- Calendar grouping is UTC until the profile gains an explicit IANA timezone.
- Android visual/device smoke remains required; source checks do not prove small-screen layout.

## Handoff
QA: run full Go tests, API owner-isolation regression, and Android smoke with empty, one-workout, and multi-week histories. Designer: verify the summary card at large font scale and narrow widths. Lead may accept only after those checks are green.
