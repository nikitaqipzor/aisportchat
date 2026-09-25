# Active workout action serialization — engineering handoff

## Goal
Prevent Finish or Cancel from overtaking an in-flight set save on slow or unavailable networks, and prevent duplicate actions from repeated taps.

## Changes
- `ActiveWorkoutScreen.tsx` now holds a synchronous ref lock across set logging, offline queue persistence, confirmation dialogs, and Finish/Cancel callbacks. React disabled state also makes Finish/Cancel unavailable while saving; the ref protects taps before the next render.
- Confirmation cancellation and Android dialog dismissal release the lock; the confirmed callback holds it until success or failure. The empty-workout Cancel button follows the same guarded path.
- A failed offline queue write leaves the set unsaved and exposes an error, allowing retry. If queue persistence succeeds but active-workout caching fails, the in-memory view is updated from the durable queue entry and displays the cache error without claiming full offline success.
- `session.ts` serializes every local queue mutation under its authenticated owner. `sync.ts` serializes concurrent flushes, acknowledges each successful operation by its exact ID, and retains anything enqueued or replaced during network calls. A server workout snapshot is cached only while holding the queue lock and only if no newer set for that workout remains pending.

## Evidence
- `node apps/mobile/src/screens/ActiveWorkoutScreen.test.mjs`: PASS. Behavioral harness uses deferred log/finish calls to verify duplicate save, Finish/Cancel during save, repeated confirmation, dialog dismissal, failure unlock, failed queue persistence, and failed cache persistence.
- `node apps/mobile/src/storage/offlineQueue.test.mjs`: PASS. Deferred network requests verify concurrent append/replacement, concurrent flush serialization, failed requests retaining operations, account changes, parallel enqueues and cancellation ordering.
- `node scripts/test-mobile-session-scope.mjs`: PASS.
- `cd apps/mobile && npm run typecheck -- --pretty false`: PASS.
- `git diff --check`: PASS.

## Risks and unresolved items
- A request can succeed on the server yet fail before local acknowledgment is persisted (network response or disk failure); the operation may replay later. Set writes are server upserts keyed by exercise and set number, but finish/cancel replay safety still depends on server semantics. Exactly once delivery cannot be guaranteed solely by the mobile queue.
- `App.tsx` Finish/Cancel handlers swallow non-network API errors and return normally, so this screen cannot tell success from failure in that path; the parent must surface or throw the failure if retry feedback is required.
- The test harness checks screen callbacks and state without mounting native React Native UI. Device QA should cover slow save followed by Finish/Cancel, Android Back dismissal, and offline storage failure after app restart.

## Handoff
QA: run Android slow-network and offline recovery flows. Lead Engineer: review app-level callback error propagation and server terminal-operation retry semantics.
