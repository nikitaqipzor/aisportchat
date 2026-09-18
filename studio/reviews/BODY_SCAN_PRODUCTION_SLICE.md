# Body Scan production slice

## Goal

Make Body Scan useful and explainable without presenting unvalidated computer
vision or medical conclusions.

## Changes

- Added deterministic, per-view capture comparison metrics for brightness,
  contrast, resolution, and a bounded consistency score.
- Added an explicit capture grade and API status that states no body inference
  is performed.
- Updated OpenAPI and the mobile client contract.
- Improved the mobile loading state, Russian quality guidance, comparison
  breakdown, and deletion of completed scans and their private media.
- Added service coverage for identical and deliberately changed captures.

## Risks

- Resolution consistency is not a proxy for camera distance or body framing.
- Physical camera permission, OEM camera behavior, and deletion should still
  be exercised on supported Android devices.

## Tests / evidence

- `go test ./internal/bodyscan ./internal/httpapi`
- `npm run typecheck` in `apps/mobile`
- OpenAPI YAML parse and Android native verification are part of final QA.

## Unresolved

- Validated pose/body-segmentation is intentionally out of scope. It must not
  be enabled without an evaluated model, consent UX, and documented error
  bounds.

## Handoff

QA should verify two completed scans under similar and dissimilar lighting,
then delete one completed scan and confirm both metadata and blobs disappear.
