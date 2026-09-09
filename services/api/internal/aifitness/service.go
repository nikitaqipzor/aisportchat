package aifitness

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/catalog"
	"github.com/example/ai-fitness-os/services/api/internal/healthdata"
	"github.com/example/ai-fitness-os/services/api/internal/nutrition"
	"github.com/example/ai-fitness-os/services/api/internal/progress"
	"github.com/example/ai-fitness-os/services/api/internal/recovery"
	"github.com/example/ai-fitness-os/services/api/internal/store"
)

type Service struct {
	store     store.Store
	nutrition *nutrition.Service
	progress  *progress.Service
	recovery  *recovery.Service
	provider  Provider
}

func NewService(st store.Store, n *nutrition.Service, p *progress.Service, provider Provider, recoveryServices ...*recovery.Service) *Service {
	var r *recovery.Service
	if len(recoveryServices) > 0 {
		r = recoveryServices[0]
	}
	return &Service{store: st, nutrition: n, progress: p, recovery: r, provider: provider}
}
func (s *Service) ProviderInfo() map[string]string {
	return map[string]string{"provider": s.provider.Name(), "model": s.provider.Model()}
}

func (s *Service) ParseFoodText(ctx context.Context, userID, text string) (FoodDraft, error) {
	text = strings.TrimSpace(text)
	if len(text) < 2 {
		return FoodDraft{}, errors.New("text must contain at least 2 characters")
	}
	if len(text) > 2000 {
		return FoodDraft{}, errors.New("text is too long")
	}
	items, err := s.provider.ExtractFoodText(ctx, text)
	if err != nil {
		return FoodDraft{}, err
	}
	return s.resolveFoodDraft(ctx, userID, "text", text, items)
}
func (s *Service) ParseFoodImage(ctx context.Context, userID, imageDataURL string) (FoodDraft, error) {
	items, err := s.provider.ExtractFoodImage(ctx, imageDataURL)
	if err != nil {
		return FoodDraft{}, err
	}
	return s.resolveFoodDraft(ctx, userID, "image", "", items)
}
func (s *Service) resolveFoodDraft(ctx context.Context, userID, source, text string, items []ExtractedFood) (FoodDraft, error) {
	if len(items) > 20 {
		items = items[:20]
	}
	out := FoodDraft{Source: source, Text: text, Items: make([]FoodDraftItem, 0, len(items)), Warning: "AI-оценка порций приблизительна. Проверь продукты и граммовку перед сохранением."}
	for _, item := range items {
		item.Name = strings.TrimSpace(item.Name)
		if item.Name == "" {
			continue
		}
		if item.QuantityG <= 0 {
			item.QuantityG = 100
		}
		if item.QuantityG > 5000 {
			item.QuantityG = 5000
		}
		item.Confidence = math.Max(0, math.Min(1, item.Confidence))
		draft := FoodDraftItem{Name: item.Name, QuantityG: round1(item.QuantityG), Confidence: round2(item.Confidence), Notes: item.Notes, MatchQuality: "unmatched"}
		candidates, _ := s.store.SearchFoodItems(ctx, userID, item.Name, 5)
		if len(candidates) == 0 {
			// Retry with the first lexical token; useful for "творог 5%" vs "творог" and inflected names.
			fields := strings.Fields(item.Name)
			if len(fields) > 0 {
				candidates, _ = s.store.SearchFoodItems(ctx, userID, strings.Trim(fields[0], ",.;:"), 5)
			}
		}
		if len(candidates) > 0 {
			best := candidates[0]
			draft.Matched = &best
			draft.MatchQuality = "catalog_match"
			factor := draft.QuantityG / 100
			draft.Calories = round1(best.Kcal100 * factor)
			draft.ProteinG = round1(best.Protein100 * factor)
			draft.FatG = round1(best.Fat100 * factor)
			draft.CarbsG = round1(best.Carbs100 * factor)
		}
		out.Items = append(out.Items, draft)
	}
	if len(out.Items) == 0 {
		return FoodDraft{}, errors.New("no food items detected")
	}
	return out, nil
}

func (s *Service) ConfirmFood(ctx context.Context, userID, mealType string, items []ConfirmFoodItem, loggedAt *time.Time) (nutrition.DaySummary, error) {
	if len(items) == 0 || len(items) > 20 {
		return nutrition.DaySummary{}, errors.New("items must contain 1-20 foods")
	}
	var out nutrition.DaySummary
	for _, item := range items {
		if strings.TrimSpace(item.FoodID) == "" || item.QuantityG <= 0 {
			return nutrition.DaySummary{}, errors.New("food_id and positive quantity_g are required")
		}
		var err error
		out, err = s.nutrition.LogFood(ctx, userID, item.FoodID, mealType, item.QuantityG, loggedAt)
		if err != nil {
			return nutrition.DaySummary{}, err
		}
	}
	return out, nil
}

func (s *Service) Context(ctx context.Context, userID string) (FitnessContext, error) {
	profile, err := s.store.GetProfile(ctx, userID)
	if err != nil {
		return FitnessContext{}, err
	}
	out := FitnessContext{Profile: profile, GeneratedAt: time.Now().UTC()}
	if goal, e := s.store.GetGoal(ctx, userID); e == nil {
		out.Goal = &goal
	}
	if prefs, e := s.store.GetTrainingPreferences(ctx, userID); e == nil {
		out.TrainingPreferences = &prefs
	}
	if _, e := s.store.GetNutritionProfile(ctx, userID); e == nil {
		if day, e2 := s.nutrition.Day(ctx, userID, time.Now().UTC()); e2 == nil {
			out.NutritionToday = day
		}
	}
	if workouts, e := s.store.ListWorkouts(ctx, userID, 8); e == nil {
		out.RecentWorkouts = workouts
	}
	if summary, e := s.progress.Summary(ctx, userID, 30, time.Now().UTC()); e == nil {
		out.Progress = summary
	}
	if records, e := s.store.ListPersonalRecords(ctx, userID, 10); e == nil {
		out.Records = records
	}
	if s.recovery != nil {
		if readiness, e := s.recovery.Summary(ctx, userID, time.Now().UTC().Format("2006-01-02")); e == nil {
			out.Readiness = readiness
		}
	}
	return out, nil
}

func (s *Service) Chat(ctx context.Context, userID, message string, history []ChatMessage) (CoachResponse, error) {
	message = strings.TrimSpace(message)
	if message == "" {
		return CoachResponse{}, errors.New("message is required")
	}
	if len(message) > 4000 {
		return CoachResponse{}, errors.New("message is too long")
	}
	if len(history) > 12 {
		history = history[len(history)-12:]
	}
	base, err := s.Context(ctx, userID)
	if err != nil {
		return CoachResponse{}, err
	}
	defs := []ToolDefinition{
		{Name: "get_profile", Description: "Get the user's body profile, fitness goal and training preferences.", Parameters: emptyObjectSchema()},
		{Name: "get_today_nutrition", Description: "Get today's deterministic calorie and macro targets, consumed values and remaining values.", Parameters: emptyObjectSchema()},
		{Name: "get_recent_workouts", Description: "Get the user's recent workout history with exercises and completed sets.", Parameters: map[string]any{"type": "object", "properties": map[string]any{"limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 10}}, "required": []string{"limit"}, "additionalProperties": false}},
		{Name: "get_progress_summary", Description: "Get weight and waist trend for a requested recent period.", Parameters: map[string]any{"type": "object", "properties": map[string]any{"days": map[string]any{"type": "integer", "minimum": 7, "maximum": 180}}, "required": []string{"days"}, "additionalProperties": false}},
		{Name: "get_personal_records", Description: "Get recent personal strength records.", Parameters: map[string]any{"type": "object", "properties": map[string]any{"limit": map[string]any{"type": "integer", "minimum": 1, "maximum": 20}}, "required": []string{"limit"}, "additionalProperties": false}},
		{Name: "get_readiness", Description: "Get today's deterministic readiness score, its factors, per-muscle recovery status, and workout adaptation multipliers. Read-only; do not claim it is a medical measurement.", Parameters: emptyObjectSchema()},
		{Name: "get_health_insights", Description: "Get wearable source provenance, sync freshness, 7/28-day personal sleep/activity baselines and deviations. Read-only fitness context; do not present it as medical diagnosis.", Parameters: emptyObjectSchema()},
	}
	return s.provider.Coach(ctx, CoachRequest{Message: message, History: history, BaseContext: base, Tools: defs, ExecuteTool: func(ctx context.Context, name string, args json.RawMessage) (string, error) {
		return s.executeTool(ctx, userID, name, args)
	}})
}
func emptyObjectSchema() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{}, "required": []string{}, "additionalProperties": false}
}
func (s *Service) executeTool(ctx context.Context, userID, name string, args json.RawMessage) (string, error) {
	var value any
	switch name {
	case "get_profile":
		p, e := s.store.GetProfile(ctx, userID)
		if e != nil {
			return "", e
		}
		g, _ := s.store.GetGoal(ctx, userID)
		prefs, _ := s.store.GetTrainingPreferences(ctx, userID)
		value = map[string]any{"profile": p, "goal": g, "training_preferences": prefs}
	case "get_today_nutrition":
		if _, e := s.store.GetNutritionProfile(ctx, userID); e != nil {
			return `{"configured":false}`, nil
		}
		d, e := s.nutrition.Day(ctx, userID, time.Now().UTC())
		if e != nil {
			return "", e
		}
		value = d
	case "get_recent_workouts":
		var a struct {
			Limit int `json:"limit"`
		}
		if json.Unmarshal(args, &a) != nil {
			return "", errors.New("invalid args")
		}
		if a.Limit < 1 {
			a.Limit = 5
		}
		if a.Limit > 10 {
			a.Limit = 10
		}
		v, e := s.store.ListWorkouts(ctx, userID, a.Limit)
		if e != nil {
			return "", e
		}
		value = coachWorkoutSummaries(v)
	case "get_progress_summary":
		var a struct {
			Days int `json:"days"`
		}
		if json.Unmarshal(args, &a) != nil {
			return "", errors.New("invalid args")
		}
		v, e := s.progress.Summary(ctx, userID, a.Days, time.Now().UTC())
		if e != nil {
			return "", e
		}
		value = v
	case "get_readiness":
		if s.recovery == nil {
			return `{"configured":false}`, nil
		}
		v, e := s.recovery.Summary(ctx, userID, time.Now().UTC().Format("2006-01-02"))
		if e != nil {
			return "", e
		}
		value = v
	case "get_health_insights":
		date := time.Now().UTC().Format("2006-01-02")
		history, e := s.store.ListHealthDailySnapshots(ctx, userID, 60)
		if e != nil {
			return "", e
		}
		var current *store.HealthDailySnapshot
		if snap, e := s.store.GetHealthDailySnapshot(ctx, userID, date); e == nil {
			copy := snap
			current = &copy
		} else if e != store.ErrNotFound {
			return "", e
		}
		sources, e := s.store.ListHealthDailySnapshotsForDate(ctx, userID, date)
		if e != nil {
			return "", e
		}
		value = healthdata.BuildInsightsWithSources(date, current, history, sources, time.Now().UTC())
	case "get_personal_records":
		var a struct {
			Limit int `json:"limit"`
		}
		if json.Unmarshal(args, &a) != nil {
			return "", errors.New("invalid args")
		}
		v, e := s.store.ListPersonalRecords(ctx, userID, a.Limit)
		if e != nil {
			return "", e
		}
		value = coachRecordSummaries(v)
	default:
		return "", fmt.Errorf("unsupported tool %q", name)
	}
	raw, err := json.Marshal(value)
	return string(raw), err
}

func (s *Service) Weekly(ctx context.Context, userID string, now time.Time) (WeeklyReport, error) {
	to := now.UTC()
	from := to.AddDate(0, 0, -7)
	stats := WeeklyStats{FromDate: from.Format("2006-01-02"), ToDate: to.Format("2006-01-02")}
	workouts, err := s.store.ListWorkouts(ctx, userID, 100)
	if err != nil {
		return WeeklyReport{}, err
	}
	for _, w := range workouts {
		when := w.Workout.CreatedAt
		if w.Workout.CompletedAt != nil {
			when = *w.Workout.CompletedAt
		}
		if w.Workout.Status == "completed" && !when.Before(from) && !when.After(to) {
			stats.CompletedWorkouts++
			stats.TrainingVolume += w.Workout.TotalVolume
		}
	}
	if _, e := s.store.GetNutritionProfile(ctx, userID); e == nil {
		np, _ := s.store.GetNutritionProfile(ctx, userID)
		stats.CalorieTarget = np.CalorieTarget
		stats.ProteinTargetG = np.ProteinTarget
		hist, e2 := s.nutrition.History(ctx, userID, 7, to)
		if e2 == nil {
			var cal, pro float64
			for _, d := range hist {
				if d.Calories > 0 {
					stats.LoggedNutritionDays++
					cal += d.Calories
					pro += d.Protein
				}
			}
			if stats.LoggedNutritionDays > 0 {
				stats.AvgCalories = round1(cal / float64(stats.LoggedNutritionDays))
				stats.AvgProteinG = round1(pro / float64(stats.LoggedNutritionDays))
			}
		}
	}
	if ps, e := s.progress.Summary(ctx, userID, 7, to); e == nil {
		stats.WeightStartKG = ps.WeightStartKG
		stats.WeightCurrentKG = ps.WeightCurrentKG
		stats.WeightDeltaKG = ps.WeightDeltaKG
	}
	if records, e := s.store.ListPersonalRecords(ctx, userID, 200); e == nil {
		for _, r := range records {
			if !r.AchievedAt.Before(from) && !r.AchievedAt.After(to) {
				stats.NewPRs++
			}
		}
	}
	if checks, e := s.store.ListRecoveryCheckIns(ctx, userID, 7); e == nil && s.recovery != nil {
		var total float64
		fromDate, toDate := from.Format("2006-01-02"), to.Format("2006-01-02")
		for _, check := range checks {
			if check.LocalDate < fromDate || check.LocalDate > toDate {
				continue
			}
			r, e3 := s.recovery.Summary(ctx, userID, check.LocalDate)
			if e3 == nil && r.CheckInCompleted {
				total += float64(r.Score)
				stats.RecoveryCheckInDays++
			}
		}
		if stats.RecoveryCheckInDays > 0 {
			stats.AvgReadiness = round1(total / float64(stats.RecoveryCheckInDays))
		}
	}
	stats.TrainingVolume = round1(stats.TrainingVolume)
	report, err := s.provider.WeeklyReport(ctx, stats)
	if err != nil {
		return WeeklyReport{}, err
	}
	report.Stats = stats
	return report, nil
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }
func round2(v float64) float64 { return math.Round(v*100) / 100 }
func sortDraft(items []FoodDraftItem) {
	sort.SliceStable(items, func(i, j int) bool { return items[i].Confidence > items[j].Confidence })
}

func coachWorkoutSummaries(items []store.WorkoutDetails) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		exercises := make([]map[string]any, 0, len(item.Exercises))
		for _, row := range item.Exercises {
			name := row.ExerciseID
			if ex, ok := catalog.ExerciseByID(row.ExerciseID); ok {
				name = ex.Name
			}
			sets := make([]store.WorkoutSet, 0)
			for _, set := range item.Sets {
				if set.WorkoutExerciseID == row.ID {
					sets = append(sets, set)
				}
			}
			exercises = append(exercises, map[string]any{
				"exercise_id": row.ExerciseID, "exercise_name": name, "sets": sets,
				"target_reps_min": row.TargetRepsMin, "target_reps_max": row.TargetRepsMax,
			})
		}
		out = append(out, map[string]any{
			"id": item.Workout.ID, "muscle": item.Workout.Muscle, "environment": item.Workout.Environment,
			"status": item.Workout.Status, "completed_at": item.Workout.CompletedAt,
			"total_volume": item.Workout.TotalVolume, "exercises": exercises,
		})
	}
	return out
}

func coachRecordSummaries(items []store.PersonalRecord) []map[string]any {
	out := make([]map[string]any, 0, len(items))
	for _, record := range items {
		name := record.ExerciseID
		if ex, ok := catalog.ExerciseByID(record.ExerciseID); ok {
			name = ex.Name
		}
		out = append(out, map[string]any{
			"exercise_id": record.ExerciseID, "exercise_name": name, "record_type": record.RecordType,
			"value": record.Value, "previous_value": record.PreviousValue, "achieved_at": record.AchievedAt,
		})
	}
	return out
}
