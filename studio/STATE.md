# Studio state

Current sprint: **4C R2 — QA Hardening**.

Lead decision: **SOURCE ACCEPTED / RELEASE CANDIDATE BLOCKED**.

Resolved high-risk audit findings:
1. Passive Health Sync is owner-scoped and isolated across logout/crash/account switch.
2. Auth endpoints have IP/identity/token throttling.
3. Production rejects weak/default JWT secrets.
4. Production rejects memory storage.
5. PostgreSQL 18 integration/migration CI gate is defined.
6. Android SDK native build CI gate is defined.
7. Nutrition and Workout critical-domain test coverage exceeds 80%.
8. All 28 mobile screens have source-level loading/error/empty, validation and accessibility hardening.
9. Mobile dependencies are locked and the strict TypeScript typecheck is green.

Strict external gates still required for an RC:
1. green real PostgreSQL 18 CI run on the candidate commit;
2. green Android native CI build on the candidate commit;
3. physical Android smoke;
4. Xiaomi Watch S3 / Mi Fitness / Health Connect calibration and cross-account device smoke.

Development policy: **feature freeze for major functionality until release blockers are closed**.
