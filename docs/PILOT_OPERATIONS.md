# Closed pilot operations

## Before a staging update

1. Deploy the reviewed candidate SHA only to the separate staging host and set
   `PILOT_RELEASE_ID` to that exact SHA. Keep secrets in the host's secret
   store or an untracked `.env`; never add `.env` to a commit.
2. Take and verify a protected backup of PostgreSQL and `/data/media` before
   applying migrations. These contain private training, health and body-scan
   data. Store backups encrypted with access limited to pilot operators, and
   use the retention period approved for the pilot. The Compose stack does not
   automate backup scheduling or encryption.
3. Apply the update with the repository's staging Compose instructions. The
   one-shot migration service serializes concurrent runners and the API waits
   for migration completion. The API health check calls `/readyz`, which must
   return success only while PostgreSQL is reachable.
4. Verify the deployed Git SHA separately from the configured release marker,
   then run the connected API preflight and record the result. The marker is
   configuration and does not prove which source revision was built.

## Backup and recovery boundary

The PostgreSQL and media Docker volumes persist across container replacement,
but persistence is not a backup. Before an update, quiesce API writes while
capturing the database and media volume so the pair is consistent. Protect
both artifacts as sensitive user data and document the operator, timestamp,
retention deadline and restore test. Do not claim account deletion removed
copies already present in backups; apply the documented retention schedule.

Do not use the PostgreSQL migration smoke test against staging or any database
with user data. It runs every down migration and drops `schema_migrations`;
the script refuses to run unless `ALLOW_DESTRUCTIVE_MIGRATION_SMOKE=1` is set.
For production-like recovery, restore a verified backup into an isolated
database and media location first, validate it, then switch staging over under
an approved maintenance window. There is no automated rollback for migrations
or coordinated volume restore in this repository.

## Readiness

Compose checks `http://127.0.0.1:8080/readyz` from the API container. The image
contains `curl` for this check. A healthy container indicates the HTTP
readiness route and its PostgreSQL ping succeeded; it does not prove media
availability, account isolation, backups, or end-to-end mobile behavior.
