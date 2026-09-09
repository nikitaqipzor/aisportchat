# AI Fitness OS — Agent Studio

This repository is developed through a role-based virtual engineering studio. Every meaningful feature passes through the same chain:

`R&D / Market → Lead Engineer → Coder → QA → Designer → Lead acceptance`

The Lead Engineer owns release acceptance. A feature is not `done` merely because it compiles.

## Non-negotiable architecture rules
1. LLMs never become the source of truth for calories, workouts, readiness, health records, rep counts, or progression.
2. Deterministic engines calculate and validate; PostgreSQL stores facts; CV/health providers measure; AI explains and orchestrates bounded tools.
3. Health/body/media data is private by default, least-permission, user-owned, and deletable.
4. Mobile writes must survive poor connectivity and be scoped to the authenticated user.
5. New native code must pass `scripts/verify_android_native.py`; backend changes must include Go tests; API changes must update OpenAPI.
6. No release claim without stating what was actually executed versus only statically verified.

## Agent handoff contract
Every role leaves an artifact in `studio/` with:
- goal;
- changes or findings;
- risks;
- tests/evidence;
- unresolved items;
- explicit handoff to the next role.

See `studio/TEAM.md` and `studio/WORKFLOW.md`.
