package recovery

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func seedUser(t *testing.T, st *store.Memory) string {
	t.Helper()
	u, err := st.CreateUser(context.Background(), "recovery@example.com", "hash")
	if err != nil {
		t.Fatal(err)
	}
	return u.ID
}

func TestReadinessHighAndLow(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	uid := seedUser(t, st)
	svc := NewService(st)
	high, err := svc.SaveCheckIn(ctx, uid, CheckInInput{Date: "2026-09-01", SleepHours: 8, SleepQuality: 5, Energy: 5, Stress: 1, MuscleSoreness: map[string]int{"quads": 1}})
	if err != nil {
		t.Fatal(err)
	}
	if high.Score < 85 || high.VolumeMultiplier != 1 {
		t.Fatalf("unexpected high readiness: %+v", high)
	}
	low, err := svc.SaveCheckIn(ctx, uid, CheckInInput{Date: "2026-09-02", SleepHours: 4.5, SleepQuality: 1, Energy: 1, Stress: 5, MuscleSoreness: map[string]int{"quads": 5}})
	if err != nil {
		t.Fatal(err)
	}
	if low.Score >= 50 || low.VolumeMultiplier >= .8 {
		t.Fatalf("unexpected low readiness: %+v", low)
	}
}

func TestMuscleRecoveryUsesWorkoutAndSoreness(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	uid := seedUser(t, st)
	// Create a completed quad workout with four sets close to the check-in date.
	d, err := st.CreateWorkout(ctx, store.Workout{UserID: uid, Muscle: "quads", Environment: "gym", Status: "planned"}, []store.WorkoutExercise{{ExerciseID: "barbell_squat", TargetSets: 4, TargetRepsMin: 5, TargetRepsMax: 8}})
	if err != nil {
		t.Fatal(err)
	}
	d, err = st.StartWorkout(ctx, uid, d.Workout.ID)
	if err != nil {
		t.Fatal(err)
	}
	for i := 1; i <= 4; i++ {
		_, err = st.UpsertWorkoutSet(ctx, uid, d.Workout.ID, store.WorkoutSet{WorkoutExerciseID: d.Exercises[0].ID, SetNumber: i, Repetitions: 6})
		if err != nil {
			t.Fatal(err)
		}
	}
	d, err = st.CompleteWorkout(ctx, uid, d.Workout.ID)
	if err != nil {
		t.Fatal(err)
	}
	// Memory store completes at now; use today's UTC date so the workout falls inside the window.
	date := time.Now().UTC().Format("2006-01-02")
	svc := NewService(st)
	r, err := svc.SaveCheckIn(ctx, uid, CheckInInput{Date: date, SleepHours: 8, SleepQuality: 4, Energy: 4, Stress: 2, MuscleSoreness: map[string]int{"quads": 4}})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range r.Muscles {
		if m.Muscle == "quads" {
			found = true
			if m.Score >= 60 {
				t.Fatalf("quad recovery should be reduced, got %+v", m)
			}
		}
	}
	if !found {
		t.Fatal("quads missing")
	}
	vol, _, _, ok := svc.AdaptationForMuscle(ctx, uid, date, "quads")
	if !ok || vol > .7 {
		t.Fatalf("expected muscle adaptation, vol=%v ok=%v", vol, ok)
	}
}

func TestValidation(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	uid := seedUser(t, st)
	svc := NewService(st)
	_, err := svc.SaveCheckIn(ctx, uid, CheckInInput{Date: "bad", SleepHours: 8, SleepQuality: 5, Energy: 5, Stress: 1})
	if err == nil {
		t.Fatal("expected validation error")
	}
	_, err = svc.SaveCheckIn(ctx, uid, CheckInInput{Date: "2026-09-01", SleepHours: 8, SleepQuality: 5, Energy: 5, Stress: 1, MuscleSoreness: map[string]int{"unknown": 3}})
	if err == nil {
		t.Fatal("expected muscle validation error")
	}
}

func TestWearableSleepOverridesManualDurationButKeepsCheckIn(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	u, _ := st.CreateUser(ctx, "wearable@example.com", "x")
	_, _ = st.UpsertHealthDailySnapshot(ctx, store.HealthDailySnapshot{UserID: u.ID, LocalDate: "2026-09-02", Provider: "health_connect", SourcePackage: "com.xiaomi.wearable", SourceLabel: "Mi Fitness / Xiaomi Watch S3", SleepMinutes: 480, CapturedAt: time.Now().UTC(), DataTypes: []string{"sleep"}})
	svc := NewService(st)
	out, err := svc.SaveCheckIn(ctx, u.ID, CheckInInput{Date: "2026-09-02", SleepHours: 4, SleepQuality: 4, Energy: 4, Stress: 2, MuscleSoreness: map[string]int{}})
	if err != nil {
		t.Fatal(err)
	}
	if out.SleepSource != "wearable" || out.Wearable == nil || out.Wearable.SleepMinutes != 480 {
		t.Fatalf("wearable not attached: %+v", out)
	}
	if out.Factors.Sleep < 85 {
		t.Fatalf("expected wearable 8h sleep to improve sleep factor, got %d", out.Factors.Sleep)
	}
}

type staleHealthStore struct {
	*store.Memory
	snapshot store.HealthDailySnapshot
}

func (s *staleHealthStore) GetHealthDailySnapshot(_ context.Context, userID, localDate string) (store.HealthDailySnapshot, error) {
	if s.snapshot.UserID == userID && s.snapshot.LocalDate == localDate {
		return s.snapshot, nil
	}
	return store.HealthDailySnapshot{}, store.ErrNotFound
}
func (s *staleHealthStore) ListHealthDailySnapshots(_ context.Context, userID string, limit int) ([]store.HealthDailySnapshot, error) {
	if s.snapshot.UserID == userID {
		return []store.HealthDailySnapshot{s.snapshot}, nil
	}
	return nil, nil
}
func (s *staleHealthStore) ListHealthDailySnapshotsForDate(_ context.Context, userID, localDate string) ([]store.HealthDailySnapshot, error) {
	if s.snapshot.UserID == userID && s.snapshot.LocalDate == localDate {
		return []store.HealthDailySnapshot{s.snapshot}, nil
	}
	return nil, nil
}

func TestStaleWearableSleepFallsBackToManualCheckIn(t *testing.T) {
	ctx := context.Background()
	mem := store.NewMemory()
	u, _ := mem.CreateUser(ctx, "stale-wearable@example.com", "x")
	date := time.Now().UTC().Format("2006-01-02")
	st := &staleHealthStore{Memory: mem, snapshot: store.HealthDailySnapshot{
		ID: "stale", UserID: u.ID, LocalDate: date, Provider: "health_connect", SourcePackage: "com.xiaomi.wearable", SourceLabel: "Mi Fitness",
		SleepMinutes: 540, DataTypes: []string{"sleep"}, CapturedAt: time.Now().UTC().Add(-48 * time.Hour), ImportedAt: time.Now().UTC().Add(-48 * time.Hour),
	}}
	svc := NewService(st)
	out, err := svc.SaveCheckIn(ctx, u.ID, CheckInInput{Date: date, SleepHours: 6, SleepQuality: 3, Energy: 4, Stress: 2, MuscleSoreness: map[string]int{}})
	if err != nil {
		t.Fatal(err)
	}
	if out.SleepSource != "check_in" {
		t.Fatalf("expected manual fallback for stale wearable, got %+v", out)
	}
	if out.HealthInsights == nil || out.HealthInsights.Freshness.Status != "stale" {
		t.Fatalf("stale health diagnostics missing: %+v", out.HealthInsights)
	}
	found := false
	for _, reason := range out.Reasons {
		if strings.Contains(reason, "устарела") {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected stale reason, got %+v", out.Reasons)
	}
}
