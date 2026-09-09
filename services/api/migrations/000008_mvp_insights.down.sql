DROP INDEX IF EXISTS idx_workouts_user_favorite;
DROP TABLE IF EXISTS personal_records;
ALTER TABLE workouts DROP COLUMN IF EXISTS is_favorite;
