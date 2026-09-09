# Sprint 4C release hardening

Owner: Lead Engineer

## Engineer
- [x] owner-scope passive Health Sync queue;
- [x] clear/suspend worker across logout/account switch;
- [x] auth rate limiter;
- [x] production secret/store guards;
- [x] PostgreSQL release test harness;
- [x] Android CI build gate;
- [x] release-preflight;
- [x] Nutrition/Workout coverage expansion.

## QA
- [x] cross-account passive sync JS regression;
- [x] race-test critical Go packages;
- [x] stale/empty/multi-source health regressions preserved;
- [x] strict preflight fails when external evidence is missing;
- [ ] observe green PostgreSQL CI job;
- [ ] observe green Android CI build;
- [ ] physical Xiaomi/device smoke.

## Design
- [x] shared button a11y fallback;
- [x] onboarding semantic controls;
- [x] Workout Preview selectors/errors;
- [x] Recovery selectors/errors;
- [x] AI Coach stop/retry/accessibility;
- [ ] whole-app selector/accessibility sweep.

## R&D
Feature expansion is frozen until release gates above are green.
