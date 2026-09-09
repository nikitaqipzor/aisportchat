package workouts

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/catalog"
	"github.com/example/ai-fitness-os/services/api/internal/model"
	"github.com/example/ai-fitness-os/services/api/internal/store"
)

type Service struct {
	store  store.Store
	engine *Engine
}

type ExerciseView struct {
	WorkoutExercise store.WorkoutExercise `json:"workout_exercise"`
	Exercise        model.Exercise        `json:"exercise"`
	Sets            []store.WorkoutSet    `json:"sets"`
}

type WorkoutView struct {
	Workout         store.Workout          `json:"workout"`
	Exercises       []ExerciseView         `json:"exercises"`
	PersonalRecords []store.PersonalRecord `json:"personal_records"`
}

type SetInput struct {
	WorkoutExerciseID string   `json:"workout_exercise_id"`
	SetNumber         int      `json:"set_number"`
	Weight            *float64 `json:"weight,omitempty"`
	Repetitions       int      `json:"repetitions"`
	RPE               *float64 `json:"rpe,omitempty"`
	RIR               *float64 `json:"rir,omitempty"`
}

type ReplaceInput struct {
	Reason string `json:"reason,omitempty"`
}

type ManualExerciseInput struct {
	ExerciseID   string   `json:"exercise_id"`
	TargetSets   int      `json:"target_sets"`
	TargetReps   int      `json:"target_reps"`
	TargetWeight *float64 `json:"target_weight,omitempty"`
	RestSeconds  int      `json:"rest_seconds"`
}

type ManualWorkoutInput struct {
	Muscle          string                `json:"muscle"`
	Environment     string                `json:"environment"`
	DurationMinutes int                   `json:"duration_minutes"`
	Exercises       []ManualExerciseInput `json:"exercises"`
}

type FinishResult struct {
	Workout             WorkoutView                 `json:"workout"`
	NextRecommendations []ProgressionRecommendation `json:"next_recommendations"`
	PersonalRecords     []store.PersonalRecord      `json:"personal_records"`
}

type ProgressionRecommendation struct {
	ExerciseID    string   `json:"exercise_id"`
	ExerciseName  string   `json:"exercise_name"`
	CurrentWeight *float64 `json:"current_weight,omitempty"`
	NextWeight    *float64 `json:"next_weight,omitempty"`
	Message       string   `json:"message"`
}

type HistoryFilter struct {
	Limit       int
	Muscle      string
	Environment string
	Status      string
	Favorite    *bool
}

type MuscleStats struct {
	Muscle               string     `json:"muscle"`
	CompletedWorkouts    int        `json:"completed_workouts"`
	TotalSets            int        `json:"total_sets"`
	TotalVolume          float64    `json:"total_volume"`
	RecentVolume         float64    `json:"recent_volume"`
	LastWorkoutVolume    float64    `json:"last_workout_volume"`
	LastWorkoutAt        *time.Time `json:"last_workout_at,omitempty"`
	DaysSinceLastWorkout *int       `json:"days_since_last_workout,omitempty"`
	PersonalRecordsCount int        `json:"personal_records_count"`
}

type ExerciseBest struct {
	ExerciseID   string   `json:"exercise_id"`
	ExerciseName string   `json:"exercise_name"`
	MaxWeight    *float64 `json:"max_weight,omitempty"`
	MaxReps      int      `json:"max_reps"`
}

type ProgressSummary struct {
	PeriodDays          int            `json:"period_days"`
	CompletedWorkouts   int            `json:"completed_workouts"`
	TrainingDays        int            `json:"training_days"`
	WorkoutsPerWeek     float64        `json:"workouts_per_week"`
	TotalSets           int            `json:"total_sets"`
	TotalVolume         float64        `json:"total_volume"`
	RecentVolume7D      float64        `json:"recent_volume_7d"`
	PreviousVolume7D    float64        `json:"previous_volume_7d"`
	WeeklyStreak        int            `json:"weekly_streak"`
	PersonalRecordCount int            `json:"personal_record_count"`
	ExerciseBests       []ExerciseBest `json:"exercise_bests"`
}

func NewService(st store.Store, engine *Engine) *Service {
	return &Service{store: st, engine: engine}
}

func (s *Service) Generate(ctx context.Context, userID string, in GenerateInput) (WorkoutView, error) {
	return s.generate(ctx, userID, in, 1, 1)
}

// CreateManual persists user-selected workout facts. It deliberately validates
// against the deterministic catalog; an AI provider is never involved.
func (s *Service) CreateManual(ctx context.Context, userID string, in ManualWorkoutInput) (WorkoutView, error) {
	status, err := s.store.GetOnboardingStatus(ctx, userID)
	if err != nil || !status.Completed {
		return WorkoutView{}, errors.New("complete onboarding before creating a workout")
	}
	prefs, err := s.store.GetTrainingPreferences(ctx, userID)
	if err != nil || !contains(prefs.Environments, in.Environment) {
		return WorkoutView{}, errors.New("selected environment is not enabled in the user profile")
	}
	if !validMuscle(in.Muscle) || in.DurationMinutes < 10 || in.DurationMinutes > 180 || len(in.Exercises) < 1 || len(in.Exercises) > 20 {
		return WorkoutView{}, errors.New("invalid manual workout")
	}
	seen := map[string]bool{}
	rows := make([]store.WorkoutExercise, 0, len(in.Exercises))
	for _, item := range in.Exercises {
		ex, ok := catalog.ExerciseByID(item.ExerciseID)
		if !ok || seen[item.ExerciseID] || !contains(ex.Environment, in.Environment) || item.TargetSets < 1 || item.TargetSets > 10 || item.TargetReps < 1 || item.TargetReps > 200 || item.RestSeconds < 0 || item.RestSeconds > 600 || item.TargetWeight != nil && (*item.TargetWeight < 0 || *item.TargetWeight > 1000) {
			return WorkoutView{}, errors.New("invalid manual exercise")
		}
		seen[item.ExerciseID] = true
		rows = append(rows, store.WorkoutExercise{ExerciseID: ex.ID, TargetSets: item.TargetSets, TargetRepsMin: item.TargetReps, TargetRepsMax: item.TargetReps, TargetWeight: cloneFloat(item.TargetWeight), RestSeconds: item.RestSeconds, ProgressionNote: "Задано вручную"})
	}
	details, err := s.store.CreateWorkout(ctx, store.Workout{UserID: userID, Muscle: in.Muscle, Environment: in.Environment, Status: "planned", DurationMinutes: in.DurationMinutes}, rows)
	if err != nil {
		return WorkoutView{}, err
	}
	return s.view(ctx, userID, details)
}

// GenerateAdapted is used by the deterministic Program Engine. Multipliers are
// intentionally not exposed on the public workout-generation endpoint, so a
// user cannot silently bypass the program rules by posting arbitrary values.
func (s *Service) GenerateAdapted(ctx context.Context, userID string, in GenerateInput, volumeMultiplier, intensityMultiplier float64) (WorkoutView, error) {
	if volumeMultiplier < 0.4 {
		volumeMultiplier = 0.4
	}
	if volumeMultiplier > 1.2 {
		volumeMultiplier = 1.2
	}
	if intensityMultiplier < 0.7 {
		intensityMultiplier = 0.7
	}
	if intensityMultiplier > 1.1 {
		intensityMultiplier = 1.1
	}
	return s.generate(ctx, userID, in, volumeMultiplier, intensityMultiplier)
}

func (s *Service) generate(ctx context.Context, userID string, in GenerateInput, volumeMultiplier, intensityMultiplier float64) (WorkoutView, error) {
	status, err := s.store.GetOnboardingStatus(ctx, userID)
	if err != nil || !status.Completed {
		return WorkoutView{}, errors.New("complete onboarding before generating a workout")
	}
	profile, _ := s.store.GetProfile(ctx, userID)
	prefs, _ := s.store.GetTrainingPreferences(ctx, userID)
	if in.Level == "" && profile.ExperienceLevel != nil {
		in.Level = *profile.ExperienceLevel
	}
	if in.DurationMinutes <= 0 {
		in.DurationMinutes = prefs.SessionMinutes
	}
	if len(in.EquipmentIDs) == 0 {
		in.EquipmentIDs = append([]string(nil), prefs.EquipmentIDs...)
	}
	if !contains(prefs.Environments, in.Environment) {
		return WorkoutView{}, errors.New("selected environment is not enabled in the user profile")
	}

	previous, err := s.store.LastCompletedWorkout(ctx, userID, in.Muscle, in.Environment)
	if err == nil {
		for _, ex := range previous.Exercises {
			in.PreviousExerciseIDs = append(in.PreviousExerciseIDs, ex.ExerciseID)
		}
	}

	generated, err := s.engine.Generate(in)
	if err != nil {
		return WorkoutView{}, err
	}

	rows := make([]store.WorkoutExercise, 0, len(generated.Exercises))
	for _, item := range generated.Exercises {
		target, note := s.progressionTarget(ctx, userID, item.Exercise, item.RepMin, item.RepMax)
		sets := int(math.Round(float64(item.Sets) * volumeMultiplier))
		if sets < 1 {
			sets = 1
		}
		if target != nil && intensityMultiplier != 1 {
			v := math.Round((*target*intensityMultiplier)*2) / 2
			target = &v
		}
		if volumeMultiplier < 0.99 || intensityMultiplier < 0.99 {
			note = fmt.Sprintf("Адаптация нагрузки: объём ×%.2f, интенсивность ×%.2f. %s", volumeMultiplier, intensityMultiplier, note)
		}
		rows = append(rows, store.WorkoutExercise{
			ExerciseID: item.Exercise.ID, TargetSets: sets, TargetRepsMin: item.RepMin,
			TargetRepsMax: item.RepMax, TargetWeight: target, RestSeconds: item.RestSecond,
			ProgressionNote: note,
		})
	}

	details, err := s.store.CreateWorkout(ctx, store.Workout{
		UserID: userID, Muscle: generated.Muscle, Environment: generated.Environment,
		Status: "planned", DurationMinutes: generated.DurationMinutes,
	}, rows)
	if err != nil {
		return WorkoutView{}, err
	}
	return s.view(ctx, userID, details)
}

func (s *Service) Active(ctx context.Context, userID string) (WorkoutView, error) {
	details, err := s.store.ActiveWorkout(ctx, userID)
	if err != nil {
		return WorkoutView{}, err
	}
	return s.view(ctx, userID, details)
}

func (s *Service) Get(ctx context.Context, userID, workoutID string) (WorkoutView, error) {
	details, err := s.store.GetWorkout(ctx, userID, workoutID)
	if err != nil {
		return WorkoutView{}, err
	}
	return s.view(ctx, userID, details)
}

func (s *Service) Start(ctx context.Context, userID, workoutID string) (WorkoutView, error) {
	details, err := s.store.StartWorkout(ctx, userID, workoutID)
	if err != nil {
		return WorkoutView{}, err
	}
	return s.view(ctx, userID, details)
}

func (s *Service) Cancel(ctx context.Context, userID, workoutID string) (WorkoutView, error) {
	details, err := s.store.CancelWorkout(ctx, userID, workoutID)
	if err != nil {
		return WorkoutView{}, err
	}
	return s.view(ctx, userID, details)
}

func (s *Service) LogSet(ctx context.Context, userID, workoutID string, in SetInput) (WorkoutView, error) {
	if in.WorkoutExerciseID == "" || in.SetNumber < 1 || in.Repetitions < 1 || in.Repetitions > 200 {
		return WorkoutView{}, errors.New("invalid set data")
	}
	if in.Weight != nil && (*in.Weight < 0 || *in.Weight > 1000) {
		return WorkoutView{}, errors.New("weight must be between 0 and 1000")
	}
	if in.RPE != nil && (*in.RPE < 1 || *in.RPE > 10) {
		return WorkoutView{}, errors.New("rpe must be between 1 and 10")
	}
	if in.RIR != nil && (*in.RIR < 0 || *in.RIR > 10) {
		return WorkoutView{}, errors.New("rir must be between 0 and 10")
	}
	details, err := s.store.UpsertWorkoutSet(ctx, userID, workoutID, store.WorkoutSet{
		WorkoutExerciseID: in.WorkoutExerciseID, SetNumber: in.SetNumber,
		Weight: cloneFloat(in.Weight), Repetitions: in.Repetitions, RPE: cloneFloat(in.RPE), RIR: cloneFloat(in.RIR),
	})
	if err != nil {
		return WorkoutView{}, err
	}
	return s.view(ctx, userID, details)
}

func (s *Service) Replace(ctx context.Context, userID, workoutID, workoutExerciseID string, _ ReplaceInput) (WorkoutView, error) {
	current, err := s.store.GetWorkout(ctx, userID, workoutID)
	if err != nil {
		return WorkoutView{}, err
	}
	var row store.WorkoutExercise
	found := false
	for _, item := range current.Exercises {
		if item.ID == workoutExerciseID {
			row = item
			found = true
			break
		}
	}
	if !found {
		return WorkoutView{}, store.ErrNotFound
	}
	old, ok := catalog.ExerciseByID(row.ExerciseID)
	if !ok {
		return WorkoutView{}, store.ErrNotFound
	}
	used := map[string]bool{}
	for _, item := range current.Exercises {
		used[item.ExerciseID] = true
	}
	candidates := make([]model.Exercise, 0)
	for _, ex := range catalog.Exercises {
		if used[ex.ID] || ex.PrimaryMuscle != old.PrimaryMuscle || !contains(ex.Environment, current.Workout.Environment) {
			continue
		}
		if ex.MovementPattern == old.MovementPattern {
			candidates = append(candidates, ex)
		}
	}
	if len(candidates) == 0 {
		for _, ex := range catalog.Exercises {
			if !used[ex.ID] && ex.PrimaryMuscle == old.PrimaryMuscle && contains(ex.Environment, current.Workout.Environment) {
				candidates = append(candidates, ex)
			}
		}
	}
	if len(candidates) == 0 {
		return WorkoutView{}, errors.New("no replacement exercise available")
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ID < candidates[j].ID })
	replacement := candidates[0]
	target, note := s.progressionTarget(ctx, userID, replacement, replacement.RepMin, replacement.RepMax)
	updated, err := s.store.ReplaceWorkoutExercise(ctx, userID, workoutID, workoutExerciseID, store.WorkoutExercise{
		ExerciseID: replacement.ID, TargetSets: replacement.DefaultSets, TargetRepsMin: replacement.RepMin,
		TargetRepsMax: replacement.RepMax, TargetWeight: target, RestSeconds: replacement.RestSeconds,
		ProgressionNote: note,
	})
	if err != nil {
		return WorkoutView{}, err
	}
	return s.view(ctx, userID, updated)
}

func (s *Service) Finish(ctx context.Context, userID, workoutID string) (FinishResult, error) {
	details, err := s.store.CompleteWorkout(ctx, userID, workoutID)
	if err != nil {
		return FinishResult{}, err
	}
	view, err := s.view(ctx, userID, details)
	if err != nil {
		return FinishResult{}, err
	}
	recommendations := make([]ProgressionRecommendation, 0, len(view.Exercises))
	for _, item := range view.Exercises {
		if len(item.Sets) == 0 {
			continue
		}
		best := item.Sets[0]
		for _, set := range item.Sets[1:] {
			if value(set.Weight) > value(best.Weight) || (value(set.Weight) == value(best.Weight) && set.Repetitions > best.Repetitions) {
				best = set
			}
		}
		next, message := nextWeight(item.Exercise, best.Weight, best.Repetitions, item.WorkoutExercise.TargetRepsMin, item.WorkoutExercise.TargetRepsMax, best.RPE, best.RIR)
		recommendations = append(recommendations, ProgressionRecommendation{
			ExerciseID: item.Exercise.ID, ExerciseName: item.Exercise.Name,
			CurrentWeight: cloneFloat(best.Weight), NextWeight: next, Message: message,
		})
	}

	records := s.detectAndSaveRecords(ctx, userID, view)
	if len(records) > 0 {
		view, _ = s.Get(ctx, userID, workoutID)
	}
	return FinishResult{Workout: view, NextRecommendations: recommendations, PersonalRecords: records}, nil
}

func (s *Service) History(ctx context.Context, userID string, filter HistoryFilter) ([]WorkoutView, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	// Pull enough rows before applying MVP filters in the domain layer. This keeps Store simple
	// and lets us move filtering into SQL later without changing the HTTP contract.
	items, err := s.store.ListWorkouts(ctx, userID, 100)
	if err != nil {
		return nil, err
	}
	out := make([]WorkoutView, 0, min(limit, len(items)))
	for _, item := range items {
		if filter.Muscle != "" && item.Workout.Muscle != filter.Muscle {
			continue
		}
		if filter.Environment != "" && item.Workout.Environment != filter.Environment {
			continue
		}
		if filter.Status != "" && item.Workout.Status != filter.Status {
			continue
		}
		if filter.Favorite != nil && item.Workout.Favorite != *filter.Favorite {
			continue
		}
		view, err := s.view(ctx, userID, item)
		if err != nil {
			return nil, err
		}
		out = append(out, view)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func (s *Service) Favorite(ctx context.Context, userID, workoutID string, favorite bool) (WorkoutView, error) {
	details, err := s.store.SetWorkoutFavorite(ctx, userID, workoutID, favorite)
	if err != nil {
		return WorkoutView{}, err
	}
	return s.view(ctx, userID, details)
}

func (s *Service) Repeat(ctx context.Context, userID, workoutID string) (WorkoutView, error) {
	source, err := s.store.GetWorkout(ctx, userID, workoutID)
	if err != nil {
		return WorkoutView{}, err
	}
	if source.Workout.Status != "completed" {
		return WorkoutView{}, errors.New("only completed workouts can be repeated")
	}
	rows := make([]store.WorkoutExercise, 0, len(source.Exercises))
	for _, row := range source.Exercises {
		ex, ok := catalog.ExerciseByID(row.ExerciseID)
		if !ok {
			continue
		}
		target, note := s.progressionTarget(ctx, userID, ex, row.TargetRepsMin, row.TargetRepsMax)
		rows = append(rows, store.WorkoutExercise{
			ExerciseID: ex.ID, TargetSets: row.TargetSets, TargetRepsMin: row.TargetRepsMin,
			TargetRepsMax: row.TargetRepsMax, TargetWeight: target, RestSeconds: row.RestSeconds,
			ProgressionNote: note,
		})
	}
	if len(rows) == 0 {
		return WorkoutView{}, errors.New("source workout has no repeatable exercises")
	}
	details, err := s.store.CreateWorkout(ctx, store.Workout{
		UserID: userID, Muscle: source.Workout.Muscle, Environment: source.Workout.Environment,
		Status: "planned", DurationMinutes: source.Workout.DurationMinutes,
	}, rows)
	if err != nil {
		return WorkoutView{}, err
	}
	return s.view(ctx, userID, details)
}

func (s *Service) MuscleStats(ctx context.Context, userID, muscle string) (MuscleStats, error) {
	if !validMuscle(muscle) {
		return MuscleStats{}, errors.New("unknown muscle")
	}
	items, err := s.store.ListWorkouts(ctx, userID, 100)
	if err != nil {
		return MuscleStats{}, err
	}
	stats := MuscleStats{Muscle: muscle}
	recentCount := 0
	for _, item := range items {
		if item.Workout.Status != "completed" || item.Workout.Muscle != muscle {
			continue
		}
		stats.CompletedWorkouts++
		stats.TotalVolume += item.Workout.TotalVolume
		stats.TotalSets += len(item.Sets)
		if recentCount < 4 {
			stats.RecentVolume += item.Workout.TotalVolume
			recentCount++
		}
		when := item.Workout.CompletedAt
		if when == nil {
			when = &item.Workout.CreatedAt
		}
		if stats.LastWorkoutAt == nil || when.After(*stats.LastWorkoutAt) {
			copy := *when
			stats.LastWorkoutAt = &copy
			stats.LastWorkoutVolume = item.Workout.TotalVolume
		}
	}
	if stats.LastWorkoutAt != nil {
		days := int(time.Since(*stats.LastWorkoutAt).Hours() / 24)
		if days < 0 {
			days = 0
		}
		stats.DaysSinceLastWorkout = &days
	}
	records, err := s.store.ListPersonalRecords(ctx, userID, 200)
	if err == nil {
		for _, record := range records {
			if ex, ok := catalog.ExerciseByID(record.ExerciseID); ok && ex.PrimaryMuscle == muscle {
				stats.PersonalRecordsCount++
			}
		}
	}
	return stats, nil
}

func (s *Service) Records(ctx context.Context, userID string, limit int) ([]store.PersonalRecord, error) {
	return s.store.ListPersonalRecords(ctx, userID, limit)
}

func (s *Service) ProgressSummary(ctx context.Context, userID string, days int, now time.Time) (ProgressSummary, error) {
	if days < 7 || days > 90 {
		days = 28
	}
	details, err := s.store.ListWorkouts(ctx, userID, 500)
	if err != nil {
		return ProgressSummary{}, err
	}
	items := make([]WorkoutView, 0, len(details))
	for _, item := range details {
		if item.Workout.Status != "completed" {
			continue
		}
		view, err := s.view(ctx, userID, item)
		if err != nil {
			return ProgressSummary{}, err
		}
		items = append(items, view)
	}
	records, err := s.store.ListPersonalRecords(ctx, userID, 200)
	if err != nil {
		return ProgressSummary{}, err
	}
	return calculateProgressSummary(items, records, days, now), nil
}

func calculateProgressSummary(items []WorkoutView, records []store.PersonalRecord, days int, now time.Time) ProgressSummary {
	now = now.UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	periodStart := todayStart.AddDate(0, 0, -(days - 1))
	recentStart := todayStart.AddDate(0, 0, -6)
	previousStart := todayStart.AddDate(0, 0, -13)
	trainingDays := map[string]bool{}
	trainingWeeks := map[string]bool{}
	bests := map[string]ExerciseBest{}
	out := ProgressSummary{PeriodDays: days, ExerciseBests: []ExerciseBest{}}
	for _, item := range items {
		when := item.Workout.CompletedAt
		if when == nil {
			when = &item.Workout.CreatedAt
		}
		if when.After(now) || when.Before(periodStart) {
			continue
		}
		if !when.Before(recentStart) {
			out.RecentVolume7D += item.Workout.TotalVolume
		} else if !when.Before(previousStart) {
			out.PreviousVolume7D += item.Workout.TotalVolume
		}
		if when.Before(periodStart) {
			continue
		}
		out.CompletedWorkouts++
		out.TotalVolume += item.Workout.TotalVolume
		trainingDays[when.Format("2006-01-02")] = true
		year, week := when.ISOWeek()
		trainingWeeks[fmt.Sprintf("%04d-%02d", year, week)] = true
		for _, exercise := range item.Exercises {
			best := bests[exercise.Exercise.ID]
			best.ExerciseID, best.ExerciseName = exercise.Exercise.ID, exercise.Exercise.Name
			for _, set := range exercise.Sets {
				out.TotalSets++
				if set.Repetitions > best.MaxReps {
					best.MaxReps = set.Repetitions
				}
				if set.Weight != nil && (best.MaxWeight == nil || *set.Weight > *best.MaxWeight) {
					value := *set.Weight
					best.MaxWeight = &value
				}
			}
			bests[exercise.Exercise.ID] = best
		}
	}
	out.TrainingDays = len(trainingDays)
	out.WorkoutsPerWeek = math.Round((float64(out.CompletedWorkouts)/(float64(days)/7))*10) / 10
	for _, record := range records {
		if !record.AchievedAt.Before(periodStart) && !record.AchievedAt.After(now) {
			out.PersonalRecordCount++
		}
	}
	for _, best := range bests {
		out.ExerciseBests = append(out.ExerciseBests, best)
	}
	sort.Slice(out.ExerciseBests, func(i, j int) bool { return out.ExerciseBests[i].ExerciseName < out.ExerciseBests[j].ExerciseName })
	if len(out.ExerciseBests) > 5 {
		out.ExerciseBests = out.ExerciseBests[:5]
	}
	weekCursor := startOfISOWeek(now)
	if !trainingWeeks[isoWeekKey(weekCursor)] {
		weekCursor = weekCursor.AddDate(0, 0, -7)
	}
	for trainingWeeks[isoWeekKey(weekCursor)] {
		out.WeeklyStreak++
		weekCursor = weekCursor.AddDate(0, 0, -7)
	}
	return out
}

func startOfISOWeek(value time.Time) time.Time {
	day := int(value.Weekday())
	if day == 0 {
		day = 7
	}
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -(day - 1))
}

func isoWeekKey(value time.Time) string {
	year, week := value.ISOWeek()
	return fmt.Sprintf("%04d-%02d", year, week)
}

func (s *Service) progressionTarget(ctx context.Context, userID string, ex model.Exercise, repMin, repMax int) (*float64, string) {
	perf, err := s.store.LastExercisePerformance(ctx, userID, ex.ID)
	if err != nil || perf.Weight == nil {
		return nil, "Первое выполнение: выбери комфортный рабочий вес и оставь 1–3 повтора в запасе."
	}
	next, message := nextWeight(ex, perf.Weight, perf.Repetitions, repMin, repMax, perf.RPE, perf.RIR)
	return next, message
}

func nextWeight(ex model.Exercise, current *float64, reps, repMin, repMax int, rpe, rir *float64) (*float64, string) {
	if current == nil || *current <= 0 || contains(ex.Equipment, "bodyweight") || contains(ex.Equipment, "resistance_band") {
		return cloneFloat(current), "Сохрани диапазон повторений и прогрессируй качеством/повторами."
	}
	weight := *current
	if reps >= repMax && !veryHard(rpe, rir) {
		weight += weightIncrement(ex)
		weight = math.Round(weight*2) / 2
		return &weight, fmt.Sprintf("Верх диапазона достигнут — попробуй %.1f кг в следующий раз.", weight)
	}
	if reps < repMin && veryHard(rpe, rir) {
		weight = math.Round(weight*0.95*2) / 2
		return &weight, fmt.Sprintf("Повторы ниже цели при высокой сложности — попробуй %.1f кг.", weight)
	}
	return &weight, fmt.Sprintf("Сохрани %.1f кг и сначала добери верх диапазона повторений.", weight)
}

func veryHard(rpe, rir *float64) bool {
	return (rpe != nil && *rpe >= 9.5) || (rir != nil && *rir <= 0.5)
}

func weightIncrement(ex model.Exercise) float64 {
	if contains(ex.Equipment, "dumbbells") {
		return 2.0
	}
	return 2.5
}

func (s *Service) view(ctx context.Context, userID string, details store.WorkoutDetails) (WorkoutView, error) {
	setMap := map[string][]store.WorkoutSet{}
	for _, set := range details.Sets {
		setMap[set.WorkoutExerciseID] = append(setMap[set.WorkoutExerciseID], set)
	}
	out := WorkoutView{Workout: details.Workout, Exercises: make([]ExerciseView, 0, len(details.Exercises)), PersonalRecords: []store.PersonalRecord{}}
	for _, row := range details.Exercises {
		ex, _ := catalog.ExerciseByID(row.ExerciseID)
		sets := setMap[row.ID]
		sort.Slice(sets, func(i, j int) bool { return sets[i].SetNumber < sets[j].SetNumber })
		out.Exercises = append(out.Exercises, ExerciseView{WorkoutExercise: row, Exercise: ex, Sets: sets})
	}
	records, err := s.store.ListWorkoutPersonalRecords(ctx, userID, details.Workout.ID)
	if err == nil {
		out.PersonalRecords = records
	} else if !errors.Is(err, store.ErrNotFound) {
		return WorkoutView{}, err
	}
	return out, nil
}

func (s *Service) detectAndSaveRecords(ctx context.Context, userID string, view WorkoutView) []store.PersonalRecord {
	out := make([]store.PersonalRecord, 0)
	for _, item := range view.Exercises {
		if len(item.Sets) == 0 {
			continue
		}
		candidates := recordCandidates(item)
		for _, candidate := range candidates {
			previous, err := s.store.BestPersonalRecord(ctx, userID, item.Exercise.ID, candidate.recordType)
			if err == nil && candidate.value <= previous.Value+0.0001 {
				continue
			}
			var previousValue *float64
			if err == nil {
				previousValue = cloneFloat(&previous.Value)
			} else if !errors.Is(err, store.ErrNotFound) {
				continue
			}
			saved, err := s.store.SavePersonalRecord(ctx, store.PersonalRecord{
				UserID: userID, WorkoutID: view.Workout.ID, WorkoutSetID: candidate.setID,
				ExerciseID: item.Exercise.ID, RecordType: candidate.recordType,
				Value: candidate.value, PreviousValue: previousValue, AchievedAt: time.Now().UTC(),
			})
			if err == nil {
				out = append(out, saved)
			}
		}
	}
	return out
}

type recordCandidate struct {
	recordType string
	value      float64
	setID      string
}

func recordCandidates(item ExerciseView) []recordCandidate {
	bestReps := item.Sets[0]
	bestWeight := item.Sets[0]
	bestE1RM := item.Sets[0]
	bestE1RMValue := estimated1RM(item.Sets[0])
	for _, set := range item.Sets[1:] {
		if set.Repetitions > bestReps.Repetitions {
			bestReps = set
		}
		if value(set.Weight) > value(bestWeight.Weight) {
			bestWeight = set
		}
		e1rm := estimated1RM(set)
		if e1rm > bestE1RMValue {
			bestE1RM = set
			bestE1RMValue = e1rm
		}
	}
	out := []recordCandidate{{recordType: "max_reps", value: float64(bestReps.Repetitions), setID: bestReps.ID}}
	if value(bestWeight.Weight) > 0 {
		out = append(out, recordCandidate{recordType: "max_weight", value: value(bestWeight.Weight), setID: bestWeight.ID})
	}
	if bestE1RMValue > 0 {
		out = append(out, recordCandidate{recordType: "estimated_1rm", value: math.Round(bestE1RMValue*10) / 10, setID: bestE1RM.ID})
	}
	return out
}

func estimated1RM(set store.WorkoutSet) float64 {
	if set.Weight == nil || *set.Weight <= 0 || set.Repetitions <= 0 {
		return 0
	}
	return *set.Weight * (1 + float64(set.Repetitions)/30.0)
}

func validMuscle(id string) bool {
	for _, muscle := range catalog.Muscles {
		if muscle.ID == id {
			return true
		}
	}
	return false
}

func cloneFloat(v *float64) *float64 {
	if v == nil {
		return nil
	}
	c := *v
	return &c
}

func value(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}
