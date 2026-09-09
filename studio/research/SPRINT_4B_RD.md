# R&D note — after Sprint 4B

## What now differentiates the product
The product can connect a wearable-derived personal baseline to a deterministic strength plan while keeping source provenance visible. The useful value is not the baseline number itself; it is the causal explanation and resulting workout action.

## Ranked next experiments
1. Readiness Explanation Timeline: show which factors moved vs yesterday and vs 28-day baseline, with exact effect on today's volume/intensity.
2. Passive Health Sync: reliable background refresh + last-success diagnostics so wearable context is present before the user opens Recovery.
3. Baseline stability: rolling median/robust outlier handling once sufficient real-device history exists.
4. Personal response learning: learn associations between prior load, subjective soreness and next-session performance without making medical claims.
5. iOS HealthKit adapter behind the same normalized health contract.

## Avoid
- inventing HRV when Mi Fitness does not provide it;
- directly integrating every wearable vendor before aggregator gaps are measured;
- using an LLM to calculate readiness or silently override deterministic adaptation.
