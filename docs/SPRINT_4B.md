# Sprint 4B — Personal Baselines & Health Provenance

## Goal
Turn Health Connect ingestion from a one-day wearable card into trustworthy personal context: compare the user to their own recent history, show where data came from, reject stale/empty inputs, and let Recovery use wearable sleep without blindly trusting it.

## Lead Engineer decisions
- Personal baselines use previous days only; the target/current day is excluded from its own baseline.
- Baselines are fitness trend context, not medical norms.
- Multiple Health Connect sources are stored separately. Resolution is deterministic; Xiaomi/Mi Fitness is the current preferred source while Xiaomi Watch S3 is the connected device.
- Empty Health Connect days do not count toward coverage/confidence.
- Stale wearable sleep is not allowed to silently override a fresh subjective check-in.
- Activity baselines are explanatory in 4B; they are not double-counted into readiness because workout-load already contributes to readiness.

## Backend
### Personal health insights
New `GET /api/v1/health/insights?date=YYYY-MM-DD` returns:
- 7-day baseline;
- 28-day baseline;
- available-day coverage;
- metric sample counts;
- current-vs-baseline deviations;
- sync freshness (`fresh`, `aging`, `stale`, `missing`);
- per-metric provenance;
- source candidates and selected source;
- confidence 0–100 + low/medium/high label;
- human-readable reasons.

Baselines currently cover:
- sleep minutes;
- steps;
- active calories;
- exercise minutes;
- resting HR when a provider exposes it.

### Multi-source resolver
Migration `000018_health_baselines_provenance` changes uniqueness from:
`user + date`

to:
`user + date + source_package`.

This prevents last-write-wins corruption when multiple Health Connect sources exist on the same day. The resolved daily view selects the configured priority policy (currently Mi Fitness first, then newest fallback source).

### Recovery integration
Readiness receives `health_insights` and uses the personal 28-day sleep baseline once enough history exists. The personal baseline influences sleep scoring, while an absolute sleep floor remains so a chronically short personal history is not treated as an ideal target.

If today's wearable snapshot is stale, Recovery falls back to manual check-in sleep and explains why.

### AI Coach
`get_health_insights` is a new read-only tool. AI can explain personal sleep/activity trends, freshness and provenance but cannot modify health facts or readiness coefficients.

## Mobile UX
Connected Devices now shows:
- 1/7/28-day sync actions;
- real progress during 28-day sync;
- skipped empty days;
- data confidence;
- 28-day coverage;
- last sync freshness;
- personal sleep/steps/calorie/exercise averages;
- today's deviation vs baseline;
- all detected sources and the selected source;
- metric provenance.

Recovery now shows:
- wearable freshness;
- 28-day personal sleep baseline;
- baseline sample-day count;
- wearable confidence;
- today's sleep deviation from personal baseline.

## QA cases added
- current day excluded from baseline;
- empty days do not inflate coverage;
- impossible/empty imports rejected;
- multi-source day keeps both rows and resolves Xiaomi deterministically;
- source conflict is visible in insights;
- stale wearable sleep falls back to manual check-in;
- cross-user health isolation from Sprint 4A remains green;
- full legacy regression gate remains green.

## Known device gate
The code path is ready, but actual Mi Fitness records must still be validated on a physical Android phone with Xiaomi Watch S3 + Health Connect permissions. No device-read claim is made from this container.
