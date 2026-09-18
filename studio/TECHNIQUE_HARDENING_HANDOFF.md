# Technique hardening handoff

## Goal
Make Technique capture safer and less error-prone before real-device calibration, without presenting heuristic feedback as medical assessment.

## Changes
- Added an explicit three-item camera preflight before live or recorded capture.
- Distinguished user cancellation, camera-permission denial, and real capture failures in mobile UX.
- Added native error codes for permission denial and back-button cancellation.
- Expanded catalog aliases that reuse the five existing calibrated movement patterns.
- Removed the unsafe broad `curl` match that could map leg/wrist curls to biceps analysis.
- Kept raw video local/temporary and retained the non-medical result disclaimer.

## Risks
- The five deterministic movement rules still require calibration on multiple physical Android devices, camera angles, body proportions, and lighting conditions.
- Alias support does not mean a new biomechanical model; variants share an existing pattern rule.
- MediaPipe/CameraX native execution cannot be fully proven by source tests alone.

## Tests/evidence
- `node scripts/test-mobile-technique-mapping.mjs`
- `go test ./internal/technique`
- `python3 scripts/verify_android_native.py`
- mobile TypeScript typecheck

## Unresolved
- Run a physical-device calibration matrix and tune thresholds only from captured landmark datasets.
- Verify Android permission denial, permanent denial, back cancellation, camera switching, and process recreation on device.

## Handoff
QA should execute the calibration matrix on at least two Android devices before raising Technique readiness above beta.
