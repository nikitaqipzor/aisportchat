ALTER TABLE user_profiles
  DROP COLUMN IF EXISTS limitations,
  DROP COLUMN IF EXISTS injuries,
  DROP COLUMN IF EXISTS age_years;
