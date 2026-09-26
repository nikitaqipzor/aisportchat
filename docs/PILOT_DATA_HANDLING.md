# Pilot data handling — engineering inventory

This is an implementation inventory for a closed pilot. It is not a published privacy notice. Before external participants are invited, the operator must publish a notice and consent flow appropriate to the deployment and decide retention, backup and third-party processing terms.

## Data locations

| Data | Primary location | Deletion path |
| --- | --- | --- |
| Account email, password hash, refresh sessions | PostgreSQL `users` and related tables | Account deletion removes the user row and dependent records. Verify in a database-backed test. |
| Profile, injuries/limitations, workouts, programs, nutrition, measurements, recovery, health snapshots and technique results | PostgreSQL tables referencing `users` | User cascade; custom food and recipes need particular verification because food entries reference food IDs. |
| Body scan photos | Private `MEDIA_ROOT` filesystem volume, with keys in `body_scan_photos` | Account deletion queues a durable media-removal job with the database deletion. A 202 response means media cleanup is pending; a 204 means live media cleanup completed. |
| Access token | Client keychain / bearer JWT | Client clears credentials; server must reject a token for a deleted user immediately. |
| Offline workouts, chat, meal action IDs, preferences and timers | Device local storage | Clear owner-scoped caches and pending writes after successful server deletion. |
| Food image sent for recognition | In-memory request, optionally sent to configured OpenAI provider | No food photo table or backend file copy was found; the provider's processing and retention require separate disclosure. |

The AI provider's Requests use `store: false` in this repository. That flag alone does not establish the provider's complete retention terms. Health Connect data is imported into user-scoped PostgreSQL snapshots and may also remain with the originating provider outside this application's deletion boundary.

## Pilot gates

1. Exercise account deletion against staging PostgreSQL and private media volume using an account with photos and all major record types. Confirm the deleted access and refresh tokens fail, a second account remains untouched, and no owned blob is left.
2. Decide and document backup retention and restore-time deletion reconciliation for database **and** media volume. A deletion from the live store does not automatically erase earlier backups.
3. Publish operator identity, purposes, categories, external processors, retention periods and contact route in the pilot's privacy materials. Present them before collecting health/body imagery. Obtain required user choices for optional health and AI features.
4. Verify device storage cleanup and retry/error behavior on actual Android devices, including interrupted deletion and offline pending writes.
5. Limit access to body images and operations logs; prevent photo bytes, JWTs and sensitive health details from entering logs.
6. Monitor pending media-removal jobs and alert on jobs that remain pending; test retry after a media-volume error or process restart.

Do not describe the feature as complete for an external pilot until the deletion implementation, tests and operational gates above have been reviewed against the deployed environment.
