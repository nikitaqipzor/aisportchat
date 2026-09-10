#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
compose_file="$repo_root/infra/staging/docker-compose.yml"

export API_DOMAIN="api.example.com"
export POSTGRES_PASSWORD="0123456789abcdef0123456789abcdef0123456789abcdef"
export AUTH_TOKEN_SECRET="0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

docker compose -f "$compose_file" config --quiet

grep -Fq 'APP_ENV: production' "$compose_file"
grep -Fq 'STORE_BACKEND: postgres' "$compose_file"
grep -Fq 'condition: service_completed_successfully' "$compose_file"
grep -Fq 'AUTH_TOKEN_SECRET: ${AUTH_TOKEN_SECRET:?set AUTH_TOKEN_SECRET}' "$compose_file"
grep -Fq '{$API_DOMAIN}' "$repo_root/infra/staging/Caddyfile"

echo "Staging deployment contract verified"
