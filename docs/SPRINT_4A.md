# Sprint 4A — Xiaomi Watch S3 / Health Connect

## Goal
Use Android Health Connect as the wearable boundary for AI Fitness OS, with Xiaomi Watch S3 / Mi Fitness as the first real source. Import normalized daily facts into Recovery instead of coupling the backend to Xiaomi Bluetooth/proprietary protocols.

## Architecture
`Xiaomi Watch S3 → Mi Fitness → Health Connect → NativeHealthConnect → normalized daily snapshot → Go API/PostgreSQL → Recovery Engine → AI read-only context`

## Least-permission Android reads
The app requests read access only for:
- steps;
- distance;
- active calories;
- sleep sessions/stages;
- exercise sessions;
- heart rate used inside exercise sessions.

The integration intentionally does not request HRV, SpO2, VO2max or other metrics that the current Xiaomi/Mi Fitness path does not reliably expose to us.

## Data model
`health_daily_snapshots` stores one normalized snapshot per user/local date:
- provider + source package/label;
- steps/distance/active calories;
- sleep total and stages;
- exercise duration/session count;
- exercise HR summary;
- available data-type list;
- capture/import timestamps.

Raw wearable streams are not persisted by the backend in this sprint.

## Duplicate prevention
Cumulative metrics use Health Connect aggregation with a preferred data origin. For Xiaomi, the preferred package is `com.xiaomi.wearable`. Backend upsert uniqueness is `(user_id, local_date)`.

## Recovery integration
Wearable sleep duration overrides manually typed sleep duration when present, while subjective sleep quality, energy, stress and soreness remain user check-in signals. Full workout adaptation still requires the subjective check-in; a wearable alone does not infer how the user feels.

## Mobile UX
`Connected Devices` provides:
- Health Connect availability state;
- Mi Fitness-installed state;
- permission rationale/request;
- Health Connect settings shortcut;
- sync today / sync 7 days;
- last normalized Xiaomi snapshot.

## Privacy
- read-only Health Connect permissions;
- least-permission scope;
- no raw wearable stream upload;
- normalized health facts are owner-scoped in API/storage;
- fitness guidance, not medical diagnosis.

## Remaining real-device gate
This environment does not have Android SDK/device access, so Health Connect permission dialogs and Xiaomi-origin data must still be validated on a physical Android phone with Mi Fitness/Health Connect configured.
