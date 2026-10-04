# Scan, technique and AI repair handoff

## Goal
Make saved results retrievable and errors actionable across Body Scan, Technique and AI Coach.

## Changes
- Body Scan history expands to show per-view capture quality and loads selected private images through the authenticated binary endpoint. Draft scans can show persisted images after reopening; image data is held only in component memory and dropped when the screen unmounts.
- Comparison failures show an error and retry rather than silently looking like missing results.
- Technique history opens stored analysis and full recommendations. Historical results cannot be applied to an unrelated currently active set.
- AI status errors show a retry action rather than a permanent connecting indicator.
- The app passes actual server-save state to the workout summary so an offline queued finish is represented truthfully.

## Evidence
- TypeScript typecheck and all mobile Node scripts pass in the integrated workspace.
- The Android native configuration verifier passes.
- Existing Body Scan and Technique scripts are source-level contracts; authenticated binary photo display still needs a device run with a working API.

## Risks and unresolved items
- FileReader/Blob decoding of authenticated photo bytes and Android image rendering have not been exercised on a physical phone.
- No local Go runtime or Android device is available here; backend and device E2E must run in CI and on a test handset.

## Handoff
QA: verify draft and history photos, changing users, token refresh during image load, and offline image errors on an installed APK. Lead: review CI and release acceptance.
