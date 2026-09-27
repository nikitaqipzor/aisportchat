//go:build cgo

package store

import (
	"context"
	"encoding/json"
	"strconv"
)

// A receipt is claimed and finalized in the same transaction as the workout
// change. Concurrent retries wait on the unique key and observe the winner.
func (p *Postgres) ApplyWorkoutOperation(ctx context.Context, userID, workoutID, operationID, kind, payloadHash string, set WorkoutSet) (WorkoutDetails, error) {
	var result WorkoutDetails
	err := p.withTx(ctx, func() error {
		claimed, err := p.queryLocked(ctx, `INSERT INTO workout_operations(user_id,operation_id,workout_id,kind,payload_hash) VALUES($1::uuid,$2,$3::uuid,$4,$5) ON CONFLICT(user_id,operation_id) DO NOTHING RETURNING operation_id`, sp(userID), sp(operationID), sp(workoutID), sp(kind), sp(payloadHash))
		if err != nil { return err }
		if len(claimed) == 0 {
			rows, err := p.queryLocked(ctx, `SELECT workout_id::text,kind,payload_hash,result::text FROM workout_operations WHERE user_id=$1::uuid AND operation_id=$2`, sp(userID), sp(operationID))
			if err != nil { return err }
			if len(rows) == 0 { return ErrInvalidState }
			if val(rows[0],0) != workoutID || val(rows[0],1) != kind || val(rows[0],2) != payloadHash { return ErrIdempotencyConflict }
			return json.Unmarshal([]byte(val(rows[0],3)), &result)
		}
		// Serialize the state check and mutation against finish/cancel and other
		// keyed operations on this workout. The receipt claim precedes this lock
		// on every keyed path, and the unique key resolves same-ID retries first.
		locked, err := p.queryLocked(ctx, `SELECT status FROM workouts WHERE id=$1::uuid AND user_id=$2::uuid FOR UPDATE`, sp(workoutID), sp(userID))
		if err != nil { return err }
		if len(locked) == 0 { return ErrNotFound }
		switch kind {
		case "log_set":
			rows, err := p.queryLocked(ctx, `SELECT 1::text FROM workout_exercises e JOIN workouts w ON w.id=e.workout_id WHERE e.id=$1::uuid AND w.id=$2::uuid AND w.user_id=$3::uuid AND w.status='active' AND $4::int BETWEEN 1 AND e.target_sets+3`, sp(set.WorkoutExerciseID), sp(workoutID), sp(userID), sp(strconv.Itoa(set.SetNumber)))
			if err != nil { return err }; if len(rows)==0 { return ErrInvalidState }
			_, err = p.queryLocked(ctx, `INSERT INTO workout_sets(workout_exercise_id,set_number,weight,repetitions,rpe,rir,completed_at) VALUES($1::uuid,$2::int,$3::numeric,$4::int,$5::numeric,$6::numeric,now()) ON CONFLICT(workout_exercise_id,set_number) DO UPDATE SET weight=EXCLUDED.weight,repetitions=EXCLUDED.repetitions,rpe=EXCLUDED.rpe,rir=EXCLUDED.rir,completed_at=now() RETURNING id::text`, sp(set.WorkoutExerciseID), sp(strconv.Itoa(set.SetNumber)), fp(set.Weight), sp(strconv.Itoa(set.Repetitions)), fp(set.RPE), fp(set.RIR))
			if err != nil { return err }
		case "finish_workout":
			rows, err := p.queryLocked(ctx, `WITH progress AS (SELECT COALESCE(SUM(LEAST(x.completed_sets,x.target_sets)),0)::numeric AS done_sets,COALESCE(SUM(x.target_sets),0)::numeric AS target_sets FROM (SELECT e.id,e.target_sets,COUNT(s.id)::int AS completed_sets FROM workout_exercises e LEFT JOIN workout_sets s ON s.workout_exercise_id=e.id WHERE e.workout_id=$1::uuid GROUP BY e.id,e.target_sets) x) UPDATE workouts w SET status='completed',completed_at=now(),duration_seconds=GREATEST(0,EXTRACT(EPOCH FROM (now()-COALESCE(started_at,now())))::int),total_volume=COALESCE((SELECT SUM(COALESCE(s.weight,0)*COALESCE(s.repetitions,0)) FROM workout_sets s JOIN workout_exercises e ON e.id=s.workout_exercise_id WHERE e.workout_id=w.id),0),completion_percent=CASE WHEN progress.target_sets>0 THEN ROUND((progress.done_sets/progress.target_sets)*100,2) ELSE 0 END,ended_early=progress.done_sets<progress.target_sets FROM progress WHERE w.id=$1::uuid AND w.user_id=$2::uuid AND w.status='active' RETURNING w.id::text`, sp(workoutID), sp(userID))
			if err != nil { return err }; if len(rows)==0 { return ErrInvalidState }
		case "cancel_workout":
			rows, err := p.queryLocked(ctx, `UPDATE workouts SET status='cancelled',completed_at=NULL,duration_seconds=CASE WHEN started_at IS NULL THEN duration_seconds ELSE GREATEST(0,EXTRACT(EPOCH FROM (now()-started_at))::int) END WHERE id=$1::uuid AND user_id=$2::uuid AND status IN ('planned','active') RETURNING id::text`, sp(workoutID), sp(userID))
			if err != nil { return err }; if len(rows)==0 { return ErrInvalidState }
		default: return ErrInvalidState
		}
		result, err = p.getWorkoutWithQuery(ctx, userID, workoutID, p.queryLocked)
		if err != nil { return err }
		encoded, err := json.Marshal(result)
		if err != nil { return err }
		_, err = p.queryLocked(ctx, `UPDATE workout_operations SET result=$3::jsonb WHERE user_id=$1::uuid AND operation_id=$2`, sp(userID), sp(operationID), sp(string(encoded)))
		return err
	})
	return result, err
}
