//go:build cgo

package store

import (
	"context"
	"strconv"
	"time"
)

func techniqueStringParam(v string) *string {
	if v == "" {
		return nil
	}
	return sp(v)
}

func techniqueIntParam(v int) *string {
	if v == 0 {
		return nil
	}
	return sp(strconv.Itoa(v))
}

func (p *Postgres) SaveTechniqueAnalysis(ctx context.Context, a TechniqueAnalysis) (TechniqueAnalysis, error) {
	if a.ID == "" {
		a.ID = newID()
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now().UTC()
	}
	if a.CaptureMode == "" {
		a.CaptureMode = "recorded"
	}
	rows, err := p.query(ctx, `INSERT INTO technique_analyses(id,user_id,exercise_key,capture_mode,workout_id,workout_exercise_id,set_number,rep_count,technique_score,rom_score,tempo_score,symmetry_score,stability_score,confidence,duration_ms,algorithm_version,result_json,created_at) VALUES($1::uuid,$2::uuid,$3,$4,$5::uuid,$6::uuid,$7::int,$8::int,$9::int,$10::int,$11::int,$12::int,$13::int,$14::numeric,$15::bigint,$16,$17::jsonb,$18::timestamptz) RETURNING id::text`,
		sp(a.ID), sp(a.UserID), sp(a.ExerciseKey), sp(a.CaptureMode), techniqueStringParam(a.WorkoutID), techniqueStringParam(a.WorkoutExerciseID), techniqueIntParam(a.SetNumber), sp(strconv.Itoa(a.RepCount)), sp(strconv.Itoa(a.TechniqueScore)), sp(strconv.Itoa(a.ROMScore)), sp(strconv.Itoa(a.TempoScore)), sp(strconv.Itoa(a.SymmetryScore)), sp(strconv.Itoa(a.StabilityScore)), sp(strconv.FormatFloat(a.Confidence, 'f', 6, 64)), sp(strconv.FormatInt(a.DurationMS, 10)), sp(a.AlgorithmVersion), sp(a.ResultJSON), sp(a.CreatedAt.Format(time.RFC3339Nano)))
	if err != nil {
		return TechniqueAnalysis{}, err
	}
	if len(rows) == 0 {
		return TechniqueAnalysis{}, ErrNotFound
	}
	return a, nil
}

func (p *Postgres) UpdateTechniqueAnalysisResult(ctx context.Context, userID, id, resultJSON string) error {
	rows, err := p.query(ctx, `UPDATE technique_analyses SET result_json=$1::jsonb WHERE id=$2::uuid AND user_id=$3::uuid RETURNING id::text`, sp(resultJSON), sp(id), sp(userID))
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return ErrNotFound
	}
	return nil
}

func scanTechnique(r []*string) (TechniqueAnalysis, error) {
	created, err := parseTime(val(r, 17))
	if err != nil {
		return TechniqueAnalysis{}, err
	}
	setNumber, _ := strconv.Atoi(val(r, 6))
	rep, _ := strconv.Atoi(val(r, 7))
	ts, _ := strconv.Atoi(val(r, 8))
	rom, _ := strconv.Atoi(val(r, 9))
	tempo, _ := strconv.Atoi(val(r, 10))
	sym, _ := strconv.Atoi(val(r, 11))
	stab, _ := strconv.Atoi(val(r, 12))
	conf, _ := strconv.ParseFloat(val(r, 13), 64)
	dur, _ := strconv.ParseInt(val(r, 14), 10, 64)
	return TechniqueAnalysis{
		ID: val(r, 0), UserID: val(r, 1), ExerciseKey: val(r, 2), CaptureMode: val(r, 3),
		WorkoutID: val(r, 4), WorkoutExerciseID: val(r, 5), SetNumber: setNumber,
		RepCount: rep, TechniqueScore: ts, ROMScore: rom, TempoScore: tempo, SymmetryScore: sym,
		StabilityScore: stab, Confidence: conf, DurationMS: dur, AlgorithmVersion: val(r, 15),
		ResultJSON: val(r, 16), CreatedAt: created,
	}, nil
}

const techniqueSelect = `SELECT id::text,user_id::text,exercise_key,capture_mode,workout_id::text,workout_exercise_id::text,set_number::text,rep_count::text,technique_score::text,rom_score::text,tempo_score::text,symmetry_score::text,stability_score::text,confidence::text,duration_ms::text,algorithm_version,result_json::text,created_at::text FROM technique_analyses`

func (p *Postgres) GetTechniqueAnalysis(ctx context.Context, userID, id string) (TechniqueAnalysis, error) {
	rows, err := p.query(ctx, techniqueSelect+` WHERE id=$1::uuid AND user_id=$2::uuid LIMIT 1`, sp(id), sp(userID))
	if err != nil {
		return TechniqueAnalysis{}, err
	}
	if len(rows) == 0 {
		return TechniqueAnalysis{}, ErrNotFound
	}
	return scanTechnique(rows[0])
}

func (p *Postgres) ListTechniqueAnalyses(ctx context.Context, userID string, limit int) ([]TechniqueAnalysis, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := p.query(ctx, techniqueSelect+` WHERE user_id=$1::uuid ORDER BY created_at DESC LIMIT $2::int`, sp(userID), sp(strconv.Itoa(limit)))
	if err != nil {
		return nil, err
	}
	out := make([]TechniqueAnalysis, 0, len(rows))
	for _, r := range rows {
		a, e := scanTechnique(r)
		if e != nil {
			return nil, e
		}
		out = append(out, a)
	}
	return out, nil
}
