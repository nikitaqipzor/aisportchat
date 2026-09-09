ALTER TABLE workouts
    ADD COLUMN IF NOT EXISTS muscle TEXT;

UPDATE workouts w
SET muscle = inferred.muscle_id
FROM (
    SELECT DISTINCT ON (we.workout_id) we.workout_id, em.muscle_id
    FROM workout_exercises we
    JOIN exercise_muscles em ON em.exercise_id = we.exercise_id AND em.role = 'primary'
    ORDER BY we.workout_id, we.position
) inferred
WHERE w.id = inferred.workout_id AND w.muscle IS NULL;

ALTER TABLE workout_exercises
    ADD COLUMN IF NOT EXISTS rest_seconds SMALLINT NOT NULL DEFAULT 90,
    ADD COLUMN IF NOT EXISTS progression_note TEXT NOT NULL DEFAULT '';

ALTER TABLE workout_sets
    DROP CONSTRAINT IF EXISTS workout_sets_repetitions_check,
    ADD CONSTRAINT workout_sets_repetitions_check CHECK (repetitions IS NULL OR repetitions BETWEEN 0 AND 200),
    DROP CONSTRAINT IF EXISTS workout_sets_weight_check,
    ADD CONSTRAINT workout_sets_weight_check CHECK (weight IS NULL OR weight BETWEEN 0 AND 1000),
    DROP CONSTRAINT IF EXISTS workout_sets_rpe_check,
    ADD CONSTRAINT workout_sets_rpe_check CHECK (rpe IS NULL OR rpe BETWEEN 1 AND 10),
    DROP CONSTRAINT IF EXISTS workout_sets_rir_check,
    ADD CONSTRAINT workout_sets_rir_check CHECK (rir IS NULL OR rir BETWEEN 0 AND 10);

CREATE INDEX IF NOT EXISTS idx_workouts_user_muscle_completed
    ON workouts(user_id, muscle, environment, completed_at DESC)
    WHERE status = 'completed';
