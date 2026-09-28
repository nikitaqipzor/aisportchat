# Backend account deletion handoff — 2026-09-26

## Goal
Authenticated account deletion removes user owned database records, revokes active and refresh tokens, and eventually removes body scan media, including unreferenced replacement photos.

## Changes
- `DELETE /api/v1/auth/account`: 204 after account and media deletion; 202 with `{"status":"media_cleanup_pending"}` after database deletion commits and media cleanup is queued. 401 for missing or deleted account, 500 for unconfirmed database failure.
- Authentication middleware checks current user existence on every protected request; signed access tokens are invalidated immediately when the user row disappears.
- Migration 000021 adds durable `pending_media_deletions` without user FK. PostgreSQL transaction inserts job and deletes user; `ON DELETE CASCADE` removes linked profile, sessions, workouts, food, health, program, technique and scan metadata. Memory store removes equivalent records.
- File and memory media stores remove only `body-scans/<user UUID>/` including orphaned old photos. Startup worker retries jobs every minute, up to 100 per pass; job stays until removal succeeds.
- OpenAPI documents response semantics and privacy limits.

## Evidence
- New Go tests: `TestAccountDeletionRemovesPrivateDataAndRevokesTokens`, `TestAccountDeletionMediaFailureQueuesRetry`, media containment and symlink tests, and PostgreSQL integration `TestPostgresAccountDeletionCascadeAndDurableMediaRetry` (requires `POSTGRES_TEST_DSN` and migrated database).
- OpenAPI YAML parsed successfully with Python PyYAML.
- Go compiler and gofmt are not installed in this workspace; tests were **not executed** here. CI must run `go test ./...` and PostgreSQL integration tests before release acceptance.

## Risks and unresolved items
- Database commit and filesystem deletion are separate resources. The durable job ensures retry after failures; 202 signals a still pending cleanup. Media must be on a persistent writable volume for the worker to reach older files.
- `os.RemoveAll` is bounded by the user UUID path and symlink parent is rejected, but the call itself is not interruptible mid traversal. It must not have untrusted writers to media root.
- Backups and external copies follow deployment retention and are outside the API purge path. Transient AI photo requests and device caches require their own policies.

## Handoff to QA / lead
1. Apply migration 000021 before deploying API, run Go test suite and PostgreSQL test with `POSTGRES_TEST_DSN`.
2. Simulate disk failure and process restart; observe 202, retained queue row, subsequent worker cleanup, and token rejection.
3. Check mobile treats both 202 and 204 as deletion success and clears device credentials; 500 leaves retry option.
