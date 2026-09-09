# Studio workflow

## Stage A — Discovery
R&D writes a short evidence-backed opportunity note. Designer states the user problem and desired flow.

## Stage B — Engineering plan
Lead Engineer defines vertical slice, data contracts, privacy boundary, rollback path and Definition of Done.

## Stage C — Implementation
Coder ships the smallest complete path: storage → service → HTTP contract → mobile/native → tests.

## Stage D — Independent review
QA runs unit/integration/regression/native checks and records blockers. Designer reviews interaction, copy, permissions, loading/error/empty states and accessibility.

## Stage E — Lead acceptance
Lead Engineer checks evidence. Any Critical/High issue goes back to Coder. Accepted sprint updates README, roadmap, build status and studio state.

## Definition of Done
- tests pass;
- OpenAPI/schema/migrations agree;
- auth + owner isolation verified;
- no unsafe health/AI inference is presented as fact;
- UI has loading/error/empty states where applicable;
- privacy and deletion/retention implications documented;
- limitations of the current build environment are explicit;
- checkpoint ZIP integrity passes.
