CREATE TABLE IF NOT EXISTS nutrition_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    goal TEXT NOT NULL CHECK (goal IN ('lose','recomp','maintain','gain','strength','endurance')),
    activity_level TEXT NOT NULL CHECK (activity_level IN ('low','light','moderate','high','athlete')),
    calculation_mode TEXT NOT NULL CHECK (calculation_mode IN ('auto','manual')),
    calorie_target INTEGER NOT NULL CHECK (calorie_target > 0),
    protein_target_g NUMERIC(8,2) NOT NULL DEFAULT 0,
    fat_target_g NUMERIC(8,2) NOT NULL DEFAULT 0,
    carb_target_g NUMERIC(8,2) NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS food_items (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    brand TEXT,
    kcal_per_100g NUMERIC(8,2) NOT NULL,
    protein_per_100g NUMERIC(8,2) NOT NULL DEFAULT 0,
    fat_per_100g NUMERIC(8,2) NOT NULL DEFAULT 0,
    carbs_per_100g NUMERIC(8,2) NOT NULL DEFAULT 0,
    fiber_per_100g NUMERIC(8,2) NOT NULL DEFAULT 0,
    serving_g NUMERIC(8,2) NOT NULL DEFAULT 100,
    source TEXT NOT NULL DEFAULT 'seed',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS food_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    food_id TEXT NOT NULL REFERENCES food_items(id),
    food_name TEXT NOT NULL,
    meal_type TEXT NOT NULL CHECK (meal_type IN ('breakfast','lunch','dinner','snack')),
    logged_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    quantity_g NUMERIC(8,2) NOT NULL CHECK (quantity_g > 0),
    calories NUMERIC(9,2) NOT NULL DEFAULT 0,
    protein_g NUMERIC(8,2) NOT NULL DEFAULT 0,
    fat_g NUMERIC(8,2) NOT NULL DEFAULT 0,
    carbs_g NUMERIC(8,2) NOT NULL DEFAULT 0,
    fiber_g NUMERIC(8,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_food_items_name ON food_items(lower(name));
CREATE INDEX IF NOT EXISTS idx_food_entries_user_logged ON food_entries(user_id, logged_at DESC);

INSERT INTO food_items(id,name,kcal_per_100g,protein_per_100g,fat_per_100g,carbs_per_100g,fiber_per_100g,serving_g,source)
VALUES
('chicken-breast','Куриная грудка',165,31,3.6,0,0,150,'seed'),
('turkey-breast','Филе индейки',135,29,1.6,0,0,150,'seed'),
('salmon','Лосось',208,20,13,0,0,150,'seed'),
('egg','Яйцо куриное',143,12.6,9.5,0.7,0,55,'seed'),
('cottage-cheese-5','Творог 5%',121,17,5,1.8,0,200,'seed'),
('greek-yogurt','Йогурт греческий 2%',73,9.9,2,3.9,0,200,'seed'),
('oats-dry','Овсяные хлопья сухие',379,13.2,6.5,67.7,10.1,80,'seed'),
('rice-cooked','Рис белый варёный',130,2.7,0.3,28.2,0.4,200,'seed'),
('buckwheat-cooked','Гречка варёная',110,4.2,1.1,21.3,2.7,200,'seed'),
('pasta-cooked','Макароны варёные',158,5.8,0.9,30.9,1.8,200,'seed'),
('potato-boiled','Картофель варёный',87,1.9,0.1,20.1,1.8,250,'seed'),
('banana','Банан',89,1.1,0.3,22.8,2.6,120,'seed'),
('apple','Яблоко',52,0.3,0.2,13.8,2.4,180,'seed'),
('avocado','Авокадо',160,2,14.7,8.5,6.7,100,'seed'),
('olive-oil','Оливковое масло',884,0,100,0,0,10,'seed'),
('almonds','Миндаль',579,21.2,49.9,21.6,12.5,30,'seed'),
('wholegrain-bread','Хлеб цельнозерновой',247,13,4.2,41,7,60,'seed'),
('milk-2','Молоко 2%',50,3.3,2,4.8,0,250,'seed'),
('whey-protein','Сывороточный протеин',390,78,6,8,0,30,'seed'),
('syrniki','Сырники',220,14,10,18,0.8,200,'seed')
ON CONFLICT (id) DO UPDATE SET
    name=EXCLUDED.name,
    kcal_per_100g=EXCLUDED.kcal_per_100g,
    protein_per_100g=EXCLUDED.protein_per_100g,
    fat_per_100g=EXCLUDED.fat_per_100g,
    carbs_per_100g=EXCLUDED.carbs_per_100g,
    fiber_per_100g=EXCLUDED.fiber_per_100g,
    serving_g=EXCLUDED.serving_g,
    source=EXCLUDED.source,
    active=true,
    updated_at=now();
