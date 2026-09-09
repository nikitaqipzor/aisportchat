CREATE TABLE IF NOT EXISTS recovery_checkins (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  local_date DATE NOT NULL,
  sleep_hours NUMERIC(4,1) NOT NULL,
  sleep_quality SMALLINT NOT NULL CHECK (sleep_quality BETWEEN 1 AND 5),
  energy SMALLINT NOT NULL CHECK (energy BETWEEN 1 AND 5),
  stress SMALLINT NOT NULL CHECK (stress BETWEEN 1 AND 5),
  muscle_soreness JSONB NOT NULL DEFAULT '{}'::jsonb,
  notes TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(user_id, local_date)
);
CREATE INDEX IF NOT EXISTS idx_recovery_checkins_user_date ON recovery_checkins(user_id, local_date DESC);
