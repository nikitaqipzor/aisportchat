# Mobile P2 repair handoff

## Goal
Repair profile defaults, AI Coach submission reliability, Technique zero-rep handling, and program history/detail freshness.

## Changes
- Athlete profile preserves an explicit empty onboarding equipment list and permits saving no equipment. It only falls back to `bodyweight` when preferences or the field are absent.
- AI Coach uses an immediate ref lock so rapid taps cannot issue two requests. Local chat writes run through a serial queue, so a slow user-message write cannot overwrite the later complete reply. Storage errors appear separately and never turn a successful provider response into a failed/retryable AI request.
- Technique only offers to copy a positive rep count into a workout and explains how to retry when zero reps are detected.
- Programs requests the API maximum of 50 history entries instead of silently showing only the first 10.
- Program detail ignores responses from an obsolete `programId` load and clears stale program data after a failed current load.

## Risks
- Program history has no cursor or offset contract. The 50-entry request removes the current first-10 truncation, but records after 50 remain inaccessible. Proper pagination requires an API response such as `{items, has_more}` plus an `offset` or stable cursor parameter, followed by a mobile “load more” flow.
- Chat persistence failures remain non-blocking and visible as a local-history warning; the server response remains visible for the current session.

## Tests and evidence
- `npm run typecheck` in `apps/mobile`.
- Android native verification is unnecessary because no native files changed.

## Unresolved
- Repeat-workout Back navigation lives in `App.tsx`, outside this ownership. Suggested narrow change: when opening the repeated workout, retain the source workout detail route/id and make the preview Back callback restore that detail instead of the generic history screen.

## Handoff
Lead engineer should validate the `App.tsx` navigation change with its owner and QA the rapid double-tap and rejected AsyncStorage cases.
