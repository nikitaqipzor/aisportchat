ALTER TABLE food_items ADD COLUMN IF NOT EXISTS barcode TEXT;
ALTER TABLE food_items ADD COLUMN IF NOT EXISTS owner_user_id UUID REFERENCES users(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_food_items_barcode ON food_items(barcode) WHERE barcode IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_food_items_owner ON food_items(owner_user_id) WHERE owner_user_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS recipes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS recipe_items (
    recipe_id UUID NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    food_id TEXT NOT NULL REFERENCES food_items(id),
    position INTEGER NOT NULL DEFAULT 0,
    quantity_g NUMERIC(8,2) NOT NULL CHECK (quantity_g > 0),
    PRIMARY KEY(recipe_id, position)
);
CREATE INDEX IF NOT EXISTS idx_recipes_user_updated ON recipes(user_id, updated_at DESC);

CREATE TABLE IF NOT EXISTS body_measurements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    logged_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    weight_kg NUMERIC(7,2),
    waist_cm NUMERIC(7,2),
    chest_cm NUMERIC(7,2),
    shoulder_cm NUMERIC(7,2),
    arm_cm NUMERIC(7,2),
    forearm_cm NUMERIC(7,2),
    hip_cm NUMERIC(7,2),
    thigh_cm NUMERIC(7,2),
    calf_cm NUMERIC(7,2),
    neck_cm NUMERIC(7,2),
    CHECK (weight_kg IS NOT NULL OR waist_cm IS NOT NULL OR chest_cm IS NOT NULL OR shoulder_cm IS NOT NULL OR arm_cm IS NOT NULL OR forearm_cm IS NOT NULL OR hip_cm IS NOT NULL OR thigh_cm IS NOT NULL OR calf_cm IS NOT NULL OR neck_cm IS NOT NULL)
);
CREATE INDEX IF NOT EXISTS idx_body_measurements_user_logged ON body_measurements(user_id, logged_at DESC);
