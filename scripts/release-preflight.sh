#!/usr/bin/env bash
set -uo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
MOBILE="$ROOT/apps/mobile"
API="$ROOT/services/api"
STRICT="${STRICT_RELEASE:-0}"
failures=0
warnings=0

ok() { printf 'PASS  %s\n' "$1"; }
warn() { printf 'WARN  %s\n' "$1"; warnings=$((warnings+1)); }
fail() { printf 'FAIL  %s\n' "$1"; failures=$((failures+1)); }

printf 'AI Fitness OS — Release Preflight\n'
printf '==================================\n'

if "$ROOT/scripts/verify.sh" >/tmp/aifitness-verify.log 2>&1; then
  ok "repository regression verifier"
else
  fail "repository regression verifier (see /tmp/aifitness-verify.log)"
fi

if (cd "$API" && go test -race -p=1 ./internal/httpapi ./internal/auth ./internal/healthdata ./internal/recovery ./internal/nutrition ./internal/workouts >/tmp/aifitness-race.log 2>&1); then
  ok "critical Go packages under race detector"
else
  fail "critical Go packages under race detector (see /tmp/aifitness-race.log)"
fi

if [[ -f "$MOBILE/package-lock.json" ]]; then
  ok "mobile npm lockfile present"
else
  if [[ "$STRICT" == "1" ]]; then fail "mobile package-lock.json missing"; else warn "mobile package-lock.json missing — clean npm build is not reproducible yet"; fi
fi

if command -v psql >/dev/null 2>&1 && [[ -n "${POSTGRES_TEST_DSN:-}" ]]; then
  if POSTGRES_TEST_DSN="$POSTGRES_TEST_DSN" "$ROOT/scripts/postgres-migration-smoke.sh" >/tmp/aifitness-pg-migrations.log 2>&1 && \
     (cd "$API" && POSTGRES_TEST_DSN="$POSTGRES_TEST_DSN" go test -tags cgo ./internal/httpapi -run TestPostgresReleaseFlow -count=1 >/tmp/aifitness-pg-http.log 2>&1); then
    ok "real PostgreSQL migration + HTTP release flow"
  else
    fail "real PostgreSQL release gate failed"
  fi
else
  if [[ "$STRICT" == "1" ]]; then fail "POSTGRES_TEST_DSN/psql unavailable — real PostgreSQL gate not executed"; else warn "real PostgreSQL gate not executed locally; GitHub CI job is configured"; fi
fi

SDK="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-}}"
if [[ -n "$SDK" && -d "$SDK" ]]; then
  ok "Android SDK detected at $SDK"
else
  if [[ "$STRICT" == "1" ]]; then fail "Android SDK unavailable"; else warn "Android SDK unavailable locally; android-debug CI job is configured"; fi
fi

if [[ -f "$ROOT/release-evidence/android-device-smoke.md" ]]; then
  ok "physical Android device smoke evidence recorded"
else
  if [[ "$STRICT" == "1" ]]; then fail "physical Android device smoke evidence missing"; else warn "physical Android/Xiaomi device smoke evidence missing"; fi
fi

printf '\nSummary: %d failure(s), %d warning(s)\n' "$failures" "$warnings"
if (( failures > 0 )); then exit 1; fi
if [[ "$STRICT" == "1" && "$warnings" -gt 0 ]]; then exit 1; fi
