# Health endpoint release identity handoff

## Goal
Expose a stable deployment identity for pilot smoke checks without adding configuration details or secrets to the health response.

## Changes
- `GET /healthz` preserves the existing `status` and `time` fields.
- A nonblank `PILOT_RELEASE_ID` is trimmed and returned as `release_id`.
- An unset or whitespace-only value omits `release_id`.
- Added Go coverage for the configured, blank, and secret-leak cases.
- Documented `/healthz` in OpenAPI with a root server override, required `status`/`time`, and optional `release_id`.

## Risks
- The release identity is visible to unauthenticated callers of `/healthz`; configure only a non-secret marker.
- No startup validation is performed; missing identity remains compatible and is represented by omission.

## Tests / evidence
- `python3 scripts/verify-api-contract.py` passed: 84 OpenAPI operations aligned with 84 router routes.
- OpenAPI YAML parsing and `git diff --check` passed.
- `go test ./internal/httpapi` could not run because Go tooling is unavailable (`gofmt: command not found`).
- Changes were reviewed statically; no runtime verification has been performed.

## Unresolved items
- Run Go formatting and the targeted API tests in an environment with Go installed.

## Handoff
Lead Engineer / QA: run `gofmt` and `go test ./internal/httpapi`, then verify the pilot deployment injects a non-secret `PILOT_RELEASE_ID`.
