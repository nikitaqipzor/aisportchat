package workouts

import (
	"testing"

	"github.com/example/ai-fitness-os/services/api/internal/catalog"
)

func TestGenerateBackGym(t *testing.T) {
	engine := NewEngine(catalog.Exercises)
	got, err := engine.Generate(GenerateInput{
		Muscle:          "back",
		Environment:     "gym",
		DurationMinutes: 45,
		Level:           "intermediate",
	})
	if err != nil {
		t.Fatalf("Generate returned error: %v", err)
	}
	if got.Muscle != "back" {
		t.Fatalf("expected back, got %s", got.Muscle)
	}
	if len(got.Exercises) == 0 {
		t.Fatal("expected at least one exercise")
	}
	for _, item := range got.Exercises {
		if item.Exercise.PrimaryMuscle != "back" {
			t.Fatalf("unexpected muscle: %s", item.Exercise.PrimaryMuscle)
		}
	}
}

func TestGenerateRejectsEmptyInput(t *testing.T) {
	engine := NewEngine(catalog.Exercises)
	if _, err := engine.Generate(GenerateInput{}); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestGenerateFiltersUnavailableEquipment(t *testing.T) {
	engine := NewEngine(catalog.Exercises)
	got, err := engine.Generate(GenerateInput{
		Muscle: "biceps", Environment: "gym", DurationMinutes: 45, Level: "intermediate",
		EquipmentIDs: []string{"dumbbells"},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range got.Exercises {
		for _, equipment := range item.Exercise.Equipment {
			if equipment != "dumbbells" && equipment != "bodyweight" {
				t.Fatalf("selected unavailable equipment %s in %s", equipment, item.Exercise.ID)
			}
		}
	}
}

func TestGenerateRotatesPreviousWorkout(t *testing.T) {
	engine := NewEngine(catalog.Exercises)
	first, err := engine.Generate(GenerateInput{Muscle: "back", Environment: "gym", DurationMinutes: 45})
	if err != nil {
		t.Fatal(err)
	}
	previous := make([]string, 0, len(first.Exercises))
	for _, item := range first.Exercises {
		previous = append(previous, item.Exercise.ID)
	}
	second, err := engine.Generate(GenerateInput{Muscle: "back", Environment: "gym", DurationMinutes: 45, PreviousExerciseIDs: previous})
	if err != nil {
		t.Fatal(err)
	}
	previousSet := map[string]bool{}
	for _, id := range previous {
		previousSet[id] = true
	}
	kept := 0
	for _, item := range second.Exercises {
		if previousSet[item.Exercise.ID] {
			kept++
		}
	}
	if kept == len(first.Exercises) {
		t.Fatalf("expected at least one exercise rotation; first=%v second=%v", previous, second.Exercises)
	}
	if kept < 3 {
		t.Fatalf("expected stable core of previous session, kept=%d", kept)
	}
}
