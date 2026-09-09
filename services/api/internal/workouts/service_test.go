package workouts

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/catalog"
	"github.com/example/ai-fitness-os/services/api/internal/model"
	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func TestCalculateProgressSummaryUsesCompletedFacts(t *testing.T) {
	now := time.Date(2026, 9, 9, 12, 0, 0, 0, time.UTC)
	weight80, weight90 := 80.0, 90.0
	makeView := func(id string, completed time.Time, volume float64, weight *float64, reps int) WorkoutView {
		return WorkoutView{
			Workout: store.Workout{ID: id, Status: "completed", CompletedAt: &completed, CreatedAt: completed, TotalVolume: volume},
			Exercises: []ExerciseView{{Exercise: model.Exercise{ID: "bench-press", Name: "Жим лёжа"}, Sets: []store.WorkoutSet{{ID: id + "-set", Weight: weight, Repetitions: reps}}}},
		}
	}
	items := []WorkoutView{
		makeView("w1", now.AddDate(0, 0, -1), 900, &weight90, 6),
		makeView("w2", now.AddDate(0, 0, -8), 800, &weight80, 10),
	}
	records := []store.PersonalRecord{{ID: "pr", AchievedAt: now.AddDate(0, 0, -1)}}
	got := calculateProgressSummary(items, records, 28, now)
	if got.CompletedWorkouts != 2 || got.TrainingDays != 2 || got.TotalSets != 2 || got.TotalVolume != 1700 {
		t.Fatalf("summary=%+v", got)
	}
	if got.RecentVolume7D != 900 || got.PreviousVolume7D != 800 || got.WeeklyStreak != 2 || got.PersonalRecordCount != 1 {
		t.Fatalf("period metrics=%+v", got)
	}
	if len(got.ExerciseBests) != 1 || got.ExerciseBests[0].MaxWeight == nil || *got.ExerciseBests[0].MaxWeight != 90 || got.ExerciseBests[0].MaxReps != 10 {
		t.Fatalf("bests=%+v", got.ExerciseBests)
	}
}

func TestCalculateProgressSummaryIncludesWholeFirstCalendarDay(t *testing.T) {
	now := time.Date(2026, 9, 9, 21, 0, 0, 0, time.UTC)
	firstDayMorning := time.Date(2026, 9, 3, 7, 0, 0, 0, time.UTC)
	items := []WorkoutView{{Workout: store.Workout{ID: "boundary", Status: "completed", CompletedAt: &firstDayMorning, CreatedAt: firstDayMorning, TotalVolume: 100}}}
	got := calculateProgressSummary(items, nil, 7, now)
	if got.CompletedWorkouts != 1 || got.TrainingDays != 1 || got.TotalVolume != 100 {
		t.Fatalf("first calendar day was excluded: %+v", got)
	}
}

func workoutFixture(t *testing.T) (*Service, *store.Memory, string) {
	t.Helper()
	ctx := context.Background()
	st := store.NewMemory()
	user, err := st.CreateUser(ctx, "workouts@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	height, weight := 183.0, 86.0
	level := "intermediate"
	if _, err := st.UpsertProfile(ctx, store.Profile{UserID: user.ID, HeightCM: &height, WeightKG: &weight, ExperienceLevel: &level}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.SetGoal(ctx, store.Goal{UserID: user.ID, GoalType: "strength"}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.SetTrainingPreferences(ctx, store.TrainingPreferences{UserID: user.ID, Environments: []string{"gym", "home"}, EquipmentIDs: []string{"barbell", "dumbbells", "bench", "cable", "bodyweight"}, WorkoutsPerWeek: 4, SessionMinutes: 45}); err != nil {
		t.Fatal(err)
	}
	if _, err := st.CompleteOnboarding(ctx, user.ID); err != nil {
		t.Fatal(err)
	}
	return NewService(st, NewEngine(catalog.Exercises)), st, user.ID
}

func TestServiceRequiresOnboardingAndAllowedEnvironment(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	user, err := st.CreateUser(ctx, "new@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(st, NewEngine(catalog.Exercises))
	if _, err := svc.Generate(ctx, user.ID, GenerateInput{Muscle: "chest", Environment: "gym", DurationMinutes: 45}); err == nil || !strings.Contains(err.Error(), "complete onboarding") {
		t.Fatalf("err=%v", err)
	}

	ready, _, readyID := workoutFixture(t)
	if _, err := ready.Generate(ctx, readyID, GenerateInput{Muscle: "chest", Environment: "bands"}); err == nil || !strings.Contains(err.Error(), "not enabled") {
		t.Fatalf("environment err=%v", err)
	}
}

func TestCreateManualWorkoutValidatesAndPersistsSelection(t *testing.T) {
	svc, st, userID := workoutFixture(t)
	ctx := context.Background()
	weight := 24.0
	created, err := svc.CreateManual(ctx, userID, ManualWorkoutInput{
		Muscle: "chest", Environment: "gym", DurationMinutes: 50,
		Exercises: []ManualExerciseInput{
			{ExerciseID: "bench_press", TargetSets: 4, TargetReps: 8, TargetWeight: &weight, RestSeconds: 120},
			{ExerciseID: "pushup", TargetSets: 3, TargetReps: 12, RestSeconds: 75},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Workout.Status != "planned" || len(created.Exercises) != 2 {
		t.Fatalf("created=%+v", created)
	}
	first := created.Exercises[0].WorkoutExercise
	if first.TargetSets != 4 || first.TargetRepsMin != 8 || first.TargetRepsMax != 8 || first.TargetWeight == nil || *first.TargetWeight != weight || first.RestSeconds != 120 {
		t.Fatalf("first=%+v", first)
	}
	other, err := st.CreateUser(ctx, "manual-other@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Get(ctx, other.ID, created.Workout.ID); !errors.Is(err, store.ErrForbidden) && !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("manual workout leaked across owners: %v", err)
	}

	bad := []ManualWorkoutInput{
		{Muscle: "chest", Environment: "gym", DurationMinutes: 50},
		{Muscle: "chest", Environment: "gym", DurationMinutes: 50, Exercises: []ManualExerciseInput{{ExerciseID: "missing", TargetSets: 3, TargetReps: 10, RestSeconds: 60}}},
		{Muscle: "chest", Environment: "gym", DurationMinutes: 50, Exercises: []ManualExerciseInput{{ExerciseID: "bench_press", TargetSets: 3, TargetReps: 10, RestSeconds: 60}, {ExerciseID: "bench_press", TargetSets: 3, TargetReps: 10, RestSeconds: 60}}},
		{Muscle: "chest", Environment: "home", DurationMinutes: 50, Exercises: []ManualExerciseInput{{ExerciseID: "bench_press", TargetSets: 3, TargetReps: 10, RestSeconds: 60}}},
	}
	for i, input := range bad {
		if _, err := svc.CreateManual(ctx, userID, input); err == nil {
			t.Fatalf("bad input %d accepted", i)
		}
	}
}

func TestWorkoutLifecycleValidationRepeatFavoriteAndStats(t *testing.T) {
	svc, _, userID := workoutFixture(t)
	ctx := context.Background()

	planned, err := svc.Generate(ctx, userID, GenerateInput{Muscle: "chest", Environment: "gym"})
	if err != nil {
		t.Fatal(err)
	}
	if planned.Workout.Status != "planned" || len(planned.Exercises) == 0 {
		t.Fatalf("planned=%+v", planned.Workout)
	}

	if _, err := svc.Repeat(ctx, userID, planned.Workout.ID); err == nil {
		t.Fatal("expected repeat planned workout rejection")
	}

	active, err := svc.Start(ctx, userID, planned.Workout.ID)
	if err != nil {
		t.Fatal(err)
	}
	gotActive, err := svc.Active(ctx, userID)
	if err != nil || gotActive.Workout.ID != active.Workout.ID {
		t.Fatalf("active=%+v err=%v", gotActive.Workout, err)
	}

	row := active.Exercises[0].WorkoutExercise
	badInputs := []SetInput{
		{WorkoutExerciseID: row.ID, SetNumber: 0, Repetitions: 8},
		{WorkoutExerciseID: row.ID, SetNumber: 1, Repetitions: 0},
		{WorkoutExerciseID: row.ID, SetNumber: 1, Repetitions: 201},
	}
	negWeight, highRPE, negRIR := -1.0, 11.0, -1.0
	badInputs = append(badInputs,
		SetInput{WorkoutExerciseID: row.ID, SetNumber: 1, Repetitions: 8, Weight: &negWeight},
		SetInput{WorkoutExerciseID: row.ID, SetNumber: 1, Repetitions: 8, RPE: &highRPE},
		SetInput{WorkoutExerciseID: row.ID, SetNumber: 1, Repetitions: 8, RIR: &negRIR},
	)
	for i, in := range badInputs {
		if _, err := svc.LogSet(ctx, userID, active.Workout.ID, in); err == nil {
			t.Fatalf("bad set %d expected error", i)
		}
	}

	weight, rpe, rir := 80.0, 8.0, 2.0
	logged, err := svc.LogSet(ctx, userID, active.Workout.ID, SetInput{WorkoutExerciseID: row.ID, SetNumber: 1, Weight: &weight, Repetitions: row.TargetRepsMax, RPE: &rpe, RIR: &rir})
	if err != nil {
		t.Fatal(err)
	}
	if len(logged.Exercises[0].Sets) != 1 {
		t.Fatalf("sets=%v", logged.Exercises[0].Sets)
	}

	fav, err := svc.Favorite(ctx, userID, active.Workout.ID, true)
	if err != nil || !fav.Workout.Favorite {
		t.Fatalf("favorite=%v err=%v", fav.Workout.Favorite, err)
	}

	finished, err := svc.Finish(ctx, userID, active.Workout.ID)
	if err != nil {
		t.Fatal(err)
	}
	if finished.Workout.Workout.Status != "completed" {
		t.Fatalf("status=%s", finished.Workout.Workout.Status)
	}
	if len(finished.NextRecommendations) == 0 {
		t.Fatal("expected progression recommendation")
	}
	if len(finished.PersonalRecords) == 0 {
		t.Fatal("expected first-workout personal records")
	}

	history, err := svc.History(ctx, userID, HistoryFilter{Muscle: "chest", Status: "completed", Favorite: boolPtr(true), Limit: 999})
	if err != nil || len(history) != 1 {
		t.Fatalf("history=%d err=%v", len(history), err)
	}

	records, err := svc.Records(ctx, userID, 50)
	if err != nil || len(records) == 0 {
		t.Fatalf("records=%d err=%v", len(records), err)
	}

	stats, err := svc.MuscleStats(ctx, userID, "chest")
	if err != nil {
		t.Fatal(err)
	}
	if stats.CompletedWorkouts != 1 || stats.TotalSets != 1 || stats.PersonalRecordsCount == 0 {
		t.Fatalf("stats=%+v", stats)
	}
	if _, err := svc.MuscleStats(ctx, userID, "unknown"); err == nil {
		t.Fatal("expected unknown muscle")
	}

	repeated, err := svc.Repeat(ctx, userID, active.Workout.ID)
	if err != nil {
		t.Fatal(err)
	}
	if repeated.Workout.Status != "planned" || repeated.Workout.ID == active.Workout.ID {
		t.Fatalf("repeat=%+v", repeated.Workout)
	}

	cancelled, err := svc.Cancel(ctx, userID, repeated.Workout.ID)
	if err != nil || cancelled.Workout.Status != "cancelled" {
		t.Fatalf("cancelled=%+v err=%v", cancelled.Workout, err)
	}
}

func TestGenerateAdaptedClampsAndReplacement(t *testing.T) {
	svc, _, userID := workoutFixture(t)
	ctx := context.Background()
	view, err := svc.GenerateAdapted(ctx, userID, GenerateInput{Muscle: "back", Environment: "gym", DurationMinutes: 60}, 0.01, 0.01)
	if err != nil {
		t.Fatal(err)
	}
	if len(view.Exercises) == 0 {
		t.Fatal("no exercises")
	}
	for _, ex := range view.Exercises {
		if ex.WorkoutExercise.TargetSets < 1 {
			t.Fatalf("sets=%d", ex.WorkoutExercise.TargetSets)
		}
		if !strings.Contains(ex.WorkoutExercise.ProgressionNote, "объём ×0.40") || !strings.Contains(ex.WorkoutExercise.ProgressionNote, "интенсивность ×0.70") {
			t.Fatalf("note=%q", ex.WorkoutExercise.ProgressionNote)
		}
	}

	first := view.Exercises[0]
	replacement, err := svc.Replace(ctx, userID, view.Workout.ID, first.WorkoutExercise.ID, ReplaceInput{Reason: "busy"})
	if err != nil {
		t.Fatal(err)
	}
	if replacement.Exercises[0].Exercise.ID == first.Exercise.ID {
		t.Fatalf("exercise not replaced: %s", first.Exercise.ID)
	}
	if _, err := svc.Replace(ctx, userID, view.Workout.ID, "missing", ReplaceInput{}); err == nil {
		t.Fatal("expected missing replacement row")
	}
}

func TestProgressionHelpers(t *testing.T) {
	barbell, ok := catalog.ExerciseByID("barbell-bench-press")
	if !ok {
		t.Skip("catalog id changed")
	}
	weight := 80.0
	rpeEasy, rirEasy := 8.0, 2.0
	next, _ := nextWeight(barbell, &weight, 12, 8, 12, &rpeEasy, &rirEasy)
	if next == nil || *next <= weight {
		t.Fatalf("expected increase: %v", next)
	}

	rpeHard, rirHard := 10.0, 0.0
	next, _ = nextWeight(barbell, &weight, 5, 8, 12, &rpeHard, &rirHard)
	if next == nil || *next >= weight {
		t.Fatalf("expected decrease: %v", next)
	}

	next, _ = nextWeight(barbell, &weight, 9, 8, 12, &rpeEasy, &rirEasy)
	if next == nil || *next != weight {
		t.Fatalf("expected hold: %v", next)
	}

	body, ok := catalog.ExerciseByID("push-up")
	if ok {
		next, _ = nextWeight(body, nil, 15, 8, 15, nil, nil)
		if next != nil {
			t.Fatalf("bodyweight next=%v", next)
		}
	}
}

func boolPtr(v bool) *bool { return &v }
