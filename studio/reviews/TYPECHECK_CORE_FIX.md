# Typecheck core fix

## Goal

Remove the React Native type errors in shared mobile files without changing screen implementations or package configuration.

## Changes

- Replaced unsupported `fontWeight` values with valid React Native weights:
  - `650` → `600` in `apps/mobile/App.tsx`.
  - both `750` values → `700` in `AIFoodDraftView.tsx`.
  - `850` → `800` in `ExerciseGuide.tsx`.
- Replaced the unavailable AsyncStorage v3 `multiRemove` call with concurrent, awaited `removeItem` calls over the same complete key set in `session.ts`.

## Risks

- The nearest supported weight can render slightly lighter than the original non-standard requested value.
- Individual storage removals are not transactional. This is equivalent for the current logout cleanup use case: every operation is awaited, and a failure is still surfaced to the caller.

## Tests / evidence

- Ran `npm run typecheck` from `apps/mobile`.
- The four edited shared files no longer report TypeScript errors.
- The command currently exits with code 2 because of remaining unsupported weights in screen files owned by parallel screen agents:
  - `ActiveWorkoutScreen.tsx`: one `650`.
  - `HomeScreen.tsx`: one `950`.
  - `RecoveryScreen.tsx`: four `950` values and one `850`.
- A root-level typecheck was not available because the repository root has no `package.json`.

## Unresolved items

- Re-run the full mobile typecheck after the screen agents resolve their remaining `fontWeight` errors.

## Handoff

Lead Engineer / QA should integrate these shared-file changes with the screen-agent work, rerun `npm run typecheck` from `apps/mobile`, and accept only after it exits successfully.
