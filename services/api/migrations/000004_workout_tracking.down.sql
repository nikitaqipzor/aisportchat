DROP INDEX IF EXISTS idx_workouts_user_muscle_completed;
ALTER TABLE workout_sets
    DROP CONSTRAINT IF EXISTS workout_sets_repetitions_check,
    DROP CONSTRAINT IF EXISTS workout_sets_weight_check,
    DROP CONSTRAINT IF EXISTS workout_sets_rpe_check,
    DROP CONSTRAINT IF EXISTS workout_sets_rir_check;
ALTER TABLE workout_exercises
    DROP COLUMN IF EXISTS progression_note,
    DROP COLUMN IF EXISTS rest_seconds;
ALTER TABLE workouts DROP COLUMN IF EXISTS muscle;
