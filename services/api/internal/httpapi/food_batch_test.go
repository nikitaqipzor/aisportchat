package httpapi

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/auth"
	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func TestFoodEndpointsRetryAfterLostResponse(t *testing.T) {
	st:=store.NewMemory()
	h:=NewServerWithDependencies(st,auth.NewTokenManager("test-secret",15*time.Minute,24*time.Hour))
	access:=registerAndOnboard(t,h,"retry-food@example.com")
	if result:=doJSON(t,h,http.MethodPut,"/api/v1/nutrition/profile",map[string]any{"goal":"maintain","activity_level":"moderate","calculation_mode":"auto"},access);result.Code!=http.StatusOK {t.Fatalf("setup: %s",result.Body.String())}
	recipeResponse:=doJSON(t,h,http.MethodPost,"/api/v1/nutrition/recipes",map[string]any{"name":"Retry breakfast","items":[]map[string]any{{"food_id":"banana","quantity_g":120},{"food_id":"oats-dry","quantity_g":80}}},access)
	if recipeResponse.Code!=http.StatusCreated {t.Fatalf("recipe: %s",recipeResponse.Body.String())}
	var recipe store.Recipe
	if err:=json.Unmarshal(recipeResponse.Body.Bytes(),&recipe);err!=nil {t.Fatal(err)}
	loggedAt:="2026-09-20T12:00:00Z"
	path:="/api/v1/nutrition/recipes/"+recipe.ID+"/log"
	batch:=map[string]any{"meal_type":"breakfast","scale":1,"logged_at":loggedAt,"idempotency_key":"recipe-network-retry-01"}
	for i:=0;i<2;i++ {
		response:=doJSON(t,h,http.MethodPost,path,batch,access)
		if response.Code!=http.StatusCreated {t.Fatalf("recipe attempt %d: %d %s",i,response.Code,response.Body.String())}
		assertFoodEntries(t,response.Body.Bytes(),2)
	}
	ai:=map[string]any{"meal_type":"snack","logged_at":loggedAt,"idempotency_key":"ai-network-retry-01","items":[]map[string]any{{"food_id":"banana","quantity_g":120},{"food_id":"apple","quantity_g":80}}}
	for i:=0;i<2;i++ {
		response:=doJSON(t,h,http.MethodPost,"/api/v1/ai/food/confirm",ai,access)
		if response.Code!=http.StatusCreated {t.Fatalf("AI attempt %d: %d %s",i,response.Code,response.Body.String())}
		assertFoodEntries(t,response.Body.Bytes(),4)
	}
	ai["meal_type"]="lunch"
	conflict:=doJSON(t,h,http.MethodPost,"/api/v1/ai/food/confirm",ai,access)
	if conflict.Code!=http.StatusConflict {t.Fatalf("changed payload response=%d body=%s",conflict.Code,conflict.Body.String())}
}

func assertFoodEntries(t *testing.T, body []byte, expected int) {
	t.Helper()
	var day struct{Entries []store.FoodEntry `json:"entries"`}
	if err:=json.Unmarshal(body,&day);err!=nil {t.Fatal(err)}
	if len(day.Entries)!=expected {t.Fatalf("expected %d entries, got %d: %s",expected,len(day.Entries),body)}
}
