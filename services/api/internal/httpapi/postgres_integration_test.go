//go:build cgo

package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/auth"
	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func TestPostgresCriticalReleaseFlow(t *testing.T) {
	dsn := os.Getenv("POSTGRES_TEST_DSN")
	if dsn == "" {
		t.Skip("POSTGRES_TEST_DSN is not set")
	}
	pg, err := store.NewPostgres(dsn)
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	defer pg.Close()
	if err := pg.Ping(context.Background()); err != nil {
		t.Fatalf("ping postgres: %v", err)
	}

	tm := auth.NewTokenManager("postgres-integration-test-secret-0123456789", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(pg, tm)
	suffix := time.Now().UTC().UnixNano()
	access := registerAndOnboard(t, h, fmt.Sprintf("pg-release-%d@example.com", suffix))

	// Profile PATCH must preserve omitted fields while still allowing explicit
	// empty arrays to clear private health notes.
	if rr := doJSON(t, h, http.MethodPatch, "/api/v1/profile", map[string]any{
		"unit_system": "imperial", "injuries": []string{"shoulder"}, "limitations": []string{"no flyes"},
	}, access); rr.Code != http.StatusOK {
		t.Fatalf("postgres profile seed=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := doJSON(t, h, http.MethodPatch, "/api/v1/profile", map[string]any{"weight_kg": 88}, access); rr.Code != http.StatusOK {
		t.Fatalf("postgres profile partial patch=%d body=%s", rr.Code, rr.Body.String())
	}
	profile := doJSON(t, h, http.MethodGet, "/api/v1/profile", nil, access)
	if profile.Code != http.StatusOK || !bytes.Contains(profile.Body.Bytes(), []byte(`"unit_system":"imperial"`)) || !bytes.Contains(profile.Body.Bytes(), []byte(`"injuries":["shoulder"]`)) || !bytes.Contains(profile.Body.Bytes(), []byte(`"limitations":["no flyes"]`)) {
		t.Fatalf("postgres profile patch lost omitted fields status=%d body=%s", profile.Code, profile.Body.String())
	}
	if rr := doJSON(t, h, http.MethodPatch, "/api/v1/profile", map[string]any{"injuries": []string{}, "limitations": []string{}}, access); rr.Code != http.StatusOK || !bytes.Contains(rr.Body.Bytes(), []byte(`"injuries":[]`)) || !bytes.Contains(rr.Body.Bytes(), []byte(`"limitations":[]`)) {
		t.Fatalf("postgres profile explicit clear=%d body=%s", rr.Code, rr.Body.String())
	}

	// The aggregate store transaction must roll back profile and goal if a later
	// training write fails (the unknown equipment id violates its FK).
	var profileEnvelope struct {
		Profile store.Profile `json:"profile"`
	}
	if err := json.Unmarshal(profile.Body.Bytes(), &profileEnvelope); err != nil {
		t.Fatal(err)
	}
	beforeGoal, err := pg.GetGoal(context.Background(), profileEnvelope.Profile.UserID)
	if err != nil {
		t.Fatalf("load goal before rollback test: %v", err)
	}
	height, weight, age := 180.0, 80.0, 31
	level := "advanced"
	_, err = pg.SaveAthleteProfile(context.Background(), profileEnvelope.Profile.UserID, store.AthleteProfileUpdate{
		Profile: store.Profile{HeightCM: &height, WeightKG: &weight, AgeYears: &age, ExperienceLevel: &level, UnitSystem: "metric", Injuries: []string{}, Limitations: []string{}},
		Goal: store.Goal{GoalType: "fat_loss"},
		TrainingPreferences: store.TrainingPreferences{Environments: []string{"gym"}, EquipmentIDs: []string{"missing-equipment-fk"}, WorkoutsPerWeek: 2, SessionMinutes: 45},
	})
	if err == nil {
		t.Fatal("expected aggregate save to fail on unknown equipment FK")
	}
	afterProfile, err := pg.GetProfile(context.Background(), profileEnvelope.Profile.UserID)
	if err != nil {
		t.Fatalf("load profile after rollback: %v", err)
	}
	afterGoal, err := pg.GetGoal(context.Background(), profileEnvelope.Profile.UserID)
	if err != nil {
		t.Fatalf("load goal after rollback: %v", err)
	}
	if afterProfile.HeightCM != nil && *afterProfile.HeightCM == height {
		t.Fatalf("profile write escaped rollback: %+v", afterProfile)
	}
	if afterGoal.GoalType != beforeGoal.GoalType {
		t.Fatalf("goal write escaped rollback: before=%s after=%s", beforeGoal.GoalType, afterGoal.GoalType)
	}

	// Workout persistence: generate -> start -> log -> finish -> history.
	generated := doJSON(t, h, http.MethodPost, "/api/v1/workouts/generate", map[string]any{
		"muscle": "chest", "environment": "gym", "duration_minutes": 45,
	}, access)
	if generated.Code != http.StatusCreated {
		t.Fatalf("postgres generate status=%d body=%s", generated.Code, generated.Body.String())
	}
	var workout struct {
		Workout struct {
			ID string `json:"id"`
		} `json:"workout"`
		Exercises []struct {
			WorkoutExercise struct {
				ID            string `json:"id"`
				TargetRepsMax int    `json:"target_reps_max"`
			} `json:"workout_exercise"`
		} `json:"exercises"`
	}
	if err := json.Unmarshal(generated.Body.Bytes(), &workout); err != nil {
		t.Fatal(err)
	}
	if workout.Workout.ID == "" || len(workout.Exercises) == 0 {
		t.Fatalf("invalid persisted workout body=%s", generated.Body.String())
	}
	if rr := doJSON(t, h, http.MethodPost, "/api/v1/workouts/"+workout.Workout.ID+"/start", map[string]any{}, access); rr.Code != http.StatusOK {
		t.Fatalf("postgres start=%d body=%s", rr.Code, rr.Body.String())
	}
	first := workout.Exercises[0].WorkoutExercise
	if rr := doJSON(t, h, http.MethodPut, "/api/v1/workouts/"+workout.Workout.ID+"/sets", map[string]any{
		"workout_exercise_id": first.ID, "set_number": 1, "weight": 80, "repetitions": first.TargetRepsMax, "rir": 2,
	}, access); rr.Code != http.StatusOK {
		t.Fatalf("postgres log set=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := doJSON(t, h, http.MethodPost, "/api/v1/workouts/"+workout.Workout.ID+"/finish", map[string]any{}, access); rr.Code != http.StatusOK {
		t.Fatalf("postgres finish=%d body=%s", rr.Code, rr.Body.String())
	}
	if rr := doJSON(t, h, http.MethodGet, "/api/v1/workouts/history", nil, access); rr.Code != http.StatusOK || !bytes.Contains(rr.Body.Bytes(), []byte(workout.Workout.ID)) {
		t.Fatalf("postgres history=%d body=%s", rr.Code, rr.Body.String())
	}

	// Nutrition persistence, custom-food ownership and diary aggregation.
	if rr := doJSON(t, h, http.MethodPut, "/api/v1/nutrition/profile", map[string]any{
		"goal": "maintain", "activity_level": "moderate", "calculation_mode": "auto",
	}, access); rr.Code != http.StatusOK {
		t.Fatalf("postgres nutrition setup=%d body=%s", rr.Code, rr.Body.String())
	}
	food := doJSON(t, h, http.MethodPost, "/api/v1/nutrition/foods/custom", map[string]any{
		"name": "Release Test Yogurt", "brand": "QA", "kcal_per_100g": 70, "protein_per_100g": 10, "fat_per_100g": 2, "carbs_per_100g": 4, "fiber_per_100g": 0, "serving_g": 200,
	}, access)
	if food.Code != http.StatusCreated {
		t.Fatalf("postgres custom food=%d body=%s", food.Code, food.Body.String())
	}
	var foodBody struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(food.Body.Bytes(), &foodBody); err != nil || foodBody.ID == "" {
		t.Fatalf("postgres custom food decode err=%v body=%s", err, food.Body.String())
	}
	entry := doJSON(t, h, http.MethodPost, "/api/v1/nutrition/entries", map[string]any{
		"food_id": foodBody.ID, "meal_type": "breakfast", "quantity_g": 200,
	}, access)
	if entry.Code != http.StatusCreated {
		t.Fatalf("postgres nutrition entry=%d body=%s", entry.Code, entry.Body.String())
	}
	if rr := doJSON(t, h, http.MethodGet, "/api/v1/nutrition/today", nil, access); rr.Code != http.StatusOK || !bytes.Contains(rr.Body.Bytes(), []byte("Release Test Yogurt")) {
		t.Fatalf("postgres nutrition today=%d body=%s", rr.Code, rr.Body.String())
	}

	// Health -> Recovery persistence and source isolation.
	date := time.Now().Format("2006-01-02")
	health := doJSON(t, h, http.MethodPut, "/api/v1/health/snapshots", map[string]any{
		"date": date, "provider": "health_connect", "source_package": "com.xiaomi.wearable", "source_label": "Mi Fitness (Xiaomi Wear)",
		"steps": 10420, "sleep_minutes": 438, "exercise_minutes": 45,
		"data_types": []string{"steps", "sleep", "exercise"}, "captured_at": time.Now().UTC().Format(time.RFC3339),
	}, access)
	if health.Code != http.StatusOK {
		t.Fatalf("postgres health import=%d body=%s", health.Code, health.Body.String())
	}
	recovery := doJSON(t, h, http.MethodPut, "/api/v1/recovery/check-in", map[string]any{
		"date": date, "sleep_hours": 4, "sleep_quality": 4, "energy": 4, "stress": 2, "muscle_soreness": map[string]int{"quads": 1},
	}, access)
	if recovery.Code != http.StatusOK || !bytes.Contains(recovery.Body.Bytes(), []byte(`"sleep_source":"wearable"`)) {
		t.Fatalf("postgres recovery=%d body=%s", recovery.Code, recovery.Body.String())
	}

	other := registerAndOnboard(t, h, fmt.Sprintf("pg-release-other-%d@example.com", suffix))
	otherHealth := doJSON(t, h, http.MethodGet, "/api/v1/health/today?date="+date, nil, other)
	if otherHealth.Code != http.StatusNoContent {
		t.Fatalf("cross-user health isolation expected 204 got=%d body=%s", otherHealth.Code, otherHealth.Body.String())
	}
	otherSearch := doJSON(t, h, http.MethodGet, "/api/v1/nutrition/foods?query=Release%20Test%20Yogurt", nil, other)
	if otherSearch.Code != http.StatusOK || bytes.Contains(otherSearch.Body.Bytes(), []byte("Release Test Yogurt")) {
		t.Fatalf("cross-user custom food leaked status=%d body=%s", otherSearch.Code, otherSearch.Body.String())
	}
}
