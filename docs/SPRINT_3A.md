# Sprint 3A — Body Scan Foundation

## Goal
Create a privacy-first photo progress pipeline that can later feed pose/body computer vision without coupling raw body photos to an LLM.

## Delivered
- draft/completed Body Scan lifecycle;
- required views: `front`, `side`, `back`;
- Android TurboModule `BodyPhotoCapture` using full-resolution system camera capture;
- EXIF rotation normalization, downsampling and JPEG compression before JS transfer;
- deterministic server-side quality gate: resolution, portrait orientation, brightness, contrast and framing warnings;
- private media storage interface with filesystem runtime adapter and in-memory test adapter;
- authenticated photo retrieval with `Cache-Control: private, no-store`;
- explicit deletion removes both scan metadata and private image blobs;
- PostgreSQL migration `000013_body_scan`;
- Docker persistent media volume `/data/media`;
- capture consistency score between latest completed scans;
- measurement deltas (weight/waist) attached to scan comparison when nearby measurements exist;
- Body Scan mobile screen under Progress with resumeable draft, retake, delete and history;
- OpenAPI `1.0.0`.

## Trust boundary
This sprint does **not** infer body-fat percentage, diagnose posture or claim muscle growth from photographs. `capture_consistency_score` measures only whether two capture sessions are reasonably comparable (lighting/dimensions/quality warnings). Visual body analysis is intentionally deferred to the dedicated CV layer.

## Photo quality rules
Hard reject:
- width < 720 px or height < 960 px;
- landscape/non-portrait capture;
- very dark average luminance;
- severe overexposure.

Warning but accepted:
- unusual portrait framing;
- low contrast.

## Storage
Development runtime uses `MEDIA_ROOT` (default `./data/media`) with restrictive file permissions. Docker mounts a dedicated `media_data` volume. Production should move the same `media.Store` contract to private object storage with managed at-rest encryption/KMS; no body-photo public URLs should be introduced.

## API
- `POST /api/v1/body-scans`
- `GET /api/v1/body-scans`
- `GET /api/v1/body-scans/{scan_id}`
- `DELETE /api/v1/body-scans/{scan_id}`
- `PUT /api/v1/body-scans/{scan_id}/photos/{view}`
- `GET /api/v1/body-scans/{scan_id}/photos/{view}`
- `POST /api/v1/body-scans/{scan_id}/complete`
- `GET /api/v1/body-scans/comparison/latest`

## Tests
- complete scan lifecycle;
- incomplete scan rejection;
- dark image rejection;
- cross-user privacy;
- private blob deletion;
- capture comparison + measurement delta;
- authenticated HTTP upload/download flow;
- existing auth/offline/native regression suite.

## Next
Sprint 3B: local pose landmarks + technique recording foundation. Start with recorded video and 3–5 exercises rather than live inference across the whole exercise catalog.
