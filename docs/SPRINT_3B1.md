# Sprint 3B.1 — Live Technique + Workout Set Linking

## Status
Source complete and regression-tested. Real Android CameraX/MediaPipe execution still requires an Android SDK + device/emulator outside this container.

## Live Android pipeline
A new TurboModule `TechniqueLive` opens a private native `TechniqueLiveActivity`.

Pipeline:

```text
CameraX PreviewView
  -> ImageAnalysis RGBA frames
  -> on-device MediaPipe Pose Landmarker
  -> skeleton overlay + framing guidance + provisional live rep counter
  -> 5 Hz bounded landmark samples
  -> React Native
  -> Go Technique Engine (authoritative recount/scoring)
```

The live activity never uploads raw video. It writes a temporary JSON result to app cache containing sampled landmarks, then the TurboModule reads it and deletes the temporary file.

CameraX is pinned to stable `1.5.3` (`core`, `camera2`, `lifecycle`, `view`).

## Live UX
The native camera screen provides:
- real CameraX preview;
- skeleton overlay;
- provisional live repetition count;
- elapsed time;
- static exercise-specific camera-angle guidance;
- dynamic prompts for low light, low pose visibility, edge clipping, or excessive distance;
- front/back camera switching;
- 60 second hard session cap;
- bounded 320-frame sampled pose payload.

The live count is deliberately provisional. The Go Technique Engine recomputes reps from the complete sampled timeline after the live session ends.

## Workout-set link
Technique analysis can now include:
- `workout_id`;
- `workout_exercise_id`;
- `set_number`;
- `capture_mode = live|recorded`.

The backend verifies:
- the workout belongs to the authenticated user;
- the workout is active;
- the workout exercise is part of that workout;
- the requested set number is within planned sets;
- the selected technique analyzer matches the actual catalog exercise.

From Active Workout, supported exercises expose **LIVE · считать повторы камерой**. After analysis, the user can use the authoritative rep count to prefill the current set without manually retyping it.

## Exercise mapping
Initial workout exercises mapped into the five calibrated analyzers:
- any supported squat variation -> `squat`;
- push-up -> `push_up`;
- curl variations -> `biceps_curl`;
- split squat/lunge variations -> `lunge`;
- overhead/shoulder press variations -> `shoulder_press`.

## Phase timing
Technique Engine version: `technique-v1.1`.

Each completed rep now stores:
- total duration;
- eccentric duration;
- concentric duration;
- ROM;
- left/right ROM;
- symmetry delta.

The result includes average eccentric/concentric timing and average repetitions per minute.

The state machine now supports both movement start conventions:
- `top` start: squat, curl, push-up, lunge;
- `bottom` start: shoulder press.

This fixes first-repetition undercounting for overhead pressing.

## Persistence
Migration `000015_technique_live` adds:
- capture mode;
- optional workout reference;
- optional workout-exercise reference;
- optional set number;
- workout-scoped index.

## API
OpenAPI version: `1.2.0`.

`POST /api/v1/technique/analyses` accepts the new optional workout link and capture mode. The response includes linkage plus phase timing metrics.

## Validation/tests
- all Go unit/integration tests;
- linked live workout-set service test;
- mismatched exercise/workout rejection;
- shoulder-press bottom-start regression;
- Technique HTTP link test;
- OpenAPI/Docker YAML parse;
- 51 TS/TSX syntax gate;
- 87 Android native static checks;
- auth refresh regression;
- cross-user offline-session isolation.

## Device gate still open
This execution environment still lacks Android SDK/Platform 37 and cannot install npm dependencies. Therefore CameraX/MediaPipe code is source-verified rather than device-compiled here. Real-device calibration remains mandatory before production scoring.

## Next
Sprint 3C: Recovery + Readiness, while real-device Technique calibration can continue in parallel once an Android build machine is available.
