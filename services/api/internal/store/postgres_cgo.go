//go:build cgo

package store

/*
#cgo pkg-config: libpq
#include <stdlib.h>
#include <libpq-fe.h>
*/
import "C"

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"unsafe"
)

type Postgres struct {
	mu   sync.Mutex
	conn *C.PGconn
	dsn  string
}

type pgErr struct {
	code string
	msg  string
}

func (e pgErr) Error() string { return e.msg }

func NewPostgres(dsn string) (*Postgres, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, errors.New("DATABASE_URL is required")
	}
	cDSN := C.CString(dsn)
	defer C.free(unsafe.Pointer(cDSN))
	conn := C.PQconnectdb(cDSN)
	if conn == nil {
		return nil, errors.New("postgres connection allocation failed")
	}
	if C.PQstatus(conn) != C.CONNECTION_OK {
		msg := strings.TrimSpace(C.GoString(C.PQerrorMessage(conn)))
		C.PQfinish(conn)
		return nil, fmt.Errorf("postgres connect: %s", msg)
	}
	return &Postgres{conn: conn, dsn: dsn}, nil
}

func (p *Postgres) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.conn != nil {
		C.PQfinish(p.conn)
		p.conn = nil
	}
}

func (p *Postgres) Ping(ctx context.Context) error {
	_, err := p.query(ctx, "SELECT 1")
	return err
}

func (p *Postgres) query(ctx context.Context, sql string, params ...*string) ([][]*string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.queryLocked(ctx, sql, params...)
}

func (p *Postgres) exec(ctx context.Context, sql string, params ...*string) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, err := p.queryLocked(ctx, sql, params...)
	return err
}

func (p *Postgres) queryLocked(ctx context.Context, sql string, params ...*string) ([][]*string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if p.conn == nil {
		return nil, errors.New("postgres connection is closed")
	}
	if C.PQstatus(p.conn) != C.CONNECTION_OK {
		C.PQreset(p.conn)
		if C.PQstatus(p.conn) != C.CONNECTION_OK {
			return nil, fmt.Errorf("postgres reconnect: %s", strings.TrimSpace(C.GoString(C.PQerrorMessage(p.conn))))
		}
	}

	cSQL := C.CString(sql)
	defer C.free(unsafe.Pointer(cSQL))

	var cValues **C.char
	var values []*C.char
	if len(params) > 0 {
		values = make([]*C.char, len(params))
		for i, v := range params {
			if v != nil {
				values[i] = C.CString(*v)
				defer C.free(unsafe.Pointer(values[i]))
			}
		}
		cValues = (**C.char)(unsafe.Pointer(&values[0]))
	}

	res := C.PQexecParams(p.conn, cSQL, C.int(len(params)), nil, cValues, nil, nil, 0)
	if res == nil {
		return nil, errors.New("postgres returned nil result")
	}
	defer C.PQclear(res)

	status := C.PQresultStatus(res)
	if status != C.PGRES_TUPLES_OK && status != C.PGRES_COMMAND_OK {
		code := ""
		if ptr := C.PQresultErrorField(res, C.PG_DIAG_SQLSTATE); ptr != nil {
			code = C.GoString(ptr)
		}
		msg := strings.TrimSpace(C.GoString(C.PQresultErrorMessage(res)))
		if msg == "" {
			msg = strings.TrimSpace(C.GoString(C.PQerrorMessage(p.conn)))
		}
		return nil, pgErr{code: code, msg: msg}
	}

	rows := int(C.PQntuples(res))
	cols := int(C.PQnfields(res))
	out := make([][]*string, 0, rows)
	for r := 0; r < rows; r++ {
		row := make([]*string, cols)
		for c := 0; c < cols; c++ {
			if C.PQgetisnull(res, C.int(r), C.int(c)) != 0 {
				continue
			}
			value := C.GoString(C.PQgetvalue(res, C.int(r), C.int(c)))
			row[c] = &value
		}
		out = append(out, row)
	}
	return out, nil
}

func (p *Postgres) withTx(ctx context.Context, fn func() error) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if _, err := p.queryLocked(ctx, "BEGIN"); err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_, _ = p.queryLocked(context.Background(), "ROLLBACK")
		}
	}()
	if err := fn(); err != nil {
		return err
	}
	if _, err := p.queryLocked(ctx, "COMMIT"); err != nil {
		return err
	}
	committed = true
	return nil
}

func (p *Postgres) CreateUser(ctx context.Context, email, passwordHash string) (User, error) {
	rows, err := p.query(ctx, `
        INSERT INTO users(email,password_hash,status)
        VALUES ($1,$2,'active')
        RETURNING id::text,email,password_hash,status,created_at::text`, sp(strings.ToLower(strings.TrimSpace(email))), sp(passwordHash))
	if err != nil {
		var pe pgErr
		if errors.As(err, &pe) && pe.code == "23505" {
			return User{}, ErrEmailExists
		}
		return User{}, err
	}
	return scanUser(rows[0])
}

func (p *Postgres) FindUserByEmail(ctx context.Context, email string) (User, error) {
	rows, err := p.query(ctx, `SELECT id::text,email,password_hash,status,created_at::text FROM users WHERE lower(email)=lower($1) LIMIT 1`, sp(strings.TrimSpace(email)))
	if err != nil {
		return User{}, err
	}
	if len(rows) == 0 {
		return User{}, ErrNotFound
	}
	return scanUser(rows[0])
}

func (p *Postgres) FindUserByID(ctx context.Context, id string) (User, error) {
	rows, err := p.query(ctx, `SELECT id::text,email,password_hash,status,created_at::text FROM users WHERE id=$1::uuid LIMIT 1`, sp(id))
	if err != nil {
		return User{}, err
	}
	if len(rows) == 0 {
		return User{}, ErrNotFound
	}
	return scanUser(rows[0])
}

func (p *Postgres) UpsertProfile(ctx context.Context, in Profile) (Profile, error) {
	var unit *string
	if in.UnitSystem != "" {
		unit = sp(in.UnitSystem)
	}
	var injuries, limitations *string
	if in.Injuries != nil {
		encoded, _ := json.Marshal(in.Injuries)
		injuries = sp(string(encoded))
	}
	if in.Limitations != nil {
		encoded, _ := json.Marshal(in.Limitations)
		limitations = sp(string(encoded))
	}
	rows, err := p.query(ctx, `
		INSERT INTO user_profiles(user_id,birth_date,gender,height_cm,weight_kg,experience_level,unit_system,age_years,injuries,limitations,updated_at)
		VALUES($1::uuid,$2::date,$3,$4::numeric,$5::numeric,$6,COALESCE($7,'metric'),$8::smallint,COALESCE($9::jsonb,'[]'::jsonb),COALESCE($10::jsonb,'[]'::jsonb),now())
		ON CONFLICT(user_id) DO UPDATE SET
          birth_date=COALESCE(EXCLUDED.birth_date,user_profiles.birth_date),
          gender=COALESCE(EXCLUDED.gender,user_profiles.gender),
          height_cm=COALESCE(EXCLUDED.height_cm,user_profiles.height_cm),
          weight_kg=COALESCE(EXCLUDED.weight_kg,user_profiles.weight_kg),
		  experience_level=COALESCE(EXCLUDED.experience_level,user_profiles.experience_level),
		  unit_system=COALESCE($7,user_profiles.unit_system),
		  age_years=COALESCE(EXCLUDED.age_years,user_profiles.age_years),
		  injuries=COALESCE($9::jsonb,user_profiles.injuries),
		  limitations=COALESCE($10::jsonb,user_profiles.limitations),
		  updated_at=now()
		RETURNING user_id::text,birth_date::text,gender,height_cm::text,weight_kg::text,experience_level,unit_system,updated_at::text,age_years::text,injuries::text,limitations::text`,
		sp(in.UserID), in.BirthDate, in.Gender, fp(in.HeightCM), fp(in.WeightKG), in.ExperienceLevel, unit, ip(in.AgeYears), injuries, limitations)
	if err != nil {
		return Profile{}, err
	}
	return scanProfile(rows[0])
}

func (p *Postgres) GetProfile(ctx context.Context, userID string) (Profile, error) {
	rows, err := p.query(ctx, `SELECT user_id::text,birth_date::text,gender,height_cm::text,weight_kg::text,experience_level,unit_system,updated_at::text,age_years::text,injuries::text,limitations::text FROM user_profiles WHERE user_id=$1::uuid`, sp(userID))
	if err != nil {
		return Profile{}, err
	}
	if len(rows) == 0 {
		return Profile{UserID: userID, UnitSystem: "metric", Injuries: []string{}, Limitations: []string{}}, nil
	}
	return scanProfile(rows[0])
}

func (p *Postgres) SetGoal(ctx context.Context, in Goal) (Goal, error) {
	rows, err := p.query(ctx, `
        INSERT INTO user_goals(user_id,goal_type,target_weight_kg,started_at,updated_at)
        VALUES($1::uuid,$2,$3::numeric,now(),now())
        ON CONFLICT(user_id) DO UPDATE SET goal_type=EXCLUDED.goal_type,target_weight_kg=EXCLUDED.target_weight_kg,updated_at=now()
        RETURNING user_id::text,goal_type,target_weight_kg::text,started_at::text`, sp(in.UserID), sp(in.GoalType), fp(in.TargetWeight))
	if err != nil {
		return Goal{}, err
	}
	return scanGoal(rows[0])
}

func (p *Postgres) GetGoal(ctx context.Context, userID string) (Goal, error) {
	rows, err := p.query(ctx, `SELECT user_id::text,goal_type,target_weight_kg::text,started_at::text FROM user_goals WHERE user_id=$1::uuid`, sp(userID))
	if err != nil {
		return Goal{}, err
	}
	if len(rows) == 0 {
		return Goal{}, ErrNotFound
	}
	return scanGoal(rows[0])
}

func (p *Postgres) SetTrainingPreferences(ctx context.Context, in TrainingPreferences) (TrainingPreferences, error) {
	err := p.withTx(ctx, func() error {
		if _, err := p.queryLocked(ctx, `
			INSERT INTO user_training_preferences(user_id,workouts_per_week,session_minutes,updated_at)
			VALUES($1::uuid,$2::smallint,$3::smallint,now())
			ON CONFLICT(user_id) DO UPDATE SET workouts_per_week=EXCLUDED.workouts_per_week,session_minutes=EXCLUDED.session_minutes,updated_at=now()`,
			sp(in.UserID), sp(strconv.Itoa(in.WorkoutsPerWeek)), sp(strconv.Itoa(in.SessionMinutes))); err != nil {
			return err
		}
		if _, err := p.queryLocked(ctx, `DELETE FROM user_training_environments WHERE user_id=$1::uuid`, sp(in.UserID)); err != nil {
			return err
		}
		for _, env := range in.Environments {
			if _, err := p.queryLocked(ctx, `INSERT INTO user_training_environments(user_id,environment) VALUES($1::uuid,$2)`, sp(in.UserID), sp(env)); err != nil {
				return err
			}
		}
		if _, err := p.queryLocked(ctx, `DELETE FROM user_equipment WHERE user_id=$1::uuid`, sp(in.UserID)); err != nil {
			return err
		}
		for _, equipmentID := range in.EquipmentIDs {
			if _, err := p.queryLocked(ctx, `INSERT INTO user_equipment(user_id,equipment_id) VALUES($1::uuid,$2)`, sp(in.UserID), sp(equipmentID)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return TrainingPreferences{}, err
	}
	return p.GetTrainingPreferences(ctx, in.UserID)
}

func (p *Postgres) GetTrainingPreferences(ctx context.Context, userID string) (TrainingPreferences, error) {
	rows, err := p.query(ctx, `
      SELECT p.user_id::text,p.workouts_per_week::text,p.session_minutes::text,p.updated_at::text,
        COALESCE((SELECT string_agg(environment,',' ORDER BY environment) FROM user_training_environments e WHERE e.user_id=p.user_id),''),
        COALESCE((SELECT string_agg(equipment_id,',' ORDER BY equipment_id) FROM user_equipment q WHERE q.user_id=p.user_id),'')
      FROM user_training_preferences p WHERE p.user_id=$1::uuid`, sp(userID))
	if err != nil {
		return TrainingPreferences{}, err
	}
	if len(rows) == 0 {
		return TrainingPreferences{}, ErrNotFound
	}
	r := rows[0]
	wpw, _ := strconv.Atoi(val(r, 1))
	mins, _ := strconv.Atoi(val(r, 2))
	updated, _ := parseTime(val(r, 3))
	return TrainingPreferences{UserID: val(r, 0), WorkoutsPerWeek: wpw, SessionMinutes: mins, UpdatedAt: updated, Environments: splitCSV(val(r, 4)), EquipmentIDs: splitCSV(val(r, 5))}, nil
}

func (p *Postgres) SaveAthleteProfile(ctx context.Context, userID string, update AthleteProfileUpdate) (AthleteProfileUpdate, error) {
	var out AthleteProfileUpdate
	err := p.withTx(ctx, func() error {
		profileIn := update.Profile
		profileIn.UserID = userID
		injuries, err := json.Marshal(profileIn.Injuries)
		if err != nil {
			return err
		}
		limitations, err := json.Marshal(profileIn.Limitations)
		if err != nil {
			return err
		}
		rows, err := p.queryLocked(ctx, `
			INSERT INTO user_profiles(user_id,birth_date,gender,height_cm,weight_kg,experience_level,unit_system,age_years,injuries,limitations,updated_at)
			VALUES($1::uuid,$2::date,$3,$4::numeric,$5::numeric,$6,$7,$8::smallint,$9::jsonb,$10::jsonb,now())
			ON CONFLICT(user_id) DO UPDATE SET birth_date=EXCLUDED.birth_date,gender=EXCLUDED.gender,
			  height_cm=EXCLUDED.height_cm,weight_kg=EXCLUDED.weight_kg,experience_level=EXCLUDED.experience_level,
			  unit_system=EXCLUDED.unit_system,age_years=EXCLUDED.age_years,injuries=EXCLUDED.injuries,
			  limitations=EXCLUDED.limitations,updated_at=now()
			RETURNING user_id::text,birth_date::text,gender,height_cm::text,weight_kg::text,experience_level,unit_system,updated_at::text,age_years::text,injuries::text,limitations::text`,
			sp(userID), profileIn.BirthDate, profileIn.Gender, fp(profileIn.HeightCM), fp(profileIn.WeightKG), profileIn.ExperienceLevel,
			sp(profileIn.UnitSystem), ip(profileIn.AgeYears), sp(string(injuries)), sp(string(limitations)))
		if err != nil {
			return err
		}
		out.Profile, err = scanProfile(rows[0])
		if err != nil {
			return err
		}

		goalIn := update.Goal
		goalIn.UserID = userID
		rows, err = p.queryLocked(ctx, `
			INSERT INTO user_goals(user_id,goal_type,target_weight_kg,started_at,updated_at)
			VALUES($1::uuid,$2,$3::numeric,now(),now())
			ON CONFLICT(user_id) DO UPDATE SET goal_type=EXCLUDED.goal_type,target_weight_kg=EXCLUDED.target_weight_kg,updated_at=now()
			RETURNING user_id::text,goal_type,target_weight_kg::text,started_at::text`, sp(userID), sp(goalIn.GoalType), fp(goalIn.TargetWeight))
		if err != nil {
			return err
		}
		out.Goal, err = scanGoal(rows[0])
		if err != nil {
			return err
		}

		prefs := update.TrainingPreferences
		prefs.UserID = userID
		if _, err = p.queryLocked(ctx, `
			INSERT INTO user_training_preferences(user_id,workouts_per_week,session_minutes,updated_at)
			VALUES($1::uuid,$2::smallint,$3::smallint,now())
			ON CONFLICT(user_id) DO UPDATE SET workouts_per_week=EXCLUDED.workouts_per_week,session_minutes=EXCLUDED.session_minutes,updated_at=now()`,
			sp(userID), sp(strconv.Itoa(prefs.WorkoutsPerWeek)), sp(strconv.Itoa(prefs.SessionMinutes))); err != nil {
			return err
		}
		if _, err = p.queryLocked(ctx, `DELETE FROM user_training_environments WHERE user_id=$1::uuid`, sp(userID)); err != nil {
			return err
		}
		for _, environment := range prefs.Environments {
			if _, err = p.queryLocked(ctx, `INSERT INTO user_training_environments(user_id,environment) VALUES($1::uuid,$2)`, sp(userID), sp(environment)); err != nil {
				return err
			}
		}
		if _, err = p.queryLocked(ctx, `DELETE FROM user_equipment WHERE user_id=$1::uuid`, sp(userID)); err != nil {
			return err
		}
		for _, equipmentID := range prefs.EquipmentIDs {
			if _, err = p.queryLocked(ctx, `INSERT INTO user_equipment(user_id,equipment_id) VALUES($1::uuid,$2)`, sp(userID), sp(equipmentID)); err != nil {
				return err
			}
		}
		rows, err = p.queryLocked(ctx, `SELECT p.user_id::text,p.workouts_per_week::text,p.session_minutes::text,p.updated_at::text,
			COALESCE((SELECT string_agg(environment,',' ORDER BY environment) FROM user_training_environments e WHERE e.user_id=p.user_id),''),
			COALESCE((SELECT string_agg(equipment_id,',' ORDER BY equipment_id) FROM user_equipment q WHERE q.user_id=p.user_id),'')
			FROM user_training_preferences p WHERE p.user_id=$1::uuid`, sp(userID))
		if err != nil {
			return err
		}
		wpw, _ := strconv.Atoi(val(rows[0], 1))
		minutes, _ := strconv.Atoi(val(rows[0], 2))
		updatedAt, _ := parseTime(val(rows[0], 3))
		out.TrainingPreferences = TrainingPreferences{UserID: userID, WorkoutsPerWeek: wpw, SessionMinutes: minutes, UpdatedAt: updatedAt, Environments: splitCSV(val(rows[0], 4)), EquipmentIDs: splitCSV(val(rows[0], 5))}
		return nil
	})
	if err != nil {
		return AthleteProfileUpdate{}, err
	}
	return out, nil
}

func (p *Postgres) GetOnboardingStatus(ctx context.Context, userID string) (OnboardingStatus, error) {
	rows, err := p.query(ctx, `
      SELECT
        EXISTS(SELECT 1 FROM user_profiles WHERE user_id=$1::uuid AND height_cm IS NOT NULL AND weight_kg IS NOT NULL AND experience_level IS NOT NULL)::text,
        EXISTS(SELECT 1 FROM user_goals WHERE user_id=$1::uuid)::text,
        EXISTS(SELECT 1 FROM user_training_preferences p WHERE p.user_id=$1::uuid AND p.workouts_per_week>0 AND p.session_minutes>0 AND EXISTS(SELECT 1 FROM user_training_environments e WHERE e.user_id=p.user_id))::text,
        EXISTS(SELECT 1 FROM user_profiles WHERE user_id=$1::uuid AND onboarding_completed_at IS NOT NULL)::text`, sp(userID))
	if err != nil {
		return OnboardingStatus{}, err
	}
	if len(rows) == 0 {
		return OnboardingStatus{}, nil
	}
	pOK := val(rows[0], 0) == "true"
	gOK := val(rows[0], 1) == "true"
	tOK := val(rows[0], 2) == "true"
	marked := val(rows[0], 3) == "true"
	return OnboardingStatus{ProfileCompleted: pOK, GoalCompleted: gOK, TrainingCompleted: tOK, Completed: marked && pOK && gOK && tOK}, nil
}

func (p *Postgres) CompleteOnboarding(ctx context.Context, userID string) (OnboardingStatus, error) {
	status, err := p.GetOnboardingStatus(ctx, userID)
	if err != nil {
		return status, err
	}
	if !status.ProfileCompleted || !status.GoalCompleted || !status.TrainingCompleted {
		return status, ErrInvalidState
	}
	if err := p.exec(ctx, `UPDATE user_profiles SET onboarding_completed_at=COALESCE(onboarding_completed_at,now()),updated_at=now() WHERE user_id=$1::uuid`, sp(userID)); err != nil {
		return status, err
	}
	status.Completed = true
	return status, nil
}

func (p *Postgres) SaveRefreshSession(ctx context.Context, in RefreshSession) error {
	return p.exec(ctx, `INSERT INTO auth_refresh_sessions(token_hash,user_id,expires_at,revoked_at) VALUES($1,$2::uuid,$3::timestamptz,$4::timestamptz) ON CONFLICT(token_hash) DO UPDATE SET user_id=EXCLUDED.user_id,expires_at=EXCLUDED.expires_at,revoked_at=EXCLUDED.revoked_at`, sp(in.TokenHash), sp(in.UserID), sp(in.ExpiresAt.Format(time.RFC3339Nano)), tp(in.RevokedAt))
}

func (p *Postgres) GetRefreshSession(ctx context.Context, tokenHash string) (RefreshSession, error) {
	rows, err := p.query(ctx, `SELECT token_hash,user_id::text,expires_at::text,revoked_at::text FROM auth_refresh_sessions WHERE token_hash=$1`, sp(tokenHash))
	if err != nil {
		return RefreshSession{}, err
	}
	if len(rows) == 0 {
		return RefreshSession{}, ErrNotFound
	}
	exp, _ := parseTime(val(rows[0], 2))
	var revoked *time.Time
	if rows[0][3] != nil {
		t, _ := parseTime(val(rows[0], 3))
		revoked = &t
	}
	return RefreshSession{TokenHash: val(rows[0], 0), UserID: val(rows[0], 1), ExpiresAt: exp, RevokedAt: revoked}, nil
}

func (p *Postgres) RevokeRefreshSession(ctx context.Context, tokenHash string) error {
	rows, err := p.query(ctx, `UPDATE auth_refresh_sessions SET revoked_at=COALESCE(revoked_at,now()) WHERE token_hash=$1 RETURNING token_hash`, sp(tokenHash))
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return ErrNotFound
	}
	return nil
}
func (p *Postgres) RevokeAllUserSessions(ctx context.Context, userID string) error {
	return p.exec(ctx, `UPDATE auth_refresh_sessions SET revoked_at=COALESCE(revoked_at,now()) WHERE user_id=$1::uuid AND revoked_at IS NULL`, sp(userID))
}

func (p *Postgres) CreateWorkout(ctx context.Context, w Workout, exercises []WorkoutExercise) (WorkoutDetails, error) {
	var workoutID string
	err := p.withTx(ctx, func() error {
		rows, err := p.queryLocked(ctx, `INSERT INTO workouts(user_id,status,environment,muscle,duration_minutes,total_volume) VALUES($1::uuid,$2,$3,$4,$5::int,0) RETURNING id::text`, sp(w.UserID), sp(defaultString(w.Status, "planned")), sp(w.Environment), sp(w.Muscle), sp(strconv.Itoa(w.DurationMinutes)))
		if err != nil {
			return err
		}
		workoutID = val(rows[0], 0)
		for i, ex := range exercises {
			pos := i + 1
			rows, err = p.queryLocked(ctx, `INSERT INTO workout_exercises(workout_id,exercise_id,position,target_sets,target_reps_min,target_reps_max,target_weight,rest_seconds,progression_note) VALUES($1::uuid,$2,$3::int,$4::int,$5::int,$6::int,$7::numeric,$8::int,$9) RETURNING id::text`, sp(workoutID), sp(ex.ExerciseID), sp(strconv.Itoa(pos)), sp(strconv.Itoa(ex.TargetSets)), sp(strconv.Itoa(ex.TargetRepsMin)), sp(strconv.Itoa(ex.TargetRepsMax)), fp(ex.TargetWeight), sp(strconv.Itoa(ex.RestSeconds)), sp(ex.ProgressionNote))
			if err != nil {
				return err
			}
			_ = rows
		}
		return nil
	})
	if err != nil {
		return WorkoutDetails{}, err
	}
	return p.GetWorkout(ctx, w.UserID, workoutID)
}

func (p *Postgres) GetWorkout(ctx context.Context, userID, workoutID string) (WorkoutDetails, error) {
	rows, err := p.query(ctx, `SELECT id::text,user_id::text,muscle,environment,status,duration_minutes::text,started_at::text,completed_at::text,duration_seconds::text,COALESCE(total_volume,0)::text,COALESCE(completion_percent,0)::text,COALESCE(ended_early,false)::text,COALESCE(is_favorite,false)::text,created_at::text FROM workouts WHERE id=$1::uuid AND user_id=$2::uuid`, sp(workoutID), sp(userID))
	if err != nil {
		return WorkoutDetails{}, err
	}
	if len(rows) == 0 {
		return WorkoutDetails{}, ErrNotFound
	}
	w, err := scanWorkout(rows[0])
	if err != nil {
		return WorkoutDetails{}, err
	}
	exRows, err := p.query(ctx, `SELECT id::text,workout_id::text,exercise_id,position::text,target_sets::text,target_reps_min::text,target_reps_max::text,target_weight::text,rest_seconds::text,COALESCE(progression_note,'') FROM workout_exercises WHERE workout_id=$1::uuid ORDER BY position`, sp(workoutID))
	if err != nil {
		return WorkoutDetails{}, err
	}
	exs := make([]WorkoutExercise, 0, len(exRows))
	for _, r := range exRows {
		ex, er := scanWorkoutExercise(r)
		if er != nil {
			return WorkoutDetails{}, er
		}
		exs = append(exs, ex)
	}
	setRows, err := p.query(ctx, `SELECT s.id::text,s.workout_exercise_id::text,s.set_number::text,s.weight::text,s.repetitions::text,s.rpe::text,s.rir::text,s.completed_at::text FROM workout_sets s JOIN workout_exercises e ON e.id=s.workout_exercise_id WHERE e.workout_id=$1::uuid ORDER BY e.position,s.set_number`, sp(workoutID))
	if err != nil {
		return WorkoutDetails{}, err
	}
	sets := make([]WorkoutSet, 0, len(setRows))
	for _, r := range setRows {
		s, er := scanWorkoutSet(r)
		if er != nil {
			return WorkoutDetails{}, er
		}
		sets = append(sets, s)
	}
	return WorkoutDetails{Workout: w, Exercises: exs, Sets: sets}, nil
}

func (p *Postgres) StartWorkout(ctx context.Context, userID, workoutID string) (WorkoutDetails, error) {
	rows, err := p.query(ctx, `UPDATE workouts SET status='active',started_at=COALESCE(started_at,now()) WHERE id=$1::uuid AND user_id=$2::uuid AND status='planned' RETURNING id::text`, sp(workoutID), sp(userID))
	if err != nil {
		return WorkoutDetails{}, err
	}
	if len(rows) == 0 {
		return WorkoutDetails{}, ErrInvalidState
	}
	return p.GetWorkout(ctx, userID, workoutID)
}

func (p *Postgres) UpsertWorkoutSet(ctx context.Context, userID, workoutID string, in WorkoutSet) (WorkoutDetails, error) {
	rows, err := p.query(ctx, `SELECT 1::text FROM workout_exercises e JOIN workouts w ON w.id=e.workout_id WHERE e.id=$1::uuid AND w.id=$2::uuid AND w.user_id=$3::uuid AND w.status='active'`, sp(in.WorkoutExerciseID), sp(workoutID), sp(userID))
	if err != nil {
		return WorkoutDetails{}, err
	}
	if len(rows) == 0 {
		return WorkoutDetails{}, ErrInvalidState
	}
	_, err = p.query(ctx, `INSERT INTO workout_sets(workout_exercise_id,set_number,weight,repetitions,rpe,rir,completed_at) VALUES($1::uuid,$2::int,$3::numeric,$4::int,$5::numeric,$6::numeric,now()) ON CONFLICT(workout_exercise_id,set_number) DO UPDATE SET weight=EXCLUDED.weight,repetitions=EXCLUDED.repetitions,rpe=EXCLUDED.rpe,rir=EXCLUDED.rir,completed_at=now() RETURNING id::text`, sp(in.WorkoutExerciseID), sp(strconv.Itoa(in.SetNumber)), fp(in.Weight), sp(strconv.Itoa(in.Repetitions)), fp(in.RPE), fp(in.RIR))
	if err != nil {
		return WorkoutDetails{}, err
	}
	return p.GetWorkout(ctx, userID, workoutID)
}

func (p *Postgres) ReplaceWorkoutExercise(ctx context.Context, userID, workoutID, workoutExerciseID string, replacement WorkoutExercise) (WorkoutDetails, error) {
	rows, err := p.query(ctx, `UPDATE workout_exercises e SET exercise_id=$1,target_sets=$2::int,target_reps_min=$3::int,target_reps_max=$4::int,target_weight=$5::numeric,rest_seconds=$6::int,progression_note=$7 FROM workouts w WHERE e.id=$8::uuid AND e.workout_id=w.id AND w.id=$9::uuid AND w.user_id=$10::uuid AND w.status IN ('planned','active') AND NOT EXISTS(SELECT 1 FROM workout_sets s WHERE s.workout_exercise_id=e.id) RETURNING e.id::text`, sp(replacement.ExerciseID), sp(strconv.Itoa(replacement.TargetSets)), sp(strconv.Itoa(replacement.TargetRepsMin)), sp(strconv.Itoa(replacement.TargetRepsMax)), fp(replacement.TargetWeight), sp(strconv.Itoa(replacement.RestSeconds)), sp(replacement.ProgressionNote), sp(workoutExerciseID), sp(workoutID), sp(userID))
	if err != nil {
		return WorkoutDetails{}, err
	}
	if len(rows) == 0 {
		return WorkoutDetails{}, ErrInvalidState
	}
	return p.GetWorkout(ctx, userID, workoutID)
}

func (p *Postgres) CompleteWorkout(ctx context.Context, userID, workoutID string) (WorkoutDetails, error) {
	rows, err := p.query(ctx, `WITH progress AS (SELECT COALESCE(SUM(LEAST(x.completed_sets,x.target_sets)),0)::numeric AS done_sets,COALESCE(SUM(x.target_sets),0)::numeric AS target_sets FROM (SELECT e.id,e.target_sets,COUNT(s.id)::int AS completed_sets FROM workout_exercises e LEFT JOIN workout_sets s ON s.workout_exercise_id=e.id WHERE e.workout_id=$1::uuid GROUP BY e.id,e.target_sets) x) UPDATE workouts w SET status='completed',completed_at=now(),duration_seconds=GREATEST(0,EXTRACT(EPOCH FROM (now()-COALESCE(started_at,now())))::int),total_volume=COALESCE((SELECT SUM(COALESCE(s.weight,0)*COALESCE(s.repetitions,0)) FROM workout_sets s JOIN workout_exercises e ON e.id=s.workout_exercise_id WHERE e.workout_id=w.id),0),completion_percent=CASE WHEN progress.target_sets>0 THEN ROUND((progress.done_sets/progress.target_sets)*100,2) ELSE 0 END,ended_early=progress.done_sets<progress.target_sets FROM progress WHERE w.id=$1::uuid AND w.user_id=$2::uuid AND w.status='active' RETURNING w.id::text`, sp(workoutID), sp(userID))
	if err != nil {
		return WorkoutDetails{}, err
	}
	if len(rows) == 0 {
		return WorkoutDetails{}, ErrInvalidState
	}
	return p.GetWorkout(ctx, userID, workoutID)
}

func (p *Postgres) CancelWorkout(ctx context.Context, userID, workoutID string) (WorkoutDetails, error) {
	rows, err := p.query(ctx, `UPDATE workouts SET status='cancelled',completed_at=NULL,duration_seconds=CASE WHEN started_at IS NULL THEN duration_seconds ELSE GREATEST(0,EXTRACT(EPOCH FROM (now()-started_at))::int) END WHERE id=$1::uuid AND user_id=$2::uuid AND status IN ('planned','active') RETURNING id::text`, sp(workoutID), sp(userID))
	if err != nil {
		return WorkoutDetails{}, err
	}
	if len(rows) == 0 {
		return WorkoutDetails{}, ErrInvalidState
	}
	return p.GetWorkout(ctx, userID, workoutID)
}

func (p *Postgres) ListWorkouts(ctx context.Context, userID string, limit int) ([]WorkoutDetails, error) {
	if limit <= 0 || limit > 500 {
		limit = 20
	}
	rows, err := p.query(ctx, `SELECT id::text FROM workouts WHERE user_id=$1::uuid ORDER BY created_at DESC LIMIT $2::int`, sp(userID), sp(strconv.Itoa(limit)))
	if err != nil {
		return nil, err
	}
	out := make([]WorkoutDetails, 0, len(rows))
	for _, r := range rows {
		d, er := p.GetWorkout(ctx, userID, val(r, 0))
		if er != nil {
			return nil, er
		}
		out = append(out, d)
	}
	return out, nil
}

func (p *Postgres) LastCompletedWorkout(ctx context.Context, userID, muscle, environment string) (WorkoutDetails, error) {
	rows, err := p.query(ctx, `SELECT id::text FROM workouts WHERE user_id=$1::uuid AND status='completed' AND muscle=$2 AND environment=$3 ORDER BY COALESCE(completed_at,created_at) DESC LIMIT 1`, sp(userID), sp(muscle), sp(environment))
	if err != nil {
		return WorkoutDetails{}, err
	}
	if len(rows) == 0 {
		return WorkoutDetails{}, ErrNotFound
	}
	return p.GetWorkout(ctx, userID, val(rows[0], 0))
}

func (p *Postgres) LastExercisePerformance(ctx context.Context, userID, exerciseID string) (ExercisePerformance, error) {
	rows, err := p.query(ctx, `SELECT e.exercise_id,s.weight::text,s.repetitions::text,s.rpe::text,s.rir::text,s.completed_at::text FROM workout_sets s JOIN workout_exercises e ON e.id=s.workout_exercise_id JOIN workouts w ON w.id=e.workout_id WHERE w.user_id=$1::uuid AND w.status='completed' AND e.exercise_id=$2 AND s.completed_at IS NOT NULL ORDER BY COALESCE(w.completed_at,w.created_at) DESC,COALESCE(s.weight,0) DESC,s.repetitions DESC LIMIT 1`, sp(userID), sp(exerciseID))
	if err != nil {
		return ExercisePerformance{}, err
	}
	if len(rows) == 0 {
		return ExercisePerformance{}, ErrNotFound
	}
	r := rows[0]
	reps, _ := strconv.Atoi(val(r, 2))
	completed, _ := parseTime(val(r, 5))
	return ExercisePerformance{ExerciseID: val(r, 0), Weight: parseFloatPtr(r[1]), Repetitions: reps, RPE: parseFloatPtr(r[3]), RIR: parseFloatPtr(r[4]), CompletedAt: completed}, nil
}

func (p *Postgres) ActiveWorkout(ctx context.Context, userID string) (WorkoutDetails, error) {
	rows, err := p.query(ctx, `SELECT id::text FROM workouts WHERE user_id=$1::uuid AND status='active' ORDER BY started_at DESC NULLS LAST,created_at DESC LIMIT 1`, sp(userID))
	if err != nil {
		return WorkoutDetails{}, err
	}
	if len(rows) == 0 {
		return WorkoutDetails{}, ErrNotFound
	}
	return p.GetWorkout(ctx, userID, val(rows[0], 0))
}

func (p *Postgres) SetWorkoutFavorite(ctx context.Context, userID, workoutID string, favorite bool) (WorkoutDetails, error) {
	rows, err := p.query(ctx, `UPDATE workouts SET is_favorite=$1::boolean WHERE id=$2::uuid AND user_id=$3::uuid RETURNING id::text`, sp(strconv.FormatBool(favorite)), sp(workoutID), sp(userID))
	if err != nil {
		return WorkoutDetails{}, err
	}
	if len(rows) == 0 {
		return WorkoutDetails{}, ErrNotFound
	}
	return p.GetWorkout(ctx, userID, workoutID)
}

func (p *Postgres) BestPersonalRecord(ctx context.Context, userID, exerciseID, recordType string) (PersonalRecord, error) {
	rows, err := p.query(ctx, `SELECT id::text,user_id::text,workout_id::text,COALESCE(workout_set_id::text,''),exercise_id,record_type,value::text,previous_value::text,achieved_at::text FROM personal_records WHERE user_id=$1::uuid AND exercise_id=$2 AND record_type=$3 ORDER BY value DESC,achieved_at DESC LIMIT 1`, sp(userID), sp(exerciseID), sp(recordType))
	if err != nil {
		return PersonalRecord{}, err
	}
	if len(rows) == 0 {
		return PersonalRecord{}, ErrNotFound
	}
	return scanPersonalRecord(rows[0])
}

func (p *Postgres) SavePersonalRecord(ctx context.Context, in PersonalRecord) (PersonalRecord, error) {
	achieved := in.AchievedAt
	if achieved.IsZero() {
		achieved = time.Now().UTC()
	}
	rows, err := p.query(ctx, `INSERT INTO personal_records(user_id,workout_id,workout_set_id,exercise_id,record_type,value,previous_value,achieved_at) VALUES($1::uuid,$2::uuid,NULLIF($3,'')::uuid,$4,$5,$6::numeric,$7::numeric,$8::timestamptz) RETURNING id::text,user_id::text,workout_id::text,COALESCE(workout_set_id::text,''),exercise_id,record_type,value::text,previous_value::text,achieved_at::text`, sp(in.UserID), sp(in.WorkoutID), sp(in.WorkoutSetID), sp(in.ExerciseID), sp(in.RecordType), sp(strconv.FormatFloat(in.Value, 'f', -1, 64)), fp(in.PreviousValue), sp(achieved.Format(time.RFC3339Nano)))
	if err != nil {
		return PersonalRecord{}, err
	}
	return scanPersonalRecord(rows[0])
}

func (p *Postgres) ListWorkoutPersonalRecords(ctx context.Context, userID, workoutID string) ([]PersonalRecord, error) {
	rows, err := p.query(ctx, `SELECT id::text,user_id::text,workout_id::text,COALESCE(workout_set_id::text,''),exercise_id,record_type,value::text,previous_value::text,achieved_at::text FROM personal_records WHERE user_id=$1::uuid AND workout_id=$2::uuid ORDER BY achieved_at DESC`, sp(userID), sp(workoutID))
	if err != nil {
		return nil, err
	}
	out := make([]PersonalRecord, 0, len(rows))
	for _, row := range rows {
		record, err := scanPersonalRecord(row)
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return out, nil
}

func (p *Postgres) ListPersonalRecords(ctx context.Context, userID string, limit int) ([]PersonalRecord, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := p.query(ctx, `SELECT id::text,user_id::text,workout_id::text,COALESCE(workout_set_id::text,''),exercise_id,record_type,value::text,previous_value::text,achieved_at::text FROM personal_records WHERE user_id=$1::uuid ORDER BY achieved_at DESC LIMIT $2::int`, sp(userID), sp(strconv.Itoa(limit)))
	if err != nil {
		return nil, err
	}
	out := make([]PersonalRecord, 0, len(rows))
	for _, row := range rows {
		record, err := scanPersonalRecord(row)
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return out, nil
}

func (p *Postgres) UpsertNutritionProfile(ctx context.Context, in NutritionProfile) (NutritionProfile, error) {
	rows, err := p.query(ctx, `
        INSERT INTO nutrition_profiles(user_id,goal,activity_level,calculation_mode,calorie_target,protein_target_g,fat_target_g,carb_target_g,updated_at)
        VALUES($1::uuid,$2,$3,$4,$5::int,$6::numeric,$7::numeric,$8::numeric,now())
        ON CONFLICT(user_id) DO UPDATE SET
          goal=EXCLUDED.goal, activity_level=EXCLUDED.activity_level, calculation_mode=EXCLUDED.calculation_mode,
          calorie_target=EXCLUDED.calorie_target, protein_target_g=EXCLUDED.protein_target_g,
          fat_target_g=EXCLUDED.fat_target_g, carb_target_g=EXCLUDED.carb_target_g, updated_at=now()
        RETURNING user_id::text,goal,activity_level,calculation_mode,calorie_target::text,protein_target_g::text,fat_target_g::text,carb_target_g::text,updated_at::text`,
		sp(in.UserID), sp(in.Goal), sp(in.ActivityLevel), sp(in.CalculationMode), sp(strconv.Itoa(in.CalorieTarget)),
		sp(strconv.FormatFloat(in.ProteinTarget, 'f', -1, 64)), sp(strconv.FormatFloat(in.FatTarget, 'f', -1, 64)), sp(strconv.FormatFloat(in.CarbTarget, 'f', -1, 64)))
	if err != nil {
		return NutritionProfile{}, err
	}
	return scanNutritionProfile(rows[0])
}

func (p *Postgres) GetNutritionProfile(ctx context.Context, userID string) (NutritionProfile, error) {
	rows, err := p.query(ctx, `SELECT user_id::text,goal,activity_level,calculation_mode,calorie_target::text,protein_target_g::text,fat_target_g::text,carb_target_g::text,updated_at::text FROM nutrition_profiles WHERE user_id=$1::uuid LIMIT 1`, sp(userID))
	if err != nil {
		return NutritionProfile{}, err
	}
	if len(rows) == 0 {
		return NutritionProfile{}, ErrNotFound
	}
	return scanNutritionProfile(rows[0])
}

func (p *Postgres) SearchFoodItems(ctx context.Context, userID, query string, limit int) ([]FoodItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	q := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	rows, err := p.query(ctx, `SELECT id,name,COALESCE(brand,''),COALESCE(barcode,''),COALESCE(owner_user_id::text,''),kcal_per_100g::text,protein_per_100g::text,fat_per_100g::text,carbs_per_100g::text,fiber_per_100g::text,serving_g::text,source FROM food_items WHERE active=true AND (owner_user_id IS NULL OR owner_user_id=$3::uuid) AND ($1='%%' OR lower(name || ' ' || COALESCE(brand,'')) LIKE $1) ORDER BY CASE WHEN owner_user_id IS NULL THEN 1 ELSE 0 END, name LIMIT $2::int`, sp(q), sp(strconv.Itoa(limit)), sp(userID))
	if err != nil {
		return nil, err
	}
	out := make([]FoodItem, 0, len(rows))
	for _, row := range rows {
		item, err := scanFoodItem(row)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

func (p *Postgres) GetFoodItem(ctx context.Context, userID, foodID string) (FoodItem, error) {
	rows, err := p.query(ctx, `SELECT id,name,COALESCE(brand,''),COALESCE(barcode,''),COALESCE(owner_user_id::text,''),kcal_per_100g::text,protein_per_100g::text,fat_per_100g::text,carbs_per_100g::text,fiber_per_100g::text,serving_g::text,source FROM food_items WHERE id=$1 AND active=true AND (owner_user_id IS NULL OR owner_user_id=$2::uuid) LIMIT 1`, sp(foodID), sp(userID))
	if err != nil {
		return FoodItem{}, err
	}
	if len(rows) == 0 {
		return FoodItem{}, ErrNotFound
	}
	return scanFoodItem(rows[0])
}

func (p *Postgres) GetFoodEntry(ctx context.Context, userID, entryID string) (FoodEntry, error) {
	rows, err := p.query(ctx, `SELECT id::text,user_id::text,food_id,food_name,meal_type,logged_at::text,quantity_g::text,calories::text,protein_g::text,fat_g::text,carbs_g::text,fiber_g::text FROM food_entries WHERE id=$1::uuid AND user_id=$2::uuid LIMIT 1`, sp(entryID), sp(userID))
	if err != nil {
		return FoodEntry{}, err
	}
	if len(rows) == 0 {
		return FoodEntry{}, ErrNotFound
	}
	return scanFoodEntry(rows[0])
}

func (p *Postgres) CreateFoodEntry(ctx context.Context, in FoodEntry) (FoodEntry, error) {
	when := in.LoggedAt
	if when.IsZero() {
		when = time.Now().UTC()
	}
	rows, err := p.query(ctx, `INSERT INTO food_entries(user_id,food_id,food_name,meal_type,logged_at,quantity_g,calories,protein_g,fat_g,carbs_g,fiber_g) VALUES($1::uuid,$2,$3,$4,$5::timestamptz,$6::numeric,$7::numeric,$8::numeric,$9::numeric,$10::numeric,$11::numeric) RETURNING id::text,user_id::text,food_id,food_name,meal_type,logged_at::text,quantity_g::text,calories::text,protein_g::text,fat_g::text,carbs_g::text,fiber_g::text`,
		sp(in.UserID), sp(in.FoodID), sp(in.FoodName), sp(in.MealType), sp(when.Format(time.RFC3339Nano)),
		sp(strconv.FormatFloat(in.QuantityG, 'f', -1, 64)), sp(strconv.FormatFloat(in.Calories, 'f', -1, 64)), sp(strconv.FormatFloat(in.Protein, 'f', -1, 64)), sp(strconv.FormatFloat(in.Fat, 'f', -1, 64)), sp(strconv.FormatFloat(in.Carbs, 'f', -1, 64)), sp(strconv.FormatFloat(in.Fiber, 'f', -1, 64)))
	if err != nil {
		return FoodEntry{}, err
	}
	return scanFoodEntry(rows[0])
}

func (p *Postgres) ListFoodEntries(ctx context.Context, userID string, from, to time.Time) ([]FoodEntry, error) {
	rows, err := p.query(ctx, `SELECT id::text,user_id::text,food_id,food_name,meal_type,logged_at::text,quantity_g::text,calories::text,protein_g::text,fat_g::text,carbs_g::text,fiber_g::text FROM food_entries WHERE user_id=$1::uuid AND logged_at >= $2::timestamptz AND logged_at <= $3::timestamptz ORDER BY logged_at,id`, sp(userID), sp(from.UTC().Format(time.RFC3339Nano)), sp(to.UTC().Format(time.RFC3339Nano)))
	if err != nil {
		return nil, err
	}
	out := make([]FoodEntry, 0, len(rows))
	for _, row := range rows {
		item, err := scanFoodEntry(row)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

func (p *Postgres) DeleteFoodEntry(ctx context.Context, userID, entryID string) error {
	rows, err := p.query(ctx, `DELETE FROM food_entries WHERE id=$1::uuid AND user_id=$2::uuid RETURNING id::text`, sp(entryID), sp(userID))
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return ErrNotFound
	}
	return nil
}

func scanNutritionProfile(r []*string) (NutritionProfile, error) {
	calories, _ := strconv.Atoi(val(r, 4))
	protein, _ := strconv.ParseFloat(defaultString(val(r, 5), "0"), 64)
	fat, _ := strconv.ParseFloat(defaultString(val(r, 6), "0"), 64)
	carbs, _ := strconv.ParseFloat(defaultString(val(r, 7), "0"), 64)
	updated, err := parseTime(val(r, 8))
	if err != nil {
		return NutritionProfile{}, err
	}
	return NutritionProfile{UserID: val(r, 0), Goal: val(r, 1), ActivityLevel: val(r, 2), CalculationMode: val(r, 3), CalorieTarget: calories, ProteinTarget: protein, FatTarget: fat, CarbTarget: carbs, UpdatedAt: updated}, nil
}

func scanFoodItem(r []*string) (FoodItem, error) {
	parse := func(i int) float64 { v, _ := strconv.ParseFloat(defaultString(val(r, i), "0"), 64); return v }
	return FoodItem{ID: val(r, 0), Name: val(r, 1), Brand: val(r, 2), Barcode: val(r, 3), OwnerUserID: val(r, 4), Kcal100: parse(5), Protein100: parse(6), Fat100: parse(7), Carbs100: parse(8), Fiber100: parse(9), ServingG: parse(10), Source: val(r, 11)}, nil
}

func scanFoodEntry(r []*string) (FoodEntry, error) {
	parse := func(i int) float64 { v, _ := strconv.ParseFloat(defaultString(val(r, i), "0"), 64); return v }
	logged, err := parseTime(val(r, 5))
	if err != nil {
		return FoodEntry{}, err
	}
	return FoodEntry{ID: val(r, 0), UserID: val(r, 1), FoodID: val(r, 2), FoodName: val(r, 3), MealType: val(r, 4), LoggedAt: logged, QuantityG: parse(6), Calories: parse(7), Protein: parse(8), Fat: parse(9), Carbs: parse(10), Fiber: parse(11)}, nil
}

func scanUser(r []*string) (User, error) {
	t, err := parseTime(val(r, 4))
	if err != nil {
		return User{}, err
	}
	return User{ID: val(r, 0), Email: val(r, 1), PasswordHash: val(r, 2), Status: val(r, 3), CreatedAt: t}, nil
}
func scanProfile(r []*string) (Profile, error) {
	t, _ := parseTime(val(r, 7))
	var injuries, limitations []string
	_ = json.Unmarshal([]byte(defaultString(val(r, 9), "[]")), &injuries)
	_ = json.Unmarshal([]byte(defaultString(val(r, 10), "[]")), &limitations)
	return Profile{UserID: val(r, 0), BirthDate: r[1], Gender: r[2], HeightCM: parseFloatPtr(r[3]), WeightKG: parseFloatPtr(r[4]), ExperienceLevel: r[5], UnitSystem: val(r, 6), UpdatedAt: t, AgeYears: parseIntPtr(r[8]), Injuries: injuries, Limitations: limitations}, nil
}
func scanGoal(r []*string) (Goal, error) {
	t, _ := parseTime(val(r, 3))
	return Goal{UserID: val(r, 0), GoalType: val(r, 1), TargetWeight: parseFloatPtr(r[2]), StartedAt: t}, nil
}

func scanPersonalRecord(r []*string) (PersonalRecord, error) {
	value, _ := strconv.ParseFloat(defaultString(val(r, 6), "0"), 64)
	achieved, err := parseTime(val(r, 8))
	if err != nil {
		return PersonalRecord{}, err
	}
	return PersonalRecord{
		ID: val(r, 0), UserID: val(r, 1), WorkoutID: val(r, 2), WorkoutSetID: val(r, 3),
		ExerciseID: val(r, 4), RecordType: val(r, 5), Value: value,
		PreviousValue: parseFloatPtr(r[7]), AchievedAt: achieved,
	}, nil
}

func scanWorkout(r []*string) (Workout, error) {
	mins, _ := strconv.Atoi(val(r, 5))
	dur := parseIntPtr(r[8])
	vol, _ := strconv.ParseFloat(defaultString(val(r, 9), "0"), 64)
	completion, _ := strconv.ParseFloat(defaultString(val(r, 10), "0"), 64)
	endedEarly := val(r, 11) == "true"
	favorite := val(r, 12) == "true"
	created, err := parseTime(val(r, 13))
	if err != nil {
		return Workout{}, err
	}
	var started, completed *time.Time
	if r[6] != nil {
		t, _ := parseTime(val(r, 6))
		started = &t
	}
	if r[7] != nil {
		t, _ := parseTime(val(r, 7))
		completed = &t
	}
	return Workout{ID: val(r, 0), UserID: val(r, 1), Muscle: val(r, 2), Environment: val(r, 3), Status: val(r, 4), DurationMinutes: mins, StartedAt: started, CompletedAt: completed, DurationSeconds: dur, TotalVolume: vol, CompletionPercent: completion, EndedEarly: endedEarly, Favorite: favorite, CreatedAt: created}, nil
}
func scanWorkoutExercise(r []*string) (WorkoutExercise, error) {
	pos, _ := strconv.Atoi(val(r, 3))
	sets, _ := strconv.Atoi(val(r, 4))
	min, _ := strconv.Atoi(val(r, 5))
	max, _ := strconv.Atoi(val(r, 6))
	rest, _ := strconv.Atoi(val(r, 8))
	return WorkoutExercise{ID: val(r, 0), WorkoutID: val(r, 1), ExerciseID: val(r, 2), Position: pos, TargetSets: sets, TargetRepsMin: min, TargetRepsMax: max, TargetWeight: parseFloatPtr(r[7]), RestSeconds: rest, ProgressionNote: val(r, 9)}, nil
}
func scanWorkoutSet(r []*string) (WorkoutSet, error) {
	n, _ := strconv.Atoi(val(r, 2))
	reps, _ := strconv.Atoi(val(r, 4))
	var completed *time.Time
	if r[7] != nil {
		t, _ := parseTime(val(r, 7))
		completed = &t
	}
	return WorkoutSet{ID: val(r, 0), WorkoutExerciseID: val(r, 1), SetNumber: n, Weight: parseFloatPtr(r[3]), Repetitions: reps, RPE: parseFloatPtr(r[5]), RIR: parseFloatPtr(r[6]), CompletedAt: completed}, nil
}

func sp(v string) *string { return &v }
func ip(v *int) *string {
	if v == nil {
		return nil
	}
	s := strconv.Itoa(*v)
	return &s
}
func fp(v *float64) *string {
	if v == nil {
		return nil
	}
	s := strconv.FormatFloat(*v, 'f', -1, 64)
	return &s
}
func tp(v *time.Time) *string {
	if v == nil {
		return nil
	}
	s := v.Format(time.RFC3339Nano)
	return &s
}
func val(r []*string, i int) string {
	if i < 0 || i >= len(r) || r[i] == nil {
		return ""
	}
	return *r[i]
}
func parseFloatPtr(v *string) *float64 {
	if v == nil {
		return nil
	}
	f, err := strconv.ParseFloat(*v, 64)
	if err != nil {
		return nil
	}
	return &f
}
func parseIntPtr(v *string) *int {
	if v == nil {
		return nil
	}
	n, err := strconv.Atoi(*v)
	if err != nil {
		return nil
	}
	return &n
}
func parseTime(v string) (time.Time, error) {
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05.999999-07", "2006-01-02 15:04:05-07", "2006-01-02 15:04:05.999999+00", "2006-01-02 15:04:05+00"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, v); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid postgres time %q", v)
}
func defaultString(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
func splitCSV(v string) []string {
	if strings.TrimSpace(v) == "" {
		return nil
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, x := range parts {
		if s := strings.TrimSpace(x); s != "" {
			out = append(out, s)
		}
	}
	return out
}
