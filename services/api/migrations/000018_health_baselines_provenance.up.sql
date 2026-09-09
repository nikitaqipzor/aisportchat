-- Keep one normalized daily snapshot per source instead of last-write-wins per day.
ALTER TABLE health_daily_snapshots
  DROP CONSTRAINT IF EXISTS health_daily_snapshots_user_id_local_date_key;

ALTER TABLE health_daily_snapshots
  ADD CONSTRAINT health_daily_snapshots_user_date_source_key
  UNIQUE(user_id, local_date, source_package);

CREATE INDEX IF NOT EXISTS idx_health_daily_snapshots_resolution
  ON health_daily_snapshots(user_id, local_date DESC, source_package, imported_at DESC);
