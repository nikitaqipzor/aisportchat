DROP INDEX IF EXISTS idx_technique_analyses_workout;
ALTER TABLE technique_analyses DROP CONSTRAINT IF EXISTS technique_analyses_set_number_check;
ALTER TABLE technique_analyses DROP CONSTRAINT IF EXISTS technique_analyses_capture_mode_check;
ALTER TABLE technique_analyses
    DROP COLUMN IF EXISTS set_number,
    DROP COLUMN IF EXISTS workout_exercise_id,
    DROP COLUMN IF EXISTS workout_id,
    DROP COLUMN IF EXISTS capture_mode;
