package store

import (
	"context"
	"encoding/json"
	"math"
)

type workoutOperationReceipt struct {
	WorkoutID string
	Kind string
	PayloadHash string
	Result []byte
}

func validWorkoutSetMeasurements(set WorkoutSet) bool {
	for _, value := range []*float64{set.Weight, set.RPE, set.RIR} {
		if value != nil && (math.IsNaN(*value) || math.IsInf(*value, 0)) { return false }
	}
	return true
}

// The memory implementation serializes keyed operations as a unit. The receipt
// keeps the original result, so a retry after a later edit cannot return that edit.
func (m *Memory) ApplyWorkoutOperation(ctx context.Context, userID, workoutID, operationID, kind, payloadHash string, set WorkoutSet) (WorkoutDetails, error) {
	m.workoutOperationMu.Lock()
	defer m.workoutOperationMu.Unlock()
	if kind == "log_set" && !validWorkoutSetMeasurements(set) { return WorkoutDetails{}, ErrInvalidState }
	key := userID + ":" + operationID
	if receipt, ok := m.workoutOperations[key]; ok {
		if receipt.WorkoutID != workoutID || receipt.Kind != kind || receipt.PayloadHash != payloadHash { return WorkoutDetails{}, ErrIdempotencyConflict }
		var details WorkoutDetails
		if err := json.Unmarshal(receipt.Result, &details); err != nil { return WorkoutDetails{}, err }
		return details, nil
	}
	var details WorkoutDetails
	var err error
	switch kind {
	case "log_set": details, err = m.UpsertWorkoutSet(ctx, userID, workoutID, set)
	case "finish_workout": details, err = m.CompleteWorkout(ctx, userID, workoutID)
	case "cancel_workout": details, err = m.CancelWorkout(ctx, userID, workoutID)
	default: return WorkoutDetails{}, ErrInvalidState
	}
	if err != nil { return WorkoutDetails{}, err }
	encoded, err := json.Marshal(details)
	if err != nil { return WorkoutDetails{}, err }
	m.workoutOperations[key] = workoutOperationReceipt{workoutID, kind, payloadHash, encoded}
	return details, nil
}
