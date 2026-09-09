# Stage 2 — athlete profile handoff

## Goal

Deliver an editable athlete profile covering goal, experience, training environment and equipment, schedule, body measurements, age, injuries and movement limitations.

## Changes

- Added an authenticated post-onboarding profile screen, reachable from the Home avatar, with load, retry, validation, save confirmation and logout states.
- Kept the existing profile, goal and training-preferences endpoints and composed their responses in the mobile flow.
- Added `age_years`, `injuries` and `limitations` to the profile contract, memory/PostgreSQL stores and migration `000019`.
- Added API validation for age (`14–100`), list size (`20`) and item length (`120`).
- Added API round-trip and invalid-input coverage.

## Risks

- The three existing write endpoints are composed sequentially, so an interrupted network can leave a partially updated profile. A future API version should expose one transactional aggregate update.
- Injury/limitation fields are user-declared context, not medical advice. Workout exclusion rules still need a dedicated deterministic mapping before these notes can automatically remove unsafe exercises.
- No physical-device visual, keyboard, large-font or TalkBack pass was available.

## Tests / evidence

- `cd apps/mobile && npm run typecheck` — passed with TypeScript 6.0.3.
- `git diff --check` — passed.
- `node scripts/verify-mobile-syntax.mjs` — blocked by its hard-coded missing global TypeScript runtime; repository-local TypeScript passed.
- Go tests/gofmt — not executed because Go is not installed in this environment. Backend files received a static review; CI must run Go formatting, tests and PostgreSQL migration coverage.

## Unresolved items

- Add deterministic injury-to-exercise contraindication codes rather than interpreting free text.
- Add screen interaction tests when the mobile test harness is introduced.
- Consider a transactional `PUT /athlete-profile` endpoint if partial save telemetry shows this is needed.

## Handoff

QA: run API tests including cross-user reads and PostgreSQL migration up/down; exercise offline failure after each of the three save calls. Designer: review 320 px width, keyboard overlap, 200% font size and TalkBack order. Lead: do not accept until Go CI and Android internal assembly are green.
