# Sprint 2A — Deterministic Nutrition Core

## Goal
Add the first Fitness 2.0 vertical slice without making an LLM responsible for calorie or macro arithmetic.

## Delivered

### Backend domain
- `internal/nutrition.Service` owns target calculation and daily aggregation.
- automatic starter targets use current body weight + selected goal + selected activity level;
- manual target mode can override calories/protein/fat/carbs;
- daily totals and remaining targets are deterministic;
- a completed workout on the same UTC day marks the nutrition day as a training day;
- 7–90 day history aggregates a date range in bulk instead of querying every day independently.

### Nutrition API
- `GET /nutrition/profile`
- `PUT /nutrition/profile`
- `GET /nutrition/today`
- `GET /nutrition/history`
- `GET /nutrition/foods`
- `POST /nutrition/entries`
- `DELETE /nutrition/entries/{entry_id}`

### Persistence
Migration `000009_nutrition_core` adds:
- `nutrition_profiles`;
- `food_items`;
- `food_entries`;
- lookup/history indexes;
- 20 starter foods for development and product testing.

The memory adapter mirrors the same contract for tests.

### Mobile
New screens:
- `NutritionScreen` — daily calories/macros, diary, training-day marker and 7-day history;
- `NutritionSetupScreen` — auto/manual targets, goal and activity level;
- `FoodSearchScreen` — food search, meal category, grams and live macro preview.

Home now exposes Nutrition as a first-class product area.

## Calculation boundary
The initial automatic calculation is deliberately an estimate, not a medical prescription. It is editable and will later be adapted from real trends. The LLM will only explain/propose bounded changes; deterministic code and user-approved values remain the source of truth.

## Not in this sprint
- barcode scanning;
- external food databases;
- saved meals/recipes;
- photo recognition;
- natural-language food logging;
- AI Coach;
- weight-trend adaptation.

These remain in Sprint 2B/2C.

## Verification
- Go unit + integration tests include a full nutrition flow.
- `go vet ./...` passes.
- OpenAPI YAML parses at version `0.6.0`.
- React Native TS/TSX source passes TypeScript syntax transpilation.
- Full React Native typecheck/native build still requires dependency installation on a developer machine.
