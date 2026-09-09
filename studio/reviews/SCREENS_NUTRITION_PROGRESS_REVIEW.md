# Nutrition / Progress / AI screens review

## Goal

Bring the assigned nutrition, progress, body-scan, history, and AI screens from prototype UX to a production-ready interaction baseline without changing navigation props or API contracts.

Reviewed screens:

- `NutritionScreen.tsx`
- `FoodSearchScreen.tsx`
- `CustomFoodScreen.tsx`
- `FoodPhotoScreen.tsx`
- `AIFoodInputScreen.tsx`
- `RecipesScreen.tsx`
- `ProgressScreen.tsx`
- `BodyScanScreen.tsx`
- `HistoryScreen.tsx`
- `AICoachScreen.tsx`
- `WeeklyAIReportScreen.tsx`

## Changes

### Shared UX baseline

- Reused `theme/tokens` and `AppButton` for color, spacing, radius, touch size, disabled state, and loading behavior.
- Added 48 px minimum touch targets to navigation, filters, list actions, meal selectors, and destructive controls.
- Added screen headers and accessible labels/roles/states for buttons, radio groups, checkboxes, progress bars, fields, loading messages, and errors.
- Kept all screens scrollable and wrapping where needed for narrow devices and larger text.
- Added explicit loading, error, retry, empty, and partial-data states instead of leaving blank content.
- Blocked duplicate submissions while network actions are active.

### Nutrition

- Added reliable initial loading and retry states to the nutrition diary.
- Made macro tracks accessible progress bars and clarified calorie over-target copy.
- Added an empty seven-day history state.
- Added validated food search, barcode lookup, portion preview, no-results recovery, and per-action busy state.
- Added strict custom-food validation for name, barcode, calories, macros, fiber, and serving size before API submission.
- Added recipe loading/search/save/log states, ingredient quantity validation, duplicate prevention, removal, and a useful first-recipe empty state.
- Added explicit AI text/photo disclaimers: AI proposes a draft; the user confirms; catalog data remains the calorie/macro source of truth.
- Disabled unsupported photo capture and supplied a recoverable fallback message.

### Progress and history

- Added initial progress loading/retry and a useful two-measurement chart empty state.
- Added range validation for weight, waist, and arm measurements, requiring at least one valid value.
- Clarified that nutrition/training correlation is observational and that pose/body comparisons are estimates rather than diagnosis.
- Added accessible, localized history filters and workout status/environment labels.
- Added history retry and filter-reset empty recovery.

### Body Scan and AI

- Fixed the Body Scan start/capture busy-state race: capture can create a draft without clearing the active capture state early.
- Added start failure handling, unsupported-camera state, first-load retry, privacy guidance, and non-medical comparison copy.
- Added AI Coach starter prompts, 1000-character input limit, minimum touch targets, error/retry/stop flow, and a health-safety disclaimer.
- Added weekly report loading/retry/pull-to-refresh, period context, partial-data copy, and explicit AI limitations.

## Risks

- `AIFoodDraftView` is a shared component outside this screen-only assignment. Its existing quantity and meal controls remain unchanged and should receive a separate accessibility pass.
- Visual QA on a physical Android device was not available in this environment. Dynamic type, TalkBack order, keyboard resize, camera picker, and very narrow-device layout still need device verification.
- API error strings are rendered to the user as supplied by the client. The client/backend should continue converting transport errors into safe user-facing messages.

## Tests / evidence

- `git diff --check -- apps/mobile/src/screens` — passed.
- TypeScript `transpileModule` syntax check for all 11 assigned TSX files — 11/11 passed.
- Targeted strict TypeScript check: `tsc --noEmit --types react`; no errors in the 11 assigned screens.
- The repository-wide typecheck remains red only on pre-existing/out-of-scope files (`App.tsx`, shared components, `ActiveWorkoutScreen`, `HomeScreen`, `RecoveryScreen`, and `storage/session.ts`).
- The default `npm run typecheck` also depends on Jest type definitions not declared by the mobile package; integration will be run by the lead after dependency/CI normalization.

## Unresolved items

- Run Android device QA at 320 dp width and with font scaling at 1.3–1.5.
- Verify TalkBack focus order, live-region announcements, and keyboard navigation on every form.
- Exercise real API failures, offline recovery, photo cancellation, permission denial, and large history datasets.
- Give `AIFoodDraftView` its own component-level production/accessibility pass.

## Handoff

Lead Engineer / QA: integrate these screen changes with the parallel screen batches, run the centralized mobile typecheck and Android smoke suite, then complete device visual/accessibility QA before release acceptance.
