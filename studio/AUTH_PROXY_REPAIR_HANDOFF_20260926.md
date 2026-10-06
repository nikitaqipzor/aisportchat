# Authentication limiter behind Caddy — 2026-09-26

## Goal
Avoid one shared login IP bucket for all users behind the staging reverse proxy without trusting client-controlled forwarding headers.

## Changes
- `TrustedProxyClientIPHandler` reads the rightmost `X-Forwarded-For` address only if the immediate TCP peer is within `AUTH_TRUSTED_PROXY_CIDRS`. Invalid configuration fails API startup; an absent configuration uses the direct peer address.
- Staging Compose pins Caddy at `172.30.239.10` on a private edge network and trusts only `/32` of that address. Caddy explicitly sets `X-Forwarded-For` to its direct client peer, discarding client-supplied values.
- Added behavioral tests for distinct users behind Caddy, an untrusted spoof attempt, invalid headers/configuration, and independent rate buckets.

## Risks and tests
- `git diff --check` and Compose YAML parsing passed locally. Go toolchain is unavailable here; CI Go race and PostgreSQL tests are required.
- Docker Compose's fixed subnet can conflict with a host network. Change both pinned Caddy IP and `AUTH_TRUSTED_PROXY_CIDRS` together if needed.
- Railway ingress proxy IP ranges have not been verified, so the runbook deliberately leaves proxy trust unset there. Verify the provider's actual source ranges and test independent rate buckets through the public URL before pilot traffic.
- Rate-limit counters remain process-local; a multi-instance deployment needs shared storage or a gateway policy.

## Handoff
Lead engineer and release owner: review trust boundary and staging network, run Go CI, then verify two real clients through deployed HTTPS Caddy.
