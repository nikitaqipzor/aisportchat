ALTER TABLE user_profiles
    ADD COLUMN IF NOT EXISTS onboarding_completed_at TIMESTAMPTZ;

CREATE TABLE user_goals (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    goal_type TEXT NOT NULL CHECK (goal_type IN ('muscle_gain','fat_loss','recomposition','strength','maintenance','endurance')),
    target_weight_kg NUMERIC(6,2),
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_training_preferences (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    workouts_per_week SMALLINT NOT NULL CHECK (workouts_per_week BETWEEN 1 AND 14),
    session_minutes SMALLINT NOT NULL CHECK (session_minutes BETWEEN 10 AND 180),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_training_environments (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    environment TEXT NOT NULL CHECK (environment IN ('home','gym','band')),
    PRIMARY KEY (user_id, environment)
);

CREATE TABLE user_equipment (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    equipment_id TEXT NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, equipment_id)
);

CREATE TABLE auth_refresh_sessions (
    token_hash TEXT PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_refresh_sessions_user ON auth_refresh_sessions(user_id, created_at DESC);
CREATE INDEX idx_refresh_sessions_expiry ON auth_refresh_sessions(expires_at) WHERE revoked_at IS NULL;

ALTER TABLE exercises
    ADD COLUMN IF NOT EXISTS default_sets SMALLINT NOT NULL DEFAULT 3,
    ADD COLUMN IF NOT EXISTS rep_min SMALLINT NOT NULL DEFAULT 8,
    ADD COLUMN IF NOT EXISTS rep_max SMALLINT NOT NULL DEFAULT 12,
    ADD COLUMN IF NOT EXISTS rest_seconds SMALLINT NOT NULL DEFAULT 90;
