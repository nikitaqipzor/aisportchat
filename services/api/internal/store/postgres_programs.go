//go:build cgo

package store

import (
	"context"
	"strconv"
	"time"
)

func (p *Postgres) CreateProgram(ctx context.Context, program Program, sessions []ProgramSession) (ProgramWithSessions, error) {
	if program.ID == "" {
		program.ID = newID()
	}
	now := time.Now().UTC()
	if program.CreatedAt.IsZero() {
		program.CreatedAt = now
	}
	program.UpdatedAt = now
	if program.Status == "" {
		program.Status = "active"
	}
	err := p.withTx(ctx, func() error {
		if _, err := p.queryLocked(ctx, `UPDATE programs SET status='archived',updated_at=now() WHERE user_id=$1::uuid AND status='active'`, sp(program.UserID)); err != nil {
			return err
		}
		if _, err := p.queryLocked(ctx, `INSERT INTO programs(id,user_id,title,goal_type,weeks,workouts_per_week,environment,status,start_date,created_at,updated_at)
          VALUES($1::uuid,$2::uuid,$3,$4,$5::smallint,$6::smallint,$7,$8,$9::date,$10::timestamptz,$11::timestamptz)`,
			sp(program.ID), sp(program.UserID), sp(program.Title), sp(program.GoalType), sp(strconv.Itoa(program.Weeks)), sp(strconv.Itoa(program.WorkoutsPerWeek)), sp(program.Environment), sp(program.Status), sp(program.StartDate.Format("2006-01-02")), sp(program.CreatedAt.Format(time.RFC3339Nano)), sp(program.UpdatedAt.Format(time.RFC3339Nano))); err != nil {
			return err
		}
		for i := range sessions {
			if sessions[i].ID == "" {
				sessions[i].ID = newID()
			}
			sessions[i].ProgramID = program.ID
			s := sessions[i]
			if _, err := p.queryLocked(ctx, `INSERT INTO program_sessions(id,program_id,week_number,day_index,planned_date,original_date,muscle,secondary_muscle,environment,status,workout_id,completed_at,is_deload,volume_multiplier,intensity_multiplier,planned_sets,adaptation_reason)
              VALUES($1::uuid,$2::uuid,$3::smallint,$4::smallint,$5::date,$6::date,$7,NULLIF($8,''),$9,$10,$11::uuid,$12::timestamptz,$13::boolean,$14::numeric,$15::numeric,$16::smallint,$17)`,
				sp(s.ID), sp(program.ID), sp(strconv.Itoa(s.WeekNumber)), sp(strconv.Itoa(s.DayIndex)), sp(s.PlannedDate.Format("2006-01-02")), sp(s.OriginalDate.Format("2006-01-02")), sp(s.Muscle), sp(s.SecondaryMuscle), sp(s.Environment), sp(s.Status), s.WorkoutID, tp(s.CompletedAt), sp(strconv.FormatBool(s.IsDeload)), sp(strconv.FormatFloat(s.VolumeMultiplier, 'f', 2, 64)), sp(strconv.FormatFloat(s.IntensityMultiplier, 'f', 2, 64)), sp(strconv.Itoa(s.PlannedSets)), sp(s.AdaptationReason)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return ProgramWithSessions{}, err
	}
	return p.GetProgram(ctx, program.UserID, program.ID)
}

func (p *Postgres) GetProgram(ctx context.Context, userID, programID string) (ProgramWithSessions, error) {
	rows, err := p.query(ctx, `SELECT id::text,user_id::text,title,goal_type,weeks::text,workouts_per_week::text,environment,status,start_date::text,created_at::text,updated_at::text FROM programs WHERE id=$1::uuid AND user_id=$2::uuid LIMIT 1`, sp(programID), sp(userID))
	if err != nil {
		return ProgramWithSessions{}, err
	}
	if len(rows) == 0 {
		return ProgramWithSessions{}, ErrNotFound
	}
	program, err := scanProgram(rows[0])
	if err != nil {
		return ProgramWithSessions{}, err
	}
	sessions, err := p.programSessions(ctx, program.ID)
	if err != nil {
		return ProgramWithSessions{}, err
	}
	return ProgramWithSessions{Program: program, Sessions: sessions}, nil
}

func (p *Postgres) GetActiveProgram(ctx context.Context, userID string) (ProgramWithSessions, error) {
	rows, err := p.query(ctx, `SELECT id::text FROM programs WHERE user_id=$1::uuid AND status='active' ORDER BY updated_at DESC LIMIT 1`, sp(userID))
	if err != nil {
		return ProgramWithSessions{}, err
	}
	if len(rows) == 0 {
		return ProgramWithSessions{}, ErrNotFound
	}
	return p.GetProgram(ctx, userID, val(rows[0], 0))
}
func (p *Postgres) ListPrograms(ctx context.Context, userID string, limit int) ([]ProgramWithSessions, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	rows, err := p.query(ctx, `SELECT id::text FROM programs WHERE user_id=$1::uuid ORDER BY updated_at DESC LIMIT $2::int`, sp(userID), sp(strconv.Itoa(limit)))
	if err != nil {
		return nil, err
	}
	out := make([]ProgramWithSessions, 0, len(rows))
	for _, r := range rows {
		item, e := p.GetProgram(ctx, userID, val(r, 0))
		if e != nil {
			return nil, e
		}
		out = append(out, item)
	}
	return out, nil
}
func (p *Postgres) GetProgramSession(ctx context.Context, userID, sessionID string) (ProgramSession, error) {
	rows, err := p.query(ctx, `SELECT s.id::text,s.program_id::text,s.week_number::text,s.day_index::text,s.planned_date::text,s.original_date::text,s.muscle,COALESCE(s.secondary_muscle,''),s.environment,s.status,s.workout_id::text,s.completed_at::text,s.is_deload::text,s.volume_multiplier::text,s.intensity_multiplier::text,s.planned_sets::text,s.adaptation_reason FROM program_sessions s JOIN programs p ON p.id=s.program_id WHERE s.id=$1::uuid AND p.user_id=$2::uuid LIMIT 1`, sp(sessionID), sp(userID))
	if err != nil {
		return ProgramSession{}, err
	}
	if len(rows) == 0 {
		return ProgramSession{}, ErrNotFound
	}
	return scanProgramSession(rows[0])
}
func (p *Postgres) UpdateProgramSession(ctx context.Context, userID string, s ProgramSession) (ProgramSession, error) {
	rows, err := p.query(ctx, `UPDATE program_sessions s SET planned_date=$1::date,status=$2,workout_id=$3::uuid,completed_at=$4::timestamptz,is_deload=$5::boolean,volume_multiplier=$6::numeric,intensity_multiplier=$7::numeric,planned_sets=$8::smallint,adaptation_reason=$9 FROM programs p WHERE s.id=$10::uuid AND p.id=s.program_id AND p.user_id=$11::uuid RETURNING s.id::text,s.program_id::text,s.week_number::text,s.day_index::text,s.planned_date::text,s.original_date::text,s.muscle,COALESCE(s.secondary_muscle,''),s.environment,s.status,s.workout_id::text,s.completed_at::text,s.is_deload::text,s.volume_multiplier::text,s.intensity_multiplier::text,s.planned_sets::text,s.adaptation_reason`, sp(s.PlannedDate.Format("2006-01-02")), sp(s.Status), s.WorkoutID, tp(s.CompletedAt), sp(strconv.FormatBool(s.IsDeload)), sp(strconv.FormatFloat(s.VolumeMultiplier, 'f', 2, 64)), sp(strconv.FormatFloat(s.IntensityMultiplier, 'f', 2, 64)), sp(strconv.Itoa(s.PlannedSets)), sp(s.AdaptationReason), sp(s.ID), sp(userID))
	if err != nil {
		return ProgramSession{}, err
	}
	if len(rows) == 0 {
		return ProgramSession{}, ErrNotFound
	}
	return scanProgramSession(rows[0])
}
func (p *Postgres) FindProgramSessionByWorkout(ctx context.Context, userID, workoutID string) (ProgramSession, error) {
	rows, err := p.query(ctx, `SELECT s.id::text FROM program_sessions s JOIN programs p ON p.id=s.program_id WHERE s.workout_id=$1::uuid AND p.user_id=$2::uuid LIMIT 1`, sp(workoutID), sp(userID))
	if err != nil {
		return ProgramSession{}, err
	}
	if len(rows) == 0 {
		return ProgramSession{}, ErrNotFound
	}
	return p.GetProgramSession(ctx, userID, val(rows[0], 0))
}
func (p *Postgres) UpdateProgramStatus(ctx context.Context, userID, programID, status string) (ProgramWithSessions, error) {
	rows, err := p.query(ctx, `UPDATE programs SET status=$1,updated_at=now() WHERE id=$2::uuid AND user_id=$3::uuid RETURNING id::text`, sp(status), sp(programID), sp(userID))
	if err != nil {
		return ProgramWithSessions{}, err
	}
	if len(rows) == 0 {
		return ProgramWithSessions{}, ErrNotFound
	}
	return p.GetProgram(ctx, userID, programID)
}
func (p *Postgres) programSessions(ctx context.Context, programID string) ([]ProgramSession, error) {
	rows, err := p.query(ctx, `SELECT id::text,program_id::text,week_number::text,day_index::text,planned_date::text,original_date::text,muscle,COALESCE(secondary_muscle,''),environment,status,workout_id::text,completed_at::text,is_deload::text,volume_multiplier::text,intensity_multiplier::text,planned_sets::text,adaptation_reason FROM program_sessions WHERE program_id=$1::uuid ORDER BY week_number,day_index`, sp(programID))
	if err != nil {
		return nil, err
	}
	out := make([]ProgramSession, 0, len(rows))
	for _, r := range rows {
		s, e := scanProgramSession(r)
		if e != nil {
			return nil, e
		}
		out = append(out, s)
	}
	return out, nil
}
func scanProgram(r []*string) (Program, error) {
	weeks, _ := strconv.Atoi(val(r, 4))
	wpw, _ := strconv.Atoi(val(r, 5))
	start, err := time.Parse("2006-01-02", val(r, 8))
	if err != nil {
		return Program{}, err
	}
	created, err := parseTime(val(r, 9))
	if err != nil {
		return Program{}, err
	}
	updated, err := parseTime(val(r, 10))
	if err != nil {
		return Program{}, err
	}
	return Program{ID: val(r, 0), UserID: val(r, 1), Title: val(r, 2), GoalType: val(r, 3), Weeks: weeks, WorkoutsPerWeek: wpw, Environment: val(r, 6), Status: val(r, 7), StartDate: start, CreatedAt: created, UpdatedAt: updated}, nil
}
func scanProgramSession(r []*string) (ProgramSession, error) {
	week, _ := strconv.Atoi(val(r, 2))
	day, _ := strconv.Atoi(val(r, 3))
	planned, err := time.Parse("2006-01-02", val(r, 4))
	if err != nil {
		return ProgramSession{}, err
	}
	original, err := time.Parse("2006-01-02", val(r, 5))
	if err != nil {
		return ProgramSession{}, err
	}
	deload, _ := strconv.ParseBool(val(r, 12))
	vol, _ := strconv.ParseFloat(val(r, 13), 64)
	intensity, _ := strconv.ParseFloat(val(r, 14), 64)
	sets, _ := strconv.Atoi(val(r, 15))
	return ProgramSession{ID: val(r, 0), ProgramID: val(r, 1), WeekNumber: week, DayIndex: day, PlannedDate: planned, OriginalDate: original, Muscle: val(r, 6), SecondaryMuscle: val(r, 7), Environment: val(r, 8), Status: val(r, 9), WorkoutID: stringPtr(r[10]), CompletedAt: timePtr(r[11]), IsDeload: deload, VolumeMultiplier: vol, IntensityMultiplier: intensity, PlannedSets: sets, AdaptationReason: val(r, 16)}, nil
}
func stringPtr(v *string) *string {
	if v == nil || *v == "" {
		return nil
	}
	x := *v
	return &x
}
func timePtr(v *string) *time.Time {
	if v == nil || *v == "" {
		return nil
	}
	t, err := parseTime(*v)
	if err != nil {
		return nil
	}
	return &t
}
