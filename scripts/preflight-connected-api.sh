#!/usr/bin/env bash
set -euo pipefail

api_base_url="${1:-}"
expected_release_id="${2:-}"
workflow_commit_sha="${3:-}"
if [[ -z "$expected_release_id" || "$expected_release_id" == *$'\n'* || "$expected_release_id" == *$'\r'* ]]; then
  echo "expected_release_id must be a non-empty single-line value" >&2
  exit 1
fi
if [[ ! "$expected_release_id" =~ ^[0-9a-f]{40}$ || ! "$workflow_commit_sha" =~ ^[0-9a-f]{40}$ || "$expected_release_id" != "$workflow_commit_sha" ]]; then
  echo "expected_release_id must exactly match the checked-out workflow commit SHA" >&2
  exit 1
fi

# Parse strictly and require the canonical HTTPS base path. In particular, avoid
# credentials, query/fragment suffixes, encoded paths, and alternate path spellings.
if ! python3 - "$api_base_url" <<'PY'
import sys
from urllib.parse import urlsplit

raw = sys.argv[1]
try:
    u = urlsplit(raw)
    port = u.port
except ValueError:
    raise SystemExit(1)
if (not raw or any(c.isspace() for c in raw) or "\\" in raw or
        u.scheme != "https" or not u.hostname or u.username is not None or
        u.password is not None or u.path != "/api/v1" or u.query or u.fragment or
        "%" in u.path or raw.endswith("/") or not u.netloc or
        (port is not None and not 1 <= port <= 65535)):
    raise SystemExit(1)
PY
then
  echo "api_base_url must be a canonical HTTPS URL ending exactly in /api/v1 (without credentials, query, or fragment)" >&2
  exit 1
fi

origin="${api_base_url%/api/v1}"
curl_bin="${CURL_BIN:-curl}"
probe_json() {
  local url="$1" body status
  body="$(mktemp)"
  # Deliberately do not follow redirects: a redirect must not move the probe to
  # a different host or conceal an unexpected endpoint.
  if ! status="$("$curl_bin" --silent --show-error --fail --retry 5 --retry-all-errors --retry-delay 2 --connect-timeout 5 --max-time 15 \
      --output "$body" --write-out '%{http_code}' "$url")"; then
    rm -f "$body"
    echo "API preflight request failed" >&2
    return 1
  fi
  if [[ "$status" != "200" ]]; then
    rm -f "$body"
    echo "API preflight endpoint returned HTTP $status (expected 200)" >&2
    return 1
  fi
  printf '%s' "$body"
}

health_file="$(probe_json "$origin/healthz")" || exit 1
if ! python3 - "$health_file" "$expected_release_id" <<'PY'
import json, sys
try:
    with open(sys.argv[1], encoding="utf-8") as f:
        payload = json.load(f)
except (OSError, UnicodeError, json.JSONDecodeError):
    raise SystemExit(1)
if not isinstance(payload, dict) or payload.get("status") != "ok" or payload.get("release_id") != sys.argv[2]:
    raise SystemExit(1)
PY
then
  rm -f "$health_file"
  echo "API health response is invalid or release_id did not match" >&2
  exit 1
fi
rm -f "$health_file"

catalog_file="$(probe_json "$api_base_url/muscles")" || exit 1
if ! python3 - "$catalog_file" <<'PY'
import json, sys
try:
    with open(sys.argv[1], encoding="utf-8") as f:
        payload = json.load(f)
except (OSError, UnicodeError, json.JSONDecodeError):
    raise SystemExit(1)
items = payload.get("items") if isinstance(payload, dict) else None
if not isinstance(items, list) or not items or not all(
    isinstance(item, dict) and isinstance(item.get("id"), str) and isinstance(item.get("name"), str)
    for item in items
):
    raise SystemExit(1)
PY
then
  rm -f "$catalog_file"
  echo "API catalog response is not the expected muscles API JSON" >&2
  exit 1
fi
rm -f "$catalog_file"
echo "Connected API preflight passed"
