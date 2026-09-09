# Sprint 3B — Technique CV Foundation

## Status
Implemented as a recorded-video technique-analysis vertical slice. Device APK execution is still pending an Android SDK/npm-capable machine.

## Privacy architecture
Raw exercise video is recorded into Android app cache and analyzed on-device with MediaPipe Pose Landmarker. The mobile app sends only timestamped normalized pose landmarks to the Go API. The temporary video file is deleted after analysis or error cleanup.

## Native Android
New TurboModule: `TechniqueVideo`.

Methods:
- `recordVideo(maxDurationSeconds)` — system camera, temporary MP4, max 90 s;
- `analyzeVideo(filePath, intervalMs)` — MediaPipe VIDEO running mode, one pose, 33 landmarks;
- `deleteVideo(filePath)` — explicit cleanup.

Android dependency is pinned to `com.google.mediapipe:tasks-vision:0.10.29`.
The build task `fetchPoseLandmarkerModel` downloads `pose_landmarker_lite.task` into app assets automatically before `preBuild`. `task/tflite` assets are configured as no-compress.

## Deterministic Technique Engine
`internal/technique` owns the calculations. LLM is not involved.

Initial analyzers:
1. squat;
2. biceps curl;
3. push-up;
4. lunge;
5. shoulder press.

Each analyzer uses exercise-specific joint angles and a state machine. A rep is accepted only after a complete top -> working phase -> bottom -> return cycle.

Output:
- rep count;
- per-rep duration;
- per-rep ROM;
- left/right ROM;
- symmetry delta;
- ROM score;
- tempo score;
- symmetry score;
- stability score;
- overall Technique Score;
- pose confidence;
- bounded feedback.

Technique Score is a training UX metric, not a medical/biomechanical diagnosis.

## API
OpenAPI version: `1.1.0`.

- `GET /api/v1/technique/exercises`
- `POST /api/v1/technique/analyses`
- `GET /api/v1/technique/analyses`
- `GET /api/v1/technique/analyses/{analysis_id}`

Pose payload limit is 6 MB. Server accepts 8..1200 monotonically timestamped frames and rejects low-visibility sequences.

## Persistence
Migration `000014_technique_cv` creates `technique_analyses` and stores algorithm version plus the immutable result JSON. Results are owner-scoped.

## Mobile UX
`Progress -> Technique` exposes:
- exercise choice;
- recording guidance;
- processing stages;
- a pose skeleton preview;
- Technique Score and metric breakdown;
- feedback;
- recent analysis history.

## Tests
- synthetic biceps-curl rep counting;
- low visibility rejection;
- per-user analysis privacy;
- HTTP catalog -> analyze -> persisted history flow;
- complete Go test/vet regression;
- TypeScript syntax gate;
- Android TurboModule/static contract verification;
- auth refresh/offline regression tests.

## Known limitations / next slice
- recorded mode only; no live CameraX overlay yet;
- initial rules are heuristic and need a real labeled exercise-video dataset for calibration;
- no per-exercise camera-angle classifier yet;
- no velocity/eccentric/concentric phase split yet;
- no linking Technique Analysis to a specific workout set yet;
- Android device build/run remains an external-machine gate in this environment.

Next: Sprint 3B.1 calibration + live overlay, then Sprint 3C recovery/readiness.
