package workouts

import (
	"errors"
	"math"
	"sort"

	"github.com/example/ai-fitness-os/services/api/internal/model"
)

type GenerateInput struct {
	Muscle              string   `json:"muscle"`
	Environment         string   `json:"environment"`
	DurationMinutes     int      `json:"duration_minutes,omitempty"`
	Level               string   `json:"level,omitempty"`
	EquipmentIDs        []string `json:"equipment_ids,omitempty"`
	PreviousExerciseIDs []string `json:"previous_exercise_ids,omitempty"`
	LocalDate           string   `json:"local_date,omitempty"`
}

type Engine struct {
	exercises []model.Exercise
}

func NewEngine(exercises []model.Exercise) *Engine {
	return &Engine{exercises: exercises}
}

func (e *Engine) Generate(in GenerateInput) (model.Workout, error) {
	if in.Muscle == "" || in.Environment == "" {
		return model.Workout{}, errors.New("muscle and environment are required")
	}
	if in.DurationMinutes <= 0 {
		in.DurationMinutes = 45
	}

	available := make(map[string]bool, len(in.EquipmentIDs)+1)
	for _, id := range in.EquipmentIDs {
		available[id] = true
	}
	available["bodyweight"] = true

	candidates := make([]model.Exercise, 0)
	for _, ex := range e.exercises {
		if ex.PrimaryMuscle != in.Muscle || !contains(ex.Environment, in.Environment) {
			continue
		}
		if in.Level == "beginner" && ex.Difficulty == "advanced" {
			continue
		}
		if len(in.EquipmentIDs) > 0 && !equipmentAvailable(ex.Equipment, available) {
			continue
		}
		candidates = append(candidates, ex)
	}
	if len(candidates) == 0 {
		return model.Workout{}, errors.New("no exercises available for selection")
	}

	previous := make(map[string]bool, len(in.PreviousExerciseIDs))
	for _, id := range in.PreviousExerciseIDs {
		previous[id] = true
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Compound != candidates[j].Compound {
			return candidates[i].Compound
		}
		return candidates[i].ID < candidates[j].ID
	})

	count := exerciseCountForDuration(in.DurationMinutes)
	if count > len(candidates) {
		count = len(candidates)
	}
	selected := rotateSelection(candidates, previous, count)

	result := model.Workout{
		Muscle:          in.Muscle,
		Environment:     in.Environment,
		DurationMinutes: in.DurationMinutes,
		Exercises:       make([]model.GeneratedExercise, 0, len(selected)),
	}
	for _, ex := range selected {
		result.Exercises = append(result.Exercises, model.GeneratedExercise{
			Exercise:   ex,
			Sets:       ex.DefaultSets,
			RepMin:     ex.RepMin,
			RepMax:     ex.RepMax,
			RestSecond: ex.RestSeconds,
		})
	}
	return result, nil
}

func rotateSelection(candidates []model.Exercise, previous map[string]bool, count int) []model.Exercise {
	if len(previous) == 0 {
		return append([]model.Exercise(nil), candidates[:count]...)
	}

	prev := make([]model.Exercise, 0, count)
	fresh := make([]model.Exercise, 0, count)
	for _, ex := range candidates {
		if previous[ex.ID] {
			prev = append(prev, ex)
		} else {
			fresh = append(fresh, ex)
		}
	}

	// Keep roughly 60–70% of the previous session and rotate the rest.
	keep := int(math.Round(float64(count) * 0.65))
	if keep < 1 {
		keep = 1
	}
	if keep > len(prev) {
		keep = len(prev)
	}
	selected := append([]model.Exercise(nil), prev[:keep]...)
	for _, ex := range fresh {
		if len(selected) == count {
			break
		}
		selected = append(selected, ex)
	}
	for _, ex := range prev[keep:] {
		if len(selected) == count {
			break
		}
		selected = append(selected, ex)
	}
	return selected
}

func exerciseCountForDuration(minutes int) int {
	switch {
	case minutes <= 20:
		return 3
	case minutes <= 35:
		return 4
	case minutes <= 60:
		return 5
	default:
		return 6
	}
}

func equipmentAvailable(required []string, available map[string]bool) bool {
	for _, id := range required {
		if !available[id] {
			return false
		}
	}
	return true
}

func contains(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}
