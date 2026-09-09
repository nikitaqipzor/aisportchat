package nutrition

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

type Service struct{ store store.Store }

func NewService(st store.Store) *Service { return &Service{store: st} }

type ProfileInput struct {
	Goal            string   `json:"goal"`
	ActivityLevel   string   `json:"activity_level"`
	CalculationMode string   `json:"calculation_mode"`
	CalorieTarget   *int     `json:"calorie_target,omitempty"`
	ProteinTarget   *float64 `json:"protein_target_g,omitempty"`
	FatTarget       *float64 `json:"fat_target_g,omitempty"`
	CarbTarget      *float64 `json:"carb_target_g,omitempty"`
}

type DaySummary struct {
	Date              string                 `json:"date"`
	Profile           store.NutritionProfile `json:"profile"`
	ConsumedCalories  float64                `json:"consumed_calories"`
	ConsumedProtein   float64                `json:"consumed_protein_g"`
	ConsumedFat       float64                `json:"consumed_fat_g"`
	ConsumedCarbs     float64                `json:"consumed_carbs_g"`
	ConsumedFiber     float64                `json:"consumed_fiber_g"`
	RemainingCalories float64                `json:"remaining_calories"`
	RemainingProtein  float64                `json:"remaining_protein_g"`
	RemainingFat      float64                `json:"remaining_fat_g"`
	RemainingCarbs    float64                `json:"remaining_carbs_g"`
	Entries           []store.FoodEntry      `json:"entries"`
	TrainingDay       bool                   `json:"training_day"`
	CompletedWorkouts int                    `json:"completed_workouts"`
}

type HistoryItem struct {
	Date             string  `json:"date"`
	Calories         float64 `json:"calories"`
	Protein          float64 `json:"protein_g"`
	Fat              float64 `json:"fat_g"`
	Carbs            float64 `json:"carbs_g"`
	TargetCalories   int     `json:"target_calories"`
	AdherencePercent float64 `json:"adherence_percent"`
	TrainingDay      bool    `json:"training_day"`
}

func (s *Service) GetProfile(ctx context.Context, userID string) (store.NutritionProfile, error) {
	return s.store.GetNutritionProfile(ctx, userID)
}

func (s *Service) SetProfile(ctx context.Context, userID string, in ProfileInput) (store.NutritionProfile, error) {
	in.Goal = strings.TrimSpace(in.Goal)
	in.ActivityLevel = strings.TrimSpace(in.ActivityLevel)
	in.CalculationMode = strings.TrimSpace(in.CalculationMode)
	if in.CalculationMode == "" {
		in.CalculationMode = "auto"
	}
	if !validGoal(in.Goal) {
		return store.NutritionProfile{}, errors.New("unsupported nutrition goal")
	}
	if !validActivity(in.ActivityLevel) {
		return store.NutritionProfile{}, errors.New("unsupported activity_level")
	}
	if in.CalculationMode != "auto" && in.CalculationMode != "manual" {
		return store.NutritionProfile{}, errors.New("calculation_mode must be auto or manual")
	}

	profile, err := s.store.GetProfile(ctx, userID)
	if err != nil {
		return store.NutritionProfile{}, err
	}
	if profile.WeightKG == nil {
		return store.NutritionProfile{}, errors.New("weight is required before nutrition setup")
	}

	calculated := CalculateTargets(*profile.WeightKG, in.Goal, in.ActivityLevel)
	calculated.UserID = userID
	calculated.CalculationMode = in.CalculationMode

	if in.CalculationMode == "manual" {
		if in.CalorieTarget == nil || in.ProteinTarget == nil || in.FatTarget == nil || in.CarbTarget == nil {
			return store.NutritionProfile{}, errors.New("manual mode requires calorie and macro targets")
		}
		if *in.CalorieTarget < 800 || *in.CalorieTarget > 8000 {
			return store.NutritionProfile{}, errors.New("calorie_target must be between 800 and 8000")
		}
		if *in.ProteinTarget < 0 || *in.FatTarget < 0 || *in.CarbTarget < 0 {
			return store.NutritionProfile{}, errors.New("macro targets cannot be negative")
		}
		calculated.CalorieTarget = *in.CalorieTarget
		calculated.ProteinTarget = round1(*in.ProteinTarget)
		calculated.FatTarget = round1(*in.FatTarget)
		calculated.CarbTarget = round1(*in.CarbTarget)
	}
	return s.store.UpsertNutritionProfile(ctx, calculated)
}

func CalculateTargets(weightKG float64, goal, activity string) store.NutritionProfile {
	kcalPerKG := map[string]float64{"low": 27, "light": 30, "moderate": 33, "high": 36, "athlete": 40}[activity]
	goalMultiplier := map[string]float64{"lose": 0.85, "recomp": 0.95, "maintain": 1, "gain": 1.10, "strength": 1.05, "endurance": 1.08}[goal]
	proteinPerKG := 1.8
	if goal == "lose" || goal == "recomp" {
		proteinPerKG = 2.0
	}
	if goal == "gain" || goal == "strength" {
		proteinPerKG = 1.9
	}
	protein := weightKG * proteinPerKG
	fat := weightKG * 0.8
	calories := int(math.Round(weightKG*kcalPerKG*goalMultiplier/10) * 10)
	carbCalories := float64(calories) - protein*4 - fat*9
	if carbCalories < 0 {
		carbCalories = 0
	}
	carbs := carbCalories / 4
	return store.NutritionProfile{
		Goal: goal, ActivityLevel: activity, CalculationMode: "auto",
		CalorieTarget: calories, ProteinTarget: round1(protein), FatTarget: round1(fat), CarbTarget: round1(carbs),
	}
}

func (s *Service) SearchFoods(ctx context.Context, userID, query string, limit int) ([]store.FoodItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	return s.store.SearchFoodItems(ctx, userID, strings.TrimSpace(query), limit)
}

func (s *Service) LogFood(ctx context.Context, userID string, foodID, mealType string, quantityG float64, loggedAt *time.Time) (DaySummary, error) {
	if !validMeal(mealType) {
		return DaySummary{}, errors.New("unsupported meal_type")
	}
	if quantityG <= 0 || quantityG > 5000 {
		return DaySummary{}, errors.New("quantity_g must be between 0 and 5000")
	}
	food, err := s.store.GetFoodItem(ctx, userID, foodID)
	if err != nil {
		return DaySummary{}, err
	}
	when := time.Now().UTC()
	if loggedAt != nil {
		when = loggedAt.UTC()
	}
	factor := quantityG / 100
	entry := store.FoodEntry{
		UserID: userID, FoodID: food.ID, FoodName: food.Name, MealType: mealType, LoggedAt: when, QuantityG: round1(quantityG),
		Calories: round1(food.Kcal100 * factor), Protein: round1(food.Protein100 * factor), Fat: round1(food.Fat100 * factor),
		Carbs: round1(food.Carbs100 * factor), Fiber: round1(food.Fiber100 * factor),
	}
	if _, err := s.store.CreateFoodEntry(ctx, entry); err != nil {
		return DaySummary{}, err
	}
	return s.Day(ctx, userID, when)
}

func (s *Service) DeleteEntry(ctx context.Context, userID, entryID string) error {
	return s.store.DeleteFoodEntry(ctx, userID, entryID)
}

func (s *Service) Day(ctx context.Context, userID string, at time.Time) (DaySummary, error) {
	date := at.UTC().Format("2006-01-02")
	start, _ := time.Parse("2006-01-02", date)
	end := start.Add(24*time.Hour - time.Nanosecond)
	nutritionProfile, err := s.store.GetNutritionProfile(ctx, userID)
	if err != nil {
		return DaySummary{}, err
	}
	entries, err := s.store.ListFoodEntries(ctx, userID, start, end)
	if err != nil {
		return DaySummary{}, err
	}
	out := DaySummary{Date: date, Profile: nutritionProfile, Entries: entries}
	for _, e := range entries {
		out.ConsumedCalories += e.Calories
		out.ConsumedProtein += e.Protein
		out.ConsumedFat += e.Fat
		out.ConsumedCarbs += e.Carbs
		out.ConsumedFiber += e.Fiber
	}
	out.ConsumedCalories = round1(out.ConsumedCalories)
	out.ConsumedProtein = round1(out.ConsumedProtein)
	out.ConsumedFat = round1(out.ConsumedFat)
	out.ConsumedCarbs = round1(out.ConsumedCarbs)
	out.ConsumedFiber = round1(out.ConsumedFiber)
	out.RemainingCalories = round1(math.Max(0, float64(nutritionProfile.CalorieTarget)-out.ConsumedCalories))
	out.RemainingProtein = round1(math.Max(0, nutritionProfile.ProteinTarget-out.ConsumedProtein))
	out.RemainingFat = round1(math.Max(0, nutritionProfile.FatTarget-out.ConsumedFat))
	out.RemainingCarbs = round1(math.Max(0, nutritionProfile.CarbTarget-out.ConsumedCarbs))
	workouts, _ := s.store.ListWorkouts(ctx, userID, 200)
	for _, w := range workouts {
		if w.Workout.Status == "completed" && w.Workout.CompletedAt != nil && w.Workout.CompletedAt.UTC().Format("2006-01-02") == date {
			out.CompletedWorkouts++
		}
	}
	out.TrainingDay = out.CompletedWorkouts > 0
	return out, nil
}

func (s *Service) History(ctx context.Context, userID string, days int, now time.Time) ([]HistoryItem, error) {
	if days <= 0 {
		days = 7
	}
	if days > 90 {
		days = 90
	}
	profile, err := s.store.GetNutritionProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	lastDate := now.UTC().Format("2006-01-02")
	lastDay, _ := time.Parse("2006-01-02", lastDate)
	firstDay := lastDay.AddDate(0, 0, -(days - 1))
	entries, err := s.store.ListFoodEntries(ctx, userID, firstDay, lastDay.Add(24*time.Hour-time.Nanosecond))
	if err != nil {
		return nil, err
	}
	workouts, err := s.store.ListWorkouts(ctx, userID, 500)
	if err != nil {
		return nil, err
	}

	byDate := make(map[string]*HistoryItem, days)
	out := make([]HistoryItem, 0, days)
	for i := 0; i < days; i++ {
		date := firstDay.AddDate(0, 0, i).Format("2006-01-02")
		out = append(out, HistoryItem{Date: date, TargetCalories: profile.CalorieTarget})
		byDate[date] = &out[len(out)-1]
	}
	for _, entry := range entries {
		date := entry.LoggedAt.UTC().Format("2006-01-02")
		item := byDate[date]
		if item == nil {
			continue
		}
		item.Calories += entry.Calories
		item.Protein += entry.Protein
		item.Fat += entry.Fat
		item.Carbs += entry.Carbs
	}
	for _, workout := range workouts {
		if workout.Workout.Status != "completed" || workout.Workout.CompletedAt == nil {
			continue
		}
		date := workout.Workout.CompletedAt.UTC().Format("2006-01-02")
		if item := byDate[date]; item != nil {
			item.TrainingDay = true
		}
	}
	for i := range out {
		out[i].Calories = round1(out[i].Calories)
		out[i].Protein = round1(out[i].Protein)
		out[i].Fat = round1(out[i].Fat)
		out[i].Carbs = round1(out[i].Carbs)
		if profile.CalorieTarget > 0 {
			out[i].AdherencePercent = round1(math.Min(200, out[i].Calories/float64(profile.CalorieTarget)*100))
		}
	}
	return out, nil
}

func validGoal(v string) bool {
	switch v {
	case "lose", "recomp", "maintain", "gain", "strength", "endurance":
		return true
	}
	return false
}
func validActivity(v string) bool {
	switch v {
	case "low", "light", "moderate", "high", "athlete":
		return true
	}
	return false
}
func validMeal(v string) bool {
	switch v {
	case "breakfast", "lunch", "dinner", "snack":
		return true
	}
	return false
}
func round1(v float64) float64 { return math.Round(v*10) / 10 }
