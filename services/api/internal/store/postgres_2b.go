//go:build cgo

package store

import (
	"context"
	"strconv"
	"strings"
	"time"
)

func (p *Postgres) FindFoodByBarcode(ctx context.Context, userID, barcode string) (FoodItem, error) {
	rows, err := p.query(ctx, `SELECT id,name,COALESCE(brand,''),COALESCE(barcode,''),COALESCE(owner_user_id::text,''),kcal_per_100g::text,protein_per_100g::text,fat_per_100g::text,carbs_per_100g::text,fiber_per_100g::text,serving_g::text,source FROM food_items WHERE active=true AND barcode=$1 AND (owner_user_id IS NULL OR owner_user_id=$2::uuid) ORDER BY owner_user_id NULLS LAST LIMIT 1`, sp(strings.TrimSpace(barcode)), sp(userID))
	if err != nil {
		return FoodItem{}, err
	}
	if len(rows) == 0 {
		return FoodItem{}, ErrNotFound
	}
	return scanFoodItem(rows[0])
}

func (p *Postgres) CreateCustomFood(ctx context.Context, userID string, food FoodItem) (FoodItem, error) {
	if food.ID == "" {
		food.ID = newID()
	}
	rows, err := p.query(ctx, `INSERT INTO food_items(id,name,brand,barcode,owner_user_id,kcal_per_100g,protein_per_100g,fat_per_100g,carbs_per_100g,fiber_per_100g,serving_g,source,active) VALUES($1,$2,NULLIF($3,''),NULLIF($4,''),$5::uuid,$6::numeric,$7::numeric,$8::numeric,$9::numeric,$10::numeric,$11::numeric,'custom',true) RETURNING id,name,COALESCE(brand,''),COALESCE(barcode,''),COALESCE(owner_user_id::text,''),kcal_per_100g::text,protein_per_100g::text,fat_per_100g::text,carbs_per_100g::text,fiber_per_100g::text,serving_g::text,source`,
		sp(food.ID), sp(food.Name), sp(food.Brand), sp(food.Barcode), sp(userID),
		sp(strconv.FormatFloat(food.Kcal100, 'f', -1, 64)), sp(strconv.FormatFloat(food.Protein100, 'f', -1, 64)), sp(strconv.FormatFloat(food.Fat100, 'f', -1, 64)), sp(strconv.FormatFloat(food.Carbs100, 'f', -1, 64)), sp(strconv.FormatFloat(food.Fiber100, 'f', -1, 64)), sp(strconv.FormatFloat(food.ServingG, 'f', -1, 64)))
	if err != nil {
		return FoodItem{}, err
	}
	return scanFoodItem(rows[0])
}

func (p *Postgres) CreateRecipe(ctx context.Context, recipe Recipe) (Recipe, error) {
	if recipe.ID == "" {
		recipe.ID = newID()
	}
	now := time.Now().UTC()
	err := p.withTx(ctx, func() error {
		if _, err := p.queryLocked(ctx, `INSERT INTO recipes(id,user_id,name,created_at,updated_at) VALUES($1::uuid,$2::uuid,$3,$4::timestamptz,$4::timestamptz)`, sp(recipe.ID), sp(recipe.UserID), sp(recipe.Name), sp(now.Format(time.RFC3339Nano))); err != nil {
			return err
		}
		for i, item := range recipe.Items {
			if _, err := p.queryLocked(ctx, `INSERT INTO recipe_items(recipe_id,food_id,position,quantity_g) VALUES($1::uuid,$2,$3::int,$4::numeric)`, sp(recipe.ID), sp(item.FoodID), sp(strconv.Itoa(i)), sp(strconv.FormatFloat(item.QuantityG, 'f', -1, 64))); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return Recipe{}, err
	}
	return p.GetRecipe(ctx, recipe.UserID, recipe.ID)
}

func (p *Postgres) ListRecipes(ctx context.Context, userID string) ([]Recipe, error) {
	rows, err := p.query(ctx, `SELECT id::text FROM recipes WHERE user_id=$1::uuid ORDER BY updated_at DESC`, sp(userID))
	if err != nil {
		return nil, err
	}
	out := make([]Recipe, 0, len(rows))
	for _, row := range rows {
		r, err := p.GetRecipe(ctx, userID, val(row, 0))
		if err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, nil
}

func (p *Postgres) GetRecipe(ctx context.Context, userID, recipeID string) (Recipe, error) {
	rows, err := p.query(ctx, `SELECT id::text,user_id::text,name,created_at::text,updated_at::text FROM recipes WHERE id=$1::uuid AND user_id=$2::uuid LIMIT 1`, sp(recipeID), sp(userID))
	if err != nil {
		return Recipe{}, err
	}
	if len(rows) == 0 {
		return Recipe{}, ErrNotFound
	}
	created, _ := parseTime(val(rows[0], 3))
	updated, _ := parseTime(val(rows[0], 4))
	recipe := Recipe{ID: val(rows[0], 0), UserID: val(rows[0], 1), Name: val(rows[0], 2), CreatedAt: created, UpdatedAt: updated}
	itemRows, err := p.query(ctx, `SELECT ri.food_id,f.name,ri.quantity_g::text,f.kcal_per_100g::text,f.protein_per_100g::text,f.fat_per_100g::text,f.carbs_per_100g::text FROM recipe_items ri JOIN food_items f ON f.id=ri.food_id WHERE ri.recipe_id=$1::uuid ORDER BY ri.position`, sp(recipeID))
	if err != nil {
		return Recipe{}, err
	}
	for _, r := range itemRows {
		q, _ := strconv.ParseFloat(val(r, 2), 64)
		factor := q / 100
		kcal, _ := strconv.ParseFloat(val(r, 3), 64)
		protein, _ := strconv.ParseFloat(val(r, 4), 64)
		fat, _ := strconv.ParseFloat(val(r, 5), 64)
		carbs, _ := strconv.ParseFloat(val(r, 6), 64)
		recipe.Items = append(recipe.Items, RecipeItem{FoodID: val(r, 0), FoodName: val(r, 1), QuantityG: q, Calories: kcal * factor, Protein: protein * factor, Fat: fat * factor, Carbs: carbs * factor})
	}
	return recipe, nil
}

func (p *Postgres) CreateBodyMeasurement(ctx context.Context, in BodyMeasurement) (BodyMeasurement, error) {
	if in.ID == "" {
		in.ID = newID()
	}
	if in.LoggedAt.IsZero() {
		in.LoggedAt = time.Now().UTC()
	}
	rows, err := p.query(ctx, `INSERT INTO body_measurements(id,user_id,logged_at,weight_kg,waist_cm,chest_cm,shoulder_cm,arm_cm,forearm_cm,hip_cm,thigh_cm,calf_cm,neck_cm) VALUES($1::uuid,$2::uuid,$3::timestamptz,$4::numeric,$5::numeric,$6::numeric,$7::numeric,$8::numeric,$9::numeric,$10::numeric,$11::numeric,$12::numeric,$13::numeric) RETURNING id::text,user_id::text,logged_at::text,weight_kg::text,waist_cm::text,chest_cm::text,shoulder_cm::text,arm_cm::text,forearm_cm::text,hip_cm::text,thigh_cm::text,calf_cm::text,neck_cm::text`, sp(in.ID), sp(in.UserID), sp(in.LoggedAt.Format(time.RFC3339Nano)), fp(in.WeightKG), fp(in.WaistCM), fp(in.ChestCM), fp(in.ShoulderCM), fp(in.ArmCM), fp(in.ForearmCM), fp(in.HipCM), fp(in.ThighCM), fp(in.CalfCM), fp(in.NeckCM))
	if err != nil {
		return BodyMeasurement{}, err
	}
	return scanBodyMeasurement(rows[0])
}

func (p *Postgres) ListBodyMeasurements(ctx context.Context, userID string, from, to time.Time) ([]BodyMeasurement, error) {
	rows, err := p.query(ctx, `SELECT id::text,user_id::text,logged_at::text,weight_kg::text,waist_cm::text,chest_cm::text,shoulder_cm::text,arm_cm::text,forearm_cm::text,hip_cm::text,thigh_cm::text,calf_cm::text,neck_cm::text FROM body_measurements WHERE user_id=$1::uuid AND logged_at >= $2::timestamptz AND logged_at <= $3::timestamptz ORDER BY logged_at`, sp(userID), sp(from.UTC().Format(time.RFC3339Nano)), sp(to.UTC().Format(time.RFC3339Nano)))
	if err != nil {
		return nil, err
	}
	out := make([]BodyMeasurement, 0, len(rows))
	for _, r := range rows {
		m, err := scanBodyMeasurement(r)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}

func scanBodyMeasurement(r []*string) (BodyMeasurement, error) {
	t, err := parseTime(val(r, 2))
	if err != nil {
		return BodyMeasurement{}, err
	}
	return BodyMeasurement{ID: val(r, 0), UserID: val(r, 1), LoggedAt: t, WeightKG: parseFloatPtr(r[3]), WaistCM: parseFloatPtr(r[4]), ChestCM: parseFloatPtr(r[5]), ShoulderCM: parseFloatPtr(r[6]), ArmCM: parseFloatPtr(r[7]), ForearmCM: parseFloatPtr(r[8]), HipCM: parseFloatPtr(r[9]), ThighCM: parseFloatPtr(r[10]), CalfCM: parseFloatPtr(r[11]), NeckCM: parseFloatPtr(r[12])}, nil
}
