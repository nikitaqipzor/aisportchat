# Account deletion: lead engineering handoff (2026-09-26)

## Goal

Enable a participant to delete an authenticated account, its database records, and private body scan files from the active pilot system, then clear device-side credentials and caches.

## Design and privacy boundary

- Server derives owner solely from verified bearer claims and rejects deleted-user access tokens on later requests.
- User-scoped PostgreSQL records are deleted by foreign-key cascades. The same database transaction writes a durable pending-media-removal record without a foreign key to the user, then commits account deletion.
- The server removes the validated user's body scan directory after commit and clears the pending record on success. Startup and periodic processing retry pending work. HTTP 204 means live media cleanup completed; 202 means cleanup is queued and the account is no longer accessible.
- The device clears the Keychain entry and owner-scoped app caches after either committed response and discloses pending media cleanup or local cleanup failure.
- Database and filesystem deletion are not one atomic transaction. The durable outbox makes post-commit media failure recoverable; operational backup retention and external AI-provider data remain outside this code path.

## Findings and unresolved pilot gates

See `docs/PILOT_DATA_HANDLING.md` for the data inventory. Before inviting external participants: publish privacy materials and optional-feature choices, decide backup retention and restore-time reconciliation, exercise deletion with staging PostgreSQL and media volume, and run an Android device smoke test. Do not claim backups or third-party records are erased by this endpoint.

## Two-stage review and evidence

1. Design and source review: rejected an initial implementation that deleted media before PostgreSQL commit. Accepted the durable outbox design for retry after a crash or volume outage. Reviewed ownership, token invalidation, directory containment, 202/204 response copy, native cache locations, and every direct `users` foreign key across migrations 000001–000020; each direct user ownership FK uses `ON DELETE CASCADE`. Cascades through workout, scan, recipe and program parents still need live database execution.
2. Integration/static QA: `python3 scripts/verify-api-contract.py` passed (83 OpenAPI operations, 84 router routes including healthz); `python3 scripts/verify-migrations.py` passed (21 migration pairs); `python3 scripts/verify_android_native.py` passed (125 checks); `npm run typecheck --prefix apps/mobile`, `node scripts/verify-mobile-syntax.mjs`, `node scripts/test-mobile-auth-refresh.mjs`, and `node apps/mobile/src/storage/offlineQueue.test.mjs` passed. `git diff --check` passed. New Go unit and PostgreSQL integration tests have been written but **cannot run locally** because Go and PostgreSQL are not installed in this workspace.

Lead disposition: **conditional, not release accepted**. Run `go test ./...` with the Go toolchain and `POSTGRES_TEST_DSN` against a migrated disposable PostgreSQL database; reproduce a 202 media outage and worker retry after restart; build and smoke-test the Android APK. A pilot using external participants remains blocked on the operational and privacy items above.

## Handoff

Backend coder → mobile coder → independent integration QA → lead engineer release decision.
