# Onboarding / account screens — production-readiness review

## Goal

Bring the onboarding, nutrition setup, and connected-device account surfaces to a consistent production-ready baseline without changing their public component props or API payloads.

Reviewed screens:

- `AuthScreen.tsx`
- `GoalScreen.tsx`
- `ProfileSetupScreen.tsx`
- `NutritionSetupScreen.tsx`
- `TrainingSetupScreen.tsx`
- `ConnectedDevicesScreen.tsx`

## Changes and findings

- Reworked visual hierarchy around one page title, supporting copy, clearly named sections, and a stable primary action.
- Reused `colors`, `spacing`, `radius`, and `control` tokens plus `AppButton`; removed screen-local hard-coded neutral/danger colours in the edited flows.
- Made long setup flows scrollable and kept their content usable on short/narrow devices. Profile measurements stack vertically below 360 px; nutrition macro inputs can wrap with a safe minimum width.
- Added explicit field labels, input hints, keyboard/return-key behaviour, length limits, locale-friendly decimal parsing, and client-side validation.
- Aligned manual nutrition calories with the backend contract (`800–8000`, integer) and reject non-finite/negative/implausibly large macro values before the request.
- Added consistent prevention of duplicate submissions and disabled interactive choices while an async action is in progress.
- Added semantic roles, labels, hints, selected/checked/disabled/busy states, live announcements, and at least 48 px interactive targets.
- Added actionable loading, partial-load error, retry, permission-denied, sync success/failure, and no-health-data states to Connected Devices.
- Replaced misleading pre-load “Mi Fitness not found” copy with an explicit unknown status.
- Improved user-facing copy: explains reversibility, distinguishes automatic from manual nutrition, avoids medical claims, and clarifies why each input is requested.

## Checklist

- [x] Public props and callback payload shapes preserved.
- [x] Existing `testID` values preserved where they existed.
- [x] Primary actions expose disabled/loading state.
- [x] Async handlers guard against duplicate invocation.
- [x] Selection controls expose radio/checkbox semantics and selected state.
- [x] Error messages use alert/live-region semantics.
- [x] Forms have visible labels and screen-reader labels/hints.
- [x] Touch targets meet the repository 48 px token baseline.
- [x] Small/short screens use scrolling; narrow measurement and macro layouts do not collapse.
- [x] Empty and partial failure states are present for health data.
- [x] Manual numeric values are validated before API submission.
- [x] The six scoped screens produce no diagnostics in the strict project typecheck.
- [ ] Physical Android screen-reader and large-font QA completed.
- [ ] iOS VoiceOver and keyboard-avoidance QA completed.
- [ ] Automated interaction tests added (the mobile package currently has no screen-test harness).

## Tests / evidence

- `cd apps/mobile && npm run typecheck` — the six scoped screens produce no diagnostics with TypeScript 6.0.3 after installing temporary, non-persisted `@types/react` and `@types/jest` required by the inherited React Native TypeScript configuration. The repository-wide command remains red on pre-existing/concurrent errors outside this task (unsupported font-weight strings, refs in `AICoachScreen`, nullability in `BodyScanScreen`, and `AsyncStorage.multiRemove`).
- `node scripts/verify-mobile-syntax.mjs` — not executed successfully in this workspace: the repository script resolves TypeScript from a global runtime path that is absent here. This is a verifier-environment issue; the stricter full mobile typecheck above passed.
- Static review performed for every interactive control, validation branch, async state, and narrow-layout branch in the six scoped files.

## Residual risks / unresolved items

- No emulator or physical-device visual pass was available. Dynamic Type (especially 150–200%), TalkBack focus order, IME overlap, and actual safe-area insets still require device QA.
- Connected Devices names `Xiaomi Watch S3` statically because the existing native status contract only reports Mi Fitness installation/permissions, not a discovered wearable model. A future native contract should return the actual device label or let this card use a generic “Mi Fitness” title.
- Error strings returned by the API/native bridge are still surfaced when they are `Error.message`; central error-code-to-localized-copy mapping is outside this screen-only task.
- Nutrition macro upper bounds are a defensive client limit (`1000 g`) while the backend currently validates only non-negative macros. The backend contract should eventually publish matching explicit maximums.

## Handoff

QA should run the six screens on the smallest supported Android viewport, at 200% font size, with TalkBack enabled, and exercise offline/timeout/permission-denied paths. Lead acceptance should keep public-prop compatibility and zero scoped TypeScript diagnostics as merge gates; the unrelated repository-wide diagnostics above must also be cleared before release. Physical accessibility verification remains the final screen-level release check.
