package httpapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/aifitness"
	"github.com/example/ai-fitness-os/services/api/internal/auth"
	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func TestRegistrationAndOnboardingFlow(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)

	register := doJSON(t, h, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": "nikita@example.com", "password": "strong-pass-123",
	}, "")
	if register.Code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", register.Code, register.Body.String())
	}

	var authBody struct {
		Tokens auth.Tokens `json:"tokens"`
	}
	if err := json.Unmarshal(register.Body.Bytes(), &authBody); err != nil {
		t.Fatal(err)
	}
	if authBody.Tokens.AccessToken == "" || authBody.Tokens.RefreshToken == "" {
		t.Fatal("expected tokens")
	}

	access := authBody.Tokens.AccessToken
	goal := doJSON(t, h, http.MethodPut, "/api/v1/profile/goal", map[string]any{"goal_type": "muscle_gain"}, access)
	if goal.Code != http.StatusOK {
		t.Fatalf("goal status=%d body=%s", goal.Code, goal.Body.String())
	}

	profile := doJSON(t, h, http.MethodPatch, "/api/v1/profile", map[string]any{
		"height_cm": 193, "weight_kg": 97, "experience_level": "intermediate", "unit_system": "metric",
		"age_years": 30, "injuries": []string{"Операция на плече"}, "limitations": []string{"Без болезненных разведений"},
	}, access)
	if profile.Code != http.StatusOK {
		t.Fatalf("profile status=%d body=%s", profile.Code, profile.Body.String())
	}
	loaded := doJSON(t, h, http.MethodGet, "/api/v1/profile", nil, access)
	if loaded.Code != http.StatusOK || !strings.Contains(loaded.Body.String(), "Операция на плече") || !strings.Contains(loaded.Body.String(), `"age_years":30`) {
		t.Fatalf("profile did not round-trip athlete details: status=%d body=%s", loaded.Code, loaded.Body.String())
	}

	prefs := doJSON(t, h, http.MethodPut, "/api/v1/profile/training-preferences", map[string]any{
		"environments": []string{"gym", "band"}, "equipment_ids": []string{"barbell", "dumbbells", "resistance_band"}, "workouts_per_week": 5, "session_minutes": 60,
	}, access)
	if prefs.Code != http.StatusOK {
		t.Fatalf("prefs status=%d body=%s", prefs.Code, prefs.Body.String())
	}

	complete := doJSON(t, h, http.MethodPost, "/api/v1/onboarding/complete", map[string]any{}, access)
	if complete.Code != http.StatusOK {
		t.Fatalf("complete status=%d body=%s", complete.Code, complete.Body.String())
	}

	status := doJSON(t, h, http.MethodGet, "/api/v1/onboarding/status", nil, access)
	if status.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", status.Code, status.Body.String())
	}
	var state store.OnboardingStatus
	if err := json.Unmarshal(status.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if !state.Completed {
		t.Fatalf("expected completed onboarding: %+v", state)
	}

	refresh := doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{"refresh_token": authBody.Tokens.RefreshToken}, "")
	if refresh.Code != http.StatusOK {
		t.Fatalf("refresh status=%d body=%s", refresh.Code, refresh.Body.String())
	}

	reused := doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{"refresh_token": authBody.Tokens.RefreshToken}, "")
	if reused.Code != http.StatusUnauthorized {
		t.Fatalf("expected rotated token to be rejected, got=%d", reused.Code)
	}
}

func TestAthleteProfileValidation(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	register := doJSON(t, h, http.MethodPost, "/api/v1/auth/register", map[string]any{"email": "profile-validation@example.com", "password": "strong-pass-123"}, "")
	var authBody struct {
		Tokens auth.Tokens `json:"tokens"`
	}
	if err := json.Unmarshal(register.Body.Bytes(), &authBody); err != nil {
		t.Fatal(err)
	}

	badAge := doJSON(t, h, http.MethodPatch, "/api/v1/profile", map[string]any{"age_years": 10}, authBody.Tokens.AccessToken)
	if badAge.Code != http.StatusUnprocessableEntity {
		t.Fatalf("bad age status=%d body=%s", badAge.Code, badAge.Body.String())
	}
	badNote := doJSON(t, h, http.MethodPatch, "/api/v1/profile", map[string]any{"injuries": []string{""}}, authBody.Tokens.AccessToken)
	if badNote.Code != http.StatusUnprocessableEntity {
		t.Fatalf("bad injury status=%d body=%s", badNote.Code, badNote.Body.String())
	}
}

func TestAthleteProfileAtomicUpdateAndValidationRollback(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	access := registerAndOnboard(t, h, "atomic-profile@example.com")

	valid := map[string]any{
		"profile": map[string]any{
			"user_id": "attacker-controlled", "height_cm": 193, "weight_kg": 97, "age_years": 30,
			"experience_level": "intermediate", "injuries": []string{"shoulder"}, "limitations": []string{"no flyes"}, "unit_system": "metric",
		},
		"goal": map[string]any{"user_id": "attacker-controlled", "goal_type": "strength"},
		"training_preferences": map[string]any{
			"user_id": "attacker-controlled", "environments": []string{"gym"}, "equipment_ids": []string{"barbell"}, "workouts_per_week": 4, "session_minutes": 75,
		},
	}
	saved := doJSON(t, h, http.MethodPut, "/api/v1/profile/athlete", valid, access)
	if saved.Code != http.StatusOK {
		t.Fatalf("atomic save status=%d body=%s", saved.Code, saved.Body.String())
	}
	if strings.Contains(saved.Body.String(), "attacker-controlled") {
		t.Fatalf("request user_id leaked into persisted aggregate: %s", saved.Body.String())
	}

	invalid := map[string]any{
		"profile": map[string]any{
			"height_cm": 180, "weight_kg": 80, "age_years": 31, "experience_level": "advanced",
			"injuries": []string{}, "limitations": []string{}, "unit_system": "metric",
		},
		"goal": map[string]any{"goal_type": "fat_loss"},
		"training_preferences": map[string]any{
			"environments": []string{"gym"}, "equipment_ids": []string{"not-real"}, "workouts_per_week": 2, "session_minutes": 45,
		},
	}
	rejected := doJSON(t, h, http.MethodPut, "/api/v1/profile/athlete", invalid, access)
	if rejected.Code != http.StatusUnprocessableEntity {
		t.Fatalf("invalid aggregate status=%d body=%s", rejected.Code, rejected.Body.String())
	}
	loaded := doJSON(t, h, http.MethodGet, "/api/v1/profile", nil, access)
	if loaded.Code != http.StatusOK || !strings.Contains(loaded.Body.String(), `"height_cm":193`) || !strings.Contains(loaded.Body.String(), `"goal_type":"strength"`) || !strings.Contains(loaded.Body.String(), `"workouts_per_week":4`) {
		t.Fatalf("invalid aggregate partially changed stored profile: status=%d body=%s", loaded.Code, loaded.Body.String())
	}
}

func TestProtectedRouteRejectsMissingToken(t *testing.T) {
	h := NewServer()
	rr := doJSON(t, h, http.MethodGet, "/api/v1/profile", nil, "")
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func doJSON(t *testing.T, h http.Handler, method, path string, payload any, access string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if access != "" {
		req.Header.Set("Authorization", "Bearer "+access)
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	return rr
}

func TestWorkoutLifecycleAndProgression(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	access := registerAndOnboard(t, h, "lifter@example.com")

	generated := doJSON(t, h, http.MethodPost, "/api/v1/workouts/generate", map[string]any{
		"muscle": "chest", "environment": "gym", "duration_minutes": 45,
	}, access)
	if generated.Code != http.StatusCreated {
		t.Fatalf("generate status=%d body=%s", generated.Code, generated.Body.String())
	}
	var workout struct {
		Workout struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"workout"`
		Exercises []struct {
			WorkoutExercise struct {
				ID            string   `json:"id"`
				TargetRepsMax int      `json:"target_reps_max"`
				TargetWeight  *float64 `json:"target_weight"`
			} `json:"workout_exercise"`
			Exercise struct {
				ID string `json:"id"`
			} `json:"exercise"`
		} `json:"exercises"`
	}
	if err := json.Unmarshal(generated.Body.Bytes(), &workout); err != nil {
		t.Fatal(err)
	}
	if workout.Workout.ID == "" || workout.Workout.Status != "planned" || len(workout.Exercises) == 0 {
		t.Fatalf("unexpected generated workout: %+v", workout)
	}

	started := doJSON(t, h, http.MethodPost, "/api/v1/workouts/"+workout.Workout.ID+"/start", map[string]any{}, access)
	if started.Code != http.StatusOK {
		t.Fatalf("start status=%d body=%s", started.Code, started.Body.String())
	}

	active := doJSON(t, h, http.MethodGet, "/api/v1/workouts/active", nil, access)
	if active.Code != http.StatusOK || !bytes.Contains(active.Body.Bytes(), []byte(workout.Workout.ID)) {
		t.Fatalf("active status=%d body=%s", active.Code, active.Body.String())
	}
	if !bytes.Contains(active.Body.Bytes(), []byte(`"instructions"`)) || !bytes.Contains(active.Body.Bytes(), []byte(`"common_mistakes"`)) {
		t.Fatalf("expected exercise guidance in active workout body=%s", active.Body.String())
	}

	first := workout.Exercises[0]
	weight := 80.0
	logged := doJSON(t, h, http.MethodPut, "/api/v1/workouts/"+workout.Workout.ID+"/sets", map[string]any{
		"workout_exercise_id": first.WorkoutExercise.ID,
		"set_number":          1,
		"weight":              weight,
		"repetitions":         first.WorkoutExercise.TargetRepsMax,
		"rir":                 2,
	}, access)
	if logged.Code != http.StatusOK {
		t.Fatalf("set status=%d body=%s", logged.Code, logged.Body.String())
	}

	finished := doJSON(t, h, http.MethodPost, "/api/v1/workouts/"+workout.Workout.ID+"/finish", map[string]any{}, access)
	if finished.Code != http.StatusOK {
		t.Fatalf("finish status=%d body=%s", finished.Code, finished.Body.String())
	}
	if !bytes.Contains(finished.Body.Bytes(), []byte(`"ended_early":true`)) {
		t.Fatalf("partial workout should be marked ended_early body=%s", finished.Body.String())
	}

	noActive := doJSON(t, h, http.MethodGet, "/api/v1/workouts/active", nil, access)
	if noActive.Code != http.StatusNoContent {
		t.Fatalf("expected no active workout after finish, got=%d body=%s", noActive.Code, noActive.Body.String())
	}

	history := doJSON(t, h, http.MethodGet, "/api/v1/workouts/history", nil, access)
	if history.Code != http.StatusOK {
		t.Fatalf("history status=%d body=%s", history.Code, history.Body.String())
	}

	next := doJSON(t, h, http.MethodPost, "/api/v1/workouts/generate", map[string]any{
		"muscle": "chest", "environment": "gym", "duration_minutes": 45,
	}, access)
	if next.Code != http.StatusCreated {
		t.Fatalf("next generate status=%d body=%s", next.Code, next.Body.String())
	}
	var nextWorkout struct {
		Exercises []struct {
			WorkoutExercise struct {
				TargetWeight *float64 `json:"target_weight"`
			} `json:"workout_exercise"`
			Exercise struct {
				ID string `json:"id"`
			} `json:"exercise"`
		} `json:"exercises"`
	}
	if err := json.Unmarshal(next.Body.Bytes(), &nextWorkout); err != nil {
		t.Fatal(err)
	}
	foundProgression := false
	for _, ex := range nextWorkout.Exercises {
		if ex.Exercise.ID == first.Exercise.ID && ex.WorkoutExercise.TargetWeight != nil && *ex.WorkoutExercise.TargetWeight > weight {
			foundProgression = true
		}
	}
	if !foundProgression {
		t.Fatalf("expected increased target weight for %s; body=%s", first.Exercise.ID, next.Body.String())
	}
}

func TestReplaceExerciseBeforeSets(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	access := registerAndOnboard(t, h, "replace@example.com")
	generated := doJSON(t, h, http.MethodPost, "/api/v1/workouts/generate", map[string]any{
		"muscle": "back", "environment": "gym", "duration_minutes": 45,
	}, access)
	if generated.Code != http.StatusCreated {
		t.Fatalf("generate status=%d body=%s", generated.Code, generated.Body.String())
	}
	var body struct {
		Workout struct {
			ID string `json:"id"`
		} `json:"workout"`
		Exercises []struct {
			WorkoutExercise struct {
				ID string `json:"id"`
			} `json:"workout_exercise"`
			Exercise struct {
				ID string `json:"id"`
			} `json:"exercise"`
		} `json:"exercises"`
	}
	_ = json.Unmarshal(generated.Body.Bytes(), &body)
	before := body.Exercises[0].Exercise.ID
	replaced := doJSON(t, h, http.MethodPost, "/api/v1/workouts/"+body.Workout.ID+"/exercises/"+body.Exercises[0].WorkoutExercise.ID+"/replace", map[string]any{"reason": "busy"}, access)
	if replaced.Code != http.StatusOK {
		t.Fatalf("replace status=%d body=%s", replaced.Code, replaced.Body.String())
	}
	if bytes.Contains(replaced.Body.Bytes(), []byte(`"exercise":{"id":"`+before+`"`)) {
		t.Fatalf("expected first exercise to change; body=%s", replaced.Body.String())
	}
}

func registerAndOnboard(t *testing.T, h http.Handler, email string) string {
	t.Helper()
	register := doJSON(t, h, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": email, "password": "strong-pass-123",
	}, "")
	if register.Code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", register.Code, register.Body.String())
	}
	var authBody struct {
		Tokens auth.Tokens `json:"tokens"`
	}
	_ = json.Unmarshal(register.Body.Bytes(), &authBody)
	access := authBody.Tokens.AccessToken
	_ = doJSON(t, h, http.MethodPut, "/api/v1/profile/goal", map[string]any{"goal_type": "muscle_gain"}, access)
	_ = doJSON(t, h, http.MethodPatch, "/api/v1/profile", map[string]any{
		"height_cm": 193, "weight_kg": 97, "experience_level": "intermediate", "unit_system": "metric",
	}, access)
	_ = doJSON(t, h, http.MethodPut, "/api/v1/profile/training-preferences", map[string]any{
		"environments":      []string{"gym", "home", "band"},
		"equipment_ids":     []string{"barbell", "bench", "rack", "dumbbells", "pullup_bar", "cable_machine", "leg_machine", "bodyweight", "resistance_band"},
		"workouts_per_week": 5, "session_minutes": 60,
	}, access)
	complete := doJSON(t, h, http.MethodPost, "/api/v1/onboarding/complete", map[string]any{}, access)
	if complete.Code != http.StatusOK {
		t.Fatalf("onboarding status=%d body=%s", complete.Code, complete.Body.String())
	}
	return access
}

func TestSprint1DInsightsFavoriteRepeatAndFilters(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	access := registerAndOnboard(t, h, "sprint1d@example.com")

	generated := doJSON(t, h, http.MethodPost, "/api/v1/workouts/generate", map[string]any{
		"muscle": "back", "environment": "gym", "duration_minutes": 45,
	}, access)
	if generated.Code != http.StatusCreated {
		t.Fatalf("generate status=%d body=%s", generated.Code, generated.Body.String())
	}
	var body struct {
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
	if err := json.Unmarshal(generated.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	_ = doJSON(t, h, http.MethodPost, "/api/v1/workouts/"+body.Workout.ID+"/start", map[string]any{}, access)
	first := body.Exercises[0]
	logged := doJSON(t, h, http.MethodPut, "/api/v1/workouts/"+body.Workout.ID+"/sets", map[string]any{
		"workout_exercise_id": first.WorkoutExercise.ID,
		"set_number":          1,
		"weight":              70,
		"repetitions":         first.WorkoutExercise.TargetRepsMax,
		"rir":                 2,
	}, access)
	if logged.Code != http.StatusOK {
		t.Fatalf("log set status=%d body=%s", logged.Code, logged.Body.String())
	}

	finished := doJSON(t, h, http.MethodPost, "/api/v1/workouts/"+body.Workout.ID+"/finish", map[string]any{}, access)
	if finished.Code != http.StatusOK {
		t.Fatalf("finish status=%d body=%s", finished.Code, finished.Body.String())
	}
	if !bytes.Contains(finished.Body.Bytes(), []byte(`"record_type":"max_weight"`)) || !bytes.Contains(finished.Body.Bytes(), []byte(`"record_type":"estimated_1rm"`)) {
		t.Fatalf("expected personal records body=%s", finished.Body.String())
	}

	favorite := doJSON(t, h, http.MethodPatch, "/api/v1/workouts/"+body.Workout.ID+"/favorite", map[string]any{"favorite": true}, access)
	if favorite.Code != http.StatusOK || !bytes.Contains(favorite.Body.Bytes(), []byte(`"favorite":true`)) {
		t.Fatalf("favorite status=%d body=%s", favorite.Code, favorite.Body.String())
	}

	filtered := doJSON(t, h, http.MethodGet, "/api/v1/workouts/history?muscle=back&environment=gym&status=completed&favorite=true", nil, access)
	if filtered.Code != http.StatusOK || !bytes.Contains(filtered.Body.Bytes(), []byte(body.Workout.ID)) {
		t.Fatalf("filtered history status=%d body=%s", filtered.Code, filtered.Body.String())
	}

	stats := doJSON(t, h, http.MethodGet, "/api/v1/muscles/back/stats", nil, access)
	if stats.Code != http.StatusOK || !bytes.Contains(stats.Body.Bytes(), []byte(`"completed_workouts":1`)) {
		t.Fatalf("muscle stats status=%d body=%s", stats.Code, stats.Body.String())
	}

	records := doJSON(t, h, http.MethodGet, "/api/v1/records?limit=10", nil, access)
	if records.Code != http.StatusOK || !bytes.Contains(records.Body.Bytes(), []byte(`"exercise_id"`)) {
		t.Fatalf("records status=%d body=%s", records.Code, records.Body.String())
	}

	repeated := doJSON(t, h, http.MethodPost, "/api/v1/workouts/"+body.Workout.ID+"/repeat", map[string]any{}, access)
	if repeated.Code != http.StatusCreated || !bytes.Contains(repeated.Body.Bytes(), []byte(`"status":"planned"`)) {
		t.Fatalf("repeat status=%d body=%s", repeated.Code, repeated.Body.String())
	}
}

func TestNutritionCoreFlow(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	access := registerAndOnboard(t, h, "nutrition@example.com")

	missing := doJSON(t, h, http.MethodGet, "/api/v1/nutrition/profile", nil, access)
	if missing.Code != http.StatusNoContent {
		t.Fatalf("expected empty nutrition profile, got=%d body=%s", missing.Code, missing.Body.String())
	}

	setup := doJSON(t, h, http.MethodPut, "/api/v1/nutrition/profile", map[string]any{
		"goal": "recomp", "activity_level": "moderate", "calculation_mode": "auto",
	}, access)
	if setup.Code != http.StatusOK {
		t.Fatalf("nutrition setup status=%d body=%s", setup.Code, setup.Body.String())
	}
	if !bytes.Contains(setup.Body.Bytes(), []byte(`"calorie_target"`)) {
		t.Fatalf("expected calculated targets body=%s", setup.Body.String())
	}

	foods := doJSON(t, h, http.MethodGet, "/api/v1/nutrition/foods?query=творог", nil, access)
	if foods.Code != http.StatusOK || !bytes.Contains(foods.Body.Bytes(), []byte("cottage-cheese-5")) {
		t.Fatalf("food search status=%d body=%s", foods.Code, foods.Body.String())
	}

	logged := doJSON(t, h, http.MethodPost, "/api/v1/nutrition/entries", map[string]any{
		"food_id": "cottage-cheese-5", "meal_type": "breakfast", "quantity_g": 200,
	}, access)
	if logged.Code != http.StatusCreated {
		t.Fatalf("log food status=%d body=%s", logged.Code, logged.Body.String())
	}
	var day struct {
		ConsumedCalories float64 `json:"consumed_calories"`
		ConsumedProtein  float64 `json:"consumed_protein_g"`
		Entries          []struct {
			ID string `json:"id"`
		} `json:"entries"`
	}
	if err := json.Unmarshal(logged.Body.Bytes(), &day); err != nil {
		t.Fatal(err)
	}
	if day.ConsumedCalories != 242 || day.ConsumedProtein != 34 || len(day.Entries) != 1 {
		t.Fatalf("unexpected nutrition day: %+v body=%s", day, logged.Body.String())
	}

	history := doJSON(t, h, http.MethodGet, "/api/v1/nutrition/history?days=7", nil, access)
	if history.Code != http.StatusOK || !bytes.Contains(history.Body.Bytes(), []byte(`"target_calories"`)) {
		t.Fatalf("nutrition history status=%d body=%s", history.Code, history.Body.String())
	}

	deleted := doJSON(t, h, http.MethodDelete, "/api/v1/nutrition/entries/"+day.Entries[0].ID, nil, access)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete entry status=%d body=%s", deleted.Code, deleted.Body.String())
	}

	today := doJSON(t, h, http.MethodGet, "/api/v1/nutrition/today", nil, access)
	if today.Code != http.StatusOK || !bytes.Contains(today.Body.Bytes(), []byte(`"consumed_calories":0`)) {
		t.Fatalf("today after delete status=%d body=%s", today.Code, today.Body.String())
	}
}

func TestNutrition2BAndProgressFlow(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	access := registerAndOnboard(t, h, "nutrition2b@example.com")

	setup := doJSON(t, h, http.MethodPut, "/api/v1/nutrition/profile", map[string]any{"goal": "maintain", "activity_level": "moderate", "calculation_mode": "auto"}, access)
	if setup.Code != http.StatusOK {
		t.Fatalf("nutrition setup=%d body=%s", setup.Code, setup.Body.String())
	}

	custom := doJSON(t, h, http.MethodPost, "/api/v1/nutrition/foods/custom", map[string]any{
		"name": "Мой йогурт", "brand": "Домашний", "barcode": "4601234567890", "kcal_per_100g": 90, "protein_per_100g": 12, "fat_per_100g": 2, "carbs_per_100g": 5, "fiber_per_100g": 0, "serving_g": 180,
	}, access)
	if custom.Code != http.StatusCreated {
		t.Fatalf("custom food=%d body=%s", custom.Code, custom.Body.String())
	}
	var food store.FoodItem
	if err := json.Unmarshal(custom.Body.Bytes(), &food); err != nil {
		t.Fatal(err)
	}
	if food.ID == "" || food.Barcode != "4601234567890" {
		t.Fatalf("bad custom food %+v", food)
	}

	otherAccess := registerAndOnboard(t, h, "nutrition2b-other@example.com")
	privateSearch := doJSON(t, h, http.MethodGet, "/api/v1/nutrition/foods?query="+"%D0%9C%D0%BE%D0%B9%20%D0%B9%D0%BE%D0%B3%D1%83%D1%80%D1%82", nil, otherAccess)
	if privateSearch.Code != http.StatusOK || bytes.Contains(privateSearch.Body.Bytes(), []byte(food.ID)) {
		t.Fatalf("custom food leaked to another user status=%d body=%s", privateSearch.Code, privateSearch.Body.String())
	}
	privateLog := doJSON(t, h, http.MethodPost, "/api/v1/nutrition/entries", map[string]any{"food_id": food.ID, "meal_type": "snack", "quantity_g": 100}, otherAccess)
	if privateLog.Code != http.StatusNotFound {
		t.Fatalf("expected foreign custom food to be inaccessible, got=%d body=%s", privateLog.Code, privateLog.Body.String())
	}

	barcode := doJSON(t, h, http.MethodGet, "/api/v1/nutrition/foods/barcode/4601234567890", nil, access)
	if barcode.Code != http.StatusOK || !bytes.Contains(barcode.Body.Bytes(), []byte("Мой йогурт")) {
		t.Fatalf("barcode=%d body=%s", barcode.Code, barcode.Body.String())
	}

	recipe := doJSON(t, h, http.MethodPost, "/api/v1/nutrition/recipes", map[string]any{"name": "Белковый завтрак", "items": []map[string]any{{"food_id": food.ID, "quantity_g": 180}, {"food_id": "banana", "quantity_g": 120}}}, access)
	if recipe.Code != http.StatusCreated {
		t.Fatalf("recipe=%d body=%s", recipe.Code, recipe.Body.String())
	}
	var rec store.Recipe
	if err := json.Unmarshal(recipe.Body.Bytes(), &rec); err != nil {
		t.Fatal(err)
	}
	if rec.ID == "" || len(rec.Items) != 2 {
		t.Fatalf("bad recipe %+v", rec)
	}

	logged := doJSON(t, h, http.MethodPost, "/api/v1/nutrition/recipes/"+rec.ID+"/log", map[string]any{"meal_type": "breakfast", "scale": 1}, access)
	if logged.Code != http.StatusCreated {
		t.Fatalf("log recipe=%d body=%s", logged.Code, logged.Body.String())
	}
	var day struct {
		Entries []store.FoodEntry `json:"entries"`
	}
	if err := json.Unmarshal(logged.Body.Bytes(), &day); err != nil {
		t.Fatal(err)
	}
	if len(day.Entries) != 2 {
		t.Fatalf("expected recipe entries body=%s", logged.Body.String())
	}

	repeated := doJSON(t, h, http.MethodPost, "/api/v1/nutrition/entries/"+day.Entries[0].ID+"/repeat", map[string]any{"meal_type": "snack"}, access)
	if repeated.Code != http.StatusCreated {
		t.Fatalf("repeat=%d body=%s", repeated.Code, repeated.Body.String())
	}

	measurement := doJSON(t, h, http.MethodPost, "/api/v1/progress/measurements", map[string]any{"weight_kg": 96.4, "waist_cm": 88.5, "arm_cm": 40.2}, access)
	if measurement.Code != http.StatusCreated {
		t.Fatalf("measurement=%d body=%s", measurement.Code, measurement.Body.String())
	}
	measurement2 := doJSON(t, h, http.MethodPost, "/api/v1/progress/measurements", map[string]any{"weight_kg": 95.9, "waist_cm": 87.9}, access)
	if measurement2.Code != http.StatusCreated {
		t.Fatalf("measurement2=%d body=%s", measurement2.Code, measurement2.Body.String())
	}

	summary := doJSON(t, h, http.MethodGet, "/api/v1/progress/summary?days=30", nil, access)
	if summary.Code != http.StatusOK || !bytes.Contains(summary.Body.Bytes(), []byte(`"weight_delta_kg":-0.5`)) {
		t.Fatalf("summary=%d body=%s", summary.Code, summary.Body.String())
	}

	correlation := doJSON(t, h, http.MethodGet, "/api/v1/nutrition/correlation?days=30", nil, access)
	if correlation.Code != http.StatusOK || !bytes.Contains(correlation.Body.Bytes(), []byte(`"logged_rest_days":1`)) {
		t.Fatalf("correlation=%d body=%s", correlation.Code, correlation.Body.String())
	}
}

type fakeAIProvider struct{}

func (fakeAIProvider) Name() string  { return "fake" }
func (fakeAIProvider) Model() string { return "fake-test-model" }
func (fakeAIProvider) ExtractFoodText(_ context.Context, _ string) ([]aifitness.ExtractedFood, error) {
	return []aifitness.ExtractedFood{{Name: "Творог", QuantityG: 200, Confidence: .96, Notes: "explicit"}, {Name: "Банан", QuantityG: 120, Confidence: .88, Notes: "estimated one banana"}}, nil
}
func (fakeAIProvider) ExtractFoodImage(_ context.Context, _ string) ([]aifitness.ExtractedFood, error) {
	return []aifitness.ExtractedFood{{Name: "Куриная грудка", QuantityG: 160, Confidence: .8, Notes: "estimated from image"}}, nil
}
func (fakeAIProvider) Coach(ctx context.Context, req aifitness.CoachRequest) (aifitness.CoachResponse, error) {
	result, err := req.ExecuteTool(ctx, "get_today_nutrition", json.RawMessage(`{}`))
	if err != nil {
		return aifitness.CoachResponse{}, err
	}
	return aifitness.CoachResponse{Message: "Контекст питания получен: " + result, Model: "fake-test-model", Provider: "fake", ToolCalls: []string{"get_today_nutrition"}}, nil
}
func (fakeAIProvider) WeeklyReport(_ context.Context, stats aifitness.WeeklyStats) (aifitness.WeeklyReport, error) {
	return aifitness.WeeklyReport{Stats: stats, Summary: "Тестовый недельный отчёт", Wins: []string{"Дневник ведётся"}, Focus: []string{"Стабильность"}, NextActions: []string{"Продолжить"}, Model: "fake-test-model", Provider: "fake"}, nil
}

func TestSprint2CAIFlow(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithAI(st, tm, fakeAIProvider{})
	access := registerAndOnboard(t, h, "ai2c@example.com")
	setup := doJSON(t, h, http.MethodPut, "/api/v1/nutrition/profile", map[string]any{"goal": "recomp", "activity_level": "moderate", "calculation_mode": "auto"}, access)
	if setup.Code != http.StatusOK {
		t.Fatalf("nutrition setup=%d body=%s", setup.Code, setup.Body.String())
	}

	parsed := doJSON(t, h, http.MethodPost, "/api/v1/ai/food/parse", map[string]any{"text": "Съел 200 г творога и банан"}, access)
	if parsed.Code != http.StatusOK {
		t.Fatalf("parse=%d body=%s", parsed.Code, parsed.Body.String())
	}
	var draft aifitness.FoodDraft
	if err := json.Unmarshal(parsed.Body.Bytes(), &draft); err != nil {
		t.Fatal(err)
	}
	if len(draft.Items) != 2 || draft.Items[0].Matched == nil || draft.Items[1].Matched == nil {
		t.Fatalf("unexpected draft %+v body=%s", draft, parsed.Body.String())
	}

	confirmed := doJSON(t, h, http.MethodPost, "/api/v1/ai/food/confirm", map[string]any{"meal_type": "breakfast", "items": []map[string]any{{"food_id": draft.Items[0].Matched.ID, "quantity_g": draft.Items[0].QuantityG}, {"food_id": draft.Items[1].Matched.ID, "quantity_g": draft.Items[1].QuantityG}}}, access)
	if confirmed.Code != http.StatusCreated || !bytes.Contains(confirmed.Body.Bytes(), []byte(`"entries"`)) {
		t.Fatalf("confirm=%d body=%s", confirmed.Code, confirmed.Body.String())
	}

	photo := doJSON(t, h, http.MethodPost, "/api/v1/ai/food/photo", map[string]any{"image_data_url": "data:image/jpeg;base64,ZmFrZQ=="}, access)
	if photo.Code != http.StatusOK || !bytes.Contains(photo.Body.Bytes(), []byte("Куриная грудка")) {
		t.Fatalf("photo=%d body=%s", photo.Code, photo.Body.String())
	}

	// Photo payloads are allowed above the normal 1 MiB JSON limit. The real mobile bridge
	// caps raw images at 5 MiB; this regression check proves the dedicated endpoint limit is used.
	largePhoto := doJSON(t, h, http.MethodPost, "/api/v1/ai/food/photo", map[string]any{"image_data_url": "data:image/jpeg;base64," + strings.Repeat("A", 2<<20)}, access)
	if largePhoto.Code != http.StatusOK {
		t.Fatalf("large photo=%d body=%s", largePhoto.Code, largePhoto.Body.String())
	}

	chat := doJSON(t, h, http.MethodPost, "/api/v1/ai/chat", map[string]any{"message": "Сколько белка осталось?"}, access)
	if chat.Code != http.StatusOK || !bytes.Contains(chat.Body.Bytes(), []byte("get_today_nutrition")) {
		t.Fatalf("chat=%d body=%s", chat.Code, chat.Body.String())
	}

	report := doJSON(t, h, http.MethodGet, "/api/v1/ai/reports/weekly", nil, access)
	if report.Code != http.StatusOK || !bytes.Contains(report.Body.Bytes(), []byte("Тестовый недельный отчёт")) {
		t.Fatalf("report=%d body=%s", report.Code, report.Body.String())
	}

	status := doJSON(t, h, http.MethodGet, "/api/v1/ai/status", nil, access)
	if status.Code != http.StatusOK || !bytes.Contains(status.Body.Bytes(), []byte(`"provider":"fake"`)) {
		t.Fatalf("status=%d body=%s", status.Code, status.Body.String())
	}
}

func TestSprint2DProgramLifecycle(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	access := registerAndOnboard(t, h, "program-api@example.com")
	start := time.Now().UTC().AddDate(0, 0, 1).Format(time.RFC3339)

	generated := doJSON(t, h, http.MethodPost, "/api/v1/programs/generate", map[string]any{
		"weeks": 4, "workouts_per_week": 3, "environment": "gym", "start_date": start,
	}, access)
	if generated.Code != http.StatusCreated {
		t.Fatalf("program generate=%d body=%s", generated.Code, generated.Body.String())
	}
	var program struct {
		Program struct {
			ID string `json:"id"`
		} `json:"program"`
		Sessions []store.ProgramSession `json:"sessions"`
	}
	if err := json.Unmarshal(generated.Body.Bytes(), &program); err != nil {
		t.Fatal(err)
	}
	if program.Program.ID == "" || len(program.Sessions) != 12 {
		t.Fatalf("unexpected program=%+v", program)
	}

	active := doJSON(t, h, http.MethodGet, "/api/v1/programs/active", nil, access)
	if active.Code != http.StatusOK {
		t.Fatalf("active=%d body=%s", active.Code, active.Body.String())
	}

	var deload *store.ProgramSession
	for i := range program.Sessions {
		if program.Sessions[i].IsDeload {
			cp := program.Sessions[i]
			deload = &cp
			break
		}
	}
	if deload == nil {
		t.Fatal("expected deload session")
	}

	createdWorkout := doJSON(t, h, http.MethodPost, "/api/v1/programs/sessions/"+deload.ID+"/workout", map[string]any{}, access)
	if createdWorkout.Code != http.StatusCreated {
		t.Fatalf("program workout=%d body=%s", createdWorkout.Code, createdWorkout.Body.String())
	}
	var linked struct {
		Session store.ProgramSession `json:"session"`
		Workout struct {
			Workout struct {
				ID string `json:"id"`
			} `json:"workout"`
			Exercises []struct {
				WorkoutExercise struct {
					TargetSets int `json:"target_sets"`
				} `json:"workout_exercise"`
			} `json:"exercises"`
		} `json:"workout"`
	}
	if err := json.Unmarshal(createdWorkout.Body.Bytes(), &linked); err != nil {
		t.Fatal(err)
	}
	if linked.Session.WorkoutID == nil || linked.Workout.Workout.ID == "" {
		t.Fatalf("workout not linked: %s", createdWorkout.Body.String())
	}
	for _, ex := range linked.Workout.Exercises {
		if ex.WorkoutExercise.TargetSets > 3 {
			t.Fatalf("expected deload-reduced sets, got %d", ex.WorkoutExercise.TargetSets)
		}
	}

	explained := doJSON(t, h, http.MethodPost, "/api/v1/programs/sessions/"+deload.ID+"/explain", map[string]any{}, access)
	if explained.Code != http.StatusOK || !bytes.Contains(explained.Body.Bytes(), []byte("explanation")) {
		t.Fatalf("explain=%d body=%s", explained.Code, explained.Body.String())
	}

	analytics := doJSON(t, h, http.MethodGet, "/api/v1/programs/"+program.Program.ID+"/analytics", nil, access)
	if analytics.Code != http.StatusOK || !bytes.Contains(analytics.Body.Bytes(), []byte("adherence_percent")) {
		t.Fatalf("analytics=%d body=%s", analytics.Code, analytics.Body.String())
	}
}

func TestWorkoutCancelAndSetValidation(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	access := registerAndOnboard(t, h, "cancel-workout@example.com")

	generated := doJSON(t, h, http.MethodPost, "/api/v1/workouts/generate", map[string]any{
		"muscle": "back", "environment": "gym", "duration_minutes": 45,
	}, access)
	if generated.Code != http.StatusCreated {
		t.Fatalf("generate status=%d body=%s", generated.Code, generated.Body.String())
	}
	var body struct {
		Workout struct {
			ID string `json:"id"`
		} `json:"workout"`
		Exercises []struct {
			WorkoutExercise struct {
				ID string `json:"id"`
			} `json:"workout_exercise"`
		} `json:"exercises"`
	}
	if err := json.Unmarshal(generated.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Exercises) == 0 {
		t.Fatal("expected exercises")
	}

	started := doJSON(t, h, http.MethodPost, "/api/v1/workouts/"+body.Workout.ID+"/start", map[string]any{}, access)
	if started.Code != http.StatusOK {
		t.Fatalf("start=%d body=%s", started.Code, started.Body.String())
	}

	zeroRep := doJSON(t, h, http.MethodPut, "/api/v1/workouts/"+body.Workout.ID+"/sets", map[string]any{
		"workout_exercise_id": body.Exercises[0].WorkoutExercise.ID,
		"set_number":          1, "repetitions": 0, "rir": 2,
	}, access)
	if zeroRep.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected zero reps to be rejected, got=%d body=%s", zeroRep.Code, zeroRep.Body.String())
	}

	cancelled := doJSON(t, h, http.MethodPost, "/api/v1/workouts/"+body.Workout.ID+"/cancel", map[string]any{}, access)
	if cancelled.Code != http.StatusOK || !bytes.Contains(cancelled.Body.Bytes(), []byte(`"status":"cancelled"`)) {
		t.Fatalf("cancel=%d body=%s", cancelled.Code, cancelled.Body.String())
	}

	active := doJSON(t, h, http.MethodGet, "/api/v1/workouts/active", nil, access)
	if active.Code != http.StatusNoContent {
		t.Fatalf("cancelled workout must not remain active: %d body=%s", active.Code, active.Body.String())
	}

	finish := doJSON(t, h, http.MethodPost, "/api/v1/workouts/"+body.Workout.ID+"/finish", map[string]any{}, access)
	if finish.Code != http.StatusConflict {
		t.Fatalf("cancelled workout must not finish: %d body=%s", finish.Code, finish.Body.String())
	}
}

func TestCancellingProgramWorkoutAllowsRecreate(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	access := registerAndOnboard(t, h, "cancel-program-workout@example.com")

	generated := doJSON(t, h, http.MethodPost, "/api/v1/programs/generate", map[string]any{
		"weeks": 4, "workouts_per_week": 3, "environment": "gym",
	}, access)
	if generated.Code != http.StatusCreated {
		t.Fatalf("program=%d body=%s", generated.Code, generated.Body.String())
	}
	var program struct {
		Sessions []store.ProgramSession `json:"sessions"`
	}
	if err := json.Unmarshal(generated.Body.Bytes(), &program); err != nil {
		t.Fatal(err)
	}
	if len(program.Sessions) == 0 {
		t.Fatal("expected sessions")
	}
	sessionID := program.Sessions[0].ID

	first := doJSON(t, h, http.MethodPost, "/api/v1/programs/sessions/"+sessionID+"/workout", map[string]any{}, access)
	if first.Code != http.StatusCreated {
		t.Fatalf("first workout=%d body=%s", first.Code, first.Body.String())
	}
	var firstBody struct {
		Workout struct {
			Workout struct {
				ID string `json:"id"`
			} `json:"workout"`
		} `json:"workout"`
	}
	if err := json.Unmarshal(first.Body.Bytes(), &firstBody); err != nil {
		t.Fatal(err)
	}
	firstID := firstBody.Workout.Workout.ID
	if firstID == "" {
		t.Fatalf("missing workout id body=%s", first.Body.String())
	}

	cancelled := doJSON(t, h, http.MethodPost, "/api/v1/workouts/"+firstID+"/cancel", map[string]any{}, access)
	if cancelled.Code != http.StatusOK {
		t.Fatalf("cancel=%d body=%s", cancelled.Code, cancelled.Body.String())
	}

	second := doJSON(t, h, http.MethodPost, "/api/v1/programs/sessions/"+sessionID+"/workout", map[string]any{}, access)
	if second.Code != http.StatusCreated {
		t.Fatalf("recreate=%d body=%s", second.Code, second.Body.String())
	}
	var secondBody struct {
		Workout struct {
			Workout struct {
				ID string `json:"id"`
			} `json:"workout"`
		} `json:"workout"`
	}
	if err := json.Unmarshal(second.Body.Bytes(), &secondBody); err != nil {
		t.Fatal(err)
	}
	if secondBody.Workout.Workout.ID == "" || secondBody.Workout.Workout.ID == firstID {
		t.Fatalf("expected a new workout after cancellation first=%s second=%s body=%s", firstID, secondBody.Workout.Workout.ID, second.Body.String())
	}
}

func testBodyScanDataURL(t *testing.T) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 800, 1200))
	for y := 0; y < 1200; y++ {
		for x := 0; x < 800; x++ {
			v := uint8(55 + ((x + y) % 150))
			img.Set(x, y, color.RGBA{R: v, G: v + 10, B: v - 5, A: 255})
		}
	}
	var b bytes.Buffer
	if err := jpeg.Encode(&b, img, &jpeg.Options{Quality: 82}); err != nil {
		t.Fatal(err)
	}
	return "data:image/jpeg;base64," + base64.StdEncoding.EncodeToString(b.Bytes())
}

func TestBodyScanHTTPFlow(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	access := registerAndOnboard(t, h, "bodyscan@example.com")
	other := registerAndOnboard(t, h, "bodyscan-other@example.com")

	created := doJSON(t, h, http.MethodPost, "/api/v1/body-scans", map[string]any{}, access)
	if created.Code != http.StatusCreated {
		t.Fatalf("create=%d body=%s", created.Code, created.Body.String())
	}
	var scan struct {
		Scan struct {
			ID string `json:"id"`
		} `json:"scan"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &scan); err != nil {
		t.Fatal(err)
	}
	if scan.Scan.ID == "" {
		t.Fatal("missing scan id")
	}

	imageData := testBodyScanDataURL(t)
	for _, view := range []string{"front", "side", "back"} {
		rr := doJSON(t, h, http.MethodPut, "/api/v1/body-scans/"+scan.Scan.ID+"/photos/"+view, map[string]any{"image_data_url": imageData}, access)
		if rr.Code != http.StatusOK {
			t.Fatalf("photo %s=%d body=%s", view, rr.Code, rr.Body.String())
		}
	}
	complete := doJSON(t, h, http.MethodPost, "/api/v1/body-scans/"+scan.Scan.ID+"/complete", map[string]any{}, access)
	if complete.Code != http.StatusOK || !bytes.Contains(complete.Body.Bytes(), []byte(`"status":"completed"`)) {
		t.Fatalf("complete=%d body=%s", complete.Code, complete.Body.String())
	}

	photo := doJSON(t, h, http.MethodGet, "/api/v1/body-scans/"+scan.Scan.ID+"/photos/front", nil, access)
	if photo.Code != http.StatusOK || photo.Header().Get("Content-Type") != "image/jpeg" || photo.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatalf("photo response=%d headers=%v", photo.Code, photo.Header())
	}
	blocked := doJSON(t, h, http.MethodGet, "/api/v1/body-scans/"+scan.Scan.ID, nil, other)
	if blocked.Code != http.StatusNotFound {
		t.Fatalf("other user should get 404, got=%d", blocked.Code)
	}
}

func TestTechniqueAnalysisHTTPFlow(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	access := registerAndOnboard(t, h, "tech-http@example.com")
	catalog := doJSON(t, h, http.MethodGet, "/api/v1/technique/exercises", nil, access)
	if catalog.Code != http.StatusOK || !bytes.Contains(catalog.Body.Bytes(), []byte(`"biceps_curl"`)) {
		t.Fatalf("catalog status=%d body=%s", catalog.Code, catalog.Body.String())
	}
	frames := make([]map[string]any, 0)
	angles := []float64{160, 150, 120, 80, 55, 80, 120, 150, 160, 150, 110, 70, 55, 75, 115, 150, 160}
	for i, a := range angles {
		frames = append(frames, techniqueFramePayload(int64(i*200), a))
	}
	analyzed := doJSON(t, h, http.MethodPost, "/api/v1/technique/analyses", map[string]any{"exercise_key": "biceps_curl", "duration_ms": 3400, "frames": frames}, access)
	if analyzed.Code != http.StatusCreated {
		t.Fatalf("analyze status=%d body=%s", analyzed.Code, analyzed.Body.String())
	}
	var result struct {
		ID             string `json:"id"`
		RepCount       int    `json:"rep_count"`
		TechniqueScore int    `json:"technique_score"`
	}
	if err := json.Unmarshal(analyzed.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.ID == "" || result.RepCount != 2 || result.TechniqueScore <= 0 {
		t.Fatalf("unexpected result %+v body=%s", result, analyzed.Body.String())
	}
	history := doJSON(t, h, http.MethodGet, "/api/v1/technique/analyses", nil, access)
	if history.Code != http.StatusOK || !bytes.Contains(history.Body.Bytes(), []byte(result.ID)) {
		t.Fatalf("history status=%d body=%s", history.Code, history.Body.String())
	}
}

func techniqueFramePayload(ts int64, degrees float64) map[string]any {
	l := make([]map[string]any, 33)
	for i := range l {
		l[i] = map[string]any{"x": .5, "y": .5, "z": 0, "visibility": .95}
	}
	setArm := func(shoulder, elbow, wrist int, offset float64) {
		bx, by := .5+offset, .5
		l[elbow] = map[string]any{"x": bx, "y": by, "z": 0, "visibility": .95}
		l[shoulder] = map[string]any{"x": bx - .15, "y": by, "z": 0, "visibility": .95}
		phi := (180 - degrees) * math.Pi / 180
		l[wrist] = map[string]any{"x": bx + .15*math.Cos(phi), "y": by + .15*math.Sin(phi), "z": 0, "visibility": .95}
	}
	setArm(11, 13, 15, -.12)
	setArm(12, 14, 16, .12)
	l[23] = map[string]any{"x": .4, "y": .7, "z": 0, "visibility": .95}
	l[24] = map[string]any{"x": .6, "y": .7, "z": 0, "visibility": .95}
	l[27] = map[string]any{"x": .4, "y": .95, "z": 0, "visibility": .95}
	l[28] = map[string]any{"x": .6, "y": .95, "z": 0, "visibility": .95}
	return map[string]any{"timestamp_ms": ts, "landmarks": l}
}

func TestTechniqueAnalysisCanLinkToLiveWorkoutSetHTTP(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	access := registerAndOnboard(t, h, "tech-linked-http@example.com")
	user, err := st.FindUserByEmail(context.Background(), "tech-linked-http@example.com")
	if err != nil {
		t.Fatal(err)
	}
	created, err := st.CreateWorkout(context.Background(), store.Workout{UserID: user.ID, Muscle: "biceps", Environment: "gym", Status: "planned", DurationMinutes: 45}, []store.WorkoutExercise{{ExerciseID: "dumbbell_curl", Position: 1, TargetSets: 3, TargetRepsMin: 8, TargetRepsMax: 12, RestSeconds: 75}})
	if err != nil {
		t.Fatal(err)
	}
	active, err := st.StartWorkout(context.Background(), user.ID, created.Workout.ID)
	if err != nil {
		t.Fatal(err)
	}
	frames := make([]map[string]any, 0)
	for i, a := range []float64{160, 150, 120, 80, 55, 80, 120, 150, 160} {
		frames = append(frames, techniqueFramePayload(int64(i*200), a))
	}
	analyzed := doJSON(t, h, http.MethodPost, "/api/v1/technique/analyses", map[string]any{
		"exercise_key": "biceps_curl", "capture_mode": "live", "duration_ms": 1800, "frames": frames,
		"workout_id": active.Workout.ID, "workout_exercise_id": active.Exercises[0].ID, "set_number": 1,
	}, access)
	if analyzed.Code != http.StatusCreated {
		t.Fatalf("linked technique status=%d body=%s", analyzed.Code, analyzed.Body.String())
	}
	if !bytes.Contains(analyzed.Body.Bytes(), []byte(`"capture_mode":"live"`)) || !bytes.Contains(analyzed.Body.Bytes(), []byte(active.Workout.ID)) {
		t.Fatalf("linked technique metadata missing body=%s", analyzed.Body.String())
	}
}

func TestRecoveryCheckInAndWorkoutAdaptation(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	access := registerAndOnboard(t, h, "recovery-http@example.com")
	date := time.Now().UTC().Format("2006-01-02")

	before := doJSON(t, h, http.MethodGet, "/api/v1/recovery/today?date="+date, nil, access)
	if before.Code != http.StatusOK || !bytes.Contains(before.Body.Bytes(), []byte(`"check_in_completed":false`)) {
		t.Fatalf("unexpected pre-checkin readiness status=%d body=%s", before.Code, before.Body.String())
	}

	check := doJSON(t, h, http.MethodPut, "/api/v1/recovery/check-in", map[string]any{
		"date": date, "sleep_hours": 4.5, "sleep_quality": 1, "energy": 1, "stress": 5,
		"muscle_soreness": map[string]any{"chest": 5, "triceps": 4, "shoulders": 4},
	}, access)
	if check.Code != http.StatusOK {
		t.Fatalf("checkin status=%d body=%s", check.Code, check.Body.String())
	}
	var readiness struct {
		Score  int     `json:"score"`
		Volume float64 `json:"volume_multiplier"`
	}
	if err := json.Unmarshal(check.Body.Bytes(), &readiness); err != nil {
		t.Fatal(err)
	}
	if readiness.Score >= 50 || readiness.Volume >= .8 {
		t.Fatalf("expected low readiness: %+v body=%s", readiness, check.Body.String())
	}

	generated := doJSON(t, h, http.MethodPost, "/api/v1/workouts/generate", map[string]any{
		"muscle": "chest", "environment": "gym", "duration_minutes": 45, "local_date": date,
	}, access)
	if generated.Code != http.StatusCreated {
		t.Fatalf("generate status=%d body=%s", generated.Code, generated.Body.String())
	}
	var out struct {
		Exercises []struct {
			WorkoutExercise struct {
				TargetSets      int    `json:"target_sets"`
				ProgressionNote string `json:"progression_note"`
			} `json:"workout_exercise"`
		} `json:"exercises"`
	}
	if err := json.Unmarshal(generated.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Exercises) == 0 {
		t.Fatal("expected adapted exercises")
	}
	adapted := false
	for _, ex := range out.Exercises {
		if strings.Contains(ex.WorkoutExercise.ProgressionNote, "Адаптация нагрузки") && ex.WorkoutExercise.TargetSets <= 2 {
			adapted = true
		}
	}
	if !adapted {
		t.Fatalf("expected reduced recovery-adapted workout body=%s", generated.Body.String())
	}

	history := doJSON(t, h, http.MethodGet, "/api/v1/recovery/history?limit=10", nil, access)
	if history.Code != http.StatusOK || !bytes.Contains(history.Body.Bytes(), []byte(date)) {
		t.Fatalf("history status=%d body=%s", history.Code, history.Body.String())
	}
}

func TestHealthConnectSnapshotFeedsRecovery(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	access := registerAndOnboard(t, h, "health-connect-http@example.com")
	date := "2026-09-02"
	imported := doJSON(t, h, http.MethodPut, "/api/v1/health/snapshots", map[string]any{
		"date": date, "provider": "health_connect", "source_package": "com.xiaomi.wearable", "source_label": "Mi Fitness (Xiaomi Wear)",
		"steps": 11240, "distance_m": 8120, "active_calories_kcal": 621, "sleep_minutes": 438, "deep_sleep_minutes": 82, "light_sleep_minutes": 251, "rem_sleep_minutes": 105, "awake_minutes": 21,
		"exercise_minutes": 54, "exercise_sessions": 1, "exercise_heart_rate_avg": 143, "exercise_heart_rate_max": 171, "data_types": []string{"steps", "distance", "active_calories", "sleep", "exercise", "heart_rate"}, "captured_at": time.Now().UTC().Add(-time.Minute).Format(time.RFC3339),
	}, access)
	if imported.Code != http.StatusOK {
		t.Fatalf("health import status=%d body=%s", imported.Code, imported.Body.String())
	}
	day := doJSON(t, h, http.MethodGet, "/api/v1/health/today?date="+date, nil, access)
	if day.Code != http.StatusOK || !bytes.Contains(day.Body.Bytes(), []byte("com.xiaomi.wearable")) {
		t.Fatalf("health today status=%d body=%s", day.Code, day.Body.String())
	}
	check := doJSON(t, h, http.MethodPut, "/api/v1/recovery/check-in", map[string]any{"date": date, "sleep_hours": 4, "sleep_quality": 4, "energy": 4, "stress": 2, "muscle_soreness": map[string]int{"quads": 1}}, access)
	if check.Code != http.StatusOK {
		t.Fatalf("recovery status=%d body=%s", check.Code, check.Body.String())
	}
	if !bytes.Contains(check.Body.Bytes(), []byte(`"sleep_source":"wearable"`)) || !bytes.Contains(check.Body.Bytes(), []byte(`"sleep_minutes":438`)) {
		t.Fatalf("wearable sleep not attached to recovery: %s", check.Body.String())
	}
}

func TestHealthInsightsHTTPBuildsBaseline(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	access := registerAndOnboard(t, h, "health-insights-http@example.com")
	for i := 1; i <= 8; i++ {
		date := time.Date(2026, 9, i, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		resp := doJSON(t, h, http.MethodPut, "/api/v1/health/snapshots", map[string]any{
			"date": date, "provider": "health_connect", "source_package": "com.xiaomi.wearable", "source_label": "Mi Fitness (Xiaomi Wear)",
			"steps": 8000 + i*100, "sleep_minutes": 420 + i*5, "exercise_minutes": 30,
			"data_types": []string{"steps", "sleep", "exercise"}, "captured_at": time.Now().UTC().Format(time.RFC3339),
		}, access)
		if resp.Code != http.StatusOK {
			t.Fatalf("import %s status=%d body=%s", date, resp.Code, resp.Body.String())
		}
	}
	out := doJSON(t, h, http.MethodGet, "/api/v1/health/insights?date=2026-09-08", nil, access)
	if out.Code != http.StatusOK {
		t.Fatalf("insights status=%d body=%s", out.Code, out.Body.String())
	}
	if !bytes.Contains(out.Body.Bytes(), []byte(`"available_days":7`)) || !bytes.Contains(out.Body.Bytes(), []byte(`"confidence_percent"`)) {
		t.Fatalf("baseline missing: %s", out.Body.String())
	}
}

func TestRecoveryTimelineHTTPExplainsScoreChange(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	access := registerAndOnboard(t, h, "recovery-timeline-http@example.com")
	for _, payload := range []map[string]any{
		{"date": "2026-09-01", "sleep_hours": 8, "sleep_quality": 5, "energy": 5, "stress": 1, "muscle_soreness": map[string]int{"quads": 1}},
		{"date": "2026-09-02", "sleep_hours": 5, "sleep_quality": 2, "energy": 2, "stress": 4, "muscle_soreness": map[string]int{"quads": 4}},
	} {
		resp := doJSON(t, h, http.MethodPut, "/api/v1/recovery/check-in", payload, access)
		if resp.Code != http.StatusOK {
			t.Fatalf("checkin status=%d body=%s", resp.Code, resp.Body.String())
		}
	}
	out := doJSON(t, h, http.MethodGet, "/api/v1/recovery/timeline?date=2026-09-02", nil, access)
	if out.Code != http.StatusOK {
		t.Fatalf("timeline status=%d body=%s", out.Code, out.Body.String())
	}
	if !bytes.Contains(out.Body.Bytes(), []byte(`"comparison_available":true`)) || !bytes.Contains(out.Body.Bytes(), []byte(`"weighted_delta_points"`)) || !bytes.Contains(out.Body.Bytes(), []byte(`"volume_change_percent"`)) {
		t.Fatalf("timeline explanation missing: %s", out.Body.String())
	}
}

func TestAuthLoginRateLimit(t *testing.T) {
	st := store.NewMemory()
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	register := doJSON(t, h, http.MethodPost, "/api/v1/auth/register", map[string]any{
		"email": "rate-limit@example.com", "password": "strong-pass-123",
	}, "")
	if register.Code != http.StatusCreated {
		t.Fatalf("register status=%d body=%s", register.Code, register.Body.String())
	}
	for i := 0; i < 8; i++ {
		rr := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
			"email": "rate-limit@example.com", "password": "wrong-pass",
		}, "")
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d expected 401 got=%d body=%s", i+1, rr.Code, rr.Body.String())
		}
		if !bytes.Contains(rr.Body.Bytes(), []byte(`"error":"invalid email or password"`)) {
			t.Fatalf("login error must not reveal account details body=%s", rr.Body.String())
		}
	}
	blocked := doJSON(t, h, http.MethodPost, "/api/v1/auth/login", map[string]any{
		"email": "rate-limit@example.com", "password": "wrong-pass",
	}, "")
	if blocked.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 got=%d body=%s", blocked.Code, blocked.Body.String())
	}
	if blocked.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
}
