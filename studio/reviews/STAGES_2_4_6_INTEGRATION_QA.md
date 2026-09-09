# Stages 2 / 4 / 6 — mobile integration QA

## Goal

Independently review the integrated athlete-profile, manual-workout and training-progress flow before lead acceptance.

## Findings and changes

- Fixed the manual-workout origin state. Opening the manual builder now clears a stale muscle selection, so Preview → Back returns to Home instead of unexpectedly entering the generated muscle-workout flow.
- Decoupled history and progress-summary requests. A progress endpoint failure no longer hides otherwise valid workout history; each section now has its own retry/error state.
- Added touched-state validation for age during onboarding, matching height and weight. An empty or invalid age now exposes an accessible inline error after interaction.
- Reviewed the route `Home → Profile`, `Home → Manual workout → Preview → Active → Summary`, and `Home → History → Workout detail` statically. Active workout completion still clears the local active-workout cache and routes through Summary; completed data is then fetched from History.

## Risks

- No physical Android device or running API/database was available for end-to-end interaction testing.
- Profile saving uses three API calls (profile, goal, training preferences); a mid-sequence network failure can leave a partial server update. The screen reports the failure and supports retry, but the backend does not expose a transactional aggregate update.
- Offline-completed workouts appear in server history only after the queued operation synchronizes, as stated by the UI.

## Tests and evidence

- `cd apps/mobile && npm run typecheck` — passed.
- `node scripts/verify-mobile-syntax.mjs` — passed: 60 files, 0 errors.
- `python3 scripts/verify_android_native.py` — passed: 125 checks.
- `git diff --check` — passed.
- Navigation, loading, empty, retry and accessibility states for the requested screens were reviewed statically.

## Unresolved items

- Run the complete profile → manual workout → finish → history scenario against the deployed HTTPS API on a physical phone.
- Confirm large-font layout and TalkBack announcements on a device.

## Handoff

Lead engineer: accept after CI and device/API smoke testing; the confirmed integration regressions found in this review are fixed in this commit.
