#!/usr/bin/env bash
set -euo pipefail

DSN="${1:-${POSTGRES_TEST_DSN:-}}"
if [[ -z "$DSN" ]]; then
  echo "POSTGRES_TEST_DSN or DSN argument is required" >&2
  exit 2
fi
if [[ "${ALLOW_DESTRUCTIVE_MIGRATION_SMOKE:-0}" != "1" ]]; then
  echo "This smoke test drops the migration ledger and rolls back every migration; set ALLOW_DESTRUCTIVE_MIGRATION_SMOKE=1 only for a disposable database" >&2
  exit 2
fi
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MIGRATIONS="$ROOT/services/api/migrations"

apply_up() {
  DATABASE_URL="$DSN" MIGRATIONS_DIR="$MIGRATIONS" "$MIGRATIONS/run.sh"
}

apply_down_all() {
  mapfile -t files < <(find "$MIGRATIONS" -maxdepth 1 -name '*.down.sql' -type f | sort -r)
  for file in "${files[@]}"; do
    name="$(basename "$file")"
    psql "$DSN" -v ON_ERROR_STOP=1 -f "$file" >/dev/null
    echo "DOWN $name"
  done
  psql "$DSN" -v ON_ERROR_STOP=1 -c "DROP TABLE IF EXISTS schema_migrations;" >/dev/null
}

echo "Migration smoke: first UP"
DATABASE_URL="$DSN" MIGRATIONS_DIR="$MIGRATIONS" "$MIGRATIONS/run.sh" &
first_runner=$!
DATABASE_URL="$DSN" MIGRATIONS_DIR="$MIGRATIONS" "$MIGRATIONS/run.sh" &
second_runner=$!
wait "$first_runner"
wait "$second_runner"
echo "Concurrent migration runners: PASS"
echo "Migration smoke: full DOWN"
apply_down_all
echo "Migration smoke: second UP"
apply_up
echo "PostgreSQL migration smoke: PASS"
