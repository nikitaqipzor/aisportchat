CREATE TABLE IF NOT EXISTS health_daily_snapshots (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  local_date DATE NOT NULL,
  provider TEXT NOT NULL DEFAULT 'health_connect',
  source_package TEXT NOT NULL,
  source_label TEXT NOT NULL DEFAULT '',
  steps BIGINT NOT NULL DEFAULT 0 CHECK (steps >= 0),
  distance_m NUMERIC(12,2) NOT NULL DEFAULT 0 CHECK (distance_m >= 0),
  active_calories_kcal NUMERIC(10,2) NOT NULL DEFAULT 0 CHECK (active_calories_kcal >= 0),
  sleep_minutes INTEGER NOT NULL DEFAULT 0 CHECK (sleep_minutes BETWEEN 0 AND 1440),
  deep_sleep_minutes INTEGER NOT NULL DEFAULT 0 CHECK (deep_sleep_minutes BETWEEN 0 AND 1440),
  light_sleep_minutes INTEGER NOT NULL DEFAULT 0 CHECK (light_sleep_minutes BETWEEN 0 AND 1440),
  rem_sleep_minutes INTEGER NOT NULL DEFAULT 0 CHECK (rem_sleep_minutes BETWEEN 0 AND 1440),
  awake_minutes INTEGER NOT NULL DEFAULT 0 CHECK (awake_minutes BETWEEN 0 AND 1440),
  exercise_minutes INTEGER NOT NULL DEFAULT 0 CHECK (exercise_minutes BETWEEN 0 AND 1440),
  exercise_sessions INTEGER NOT NULL DEFAULT 0 CHECK (exercise_sessions >= 0),
  exercise_heart_rate_avg NUMERIC(6,2),
  exercise_heart_rate_max NUMERIC(6,2),
  resting_heart_rate NUMERIC(6,2),
  data_types JSONB NOT NULL DEFAULT '[]'::jsonb,
  captured_at TIMESTAMPTZ NOT NULL,
  imported_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(user_id, local_date)
);
CREATE INDEX IF NOT EXISTS idx_health_daily_snapshots_user_date ON health_daily_snapshots(user_id, local_date DESC);
CREATE INDEX IF NOT EXISTS idx_health_daily_snapshots_origin ON health_daily_snapshots(user_id, source_package, local_date DESC);
