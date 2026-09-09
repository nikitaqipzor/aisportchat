package technique

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func TestAnalyzeBicepsCurlCountsRepsAndPersists(t *testing.T) {
	st := store.NewMemory()
	u, err := st.CreateUser(context.Background(), "tech@test.dev", "hash")
	if err != nil {
		t.Fatal(err)
	}
	svc := NewService(st)
	frames := []Frame{}
	ts := int64(0)
	angles := []float64{160, 150, 120, 80, 55, 80, 120, 150, 160, 150, 110, 70, 55, 75, 115, 150, 160}
	for _, a := range angles {
		frames = append(frames, elbowFrame(ts, a))
		ts += 200
	}
	out, err := svc.Analyze(context.Background(), u.ID, AnalyzeInput{ExerciseKey: "biceps_curl", DurationMS: ts, Frames: frames})
	if err != nil {
		t.Fatal(err)
	}
	if out.RepCount != 2 {
		t.Fatalf("expected 2 reps, got %d", out.RepCount)
	}
	if out.TechniqueScore <= 0 || out.ROMScore < 70 {
		t.Fatalf("unexpected score %+v", out)
	}
	if out.AverageEccentricMS <= 0 || out.AverageConcentricMS <= 0 || out.AverageRepRPM <= 0 {
		t.Fatalf("expected phase timing metrics, got %+v", out)
	}
	if len(out.Reps) == 0 || out.Reps[0].EccentricMS <= 0 || out.Reps[0].ConcentricMS <= 0 {
		t.Fatalf("expected per-rep phase metrics, got %+v", out.Reps)
	}
	got, err := svc.Get(context.Background(), u.ID, out.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.RepCount != 2 {
		t.Fatalf("persisted reps=%d", got.RepCount)
	}
}

func TestAnalyzeRejectsLowVisibility(t *testing.T) {
	st := store.NewMemory()
	u, _ := st.CreateUser(context.Background(), "low@test.dev", "hash")
	svc := NewService(st)
	frames := []Frame{}
	for i := 0; i < 10; i++ {
		f := elbowFrame(int64(i*200), 160)
		for j := range f.Landmarks {
			f.Landmarks[j].Visibility = 0.1
		}
		frames = append(frames, f)
	}
	if _, err := svc.Analyze(context.Background(), u.ID, AnalyzeInput{ExerciseKey: "biceps_curl", Frames: frames}); err == nil {
		t.Fatal("expected visibility error")
	}
}

func TestSupportedExercisesHasInitialFive(t *testing.T) {
	if len(SupportedExercises()) != 5 {
		t.Fatalf("got %d", len(SupportedExercises()))
	}
}

func elbowFrame(ts int64, degrees float64) Frame {
	l := make([]Landmark, 33)
	for i := range l {
		l[i] = Landmark{X: .5, Y: .5, Visibility: .95}
	}
	// shoulder -> elbow -> wrist; left vector points left, wrist vector rotates to form requested angle.
	setArm := func(shoulder, elbow, wrist int, offset float64) {
		bx, by := .5+offset, .5
		l[elbow] = Landmark{X: bx, Y: by, Visibility: .95}
		l[shoulder] = Landmark{X: bx - .15, Y: by, Visibility: .95}
		phi := (180 - degrees) * math.Pi / 180
		l[wrist] = Landmark{X: bx + .15*math.Cos(phi), Y: by + .15*math.Sin(phi), Visibility: .95}
	}
	setArm(11, 13, 15, -.12)
	setArm(12, 14, 16, .12)
	// stable trunk
	l[23] = Landmark{X: .4, Y: .7, Visibility: .95}
	l[24] = Landmark{X: .6, Y: .7, Visibility: .95}
	l[27] = Landmark{X: .4, Y: .95, Visibility: .95}
	l[28] = Landmark{X: .6, Y: .95, Visibility: .95}
	return Frame{TimestampMS: ts, Landmarks: l}
}

func TestTechniqueAnalysisIsPrivatePerUser(t *testing.T) {
	st := store.NewMemory()
	u1, _ := st.CreateUser(context.Background(), "owner-tech@test.dev", "hash")
	u2, _ := st.CreateUser(context.Background(), "other-tech@test.dev", "hash")
	svc := NewService(st)
	frames := []Frame{}
	angles := []float64{160, 145, 110, 70, 55, 80, 120, 155, 160}
	for i, a := range angles {
		frames = append(frames, elbowFrame(int64(i*200), a))
	}
	out, err := svc.Analyze(context.Background(), u1.ID, AnalyzeInput{ExerciseKey: "biceps_curl", Frames: frames})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Get(context.Background(), u2.ID, out.ID); !errors.Is(err, store.ErrNotFound) {
		t.Fatalf("expected private result, got %v", err)
	}
}

func TestAnalyzeLinksLiveResultToActiveWorkoutSet(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	u, _ := st.CreateUser(ctx, "linked-tech@test.dev", "hash")
	created, err := st.CreateWorkout(ctx, store.Workout{UserID: u.ID, Muscle: "biceps", Environment: "gym", Status: "planned", DurationMinutes: 45}, []store.WorkoutExercise{{ExerciseID: "dumbbell_curl", Position: 1, TargetSets: 3, TargetRepsMin: 8, TargetRepsMax: 12, RestSeconds: 75}})
	if err != nil {
		t.Fatal(err)
	}
	active, err := st.StartWorkout(ctx, u.ID, created.Workout.ID)
	if err != nil {
		t.Fatal(err)
	}
	frames := []Frame{}
	angles := []float64{160, 145, 110, 70, 55, 80, 120, 155, 160}
	for i, a := range angles {
		frames = append(frames, elbowFrame(int64(i*200), a))
	}
	result, err := NewService(st).Analyze(ctx, u.ID, AnalyzeInput{
		ExerciseKey: "biceps_curl", CaptureMode: "live", WorkoutID: active.Workout.ID,
		WorkoutExerciseID: active.Exercises[0].ID, SetNumber: 1, Frames: frames,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.CaptureMode != "live" || result.WorkoutID != active.Workout.ID || result.WorkoutExerciseID != active.Exercises[0].ID || result.SetNumber != 1 {
		t.Fatalf("link not preserved: %+v", result)
	}
}

func TestAnalyzeRejectsMismatchedWorkoutExercise(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	u, _ := st.CreateUser(ctx, "mismatch-tech@test.dev", "hash")
	created, _ := st.CreateWorkout(ctx, store.Workout{UserID: u.ID, Muscle: "quads", Environment: "gym", Status: "planned", DurationMinutes: 45}, []store.WorkoutExercise{{ExerciseID: "barbell_squat", Position: 1, TargetSets: 3, TargetRepsMin: 5, TargetRepsMax: 8, RestSeconds: 120}})
	active, _ := st.StartWorkout(ctx, u.ID, created.Workout.ID)
	frames := []Frame{}
	for i, a := range []float64{160, 145, 110, 70, 55, 80, 120, 155, 160} {
		frames = append(frames, elbowFrame(int64(i*200), a))
	}
	_, err := NewService(st).Analyze(ctx, u.ID, AnalyzeInput{ExerciseKey: "biceps_curl", CaptureMode: "live", WorkoutID: active.Workout.ID, WorkoutExerciseID: active.Exercises[0].ID, SetNumber: 1, Frames: frames})
	if err == nil {
		t.Fatal("expected workout/technique mismatch")
	}
}

func TestShoulderPressCountsFirstRepFromBottomStart(t *testing.T) {
	st := store.NewMemory()
	u, _ := st.CreateUser(context.Background(), "press-tech@test.dev", "hash")
	frames := []Frame{}
	for i, a := range []float64{90, 100, 120, 145, 160, 145, 120, 100, 90} {
		frames = append(frames, elbowFrame(int64(i*200), a))
	}
	out, err := NewService(st).Analyze(context.Background(), u.ID, AnalyzeInput{ExerciseKey: "shoulder_press", Frames: frames})
	if err != nil {
		t.Fatal(err)
	}
	if out.RepCount != 1 {
		t.Fatalf("expected first press rep to count from bottom start, got %d", out.RepCount)
	}
	if len(out.Reps) != 1 || out.Reps[0].ConcentricMS <= 0 || out.Reps[0].EccentricMS <= 0 {
		t.Fatalf("unexpected phase metrics %+v", out.Reps)
	}
}
