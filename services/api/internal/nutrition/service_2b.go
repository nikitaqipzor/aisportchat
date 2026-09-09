package nutrition

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

type CustomFoodInput struct {
	Name       string  `json:"name"`
	Brand      string  `json:"brand,omitempty"`
	Barcode    string  `json:"barcode,omitempty"`
	Kcal100    float64 `json:"kcal_per_100g"`
	Protein100 float64 `json:"protein_per_100g"`
	Fat100     float64 `json:"fat_per_100g"`
	Carbs100   float64 `json:"carbs_per_100g"`
	Fiber100   float64 `json:"fiber_per_100g"`
	ServingG   float64 `json:"serving_g"`
}

type RecipeItemInput struct {
	FoodID    string  `json:"food_id"`
	QuantityG float64 `json:"quantity_g"`
}

type RecipeInput struct {
	Name  string            `json:"name"`
	Items []RecipeItemInput `json:"items"`
}

type NutritionTrainingCorrelation struct {
	Days                     int     `json:"days"`
	LoggedTrainingDays       int     `json:"logged_training_days"`
	LoggedRestDays           int     `json:"logged_rest_days"`
	AvgTrainingCalories      float64 `json:"avg_training_calories"`
	AvgRestCalories          float64 `json:"avg_rest_calories"`
	AvgTrainingProteinG      float64 `json:"avg_training_protein_g"`
	AvgRestProteinG          float64 `json:"avg_rest_protein_g"`
	TrainingCalorieAdherence float64 `json:"training_calorie_adherence_percent"`
	RestCalorieAdherence     float64 `json:"rest_calorie_adherence_percent"`
}

func (s *Service) CreateCustomFood(ctx context.Context, userID string, in CustomFoodInput) (store.FoodItem, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Brand = strings.TrimSpace(in.Brand)
	in.Barcode = strings.TrimSpace(in.Barcode)
	if len(in.Name) < 2 || len(in.Name) > 120 {
		return store.FoodItem{}, errors.New("name must be 2-120 characters")
	}
	if in.Kcal100 < 0 || in.Kcal100 > 1000 {
		return store.FoodItem{}, errors.New("kcal_per_100g must be between 0 and 1000")
	}
	for _, value := range []float64{in.Protein100, in.Fat100, in.Carbs100, in.Fiber100} {
		if value < 0 || value > 100 {
			return store.FoodItem{}, errors.New("macros per 100g must be between 0 and 100")
		}
	}
	if in.ServingG <= 0 {
		in.ServingG = 100
	}
	if in.ServingG > 5000 {
		return store.FoodItem{}, errors.New("serving_g must be <= 5000")
	}
	if in.Barcode != "" {
		if existing, err := s.store.FindFoodByBarcode(ctx, userID, in.Barcode); err == nil {
			return existing, errors.New("barcode already exists")
		} else if !errors.Is(err, store.ErrNotFound) {
			return store.FoodItem{}, err
		}
	}
	return s.store.CreateCustomFood(ctx, userID, store.FoodItem{Name: in.Name, Brand: in.Brand, Barcode: in.Barcode, Kcal100: round1(in.Kcal100), Protein100: round1(in.Protein100), Fat100: round1(in.Fat100), Carbs100: round1(in.Carbs100), Fiber100: round1(in.Fiber100), ServingG: round1(in.ServingG), Source: "custom"})
}

func (s *Service) FoodByBarcode(ctx context.Context, userID, barcode string) (store.FoodItem, error) {
	barcode = strings.TrimSpace(barcode)
	if barcode == "" {
		return store.FoodItem{}, errors.New("barcode is required")
	}
	return s.store.FindFoodByBarcode(ctx, userID, barcode)
}

func (s *Service) RepeatEntry(ctx context.Context, userID, entryID, mealType string, loggedAt *time.Time) (DaySummary, error) {
	previous, err := s.store.GetFoodEntry(ctx, userID, entryID)
	if err != nil {
		return DaySummary{}, err
	}
	if mealType == "" {
		mealType = previous.MealType
	}
	return s.LogFood(ctx, userID, previous.FoodID, mealType, previous.QuantityG, loggedAt)
}

func (s *Service) CreateRecipe(ctx context.Context, userID string, in RecipeInput) (store.Recipe, error) {
	in.Name = strings.TrimSpace(in.Name)
	if len(in.Name) < 2 || len(in.Name) > 120 {
		return store.Recipe{}, errors.New("recipe name must be 2-120 characters")
	}
	if len(in.Items) == 0 || len(in.Items) > 30 {
		return store.Recipe{}, errors.New("recipe must contain 1-30 items")
	}
	recipe := store.Recipe{UserID: userID, Name: in.Name, Items: make([]store.RecipeItem, 0, len(in.Items))}
	for _, raw := range in.Items {
		if raw.QuantityG <= 0 || raw.QuantityG > 5000 {
			return store.Recipe{}, errors.New("recipe quantity_g must be between 0 and 5000")
		}
		food, err := s.store.GetFoodItem(ctx, userID, raw.FoodID)
		if err != nil {
			return store.Recipe{}, err
		}
		factor := raw.QuantityG / 100
		recipe.Items = append(recipe.Items, store.RecipeItem{FoodID: food.ID, FoodName: food.Name, QuantityG: round1(raw.QuantityG), Calories: round1(food.Kcal100 * factor), Protein: round1(food.Protein100 * factor), Fat: round1(food.Fat100 * factor), Carbs: round1(food.Carbs100 * factor)})
	}
	return s.store.CreateRecipe(ctx, recipe)
}

func (s *Service) ListRecipes(ctx context.Context, userID string) ([]store.Recipe, error) {
	return s.store.ListRecipes(ctx, userID)
}

func (s *Service) LogRecipe(ctx context.Context, userID, recipeID, mealType string, scale float64, loggedAt *time.Time) (DaySummary, error) {
	if !validMeal(mealType) {
		return DaySummary{}, errors.New("unsupported meal_type")
	}
	if scale <= 0 {
		scale = 1
	}
	if scale > 10 {
		return DaySummary{}, errors.New("scale must be <= 10")
	}
	recipe, err := s.store.GetRecipe(ctx, userID, recipeID)
	if err != nil {
		return DaySummary{}, err
	}
	when := time.Now().UTC()
	if loggedAt != nil {
		when = loggedAt.UTC()
	}
	for _, item := range recipe.Items {
		food, err := s.store.GetFoodItem(ctx, userID, item.FoodID)
		if err != nil {
			return DaySummary{}, err
		}
		q := round1(item.QuantityG * scale)
		factor := q / 100
		_, err = s.store.CreateFoodEntry(ctx, store.FoodEntry{UserID: userID, FoodID: food.ID, FoodName: recipe.Name + " · " + food.Name, MealType: mealType, LoggedAt: when, QuantityG: q, Calories: round1(food.Kcal100 * factor), Protein: round1(food.Protein100 * factor), Fat: round1(food.Fat100 * factor), Carbs: round1(food.Carbs100 * factor), Fiber: round1(food.Fiber100 * factor)})
		if err != nil {
			return DaySummary{}, err
		}
	}
	return s.Day(ctx, userID, when)
}

func (s *Service) Correlation(ctx context.Context, userID string, days int, now time.Time) (NutritionTrainingCorrelation, error) {
	if days <= 0 {
		days = 30
	}
	if days > 180 {
		days = 180
	}
	history, err := s.History(ctx, userID, days, now)
	if err != nil {
		return NutritionTrainingCorrelation{}, err
	}
	profile, err := s.store.GetNutritionProfile(ctx, userID)
	if err != nil {
		return NutritionTrainingCorrelation{}, err
	}
	var out NutritionTrainingCorrelation
	out.Days = days
	var tCal, rCal, tPro, rPro float64
	for _, d := range history {
		if d.Calories <= 0 {
			continue
		}
		if d.TrainingDay {
			out.LoggedTrainingDays++
			tCal += d.Calories
			tPro += d.Protein
		} else {
			out.LoggedRestDays++
			rCal += d.Calories
			rPro += d.Protein
		}
	}
	if out.LoggedTrainingDays > 0 {
		out.AvgTrainingCalories = round1(tCal / float64(out.LoggedTrainingDays))
		out.AvgTrainingProteinG = round1(tPro / float64(out.LoggedTrainingDays))
		out.TrainingCalorieAdherence = round1(out.AvgTrainingCalories / float64(profile.CalorieTarget) * 100)
	}
	if out.LoggedRestDays > 0 {
		out.AvgRestCalories = round1(rCal / float64(out.LoggedRestDays))
		out.AvgRestProteinG = round1(rPro / float64(out.LoggedRestDays))
		out.RestCalorieAdherence = round1(out.AvgRestCalories / float64(profile.CalorieTarget) * 100)
	}
	out.TrainingCalorieAdherence = math.Min(200, out.TrainingCalorieAdherence)
	out.RestCalorieAdherence = math.Min(200, out.RestCalorieAdherence)
	return out, nil
}
