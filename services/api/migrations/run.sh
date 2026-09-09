#!/bin/sh
set -eu

: "${PGHOST:=postgres}"
: "${PGPORT:=5432}"
: "${PGUSER:=fitness}"
: "${PGDATABASE:=fitness}"

psql -v ON_ERROR_STOP=1 \
  -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" \
  -c "CREATE TABLE IF NOT EXISTS schema_migrations (filename TEXT PRIMARY KEY, applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW());"

for file in /migrations/*.up.sql; do
  [ -f "$file" ] || continue
  name="$(basename "$file")"
  applied="$(psql -Atq -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE" -c "SELECT 1 FROM schema_migrations WHERE filename = '$name' LIMIT 1;")"

  if [ "$applied" = "1" ]; then
    echo "Skipping $name (already applied)"
    continue
  fi

  echo "Applying $name"
  {
    echo "BEGIN;"
    cat "$file"
    printf "\nINSERT INTO schema_migrations(filename) VALUES ('%s');\n" "$name"
    echo "COMMIT;"
  } | psql -v ON_ERROR_STOP=1 -h "$PGHOST" -p "$PGPORT" -U "$PGUSER" -d "$PGDATABASE"
done
