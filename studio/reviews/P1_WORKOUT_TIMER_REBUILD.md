# P1 workout and rest-timer rebuild handoff

## Goal

Remove the silent Program Detail start failure and make the active-workout rest timer reliable across backgrounding, Technique navigation, storage failures and overlapping user actions.

## Changes

- Program Detail now clears stale feedback, contains start failures, exposes a user-facing error and always releases the per-session busy state.
- Rest countdowns derive from an absolute wall-clock deadline instead of decrementing a JavaScript counter.
- A minimal snapshot contains only owner ID, workout ID, exercise label and deadline. Restore rejects expired, cross-owner and cross-workout state.
- Active Workout restores on mount and every return to the foreground through `AppState`.
- Ordinary unmounts, including Technique navigation, only suppress the visible timer; they do not clear persistence or cancel the Android alarm.
- Skip, finish, cancel and natural expiry invalidate synchronously, then clear persistence and cancel the native alarm.
- Timer side effects use one failure-contained FIFO. A rejected storage/native task cannot poison later work, and skip followed immediately by a new start is ordered correctly.
- Every awaited prepare/save/schedule stage checks the invalidation generation and compensates by clearing storage and cancelling the alarm if a destructive action interrupted it.
- AsyncStorage failures degrade to owner-scoped in-memory state plus the native alarm.

## Risks

- Native alarm delivery and Android notification permission behaviour still require a physical-device pass.
- If both AsyncStorage and the native notification bridge are unavailable, the timer remains correct only while the JavaScript process remains alive.
- iOS has no native rest notification bridge; the persisted wall-clock countdown still restores when the app resumes.

## Tests / evidence

- `node scripts/test-mobile-rest-timer.mjs` — PASS: deadline rounding, restore and expiry, destructive interruption during owner prepare/save/native schedule, failed first FIFO task followed by a valid task, unmount preservation, and skip→immediate start ordering.
- `npm --prefix apps/mobile run typecheck` — PASS.
- `node scripts/verify-mobile-syntax.mjs` — PASS, 61 files and zero errors.
- `python3 scripts/verify_android_native.py` — PASS, 125 checks.
- `git diff --check` — PASS.

## Unresolved items

- Run an Android device test that backgrounds the app beyond the deadline, opens/closes Technique during rest, denies notification permission, and rapidly taps skip then completes another set.
- Full Gradle APK assembly is an integration/CI gate and was not part of this isolated package.

## Handoff

Integration should merge this package without taking changes to `App.tsx`, `api/client.ts` or `storage/session.ts`, run the complete mobile regression suite, then verify the notification and foreground reconciliation on a physical Android phone.
