package programs

import (
	"context"
	"testing"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func programTestUser(t *testing.T) (*store.Memory, string) {
	t.Helper()
	ctx := context.Background()
	st := store.NewMemory()
	u, err := st.CreateUser(ctx, "program@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	h, w, l := 193.0, 97.0, "intermediate"
	if _, err = st.UpsertProfile(ctx, store.Profile{UserID: u.ID, HeightCM: &h, WeightKG: &w, ExperienceLevel: &l, UnitSystem: "metric"}); err != nil {
		t.Fatal(err)
	}
	if _, err = st.SetGoal(ctx, store.Goal{UserID: u.ID, GoalType: "gain_muscle"}); err != nil {
		t.Fatal(err)
	}
	if _, err = st.SetTrainingPreferences(ctx, store.TrainingPreferences{UserID: u.ID, Environments: []string{"gym", "home"}, EquipmentIDs: []string{"barbell", "dumbbells", "bench", "cable_machine", "rack", "leg_machine"}, WorkoutsPerWeek: 3, SessionMinutes: 60}); err != nil {
		t.Fatal(err)
	}
	if _, err = st.CompleteOnboarding(ctx, u.ID); err != nil {
		t.Fatal(err)
	}
	return st, u.ID
}

func TestGenerateProgramCreatesCalendarAndDeload(t *testing.T) {
	st, userID := programTestUser(t)
	svc := NewService(st)
	start := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)
	out, err := svc.Generate(context.Background(), userID, GenerateInput{Weeks: 4, WorkoutsPerWeek: 3, Environment: "gym", StartDate: &start})
	if err != nil {
		t.Fatal(err)
	}
	if got := len(out.Sessions); got != 12 {
		t.Fatalf("sessions=%d want 12", got)
	}
	if out.Program.Status != "active" || out.Program.Weeks != 4 {
		t.Fatalf("bad program: %+v", out.Program)
	}
	deload := 0
	for _, s := range out.Sessions {
		if s.WeekNumber == 4 {
			if !s.IsDeload {
				t.Fatalf("week 4 session is not deload: %+v", s)
			}
			if s.VolumeMultiplier != 0.6 || s.IntensityMultiplier != 0.9 {
				t.Fatalf("bad deload multipliers: %+v", s)
			}
			deload++
		} else if s.IsDeload {
			t.Fatalf("unexpected deload week %d", s.WeekNumber)
		}
	}
	if deload != 3 {
		t.Fatalf("deload sessions=%d", deload)
	}
}

func TestProgramSessionLifecycleAndAnalytics(t *testing.T) {
	ctx := context.Background()
	st, userID := programTestUser(t)
	svc := NewService(st)
	start := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)
	out, err := svc.Generate(ctx, userID, GenerateInput{Weeks: 4, WorkoutsPerWeek: 3, Environment: "gym", StartDate: &start})
	if err != nil {
		t.Fatal(err)
	}
	session := out.Sessions[0]
	details, err := st.CreateWorkout(ctx, store.Workout{UserID: userID, Muscle: session.Muscle, Environment: session.Environment, Status: "planned", DurationMinutes: 60}, []store.WorkoutExercise{{ExerciseID: "bench_press", TargetSets: 4, TargetRepsMin: 6, TargetRepsMax: 10, RestSeconds: 120}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.LinkWorkout(ctx, userID, session.ID, details.Workout.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = st.StartWorkout(ctx, userID, details.Workout.ID); err != nil {
		t.Fatal(err)
	}
	active, _ := st.GetWorkout(ctx, userID, details.Workout.ID)
	exID := active.Exercises[0].ID
	weight := 80.0
	if _, err = st.UpsertWorkoutSet(ctx, userID, details.Workout.ID, store.WorkoutSet{WorkoutExerciseID: exID, SetNumber: 1, Weight: &weight, Repetitions: 8}); err != nil {
		t.Fatal(err)
	}
	if _, err = st.CompleteWorkout(ctx, userID, details.Workout.ID); err != nil {
		t.Fatal(err)
	}
	completed, err := svc.CompleteByWorkout(ctx, userID, details.Workout.ID)
	if err != nil {
		t.Fatal(err)
	}
	if completed.Status != "completed" || completed.CompletedAt == nil {
		t.Fatalf("session not completed: %+v", completed)
	}
	analytics, err := svc.Analytics(ctx, userID, out.Program.ID, start.AddDate(0, 0, 2))
	if err != nil {
		t.Fatal(err)
	}
	if analytics.CompletedSessions != 1 {
		t.Fatalf("completed=%d", analytics.CompletedSessions)
	}
	if analytics.CompletedSetsByMuscle[session.Muscle] != 1 {
		t.Fatalf("completed sets=%v", analytics.CompletedSetsByMuscle)
	}
}

func TestMissedSessionCanAutoReschedule(t *testing.T) {
	ctx := context.Background()
	st, userID := programTestUser(t)
	svc := NewService(st)
	start := time.Now().UTC().AddDate(0, 0, -10)
	out, err := svc.Generate(ctx, userID, GenerateInput{Weeks: 4, WorkoutsPerWeek: 2, Environment: "gym", StartDate: &start})
	if err != nil {
		t.Fatal(err)
	}
	active, err := svc.Active(ctx, userID)
	if err != nil {
		t.Fatal(err)
	}
	var missed *store.ProgramSession
	for i := range active.Sessions {
		if active.Sessions[i].Status == "missed" {
			cp := active.Sessions[i]
			missed = &cp
			break
		}
	}
	if missed == nil {
		t.Fatalf("expected at least one missed session, sessions=%+v", active.Sessions)
	}
	moved, err := svc.AutoRescheduleMissed(ctx, userID, missed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if moved.Status != "rescheduled" {
		t.Fatalf("status=%s", moved.Status)
	}
	if !moved.PlannedDate.After(time.Now().UTC().Add(-24 * time.Hour)) {
		t.Fatalf("new date is not future-ish: %s", moved.PlannedDate)
	}
	_ = out
}

func TestMidweekRequestStartsNextMondayWithoutDuplicateDates(t *testing.T) {
	st, userID := programTestUser(t)
	svc := NewService(st)
	friday := time.Date(2026, 8, 14, 0, 0, 0, 0, time.UTC)
	out, err := svc.Generate(context.Background(), userID, GenerateInput{Weeks: 4, WorkoutsPerWeek: 3, Environment: "gym", StartDate: &friday})
	if err != nil {
		t.Fatal(err)
	}
	wantMonday := time.Date(2026, 8, 17, 0, 0, 0, 0, time.UTC)
	if !out.Program.StartDate.Equal(wantMonday) {
		t.Fatalf("start=%s want=%s", out.Program.StartDate, wantMonday)
	}
	seen := map[string]bool{}
	for _, session := range out.Sessions {
		key := session.PlannedDate.Format("2006-01-02")
		if seen[key] {
			t.Fatalf("duplicate planned date %s", key)
		}
		seen[key] = true
	}
}
