# Five-screen fixes: lead handoff

## Goal

Close the five priority findings from the 30-screen audit: reachable planned workouts, safe food retries, correct nutrition calendar days, trustworthy device status, and accessible mobile navigation.

## Changes and evidence

- Program sessions reopen their linked workout, including planned and active states; the preview returns to its source and the program remains usable if optional analytics fails.
- Recipe and AI confirmations use atomic batch storage with user-scoped retry keys. Reconnection inside a PostgreSQL transaction is prohibited.
- Nutrition day/history and returned summaries accept optional IANA `time_zone`; existing API callers still default to UTC. The mobile client handles late responses, daylight-saving boundaries, and an open screen crossing midnight.
- Connected devices distinguish stale readings and recheck live permissions. The Home CTA, bottom tabs, body map, and technique guide have clearer labels and larger touch targets.
- Local TypeScript check, 17 mobile regression scripts, OpenAPI YAML parsing, Android native static verification, and `git diff --check` passed.

## Risks and release gate

- GitHub Actions must pass backend Go + PostgreSQL integration, static mobile, and Android internal APK before merging. Local Go, PostgreSQL, Docker and Android SDK are unavailable in the workspace.
- Physical-device rendering, 320dp/large text, TalkBack, and Health Connect permissions remain to be tested.
- Offline food actions retain retry identity but do not yet use a durable send queue. A planned workout from an archived program can still be started from History; the archived restriction currently lives in the program UI only.

## Handoff

Lead engineer: review the draft PR's four CI jobs, then merge only if all pass. Product QA: test the connected APK against a real backend and a physical Android device before wider invitations.
