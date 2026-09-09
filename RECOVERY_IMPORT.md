# Sprint 4C R2 recovery import

Target branch: `recovery/sprint-4c-r2-full`

Source checkpoint: AI Fitness OS Sprint 4C R2.

Import policy:
- restore application source at repository root;
- include `apps/`, `services/`, `scripts/`, `infra/`, `.github/` and root configuration;
- keep `apps/mobile/android/gradle/wrapper/gradle-wrapper.jar`;
- do **not** commit `apps/mobile/android/app/debug.keystore`;
- omit the non-build-critical `docs/design/audit-preview-concept.png` from the public repository;
- after source expansion, run the repository `release-gate` GitHub Actions workflow and use its Android debug artifact as the first APK evidence.

Local recovery core archive prepared from the checkpoint: `r2-core.zip`.

The existing incomplete recovery branch is not the source of truth. This branch is the clean restart point.
