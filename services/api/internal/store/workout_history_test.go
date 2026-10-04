package store

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestWorkoutHistoryFiltersBeforePagination(t *testing.T) {
	ctx := context.Background()
	m := NewMemory()
	for i := 0; i < 121; i++ {
		id := fmt.Sprintf("workout-%03d", i)
		status := "planned"
		if i%2 == 0 {
			status = "completed"
		}
		m.workouts[id] = Workout{ID: id, UserID: "owner", Muscle: "chest", Environment: "gym", Status: status, CreatedAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(i) * time.Minute)}
	}
	m.workouts["other-owner"] = Workout{ID: "other-owner", UserID: "other", Muscle: "chest", Status: "completed", CreatedAt: time.Now()}
	first, err := m.ListWorkoutHistory(ctx, "owner", WorkoutHistoryFilter{Limit: 50, Status: "completed"})
	if err != nil || len(first) != 50 {
		t.Fatalf("first page: length=%d err=%v", len(first), err)
	}
	second, err := m.ListWorkoutHistory(ctx, "owner", WorkoutHistoryFilter{Limit: 50, Offset: 50, Status: "completed"})
	if err != nil || len(second) != 11 {
		t.Fatalf("second page: length=%d err=%v", len(second), err)
	}
	if first[49].Workout.ID == second[0].Workout.ID || second[0].Workout.ID != "workout-020" {
		t.Fatalf("pagination boundary: %s / %s", first[49].Workout.ID, second[0].Workout.ID)
	}
}
