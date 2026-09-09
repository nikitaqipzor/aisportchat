CREATE TABLE technique_analyses (
 id uuid PRIMARY KEY, user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE, exercise_key text NOT NULL, rep_count integer NOT NULL DEFAULT 0, technique_score integer NOT NULL DEFAULT 0, rom_score integer NOT NULL DEFAULT 0, tempo_score integer NOT NULL DEFAULT 0, symmetry_score integer NOT NULL DEFAULT 0, stability_score integer NOT NULL DEFAULT 0, confidence numeric NOT NULL DEFAULT 0, duration_ms bigint NOT NULL DEFAULT 0, algorithm_version text NOT NULL, result_json jsonb NOT NULL DEFAULT '{}'::jsonb, created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_technique_analyses_user_created ON technique_analyses(user_id,created_at DESC);
