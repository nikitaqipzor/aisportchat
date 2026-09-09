#!/usr/bin/env bash
set -euo pipefail

DSN="${1:-${POSTGRES_TEST_DSN:-}}"
if [[ -z "$DSN" ]]; then
  echo "POSTGRES_TEST_DSN or DSN argument is required" >&2
  exit 2
fi
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MIGRATIONS="$ROOT/services/api/migrations"

apply_up() {
  psql "$DSN" -v ON_ERROR_STOP=1 -c "CREATE TABLE IF NOT EXISTS schema_migrations (filename TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW());" >/dev/null
  for file in "$MIGRATIONS"/*.up.sql; do
    name="$(basename "$file")"
    applied="$(psql "$DSN" -Atqc "SELECT 1 FROM schema_migrations WHERE filename='$name' LIMIT 1")"
    [[ "$applied" == "1" ]] && continue
    { echo 'BEGIN;'; cat "$file"; printf "\nINSERT INTO schema_migrations(filename) VALUES ('%s');\n" "$name"; echo 'COMMIT;'; } | psql "$DSN" -v ON_ERROR_STOP=1 >/dev/null
    echo "UP   $name"
  done
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
apply_up
echo "Migration smoke: full DOWN"
apply_down_all
echo "Migration smoke: second UP"
apply_up
echo "PostgreSQL migration smoke: PASS"
