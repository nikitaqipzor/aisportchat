# Sprint 4A task board — Xiaomi Watch S3 / Health Connect

## R&D — done
- Confirm Health Connect as Android aggregation boundary.
- Confirm Mi Fitness package identifier used as preferred data origin.
- Keep unsupported Xiaomi metrics out of requested permissions.

## Lead Engineer — done
- Define least-permission read-only provider.
- Persist normalized daily snapshots instead of raw wearable streams.
- Recovery consumes wearable sleep but still requires subjective check-in for full adaptation.
- Prevent cumulative-metric double counting by using Health Connect aggregation and a preferred origin.

## Coder — done
- Health Connect TurboModule + rationale activity.
- Daily health snapshot API/store/migration.
- Connected Devices UI and 1/7-day sync.
- Recovery wearable context.

## QA — done
- user isolation;
- impossible snapshot validation;
- wearable sleep precedence;
- HTTP import → recovery;
- mobile/native regression.

## Designer — done
- explain permissions before system dialog;
- show source label and last imported metrics;
- make missing Health Connect / missing Mi Fitness states actionable;
- avoid presenting wearable data as medical diagnosis.
