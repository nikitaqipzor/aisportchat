# Nutrition P2 repair handoff

## Goal

Repair custom food validation and nutrition navigation state loss in the mobile app without changing the API.

## Changes

- `CustomFoodScreen.tsx`
  - Rejects blank calorie and macro fields instead of coercing an empty string to zero.
  - Matches API limits: 0–1000 kcal, 0–100 g for each macro/fiber field, and a positive serving up to 5000 g.
  - Matches the API name limit of 120 characters.
- `FoodSearchScreen.tsx` and `App.tsx`
  - Keep the text search query above the route boundary.
  - Returning from custom food recreates the search screen with the same query and reruns that search.
  - Clear the stored query after leaving the flow or logging food.
- `AIFoodInputScreen.tsx` and `App.tsx`
  - Keep the AI text draft above the text/photo route boundary.
  - Returning from photo restores the entered text.
  - Clear the draft after completion or an explicit return to nutrition.
- `App.tsx`
  - Android hardware Back from `customFood` now returns to `foodSearch`, matching the onscreen Back action.

`FoodPhotoScreen.tsx` did not require a direct change; its existing `onBack` route now restores the lifted text draft.

## Risks

- Parsed AI recognition results are not retained across a switch to photo; the requested entered text is retained and can be parsed again.
- Search results are fetched again when returning from custom food rather than cached, so an offline return may show the preserved query with a request error.

## Tests and evidence

- `cd apps/mobile && npm run typecheck` — passed (`tsc --noEmit`).
- `node scripts/test-mobile-nutrition-regressions.mjs` — passed, including explicit checks for local AI draft invalidation followed by lifted text preservation.
- `git diff --check` — passed.
- Compared client constraints with `services/api/internal/nutrition/service_2b.go` and `services/api/openapi/openapi.yaml`.
- No native code was changed, so `scripts/verify_android_native.py` was not applicable.

## Unresolved items

- No automated navigation harness exists in the mobile package; route behavior was statically verified.

## Handoff

QA should manually cover: blank B/J/U rejection; boundary values for custom food; search → custom food → onscreen and hardware Back; AI text → photo → Back; successful logging clearing retained drafts.
