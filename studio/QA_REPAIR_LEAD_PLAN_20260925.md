# QA repair plan — 2026-09-25

## Goal and acceptance
Repair user journeys found in the full screen audit. The lead accepts each slice after an implementation review (data ownership, concurrency, input limits, failure handling) and a separate regression review (targeted tests, API contracts and cross-screen navigation). Real Android taps require a device or emulator and a connected backend; static checks alone cannot establish device acceptance.

| Priority | Owner/model | Scope | Acceptance |
| --- | --- | --- | --- |
| P1 | archived_program / gpt-6-sol | program session HTTP and OpenAPI | archived sessions cannot create or return active workouts; backend regression covers rejection and no insert |
| P1 | bodyscan / gpt-6-sol | body scan service and tests | metadata failure does not remove prior photo; new blob cleanup tested |
| P1 | active_workout / gpt-6-sol | active workout screen | finish and cancel cannot overtake an in-flight set save; retry and failed save remain visible |
| P2 | nutrition_p2 / gpt-5.6-sol | nutrition screens and nutrition navigation | correct limits and draft retention; system Back matches screen Back |
| P2 | other_mobile_p2 / gpt-5.6-sol | athlete profile, AI coach, technique, programs | no-equipment state, duplicate send, zero repetitions, complete program list |
| P2 | lead | repeat workout navigation, CI device-display gate | Back returns to original detail; regression runs in CI |

These are callable models in the current workspace. Although other versions have been discussed, no 5.5 or 5 model override is exposed here. Agent files must remain disjoint; App.tsx belongs to nutrition_p2 until its handoff. Every owner records evidence and unresolved risks in its own `studio/` handoff.

## Review and integration
1. Check each diff for user ownership, transactional order, network and lifecycle races, and realistic error handling. Request corrections from its owner.
2. Run the targeted regression gates, then mobile syntax/typecheck, Go unit/race tests where available, native verification if native files changed, and verify OpenAPI alignment.
3. Inspect combined navigation after screens and App.tsx land; document device testing still required and hand off to remote PR owner.
