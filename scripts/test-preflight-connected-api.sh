#!/usr/bin/env bash
set -euo pipefail
root="$(cd "$(dirname "$0")/.." && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cat >"$tmp/curl" <<'MOCK'
#!/usr/bin/env bash
set -euo pipefail
out=""; url=""
while (($#)); do
  case "$1" in
    --output) out="$2"; shift 2 ;;
    --write-out) shift 2 ;;
    *) url="$1"; shift ;;
  esac
done
case "$url" in
  */healthz)
    case "${MOCK_MODE:-ok}" in
      html) printf '<html>oops</html>' >"$out" ;;
      wrong-release) printf '{"status":"ok","release_id":"other"}' >"$out" ;;
      missing-release) printf '{"status":"ok"}' >"$out" ;;
      *) printf '{"status":"ok","release_id":"0123456789abcdef0123456789abcdef01234567"}' >"$out" ;;
    esac
    ;;
  */api/v1/muscles)
    case "${MOCK_MODE:-ok}" in
      wrong-api) printf '{"message":"not this API"}' >"$out" ;;
      *) printf '{"items":[{"id":"back","name":"Back"}]}' >"$out" ;;
    esac
    ;;
  *) exit 2 ;;
esac
printf '200'
MOCK
chmod +x "$tmp/curl"
export CURL_BIN="$tmp/curl"
pass() { "$root/scripts/preflight-connected-api.sh" "$1" "$2" "$3" >/dev/null; }
fail() { if pass "$1" "$2" "$3" >/dev/null 2>&1; then echo "unexpected preflight success: $1" >&2; exit 1; fi; }

release='0123456789abcdef0123456789abcdef01234567'
pass 'https://api.example.test/api/v1' "$release" "$release"
for bad in \
  'http://api.example.test/api/v1' \
  'https://user:pass@api.example.test/api/v1' \
  'https://api.example.test/api/v1?x=1' \
  'https://api.example.test/api/v1#frag' \
  'https://api.example.test/api/v1/' \
  'https://api.example.test/other'; do
  fail "$bad" "$release" "$release"
done
fail 'https://api.example.test/api/v1' "$release" 1123456789abcdef0123456789abcdef01234567
fail 'https://api.example.test/api/v1' abc123 abc123
for mode in html wrong-release missing-release wrong-api; do
  export MOCK_MODE="$mode"
  fail 'https://api.example.test/api/v1' "$release" "$release"
done
echo "Connected API preflight tests passed"
