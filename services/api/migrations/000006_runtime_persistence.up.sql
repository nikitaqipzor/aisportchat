ALTER TABLE workouts
    ADD COLUMN IF NOT EXISTS muscle TEXT NOT NULL DEFAULT 'full_body',
    ADD COLUMN IF NOT EXISTS duration_minutes INTEGER NOT NULL DEFAULT 45;

ALTER TABLE workout_exercises
    ADD COLUMN IF NOT EXISTS rest_seconds INTEGER NOT NULL DEFAULT 90,
    ADD COLUMN IF NOT EXISTS progression_note TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_workouts_user_status_started
    ON workouts(user_id, status, started_at DESC);
