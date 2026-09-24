# Nutrition date and request race handoff

## Goal
Keep the selected nutrition day and food search results in sync with the user's latest request and the phone's calendar, including UTC midnight and daylight-saving transitions.

## Changes
- `NutritionScreen` invalidates earlier profile/day/history requests synchronously when a different day opens or the screen unmounts; older responses cannot change selected day, history, errors or trigger setup.
- A foreground return and a midnight timer advance the automatically followed "today" to the new local day. Manually selected archive dates stay selected. Pull-to-refresh reconciles the date too.
- Repeating a food entry after local midnight ignores the write endpoint's potentially UTC day summary and reloads the local day through the time-zone-aware read contract.
- `FoodSearchScreen` clears a previous selection when the user edits a search/barcode or submits a new request; late responses from older queries/barcodes cannot replace the new results.
- Direct food logging and undo submit the client's IANA zone to the additive write endpoint.
- `domain/date.ts` exposes the runtime's IANA time zone, defaulting to UTC if unavailable.
- `nutrition.Service` has additive `DayInLocation`, `HistoryInLocation`, `LogFoodInLocation` methods using calendar boundaries from `time.Date(..., loc)` and `AddDate`, preserving existing UTC defaults. UTC storage remains unchanged.

## Integration needed in shared API files
- `GET /nutrition/today` and `GET /nutrition/history` accept optional `time_zone=Europe/Moscow`. `POST /nutrition/entries` likewise accepts this optional query. Default `UTC` when missing; unknown/invalid IANA zone returns HTTP 400. `time.LoadLocation` parses, with bundled `time/tzdata` for minimal images.
- The optional `date` query on `/nutrition/today` must be parsed with `time.ParseInLocation("2006-01-02", rawDate, loc)` so a calendar date is interpreted in the caller's time zone. With no date use `time.Now()`.
- Route to `DayInLocation(ctx, userID, at, loc)`, `HistoryInLocation(ctx, userID, days, time.Now(), loc)` and `LogFoodInLocation(ctx, userID, foodID, mealType, grams, loggedAt, loc)`.
- Update `apps/mobile/src/api/client.ts` to accept optional third `timeZone?: string` on `nutritionToday` and `nutritionHistory` and append `time_zone` query encoded with `encodeURIComponent`; the screen already sends this argument.
- Update OpenAPI query parameters for all three endpoints, including UTC default and IANA/DST behavior.

## Risks and unresolved items
- The runtime falls back to UTC if `Intl.DateTimeFormat().resolvedOptions().timeZone` is unavailable; this path needs a physical Android test.
- Other write endpoints (`/nutrition/entries/{entry_id}/repeat`, `/nutrition/recipes/{recipe_id}/log`, AI food confirmation) still produce a UTC `DaySummary` near midnight until optional time zone is routed through them. Current navigation reloads NutritionScreen and obtains the correct local-day summary after these writes. Plan an endpoint consistency follow-up if clients consume returned summaries.
- Date/history server aggregation is correct by IANA zone. The store queries use an inclusive upper bound, so the service queries through the exact next local midnight and excludes its entries in memory; this avoids a PostgreSQL microsecond rounding leak from `next midnight - 1 ns`. Boundary tests cover the exact next midnight.

## Tests / next role
- Go test `TestNutritionLocalCalendarDayAcrossUTCDate` covers +3 and -5, and `TestNutritionLocalDayHandlesDSTBoundary` covers 23/25-hour dates and history grouping.
- Node test `node scripts/test-mobile-nutrition-calendar.mjs` checks client IANA selection and date shifts around DST.
- Its rollover cases cover local midnight, manual archive selection, and a foreground event without a day change.
- `node scripts/test-mobile-core-screens.mjs`, `node scripts/test-mobile-nutrition-regressions.mjs`, `node scripts/test-mobile-nutrition-calendar.mjs`, and mobile `tsc --noEmit` passed locally after shared client integration.
- `go test ./internal/nutrition` is required in a Go-enabled environment; this workspace currently lacks Go.
- Lead engineer: integrate shared router/client/OpenAPI code, run Go/mobile typecheck/regressions, then device QA using a working backend and clock near local midnight.
