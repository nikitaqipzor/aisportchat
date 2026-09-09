# Training screens review

## Goal

Harden the training flow screens for production use without changing their public props or API contracts.

## Changes

- Completed `MuscleDetailScreen` states for loading, load failure with retry, empty history, and workout-generation failure.
- Completed `ProgramSetupScreen` profile loading/failure/retry flow and explicit empty state when no training environment is configured.
- Added stable test IDs, semantic headers, radio-group semantics, selected/disabled/busy accessibility state, descriptive labels, and 48 px minimum touch targets.
- Made option groups wrap on narrow screens and added bottom-safe scroll content spacing.
- Added valid 1–7 workouts-per-week choices so profile values at either boundary remain selectable.
- Reused shared theme tokens and `AppButton` states for consistent interaction feedback.
- Replaced unsupported React Native font weights (`650`, `850`, `950`) in the assigned training files with supported values.

## Risks

- Screen-reader wording and dynamic-type layout still need verification on physical iOS and Android devices.
- The profile flow has no direct navigation callback from program setup to training preferences; the empty state therefore explains the required action and keeps creation disabled.
- Recovery data on the muscle screen remains optional by design: a recovery request failure does not block workout statistics or generation.

## Tests / evidence

- `cd apps/mobile && npm run typecheck` — passed (`tsc --noEmit`, exit code 0).
- Confirmed no remaining `650`, `850`, or `950` font weights in `ActiveWorkoutScreen`, `HomeScreen`, or `RecoveryScreen`.

## Unresolved items

- Physical-device visual QA, TalkBack/VoiceOver QA, and automated interaction tests were not executed in this environment.

## Handoff

Lead Engineer / QA: run the complete mobile release gate and exercise the training flow on compact and large-font device profiles before release acceptance.
