-- Collapse multiple source rows to the most recently imported row before restoring legacy uniqueness.
DELETE FROM health_daily_snapshots older
USING health_daily_snapshots newer
WHERE older.user_id = newer.user_id
  AND older.local_date = newer.local_date
  AND (older.imported_at < newer.imported_at OR (older.imported_at = newer.imported_at AND older.id::text < newer.id::text));

ALTER TABLE health_daily_snapshots
  DROP CONSTRAINT IF EXISTS health_daily_snapshots_user_date_source_key;
ALTER TABLE health_daily_snapshots
  ADD CONSTRAINT health_daily_snapshots_user_id_local_date_key UNIQUE(user_id, local_date);
DROP INDEX IF EXISTS idx_health_daily_snapshots_resolution;
