#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/../apps/mobile"
if [[ -f package-lock.json ]]; then
  echo "Using reproducible npm ci from package-lock.json"
  npm ci --no-audit --no-fund
else
  echo "WARNING: package-lock.json is missing; using npm install for build validation only." >&2
  echo "A committed lockfile remains mandatory before production release." >&2
  npm install --no-audit --no-fund
fi
