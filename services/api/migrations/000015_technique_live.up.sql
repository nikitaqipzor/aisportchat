ALTER TABLE technique_analyses
    ADD COLUMN IF NOT EXISTS capture_mode text NOT NULL DEFAULT 'recorded',
    ADD COLUMN IF NOT EXISTS workout_id uuid REFERENCES workouts(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS workout_exercise_id uuid REFERENCES workout_exercises(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS set_number integer;

ALTER TABLE technique_analyses
    DROP CONSTRAINT IF EXISTS technique_analyses_capture_mode_check;
ALTER TABLE technique_analyses
    ADD CONSTRAINT technique_analyses_capture_mode_check CHECK (capture_mode IN ('recorded','live'));

ALTER TABLE technique_analyses
    DROP CONSTRAINT IF EXISTS technique_analyses_set_number_check;
ALTER TABLE technique_analyses
    ADD CONSTRAINT technique_analyses_set_number_check CHECK (set_number IS NULL OR set_number BETWEEN 1 AND 100);

CREATE INDEX IF NOT EXISTS idx_technique_analyses_workout
    ON technique_analyses(user_id, workout_id, created_at DESC)
    WHERE workout_id IS NOT NULL;
