package store

import (
	"context"
	"errors"
	"math"
	"sync"
	"testing"
)

func TestWorkoutOperationReceiptsSurviveAmbiguousAcknowledgmentAndStaleReplay(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()
	user, err := m.CreateUser(ctx, "offline@example.com", "hash")
	if err != nil { t.Fatal(err) }
	other, err := m.CreateUser(ctx, "other@example.com", "hash")
	if err != nil { t.Fatal(err) }
	created, err := m.CreateWorkout(ctx, Workout{UserID: user.ID, Status: "planned", Muscle: "chest", Environment: "gym"}, []WorkoutExercise{{ExerciseID: "bench-press", TargetSets: 3, TargetRepsMin: 6, TargetRepsMax: 10}})
	if err != nil { t.Fatal(err) }
	wid, eid := created.Workout.ID, created.Exercises[0].ID
	if _, err = m.StartWorkout(ctx, user.ID, wid); err != nil { t.Fatal(err) }
	first := WorkoutSet{WorkoutExerciseID: eid, SetNumber: 1, Repetitions: 8}
	const firstID = "offline-first-0001"
	var wg sync.WaitGroup
	results := make(chan error, 12)
	for i := 0; i < 12; i++ { wg.Add(1); go func() { defer wg.Done(); _, e := m.ApplyWorkoutOperation(ctx, user.ID, wid, firstID, "log_set", "first-payload", first); results <- e }() }
	wg.Wait(); close(results)
	for e := range results { if e != nil { t.Fatalf("concurrent retry: %v", e) } }
	second := WorkoutSet{WorkoutExerciseID: eid, SetNumber: 1, Repetitions: 12}
	if _, err := m.ApplyWorkoutOperation(ctx, user.ID, wid, "offline-second-0002", "log_set", "second-payload", second); err != nil { t.Fatal(err) }
	// The client lost the first response; its delayed retry must return that
	// receipt without changing the now-newer value.
	replay, err := m.ApplyWorkoutOperation(ctx, user.ID, wid, firstID, "log_set", "first-payload", first)
	if err != nil || replay.Sets[0].Repetitions != 8 { t.Fatalf("old receipt=%+v err=%v", replay.Sets, err) }
	current, err := m.GetWorkout(ctx, user.ID, wid)
	if err != nil || len(current.Sets) != 1 || current.Sets[0].Repetitions != 12 { t.Fatalf("stale replay overwrote newer set: %+v %v", current.Sets, err) }
	if _, err := m.ApplyWorkoutOperation(ctx, user.ID, wid, firstID, "log_set", "changed-payload", first); !errors.Is(err, ErrIdempotencyConflict) { t.Fatalf("payload reuse: %v", err) }
	if _, err := m.ApplyWorkoutOperation(ctx, user.ID, wid, firstID, "cancel_workout", "first-payload", WorkoutSet{}); !errors.Is(err, ErrIdempotencyConflict) { t.Fatalf("action reuse: %v", err) }
	if _, err := m.ApplyWorkoutOperation(ctx, other.ID, wid, firstID, "log_set", "first-payload", first); !errors.Is(err, ErrForbidden) { t.Fatalf("other user's key: %v", err) }
	finished, err := m.ApplyWorkoutOperation(ctx, user.ID, wid, "offline-finish-0003", "finish_workout", "finish-payload", WorkoutSet{})
	if err != nil || finished.Workout.Status != "completed" { t.Fatalf("finish=%+v err=%v", finished.Workout, err) }
	finishedAgain, err := m.ApplyWorkoutOperation(ctx, user.ID, wid, "offline-finish-0003", "finish_workout", "finish-payload", WorkoutSet{})
	if err != nil || finishedAgain.Workout.Status != "completed" || !finishedAgain.Workout.CompletedAt.Equal(*finished.Workout.CompletedAt) { t.Fatalf("finish retry=%+v err=%v", finishedAgain.Workout, err) }
	secondWorkout, err := m.CreateWorkout(ctx, Workout{UserID: user.ID, Status: "planned", Muscle: "back", Environment: "gym"}, nil)
	if err != nil { t.Fatal(err) }
	cancelled, err := m.ApplyWorkoutOperation(ctx, user.ID, secondWorkout.Workout.ID, "offline-cancel-0004", "cancel_workout", "cancel-payload", WorkoutSet{})
	if err != nil || cancelled.Workout.Status != "cancelled" { t.Fatalf("cancel=%+v err=%v", cancelled.Workout, err) }
	again, err := m.ApplyWorkoutOperation(ctx, user.ID, secondWorkout.Workout.ID, "offline-cancel-0004", "cancel_workout", "cancel-payload", WorkoutSet{})
	if err != nil || again.Workout.Status != "cancelled" { t.Fatalf("cancel retry=%+v err=%v", again.Workout, err) }
}

func TestDeleteAccountRemovesWorkoutOperationReceipts(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()
	user, err := m.CreateUser(ctx, "delete-workout-receipt@example.com", "hash")
	if err != nil { t.Fatal(err) }
	created, err := m.CreateWorkout(ctx, Workout{UserID: user.ID, Status: "planned", Muscle: "chest", Environment: "gym"}, nil)
	if err != nil { t.Fatal(err) }
	const operationID = "delete-receipt-0001"
	if _, err := m.ApplyWorkoutOperation(ctx, user.ID, created.Workout.ID, operationID, "cancel_workout", "hash", WorkoutSet{}); err != nil { t.Fatal(err) }
	if len(m.workoutOperations) != 1 { t.Fatalf("expected receipt, got %d", len(m.workoutOperations)) }
	if err := m.DeleteAccount(ctx, user.ID, func(context.Context) error { return nil }); err != nil { t.Fatal(err) }
	if len(m.workoutOperations) != 0 { t.Fatalf("deleted account retained %d workout receipts", len(m.workoutOperations)) }
	if _, err := m.ApplyWorkoutOperation(ctx, user.ID, created.Workout.ID, operationID, "cancel_workout", "hash", WorkoutSet{}); !errors.Is(err, ErrNotFound) { t.Fatalf("deleted account replay: %v", err) }
}

func TestInvalidWorkoutOperationDoesNotWriteSetOrClaimReceipt(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()
	user, err := m.CreateUser(ctx, "nonfinite-operation@example.com", "hash")
	if err != nil { t.Fatal(err) }
	created, err := m.CreateWorkout(ctx, Workout{UserID: user.ID, Status: "planned", Muscle: "chest", Environment: "gym"}, []WorkoutExercise{{ExerciseID: "bench-press", TargetSets: 3, TargetRepsMin: 6, TargetRepsMax: 10}})
	if err != nil { t.Fatal(err) }
	wid, eid := created.Workout.ID, created.Exercises[0].ID
	if _, err := m.StartWorkout(ctx, user.ID, wid); err != nil { t.Fatal(err) }
	const operationID = "nonfinite-set-0001"
	nan := math.NaN()
	set := WorkoutSet{WorkoutExerciseID: eid, SetNumber: 1, Repetitions: 8, Weight: &nan}
	if _, err := m.UpsertWorkoutSet(ctx, user.ID, wid, set); !errors.Is(err, ErrInvalidState) { t.Fatalf("direct NaN set: %v", err) }
	if _, err := m.ApplyWorkoutOperation(ctx, user.ID, wid, operationID, "log_set", "payload", set); !errors.Is(err, ErrInvalidState) { t.Fatalf("NaN set: %v", err) }
	current, err := m.GetWorkout(ctx, user.ID, wid)
	if err != nil || len(current.Sets) != 0 || len(m.workoutOperations) != 0 { t.Fatalf("invalid write changed state: sets=%v receipts=%d err=%v", current.Sets, len(m.workoutOperations), err) }
	weight := 50.0
	set.Weight = &weight
	if _, err := m.ApplyWorkoutOperation(ctx, user.ID, wid, operationID, "log_set", "payload", set); err != nil { t.Fatalf("corrected retry: %v", err) }
}
