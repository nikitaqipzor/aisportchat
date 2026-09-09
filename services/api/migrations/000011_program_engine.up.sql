CREATE TABLE IF NOT EXISTS programs (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title text NOT NULL,
  goal_type text NOT NULL,
  weeks smallint NOT NULL CHECK (weeks IN (4,8,12)),
  workouts_per_week smallint NOT NULL CHECK (workouts_per_week BETWEEN 1 AND 7),
  environment text NOT NULL CHECK (environment IN ('home','gym','band')),
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('active','completed','archived')),
  start_date date NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_active_program_per_user ON programs(user_id) WHERE status='active';
CREATE INDEX IF NOT EXISTS idx_programs_user_updated ON programs(user_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS program_sessions (
  id uuid PRIMARY KEY,
  program_id uuid NOT NULL REFERENCES programs(id) ON DELETE CASCADE,
  week_number smallint NOT NULL,
  day_index smallint NOT NULL,
  planned_date date NOT NULL,
  original_date date NOT NULL,
  muscle text NOT NULL,
  secondary_muscle text,
  environment text NOT NULL CHECK (environment IN ('home','gym','band')),
  status text NOT NULL DEFAULT 'planned' CHECK (status IN ('planned','rescheduled','missed','completed','skipped')),
  workout_id uuid REFERENCES workouts(id) ON DELETE SET NULL,
  completed_at timestamptz,
  is_deload boolean NOT NULL DEFAULT false,
  volume_multiplier numeric(5,2) NOT NULL DEFAULT 1.00,
  intensity_multiplier numeric(5,2) NOT NULL DEFAULT 1.00,
  planned_sets smallint NOT NULL DEFAULT 8,
  adaptation_reason text NOT NULL DEFAULT ''
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_program_session_slot ON program_sessions(program_id, week_number, day_index);
CREATE UNIQUE INDEX IF NOT EXISTS uq_program_session_workout ON program_sessions(workout_id) WHERE workout_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_program_sessions_program_date ON program_sessions(program_id, planned_date);
