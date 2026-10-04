package nutrition

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

type FoodBatchItem struct {
	FoodID string `json:"food_id"`
	QuantityG float64 `json:"quantity_g"`
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
	return s.RepeatEntryWithKey(ctx,userID,entryID,mealType,loggedAt,"")
}

func (s *Service) RepeatEntryWithKey(ctx context.Context, userID, entryID, mealType string, loggedAt *time.Time, operationKey string) (DaySummary, error) {
	previous, err := s.store.GetFoodEntry(ctx, userID, entryID)
	if err != nil {
		return DaySummary{}, err
	}
	if mealType == "" {
		mealType = previous.MealType
	}
	return s.LogFoodBatch(ctx,userID,mealType,[]FoodBatchItem{{FoodID:previous.FoodID,QuantityG:previous.QuantityG}},loggedAt,operationKey,"repeat:"+entryID,"")
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

func (s *Service) LogRecipe(ctx context.Context, userID, recipeID, mealType string, scale float64, loggedAt *time.Time, operationKeys ...string) (DaySummary, error) {
	key:=""
	if len(operationKeys)>0 {key=operationKeys[0]}
	return s.LogRecipeInLocation(ctx,userID,recipeID,mealType,scale,loggedAt,key,time.UTC)
}

func (s *Service) LogRecipeInLocation(ctx context.Context, userID, recipeID, mealType string, scale float64, loggedAt *time.Time, operationKey string, loc *time.Location) (DaySummary, error) {
	if !validMeal(mealType) {
		return DaySummary{}, errors.New("unsupported meal_type")
	}
	if scale <= 0 {
		scale = 1
	}
	if math.IsNaN(scale) || math.IsInf(scale, 0) || scale > 10 {
		return DaySummary{}, errors.New("scale must be <= 10")
	}
	recipe, err := s.store.GetRecipe(ctx, userID, recipeID)
	if err != nil {
		return DaySummary{}, err
	}
	items := make([]FoodBatchItem,0,len(recipe.Items))
	for _, item := range recipe.Items {
		items=append(items,FoodBatchItem{FoodID:item.FoodID,QuantityG:round1(item.QuantityG*scale)})
	}
	return s.LogFoodBatchInLocation(ctx,userID,mealType,items,loggedAt,operationKey,"recipe:"+recipeID,recipe.Name+" · ",loc)
}

// LogFoodBatch prevalidates the entire meal and delegates the all-or-nothing
// write and deduplication to the selected store. Source scopes a key across
// recipe and AI endpoints; separate user actions must use separate keys.
func (s *Service) LogFoodBatch(ctx context.Context,userID,mealType string,items []FoodBatchItem,loggedAt *time.Time,operationKey,source,namePrefix string) (DaySummary,error) {
	return s.LogFoodBatchInLocation(ctx,userID,mealType,items,loggedAt,operationKey,source,namePrefix,time.UTC)
}

func (s *Service) LogFoodBatchInLocation(ctx context.Context,userID,mealType string,items []FoodBatchItem,loggedAt *time.Time,operationKey,source,namePrefix string,loc *time.Location) (DaySummary,error) {
	if !validMeal(mealType) {return DaySummary{},errors.New("unsupported meal_type")}
	if len(items)==0 || len(items)>30 {return DaySummary{},errors.New("items must contain 1-30 foods")}
	if err:=validateFoodOperationKey(operationKey);err!=nil {return DaySummary{},err}
	when:=time.Now().UTC()
	var requestedAt *time.Time
	if loggedAt!=nil {utc:=loggedAt.UTC();when=utc;requestedAt=&utc}
	maxQuantityG:=5000.0
	if strings.HasPrefix(source,"recipe:") {maxQuantityG=50000}
	prepared:=make([]store.FoodEntry,0,len(items))
	for _,item:=range items {
		if strings.TrimSpace(item.FoodID)=="" || math.IsNaN(item.QuantityG) || math.IsInf(item.QuantityG,0) || item.QuantityG<=0 || item.QuantityG>maxQuantityG {return DaySummary{},errors.New("food_id and a supported positive quantity_g are required")}
		food,err:=s.store.GetFoodItem(ctx,userID,item.FoodID)
		if err!=nil {return DaySummary{},err}
		q:=round1(item.QuantityG)
		if q<=0 {return DaySummary{},errors.New("quantity_g is too small")}
		factor:=q/100
		prepared=append(prepared,store.FoodEntry{UserID:userID,FoodID:food.ID,FoodName:namePrefix+food.Name,MealType:mealType,LoggedAt:when,QuantityG:q,Calories:round1(food.Kcal100*factor),Protein:round1(food.Protein100*factor),Fat:round1(food.Fat100*factor),Carbs:round1(food.Carbs100*factor),Fiber:round1(food.Fiber100*factor)})
	}
	fingerprint:=struct{Source string;MealType string;Items []FoodBatchItem;LoggedAt *time.Time}{source,mealType,items,requestedAt}
	encoded,err:=json.Marshal(fingerprint)
	if err!=nil {return DaySummary{},err}
	digest:=sha256.Sum256(encoded)
	savedAt,err:=s.store.CreateFoodEntries(ctx,userID,operationKey,hex.EncodeToString(digest[:]),prepared)
	if err!=nil {return DaySummary{},err}
	return s.DayInLocation(ctx,userID,savedAt,loc)
}

func validateFoodOperationKey(key string) error {
	if key=="" {return nil} // Legacy callers still receive atomic batches.
	if len(key)<8 || len(key)>128 {return errors.New("idempotency_key must be 8-128 ASCII letters, numbers, '-' or '_'")}
	for _,c:=range key {if !((c>='a'&&c<='z')||(c>='A'&&c<='Z')||(c>='0'&&c<='9')||c=='-'||c=='_') {return errors.New("idempotency_key must be 8-128 ASCII letters, numbers, '-' or '_'")}}
	return nil
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
