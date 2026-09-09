# Sprint 2B — Nutrition Expansion & Body Progress

## Goal
Turn Nutrition Core into a reusable daily system and establish the structured body-progress data that later AI agents will consume.

## Delivered
- user-owned custom foods with privacy isolation;
- optional barcode field and exact barcode lookup;
- custom/global food search isolation (global + own only);
- saved recipes with multiple ingredients;
- one-tap recipe logging;
- repeat previous food entry;
- body measurements (weight, waist, chest, shoulders, arm, forearm, hip, thigh, calf, neck);
- body weight measurement synchronizes the canonical user profile weight;
- 30-day progress summary and measurement history;
- training-day vs rest-day nutrition correlation;
- mobile screens for custom foods, recipes and body progress;
- lightweight native progress visualization without a chart dependency;
- migration `000010_nutrition_progress`;
- OpenAPI `0.7.0`;
- integration test for custom food → barcode → recipe → repeat → measurement → progress/correlation.

## Important architecture choices
1. LLM does not calculate calories/macros or mutate measurements. These remain deterministic domain data.
2. Custom food ownership is enforced in the store contract, not only in UI filtering.
3. Body measurements are stored independently from photos/body scan so AI Fitness can add computer vision later without changing the measurement history model.
4. Correlation only averages days with logged nutrition data to avoid treating missing food logs as zero-calorie days.

## Next
Sprint 2C: AI Food Input, food-photo intake pipeline, AI Coach tool layer, user fitness context, weekly deterministic facts + AI narrative report.
