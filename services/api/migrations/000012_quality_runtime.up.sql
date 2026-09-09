ALTER TABLE workouts
  ADD COLUMN IF NOT EXISTS completion_percent numeric(5,2) NOT NULL DEFAULT 0,
  ADD COLUMN IF NOT EXISTS ended_early boolean NOT NULL DEFAULT false;
