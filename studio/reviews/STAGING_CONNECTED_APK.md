# Staging + connected APK handoff

## Goal

Provide a reproducible path from the repository to a public HTTPS API and an Android preview that is connected to that API.

## Changes

- added a staging Compose topology with PostgreSQL 18, Redis 8, one-shot migrations, the Go API and Caddy TLS termination;
- made database, media and Caddy state persistent;
- required production mode, PostgreSQL and non-default secrets;
- added a deployment contract check to the release gate;
- added a manually dispatched connected APK workflow that validates/probes the API URL before embedding it;
- documented deploy, update, APK and recovery procedures.

## Risks

- a real VPS, DNS name and user-owned secrets are still required;
- backups and restores are operational procedures and are not automated in this slice;
- the staging topology runs one API instance and is not highly available;
- physical Android verification remains external evidence.

## Tests / evidence

- `scripts/verify-staging-deploy.sh` validates the rendered Compose configuration and required security wiring;
- normal CI still runs backend, PostgreSQL, mobile and Android gates;
- connected APK workflow refuses HTTP and unavailable APIs.

## Unresolved items

- provision the actual server and DNS;
- run the workflow with its public API URL;
- complete physical registration/profile/workout/history smoke testing.

## Handoff

Lead Engineer should accept only after CI passes. Operations then supplies the VPS/DNS values without committing secrets.
