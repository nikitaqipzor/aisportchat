# Device status and report dates: engineering handoff

Goal: prevent false synchronized-watch claims on Home, calendar-date shifts in the weekly report, and loss of fractional kilometres.

Changes: Home checks native Health Connect status and current-day readiness freshness before calling watch data updated; removes invented Xiaomi Watch S3 model. Devices formats metres as fractional kilometres. Weekly report formats date-only values without UTC timestamp conversion. Added a display regression contract script.

Risks: Native Health Connect permissions and actual camera/watch behavior need Android device verification. Home checks status on screen mount; while Home remains mounted, an external permission revoke will only be visible after reopening the screen. No server endpoint or native action currently supports deletion of imported health snapshots; the existing system-settings action and passive-sync toggle do not erase server history, so no misleading "disconnect and delete" button was added.

Tests/evidence: `node scripts/test-mobile-device-display.mjs`, `node scripts/test-mobile-devices-status.mjs`, `node scripts/test-mobile-core-screens.mjs`, `node scripts/test-mobile-design-accessibility.mjs`, and `git diff --check` passed. Mobile `npm run typecheck` currently fails in concurrently edited `App.tsx` / `WorkoutSummaryScreen.tsx` prop `savedOnServer`; unrelated to these screens and awaiting integration.

Unresolved: privacy deletion needs an explicit authenticated server deletion contract plus native-side clearing policy; verify real device status and rendered report in negative UTC offset.

Handoff: QA validates permission revocation, aging/stale/fresh status, zero and 500 m distance, and negative UTC offset report dates; Lead Engineer owns release acceptance.
