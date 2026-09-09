CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    birth_date DATE,
    gender TEXT,
    height_cm NUMERIC(5,2),
    weight_kg NUMERIC(6,2),
    experience_level TEXT,
    unit_system TEXT NOT NULL DEFAULT 'metric',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE muscles (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    body_region TEXT NOT NULL
);

CREATE TABLE equipment (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    category TEXT NOT NULL
);

CREATE TABLE exercises (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    difficulty TEXT NOT NULL,
    movement_pattern TEXT NOT NULL,
    exercise_type TEXT NOT NULL,
    compound BOOLEAN NOT NULL DEFAULT false,
    instructions JSONB NOT NULL DEFAULT '[]'::jsonb,
    video_url TEXT,
    thumbnail_url TEXT,
    status TEXT NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE exercise_muscles (
    exercise_id TEXT NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    muscle_id TEXT NOT NULL REFERENCES muscles(id) ON DELETE CASCADE,
    role TEXT NOT NULL CHECK (role IN ('primary', 'secondary')),
    contribution_weight NUMERIC(4,3),
    PRIMARY KEY (exercise_id, muscle_id, role)
);

CREATE TABLE exercise_equipment (
    exercise_id TEXT NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    equipment_id TEXT NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    PRIMARY KEY (exercise_id, equipment_id)
);

CREATE TABLE exercise_environments (
    exercise_id TEXT NOT NULL REFERENCES exercises(id) ON DELETE CASCADE,
    environment TEXT NOT NULL CHECK (environment IN ('home', 'gym', 'band')),
    PRIMARY KEY (exercise_id, environment)
);

CREATE TABLE workouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL CHECK (status IN ('planned', 'active', 'completed', 'cancelled')),
    environment TEXT NOT NULL,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_seconds INTEGER,
    total_volume NUMERIC(12,2),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE workout_exercises (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_id UUID NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    exercise_id TEXT NOT NULL REFERENCES exercises(id),
    position INTEGER NOT NULL,
    target_sets INTEGER NOT NULL,
    target_reps_min INTEGER,
    target_reps_max INTEGER,
    target_weight NUMERIC(8,2),
    UNIQUE (workout_id, position)
);

CREATE TABLE workout_sets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workout_exercise_id UUID NOT NULL REFERENCES workout_exercises(id) ON DELETE CASCADE,
    set_number INTEGER NOT NULL,
    weight NUMERIC(8,2),
    repetitions INTEGER,
    rpe NUMERIC(3,1),
    rir NUMERIC(3,1),
    completed_at TIMESTAMPTZ,
    UNIQUE (workout_exercise_id, set_number)
);

CREATE INDEX idx_workouts_user_created ON workouts(user_id, created_at DESC);
CREATE INDEX idx_workout_exercises_workout ON workout_exercises(workout_id, position);
