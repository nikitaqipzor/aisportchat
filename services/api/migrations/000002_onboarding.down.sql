ALTER TABLE exercises
    DROP COLUMN IF EXISTS rest_seconds,
    DROP COLUMN IF EXISTS rep_max,
    DROP COLUMN IF EXISTS rep_min,
    DROP COLUMN IF EXISTS default_sets;
DROP TABLE IF EXISTS auth_refresh_sessions;
DROP TABLE IF EXISTS user_equipment;
DROP TABLE IF EXISTS user_training_environments;
DROP TABLE IF EXISTS user_training_preferences;
DROP TABLE IF EXISTS user_goals;
ALTER TABLE user_profiles DROP COLUMN IF EXISTS onboarding_completed_at;
