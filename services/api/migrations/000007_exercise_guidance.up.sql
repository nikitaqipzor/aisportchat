ALTER TABLE exercises
    ADD COLUMN IF NOT EXISTS common_mistakes JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS technique_tips JSONB NOT NULL DEFAULT '[]'::jsonb;
