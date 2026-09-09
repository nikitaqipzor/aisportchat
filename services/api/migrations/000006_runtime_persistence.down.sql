DROP INDEX IF EXISTS idx_workouts_user_status_started;
ALTER TABLE workout_exercises DROP COLUMN IF EXISTS progression_note;
ALTER TABLE workout_exercises DROP COLUMN IF EXISTS rest_seconds;
ALTER TABLE workouts DROP COLUMN IF EXISTS duration_minutes;
ALTER TABLE workouts DROP COLUMN IF EXISTS muscle;
