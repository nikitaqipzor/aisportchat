# Changelog — Sprint 4C R1

## Security / privacy
- owner-scoped encrypted passive Health Connect queues;
- logout/account-switch worker isolation;
- legacy unscoped pending health cache invalidation;
- auth IP/account/token rate limiting;
- production JWT secret validation;
- production PostgreSQL-only store enforcement.

## Reliability
- PostgreSQL 18 CI integration flow;
- migration `up → down → up` smoke harness;
- Android SDK native build CI job;
- strict release-preflight and evidence directory.

## Quality
- Nutrition coverage 85.5%;
- Workout coverage 80.3%;
- backend total coverage 56.2%;
- critical race detector pass;
- first broad accessibility/testID pass;
- AI Coach stop/retry/abort UX.

## Product
- passive Health Connect sync foundation;
- deterministic Readiness Explanation Timeline;
- health context separated from causal readiness factors.

## Known blockers
- package-lock cannot be generated in this environment;
- real PostgreSQL CI execution evidence pending;
- Android build evidence pending;
- physical Xiaomi Watch S3 / Health Connect validation pending.
