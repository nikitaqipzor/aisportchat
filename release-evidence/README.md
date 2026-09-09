# Release evidence

This directory contains evidence produced outside source-level tests.

Required before a production Release Candidate:

- `android-device-smoke.md` — physical Android device run covering install, auth, active workout, offline restore, rest notification, Health Connect permissions/sync, Xiaomi Watch S3 data, Technique CameraX, logout/login isolation.
- GitHub Actions green run for backend PostgreSQL 18 and Android `assembleDebug`/release pipeline.
- mobile `package-lock.json` committed and `npm ci` green from a clean workspace.

Do not create placeholder evidence files just to satisfy preflight.
