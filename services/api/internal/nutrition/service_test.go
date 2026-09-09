package nutrition

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func TestCalculateTargets(t *testing.T) {
	got := CalculateTargets(100, "recomp", "moderate")
	if got.CalorieTarget != 3140 {
		t.Fatalf("calories=%d", got.CalorieTarget)
	}
	if got.ProteinTarget != 200 {
		t.Fatalf("protein=%v", got.ProteinTarget)
	}
	if got.FatTarget != 80 {
		t.Fatalf("fat=%v", got.FatTarget)
	}
	if got.CarbTarget <= 0 {
		t.Fatal("expected carbs")
	}
}

func nutritionFixture(t *testing.T, weight *float64) (*Service, *store.Memory, string) {
	t.Helper()
	ctx := context.Background()
	st := store.NewMemory()
	user, err := st.CreateUser(ctx, "nutrition@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.UpsertProfile(ctx, store.Profile{UserID: user.ID, WeightKG: weight}); err != nil {
		t.Fatal(err)
	}
	return NewService(st), st, user.ID
}

func setAutoNutrition(t *testing.T, svc *Service, userID string) store.NutritionProfile {
	t.Helper()
	got, err := svc.SetProfile(context.Background(), userID, ProfileInput{Goal: "maintain", ActivityLevel: "moderate", CalculationMode: "auto"})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestSetProfileValidationAndManualTargets(t *testing.T) {
	weight := 90.0
	svc, _, userID := nutritionFixture(t, &weight)
	ctx := context.Background()

	cases := []struct {
		name string
		in   ProfileInput
		want string
	}{
		{"goal", ProfileInput{Goal: "bulk-forever", ActivityLevel: "moderate"}, "unsupported nutrition goal"},
		{"activity", ProfileInput{Goal: "maintain", ActivityLevel: "extreme"}, "unsupported activity_level"},
		{"mode", ProfileInput{Goal: "maintain", ActivityLevel: "moderate", CalculationMode: "magic"}, "calculation_mode"},
		{"manual-missing", ProfileInput{Goal: "maintain", ActivityLevel: "moderate", CalculationMode: "manual"}, "requires calorie and macro"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.SetProfile(ctx, userID, tc.in)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("err=%v want %q", err, tc.want)
			}
		})
	}

	p, f, c := 180.04, 70.06, 250.05
	low := 799
	_, err := svc.SetProfile(ctx, userID, ProfileInput{Goal: "maintain", ActivityLevel: "moderate", CalculationMode: "manual", CalorieTarget: &low, ProteinTarget: &p, FatTarget: &f, CarbTarget: &c})
	if err == nil || !strings.Contains(err.Error(), "between 800 and 8000") {
		t.Fatalf("low calories err=%v", err)
	}

	kcal := 2400
	negative := -1.0
	_, err = svc.SetProfile(ctx, userID, ProfileInput{Goal: "maintain", ActivityLevel: "moderate", CalculationMode: "manual", CalorieTarget: &kcal, ProteinTarget: &negative, FatTarget: &f, CarbTarget: &c})
	if err == nil || !strings.Contains(err.Error(), "cannot be negative") {
		t.Fatalf("negative macro err=%v", err)
	}

	got, err := svc.SetProfile(ctx, userID, ProfileInput{Goal: "maintain", ActivityLevel: "moderate", CalculationMode: "manual", CalorieTarget: &kcal, ProteinTarget: &p, FatTarget: &f, CarbTarget: &c})
	if err != nil {
		t.Fatal(err)
	}
	if got.CalorieTarget != 2400 || got.ProteinTarget != 180 || got.FatTarget != 70.1 || got.CarbTarget != 250.1 {
		t.Fatalf("manual targets=%+v", got)
	}
}

func TestSetProfileRequiresWeight(t *testing.T) {
	svc, _, userID := nutritionFixture(t, nil)
	_, err := svc.SetProfile(context.Background(), userID, ProfileInput{Goal: "maintain", ActivityLevel: "moderate"})
	if err == nil || !strings.Contains(err.Error(), "weight is required") {
		t.Fatalf("err=%v", err)
	}
}

func TestLogFoodDayDeleteAndSearch(t *testing.T) {
	weight := 85.0
	svc, _, userID := nutritionFixture(t, &weight)
	setAutoNutrition(t, svc, userID)
	ctx := context.Background()
	when := time.Date(2026, 9, 2, 10, 0, 0, 0, time.UTC)

	if _, err := svc.LogFood(ctx, userID, "banana", "invalid", 120, &when); err == nil {
		t.Fatal("expected invalid meal")
	}
	if _, err := svc.LogFood(ctx, userID, "banana", "snack", 0, &when); err == nil {
		t.Fatal("expected invalid quantity")
	}

	day, err := svc.LogFood(ctx, userID, "banana", "snack", 120, &when)
	if err != nil {
		t.Fatal(err)
	}
	if len(day.Entries) != 1 || day.ConsumedCalories != 106.8 || day.ConsumedCarbs != 27.4 {
		t.Fatalf("day=%+v", day)
	}
	if day.RemainingCalories <= 0 {
		t.Fatalf("remaining=%v", day.RemainingCalories)
	}

	foods, err := svc.SearchFoods(ctx, userID, "банан", 999)
	if err != nil || len(foods) != 1 || foods[0].ID != "banana" {
		t.Fatalf("foods=%v err=%v", foods, err)
	}

	if err := svc.DeleteEntry(ctx, userID, day.Entries[0].ID); err != nil {
		t.Fatal(err)
	}
	after, err := svc.Day(ctx, userID, when)
	if err != nil {
		t.Fatal(err)
	}
	if len(after.Entries) != 0 || after.ConsumedCalories != 0 {
		t.Fatalf("after=%+v", after)
	}
}

func TestCustomFoodBarcodePrivacyAndValidation(t *testing.T) {
	weight := 80.0
	svc, st, userID := nutritionFixture(t, &weight)
	ctx := context.Background()

	invalid := []CustomFoodInput{
		{Name: "x", Kcal100: 100, ServingG: 100},
		{Name: "Valid", Kcal100: 1001, ServingG: 100},
		{Name: "Valid", Kcal100: 100, Protein100: 101, ServingG: 100},
		{Name: "Valid", Kcal100: 100, ServingG: 5001},
	}
	for i, in := range invalid {
		if _, err := svc.CreateCustomFood(ctx, userID, in); err == nil {
			t.Fatalf("case %d expected error", i)
		}
	}

	food, err := svc.CreateCustomFood(ctx, userID, CustomFoodInput{Name: "  My Curd  ", Brand: " Brand ", Barcode: "460000000001", Kcal100: 121.04, Protein100: 17.04, Fat100: 5.04, Carbs100: 1.84})
	if err != nil {
		t.Fatal(err)
	}
	if food.Name != "My Curd" || food.ServingG != 100 || food.Source != "custom" {
		t.Fatalf("food=%+v", food)
	}

	found, err := svc.FoodByBarcode(ctx, userID, " 460000000001 ")
	if err != nil || found.ID != food.ID {
		t.Fatalf("found=%+v err=%v", found, err)
	}
	if _, err := svc.FoodByBarcode(ctx, userID, " "); err == nil {
		t.Fatal("expected blank barcode error")
	}
	if _, err := svc.CreateCustomFood(ctx, userID, CustomFoodInput{Name: "Duplicate", Barcode: "460000000001", Kcal100: 100, ServingG: 100}); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("duplicate err=%v", err)
	}

	other, err := st.CreateUser(ctx, "other@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	otherSvc := NewService(st)
	if _, err := otherSvc.FoodByBarcode(ctx, other.ID, "460000000001"); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("cross user barcode err=%v", err)
	}
	search, err := otherSvc.SearchFoods(ctx, other.ID, "My Curd", 30)
	if err != nil || len(search) != 0 {
		t.Fatalf("cross user search=%v err=%v", search, err)
	}
}

func TestRecipeRepeatHistoryAndCorrelation(t *testing.T) {
	weight := 82.0
	svc, _, userID := nutritionFixture(t, &weight)
	profile := setAutoNutrition(t, svc, userID)
	ctx := context.Background()
	when := time.Date(2026, 9, 2, 8, 0, 0, 0, time.UTC)

	if _, err := svc.CreateRecipe(ctx, userID, RecipeInput{Name: "x"}); err == nil {
		t.Fatal("expected short name error")
	}
	if _, err := svc.CreateRecipe(ctx, userID, RecipeInput{Name: "Breakfast", Items: []RecipeItemInput{{FoodID: "banana", QuantityG: 0}}}); err == nil {
		t.Fatal("expected recipe quantity error")
	}

	recipe, err := svc.CreateRecipe(ctx, userID, RecipeInput{Name: "Breakfast", Items: []RecipeItemInput{{FoodID: "banana", QuantityG: 120}, {FoodID: "oats-dry", QuantityG: 80}}})
	if err != nil {
		t.Fatal(err)
	}
	if len(recipe.Items) != 2 || recipe.Items[0].Calories <= 0 {
		t.Fatalf("recipe=%+v", recipe)
	}

	if _, err := svc.LogRecipe(ctx, userID, recipe.ID, "invalid", 1, &when); err == nil {
		t.Fatal("expected meal error")
	}
	if _, err := svc.LogRecipe(ctx, userID, recipe.ID, "breakfast", 11, &when); err == nil {
		t.Fatal("expected scale error")
	}
	day, err := svc.LogRecipe(ctx, userID, recipe.ID, "breakfast", 0, &when)
	if err != nil {
		t.Fatal(err)
	}
	if len(day.Entries) != 2 {
		t.Fatalf("entries=%d", len(day.Entries))
	}

	later := when.Add(2 * time.Hour)
	repeated, err := svc.RepeatEntry(ctx, userID, day.Entries[0].ID, "snack", &later)
	if err != nil {
		t.Fatal(err)
	}
	if len(repeated.Entries) != 3 {
		t.Fatalf("repeated entries=%d", len(repeated.Entries))
	}

	hist, err := svc.History(ctx, userID, 0, when)
	if err != nil {
		t.Fatal(err)
	}
	if len(hist) != 7 || hist[len(hist)-1].Calories <= 0 {
		t.Fatalf("history=%+v", hist)
	}

	corr, err := svc.Correlation(ctx, userID, 999, when)
	if err != nil {
		t.Fatal(err)
	}
	if corr.Days != 180 || corr.LoggedRestDays != 1 || corr.AvgRestCalories <= 0 {
		t.Fatalf("corr=%+v", corr)
	}
	if corr.RestCalorieAdherence < 0 || corr.RestCalorieAdherence > 200 {
		t.Fatalf("adherence=%v target=%d", corr.RestCalorieAdherence, profile.CalorieTarget)
	}
}
