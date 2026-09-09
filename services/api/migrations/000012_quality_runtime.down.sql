ALTER TABLE workouts
  DROP COLUMN IF EXISTS ended_early,
  DROP COLUMN IF EXISTS completion_percent;
