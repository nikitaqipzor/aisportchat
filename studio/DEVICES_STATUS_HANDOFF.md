# Devices screen date and permission status handoff

## Goal
Ensure a previous day's Health Connect import is never presented as today's measurements and revoked permissions cannot look like a completed connection.

## Changes
- The device screen fetches current-day records plus recent history and picks a snapshot by its actual date, preferring today's record. Historical measurements are labeled with their date and invite an update; comparisons with today's baseline are hidden for historical measurements.
- Connection progress now depends on the latest native Health Connect status. It rechecks Android permissions when the application becomes active, before a sync begins, after a sync finishes, and after a sync error. Revoked access reverts the progress indicator to the permissions step and shows a warning over previously saved data.
- Overlapping loading requests cannot overwrite a newer native status; a foreground reload cannot race an ongoing sync into a misleading state.
- Added isolated status and date-selection regression tests runnable without a watch or Android device.

## Risks
- Real Mi Fitness → Health Connect timing and Android permission revocation must still be tested on a physical device. Repository tests simulate native statuses and check source bindings; they do not execute Android's permission UI.
- Historical records remain readable after access is revoked, since they are already saved under the authenticated account; the screen labels them as historical and disables further native syncing until permission returns.
- The 28-day import remains sequential and may take time on slow devices; this change does not address batching or progress performance.

## Tests/evidence
- `node scripts/test-mobile-devices-status.mjs` — pass (today, yesterday, future-date rejection, revoked/unknown permissions).
- `node scripts/test-mobile-health-passive-sync.mjs` — pass (authenticated owner isolation).
- `cd apps/mobile && npm run typecheck` — pass.
- `python3 scripts/verify_android_native.py` — pass, 125 checks.
- `node scripts/test-mobile-core-screens.mjs` — fails currently on another agent's concurrent nutrition-screen source assertion (`past nutrition day flow is incomplete`); devices assertions were not reached.

## Handoff
QA should install an internal Android build and verify date rollover, sync returning only yesterday's records, permission revocation through Android settings while the screen is open, permission grant recovery, airplane-mode reconnect, and device screen resume after a seven-day import.
