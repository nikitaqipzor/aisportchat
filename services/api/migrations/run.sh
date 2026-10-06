#!/bin/sh
set -eu

: "${PGHOST:=postgres}"
: "${PGPORT:=5432}"
: "${PGUSER:=fitness}"
: "${PGDATABASE:=fitness}"
: "${MIGRATIONS_DIR:=/migrations}"

run_psql() {
  if [ -n "${DATABASE_URL:-}" ]; then
    psql "$DATABASE_URL" "$@"
  else
    psql -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" "$@"
  fi
}

run_psql -v ON_ERROR_STOP=1 \
  -c "BEGIN; SELECT pg_advisory_xact_lock(20260928, 23001); CREATE TABLE IF NOT EXISTS schema_migrations (filename TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()); COMMIT;"

for file in "$MIGRATIONS_DIR"/*.up.sql; do
  [ -f "$file" ] || continue
  name="$(basename "$file")"
  # Serialize deploys and check the ledger only after acquiring the lock.
  # The transaction-scoped lock is released with this migration's commit.
  {
    echo "BEGIN;"
    echo "SELECT pg_advisory_xact_lock(20260928, 23001);"
    printf "SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE filename = '%s') AS already_applied \\gset\n" "$name"
    printf '%s\n' '\if :already_applied'
    printf '\\echo Skipping %s (already applied)\n' "$name"
    echo "ROLLBACK;"
    printf '%s\n' '\else'
    printf '\\echo Applying %s\n' "$name"
    cat "$file"
    printf "\nINSERT INTO schema_migrations(filename) VALUES ('%s');\n" "$name"
    echo "COMMIT;"
    printf '%s\n' '\endif'
  } | run_psql -v ON_ERROR_STOP=1
done
