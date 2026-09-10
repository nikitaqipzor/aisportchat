# P1 nutrition rebuild handoff

## Goal

Restore the approved nutrition hardening package without changing application navigation, session ownership, or the shared transport layer.

## Changes

- Nutrition settings load the current server profile before rendering editable values.
- A profile-load failure renders a retry state and never exposes saveable defaults.
- Repeating an entry aligns both the visible selected date and diary data to the date returned by the server.
- AI quantities reject blank, non-finite, non-positive, and over-5000-gram input.
- Calories and macros update deterministically from the matched catalog food's per-100-gram values.
- Editing AI text invalidates the previous draft; late responses from an obsolete request are ignored.
- Replacing a photo/text draft resets its local quantity fields.
- Added a domain helper and behavioral/static regression coverage.

## Risks

- Physical-device keyboard, large-font, and TalkBack behavior remains to be verified.
- The repeat endpoint defines the destination date; the client intentionally trusts `NutritionDay.date` from that response.
- Catalog macro display assumes the existing API contract's per-100-gram values are authoritative.

## Tests / evidence

- `node scripts/test-mobile-nutrition-regressions.mjs`
- `node scripts/verify-mobile-syntax.mjs`
- `npm run typecheck`
- `python3 scripts/verify_android_native.py`
- `git diff --check`

Exact results and environment limitations are recorded in the commit handoff message.

## Unresolved items

- Run nutrition setup, past-day repeat, AI text edit, and quantity validation on a physical Android device against staging.
- Confirm decimal keyboard behavior on the target Xiaomi device.

## Handoff

Lead Engineer / QA: integrate this branch without squashing away the regression script, rerun the complete release gate, and execute the four device scenarios above before personal dogfood acceptance.
