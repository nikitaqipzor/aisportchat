ALTER TABLE workouts
    ADD COLUMN IF NOT EXISTS is_favorite BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE IF NOT EXISTS personal_records (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    workout_id UUID NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    workout_set_id UUID REFERENCES workout_sets(id) ON DELETE SET NULL,
    exercise_id TEXT NOT NULL REFERENCES exercises(id) ON DELETE RESTRICT,
    record_type TEXT NOT NULL CHECK (record_type IN ('max_weight', 'max_reps', 'estimated_1rm')),
    value NUMERIC(12,3) NOT NULL CHECK (value >= 0),
    previous_value NUMERIC(12,3),
    achieved_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_personal_records_user_recent
    ON personal_records(user_id, achieved_at DESC);

CREATE INDEX IF NOT EXISTS idx_personal_records_best
    ON personal_records(user_id, exercise_id, record_type, value DESC, achieved_at DESC);

CREATE INDEX IF NOT EXISTS idx_personal_records_workout
    ON personal_records(workout_id, achieved_at DESC);

CREATE INDEX IF NOT EXISTS idx_workouts_user_favorite
    ON workouts(user_id, is_favorite, created_at DESC)
    WHERE is_favorite = TRUE;
