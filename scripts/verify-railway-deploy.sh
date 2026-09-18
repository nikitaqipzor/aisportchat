#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dockerfile="$repo_root/services/api/Dockerfile"
migration_runner="$repo_root/services/api/migrations/run.sh"

grep -Fq 'port := resolvePort()' "$repo_root/services/api/cmd/api/main.go"
grep -Fq 'env("PORT", env("API_PORT", "8080"))' "$repo_root/services/api/cmd/api/main.go"
grep -Fq 'postgresql-client' "$dockerfile"
grep -Fq 'COPY services/api/migrations /migrations' "$dockerfile"
grep -Fq 'DATABASE_URL' "$migration_runner"
grep -Fq 'MIGRATIONS_DIR' "$migration_runner"
grep -Fq 'apps' "$repo_root/.dockerignore"
grep -Fq '/healthz' "$repo_root/docs/RAILWAY_DEPLOYMENT.md"
grep -Fq '/data/media' "$repo_root/docs/RAILWAY_DEPLOYMENT.md"

echo "Railway deployment contract verified"
