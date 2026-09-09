#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

docker compose -f "$ROOT/infra/docker-compose.yml" run --rm migrate
