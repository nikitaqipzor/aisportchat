# AI Fitness OS

**Current checkpoint: Phase 4 / Sprint 4C R2 — Passive Health Sync, Readiness Timeline & Release Hardening.**

Mobile-first fitness platform with deterministic workout, nutrition and program engines, bounded AI tools, native Android infrastructure and a privacy-first body progress and on-device pose-analysis pipeline.

## Implemented
- registration + onboarding;
- home/gym/band workout generation;
- active workout, sets, RPE/RIR, rest timer, offline queue and restore;
- history, PRs, progressive overload and programs with deload/adherence;
- nutrition targets, food diary, custom foods, recipes, measurements and trends;
- AI food parsing/photo draft, read-only AI Coach and weekly report;
- React Native 0.87 Android native scaffold + TurboModules;
- Body Scan front/side/back capture, quality checks, private storage, delete and comparison consistency.
- Recorded + live technique analysis: CameraX preview, Android MediaPipe pose extraction, skeleton overlay, live rep counter, ROM/tempo/symmetry/stability scores, phase timing and history.
- Technique results can be linked to an active workout exercise/set and prefill camera-counted repetitions back into the workout.
- Daily Recovery check-in, readiness 0–100, per-muscle recovery map and deterministic workout adaptation are implemented.
- Android Health Connect ingestion is implemented with Xiaomi Watch S3 / Mi Fitness as the first preferred source: steps, distance, active calories, sleep, exercise sessions and exercise heart-rate summaries feed normalized daily health snapshots and Recovery.
- Health data now has separate per-source persistence, deterministic conflict resolution, 7/28-day personal baselines, freshness/confidence/provenance diagnostics and stale-data fallback. AI Coach can explain this context through a read-only tool.

## Architecture rule
LLM is not the source of truth. Deterministic engines calculate and validate, PostgreSQL stores facts, CV measures visual signals, and AI explains/orchestrates bounded actions.

## Repository
- `apps/mobile` — React Native + native Android project
- `services/api` — Go modular monolith
- `services/api/migrations` — PostgreSQL migrations
- `services/api/openapi` — OpenAPI contract
- `infra` — PostgreSQL / Redis / migration / API compose stack
- `docs` — sprint and architecture documentation

## Verify
```bash
./scripts/verify.sh
```

## Release preflight
```bash
./scripts/release-preflight.sh
STRICT_RELEASE=1 ./scripts/release-preflight.sh
```

The strict command intentionally blocks a Release Candidate until dependency lockfile, real PostgreSQL, Android build and physical-device evidence are present.

## Run API stack
```bash
docker compose -f infra/docker-compose.yml up --build
```

Body scan media is stored under `MEDIA_ROOT`; Docker uses a private persistent `/data/media` volume. Production should replace the filesystem adapter with private encrypted object storage using the same `media.Store` interface.

## Android
```bash
cd apps/mobile
npm install
npm start
npm run android
```

A debug APK can be built with:
```bash
npm run android:assemble:debug
```

The current execution environment used to build this checkpoint does not contain Android SDK or npm registry access, so source/native contract verification is automated but an APK is not falsely claimed as produced here.

## Current API
OpenAPI version: **1.6.1**.

See `docs/SPRINT_4C.md` and `docs/RELEASE_READINESS.md` for current release gates, `docs/SPRINT_4B.md` for personal baselines/provenance, `studio/` for the permanent virtual team workflow, and `docs/ROADMAP.md` for the Fitness OS phase.
