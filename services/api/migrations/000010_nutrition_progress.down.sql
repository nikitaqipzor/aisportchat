DROP TABLE IF EXISTS body_measurements;
DROP TABLE IF EXISTS recipe_items;
DROP TABLE IF EXISTS recipes;
DROP INDEX IF EXISTS idx_food_items_owner;
DROP INDEX IF EXISTS idx_food_items_barcode;
ALTER TABLE food_items DROP COLUMN IF EXISTS owner_user_id;
ALTER TABLE food_items DROP COLUMN IF EXISTS barcode;
