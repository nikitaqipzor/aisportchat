# Nutrition, devices and analytics screen handoff

## Goal

Finish the remaining core mobile surfaces without inserting demo values or moving deterministic calculations into the client.

## Changes

- Nutrition supports explicit day navigation, selection from the seven-day history and a guarded return-to-today flow for new entries.
- Connected Devices presents Mi Fitness → Health Connect → permissions → first sync as a four-step setup state.
- Progress is now Athletica Analytics with Overview, Body and Nutrition tabs.
- Analytics combines workout facts, body measurements, nutrition correlation and Health Connect insights through independent requests, so one unavailable domain does not hide the others.
- Analytics links directly to device setup when wearable data is absent.
- Stable test IDs and a source contract regression were added for the new navigation.

## Risks

- Past-day food entry creation remains intentionally disabled because the current add-food screens do not carry a selected diary date.
- Health Connect can only be proven on a physical Android device with Mi Fitness data.
- Analytics uses existing 28/30-day API windows; custom ranges are deferred.
- Nutrition correlation is observational and is labelled accordingly.

## Tests / evidence

- mobile TypeScript typecheck;
- mobile syntax verifier;
- Android native verifier;
- nutrition/devices/analytics screen contract;
- release-gate Android APK build.

## Unresolved items

- physical-device layout and Health Connect permission smoke;
- live API data validation after staging deployment;
- optional custom analytics periods and export.

## Handoff

QA should verify small-screen scrolling, date boundaries, denied permissions, partial API failures and TalkBack labels before release acceptance.
