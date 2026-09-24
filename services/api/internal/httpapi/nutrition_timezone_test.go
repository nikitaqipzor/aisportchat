package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/auth"
	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func TestNutritionCalendarZoneHTTP(t *testing.T) {
	h := NewServerWithDependencies(store.NewMemory(), auth.NewTokenManager("zone-test-secret", 15*time.Minute, time.Hour))
	access := registerAndOnboard(t, h, "zone-test@example.com")
	setup := doJSON(t, h, http.MethodPut, "/api/v1/nutrition/profile", map[string]any{
		"goal": "recomp", "activity_level": "moderate", "calculation_mode": "auto",
	}, access)
	if setup.Code != http.StatusOK {
		t.Fatalf("set profile status=%d body=%s", setup.Code, setup.Body.String())
	}

	// At this instant, Moscow is already on the next calendar day.
	logged := doJSON(t, h, http.MethodPost, "/api/v1/nutrition/entries?time_zone=Europe%2FMoscow", map[string]any{
		"food_id": "cottage-cheese-5", "meal_type": "breakfast", "quantity_g": 100,
		"logged_at": "2026-09-23T22:30:00Z",
	}, access)
	if logged.Code != http.StatusCreated {
		t.Fatalf("log status=%d body=%s", logged.Code, logged.Body.String())
	}
	var local struct {
		Date    string `json:"date"`
		Entries []any  `json:"entries"`
	}
	if err := json.Unmarshal(logged.Body.Bytes(), &local); err != nil {
		t.Fatal(err)
	}
	if local.Date != "2026-09-24" || len(local.Entries) != 1 {
		t.Fatalf("wrong local day: %s", logged.Body.String())
	}

	utc := doJSON(t, h, http.MethodGet, "/api/v1/nutrition/today?date=2026-09-23", nil, access)
	localDay := doJSON(t, h, http.MethodGet, "/api/v1/nutrition/today?date=2026-09-24&time_zone=Europe%2FMoscow", nil, access)
	if utc.Code != http.StatusOK || localDay.Code != http.StatusOK ||
		!strings.Contains(utc.Body.String(), `"date":"2026-09-23"`) ||
		!strings.Contains(localDay.Body.String(), `"date":"2026-09-24"`) ||
		!strings.Contains(localDay.Body.String(), `"quantity_g":100`) {
		t.Fatalf("UTC=%d %s local=%d %s", utc.Code, utc.Body.String(), localDay.Code, localDay.Body.String())
	}
	// Batch endpoints use the same calendar convention as the single food write.
	recipe := doJSON(t, h, http.MethodPost, "/api/v1/nutrition/recipes", map[string]any{
		"name": "Завтрак у полуночи", "items": []map[string]any{{"food_id": "cottage-cheese-5", "quantity_g": 100}},
	}, access)
	if recipe.Code != http.StatusCreated {
		t.Fatalf("recipe status=%d body=%s", recipe.Code, recipe.Body.String())
	}
	var created struct { ID string `json:"id"` }
	if err := json.Unmarshal(recipe.Body.Bytes(), &created); err != nil || created.ID == "" {
		t.Fatalf("recipe ID: %v %s", err, recipe.Body.String())
	}
	recipeLog := doJSON(t, h, http.MethodPost, "/api/v1/nutrition/recipes/"+created.ID+"/log?time_zone=Europe%2FMoscow", map[string]any{
		"meal_type": "breakfast", "logged_at": "2026-09-23T22:30:00Z", "idempotency_key": "test-recipe-zone-1",
	}, access)
	if recipeLog.Code != http.StatusCreated || !strings.Contains(recipeLog.Body.String(), `"date":"2026-09-24"`) {
		t.Fatalf("recipe local day=%d %s", recipeLog.Code, recipeLog.Body.String())
	}
	aiLog := doJSON(t, h, http.MethodPost, "/api/v1/ai/food/confirm?time_zone=Europe%2FMoscow", map[string]any{
		"meal_type": "lunch", "items": []map[string]any{{"food_id": "cottage-cheese-5", "quantity_g": 100}},
		"logged_at": "2026-09-23T22:30:00Z", "idempotency_key": "test-ai-zone-1",
	}, access)
	if aiLog.Code != http.StatusCreated || !strings.Contains(aiLog.Body.String(), `"date":"2026-09-24"`) {
		t.Fatalf("AI local day=%d %s", aiLog.Code, aiLog.Body.String())
	}
	for _, path := range []string{
		"/api/v1/nutrition/today?time_zone=Not%2FAZone",
		"/api/v1/nutrition/today?time_zone=Local",
		"/api/v1/nutrition/history?time_zone=Not%2FAZone",
		"/api/v1/nutrition/entries?time_zone=Not%2FAZone",
	} {
		method := http.MethodGet
		if strings.HasPrefix(path, "/api/v1/nutrition/entries") {
			method = http.MethodPost
		}
		response := doJSON(t, h, method, path, map[string]any{}, access)
		if response.Code != http.StatusBadRequest {
			t.Errorf("invalid time zone accepted: %s status=%d", path, response.Code)
		}
	}
}
