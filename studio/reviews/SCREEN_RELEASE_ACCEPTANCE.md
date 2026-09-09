# Screen and release-gate integration acceptance

## Goal

Integrate the three parallel screen-hardening batches with the backend, static-mobile and Android CI repairs, and establish the evidence required before merging the recovery branch into `main`.

## Changes and findings

- Reviewed all 28 mobile screens in three non-overlapping batches: onboarding/account, training/programs, and nutrition/progress/AI.
- Added or normalized loading, error, retry, empty, validation, disabled/busy, accessibility and narrow-screen behavior while preserving public screen props and API payloads.
- Added the missing React and Jest type packages and committed `package-lock.json`, allowing CI to use reproducible `npm ci` instead of an unconstrained install.
- Centralized TypeScript discovery for the five mobile verification scripts so they work with project-local or CI-global TypeScript.
- Replaced unsupported React Native font weights and the removed AsyncStorage `multiRemove` API.
- Replaced the Android verifier's external `jar` dependency with Python ZIP inspection.
- Aligned Android CI with JDK 21, which matches the bytecode level of the committed Gradle wrapper.

## Tests / evidence

- `npm run typecheck` — passed.
- `node scripts/verify-mobile-syntax.mjs` — 58 files, 0 errors.
- Auth refresh, session scope, technique mapping and passive-health owner-isolation regression scripts — passed.
- OpenAPI/router contract — 78 operations aligned; only `/healthz` is intentionally outside OpenAPI.
- Migration structure — 18 up/down pairs, sequence `000001`–`000018`.
- Android native static gate — 122 checks passed.
- `git diff --check` — passed.

## Risks and unresolved items

- Local Go/PostgreSQL execution and Android APK assembly are unavailable in this workspace; GitHub Actions is the authoritative automated evidence for those jobs.
- Physical Android/Xiaomi, TalkBack, large-font, keyboard and camera/permission flows remain required before an RC or production-release claim.
- Source-level completion does not replace device visual QA; the screen batches are accepted for merge once all GitHub Actions jobs pass.

## Handoff

Publish the integrated commit to `recovery/r2-clean-restart-2`, require the complete `release-gate` workflow to pass, then merge the reviewed pull request into `main` without bypassing branch checks.
