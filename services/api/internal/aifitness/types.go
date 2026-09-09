package aifitness

import (
	"context"
	"encoding/json"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

type ExtractedFood struct {
	Name       string  `json:"name"`
	QuantityG  float64 `json:"quantity_g"`
	Confidence float64 `json:"confidence"`
	Notes      string  `json:"notes"`
}

type FoodDraftItem struct {
	Name         string          `json:"name"`
	QuantityG    float64         `json:"quantity_g"`
	Confidence   float64         `json:"confidence"`
	Notes        string          `json:"notes,omitempty"`
	Matched      *store.FoodItem `json:"matched_food,omitempty"`
	MatchQuality string          `json:"match_quality"`
	Calories     float64         `json:"estimated_calories,omitempty"`
	ProteinG     float64         `json:"estimated_protein_g,omitempty"`
	FatG         float64         `json:"estimated_fat_g,omitempty"`
	CarbsG       float64         `json:"estimated_carbs_g,omitempty"`
}

type FoodDraft struct {
	Source  string          `json:"source"`
	Text    string          `json:"text,omitempty"`
	Items   []FoodDraftItem `json:"items"`
	Warning string          `json:"warning,omitempty"`
}

type ConfirmFoodItem struct {
	FoodID    string  `json:"food_id"`
	QuantityG float64 `json:"quantity_g"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type FitnessContext struct {
	Profile             store.Profile              `json:"profile"`
	Goal                *store.Goal                `json:"goal,omitempty"`
	TrainingPreferences *store.TrainingPreferences `json:"training_preferences,omitempty"`
	NutritionToday      any                        `json:"nutrition_today,omitempty"`
	RecentWorkouts      []store.WorkoutDetails     `json:"recent_workouts,omitempty"`
	Progress            any                        `json:"progress,omitempty"`
	Records             []store.PersonalRecord     `json:"personal_records,omitempty"`
	Readiness           any                        `json:"readiness,omitempty"`
	GeneratedAt         time.Time                  `json:"generated_at"`
}

type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type ToolExecutor func(ctx context.Context, name string, args json.RawMessage) (string, error)

type CoachRequest struct {
	Message     string
	History     []ChatMessage
	BaseContext FitnessContext
	Tools       []ToolDefinition
	ExecuteTool ToolExecutor
}

type CoachResponse struct {
	Message   string   `json:"message"`
	Model     string   `json:"model"`
	Provider  string   `json:"provider"`
	ToolCalls []string `json:"tool_calls,omitempty"`
}

type WeeklyStats struct {
	FromDate            string   `json:"from_date"`
	ToDate              string   `json:"to_date"`
	CompletedWorkouts   int      `json:"completed_workouts"`
	TrainingVolume      float64  `json:"training_volume"`
	LoggedNutritionDays int      `json:"logged_nutrition_days"`
	AvgCalories         float64  `json:"avg_calories"`
	AvgProteinG         float64  `json:"avg_protein_g"`
	CalorieTarget       int      `json:"calorie_target,omitempty"`
	ProteinTargetG      float64  `json:"protein_target_g,omitempty"`
	WeightStartKG       *float64 `json:"weight_start_kg,omitempty"`
	WeightCurrentKG     *float64 `json:"weight_current_kg,omitempty"`
	WeightDeltaKG       *float64 `json:"weight_delta_kg,omitempty"`
	NewPRs              int      `json:"new_prs"`
	RecoveryCheckInDays int      `json:"recovery_checkin_days"`
	AvgReadiness        float64  `json:"avg_readiness"`
}

type WeeklyReport struct {
	Stats       WeeklyStats `json:"stats"`
	Summary     string      `json:"summary"`
	Wins        []string    `json:"wins"`
	Focus       []string    `json:"focus"`
	NextActions []string    `json:"next_actions"`
	Model       string      `json:"model"`
	Provider    string      `json:"provider"`
}

type Provider interface {
	Name() string
	Model() string
	ExtractFoodText(ctx context.Context, text string) ([]ExtractedFood, error)
	ExtractFoodImage(ctx context.Context, imageDataURL string) ([]ExtractedFood, error)
	Coach(ctx context.Context, req CoachRequest) (CoachResponse, error)
	WeeklyReport(ctx context.Context, stats WeeklyStats) (WeeklyReport, error)
}
