//go:build cgo

package store

import (
	"context"
	"encoding/json"
	"strconv"
	"time"
)

func fptr(v *float64) *string {
	if v == nil {
		return nil
	}
	s := strconv.FormatFloat(*v, 'f', 2, 64)
	return &s
}

func parseHealthFloatPtr(v *string) *float64 {
	if v == nil {
		return nil
	}
	f, err := strconv.ParseFloat(*v, 64)
	if err != nil {
		return nil
	}
	return &f
}

func (p *Postgres) UpsertHealthDailySnapshot(ctx context.Context, in HealthDailySnapshot) (HealthDailySnapshot, error) {
	if in.ID == "" {
		in.ID = newID()
	}
	if in.CapturedAt.IsZero() {
		in.CapturedAt = time.Now().UTC()
	}
	types, err := json.Marshal(in.DataTypes)
	if err != nil {
		return HealthDailySnapshot{}, err
	}
	rows, err := p.query(ctx, `
		INSERT INTO health_daily_snapshots(
		 id,user_id,local_date,provider,source_package,source_label,steps,distance_m,active_calories_kcal,
		 sleep_minutes,deep_sleep_minutes,light_sleep_minutes,rem_sleep_minutes,awake_minutes,
		 exercise_minutes,exercise_sessions,exercise_heart_rate_avg,exercise_heart_rate_max,resting_heart_rate,data_types,captured_at,imported_at)
		VALUES($1::uuid,$2::uuid,$3::date,$4,$5,$6,$7::bigint,$8::numeric,$9::numeric,$10::int,$11::int,$12::int,$13::int,$14::int,$15::int,$16::int,$17::numeric,$18::numeric,$19::numeric,$20::jsonb,$21::timestamptz,now())
		ON CONFLICT(user_id,local_date,source_package) DO UPDATE SET
		 provider=EXCLUDED.provider,source_package=EXCLUDED.source_package,source_label=EXCLUDED.source_label,
		 steps=EXCLUDED.steps,distance_m=EXCLUDED.distance_m,active_calories_kcal=EXCLUDED.active_calories_kcal,
		 sleep_minutes=EXCLUDED.sleep_minutes,deep_sleep_minutes=EXCLUDED.deep_sleep_minutes,light_sleep_minutes=EXCLUDED.light_sleep_minutes,
		 rem_sleep_minutes=EXCLUDED.rem_sleep_minutes,awake_minutes=EXCLUDED.awake_minutes,exercise_minutes=EXCLUDED.exercise_minutes,
		 exercise_sessions=EXCLUDED.exercise_sessions,exercise_heart_rate_avg=EXCLUDED.exercise_heart_rate_avg,
		 exercise_heart_rate_max=EXCLUDED.exercise_heart_rate_max,resting_heart_rate=EXCLUDED.resting_heart_rate,
		 data_types=EXCLUDED.data_types,captured_at=EXCLUDED.captured_at,imported_at=now()
		RETURNING id::text,user_id::text,local_date::text,provider,source_package,source_label,steps::text,distance_m::text,active_calories_kcal::text,
		 sleep_minutes::text,deep_sleep_minutes::text,light_sleep_minutes::text,rem_sleep_minutes::text,awake_minutes::text,
		 exercise_minutes::text,exercise_sessions::text,exercise_heart_rate_avg::text,exercise_heart_rate_max::text,resting_heart_rate::text,
		 data_types::text,captured_at::text,imported_at::text`,
		sp(in.ID), sp(in.UserID), sp(in.LocalDate), sp(in.Provider), sp(in.SourcePackage), sp(in.SourceLabel),
		sp(strconv.FormatInt(in.Steps, 10)), sp(strconv.FormatFloat(in.DistanceM, 'f', 2, 64)), sp(strconv.FormatFloat(in.ActiveCaloriesKcal, 'f', 2, 64)),
		sp(strconv.Itoa(in.SleepMinutes)), sp(strconv.Itoa(in.DeepSleepMinutes)), sp(strconv.Itoa(in.LightSleepMinutes)), sp(strconv.Itoa(in.REMSleepMinutes)), sp(strconv.Itoa(in.AwakeMinutes)),
		sp(strconv.Itoa(in.ExerciseMinutes)), sp(strconv.Itoa(in.ExerciseSessions)), fptr(in.ExerciseHeartRateAvg), fptr(in.ExerciseHeartRateMax), fptr(in.RestingHeartRate), sp(string(types)), sp(in.CapturedAt.Format(time.RFC3339Nano)))
	if err != nil {
		return HealthDailySnapshot{}, err
	}
	if len(rows) == 0 {
		return HealthDailySnapshot{}, ErrNotFound
	}
	return scanHealthSnapshot(rows[0])
}

const healthSnapshotSelect = `SELECT id::text,user_id::text,local_date::text,provider,source_package,source_label,steps::text,distance_m::text,active_calories_kcal::text,sleep_minutes::text,deep_sleep_minutes::text,light_sleep_minutes::text,rem_sleep_minutes::text,awake_minutes::text,exercise_minutes::text,exercise_sessions::text,exercise_heart_rate_avg::text,exercise_heart_rate_max::text,resting_heart_rate::text,data_types::text,captured_at::text,imported_at::text FROM health_daily_snapshots`

func scanHealthSnapshot(r []*string) (HealthDailySnapshot, error) {
	steps, _ := strconv.ParseInt(val(r, 6), 10, 64)
	distance, _ := strconv.ParseFloat(val(r, 7), 64)
	calories, _ := strconv.ParseFloat(val(r, 8), 64)
	sleep, _ := strconv.Atoi(val(r, 9))
	deep, _ := strconv.Atoi(val(r, 10))
	light, _ := strconv.Atoi(val(r, 11))
	rem, _ := strconv.Atoi(val(r, 12))
	awake, _ := strconv.Atoi(val(r, 13))
	exercise, _ := strconv.Atoi(val(r, 14))
	sessions, _ := strconv.Atoi(val(r, 15))
	captured, err := parseTime(val(r, 20))
	if err != nil {
		return HealthDailySnapshot{}, err
	}
	imported, err := parseTime(val(r, 21))
	if err != nil {
		return HealthDailySnapshot{}, err
	}
	types := []string{}
	if raw := val(r, 19); raw != "" {
		_ = json.Unmarshal([]byte(raw), &types)
	}
	return HealthDailySnapshot{ID: val(r, 0), UserID: val(r, 1), LocalDate: val(r, 2), Provider: val(r, 3), SourcePackage: val(r, 4), SourceLabel: val(r, 5), Steps: steps, DistanceM: distance, ActiveCaloriesKcal: calories, SleepMinutes: sleep, DeepSleepMinutes: deep, LightSleepMinutes: light, REMSleepMinutes: rem, AwakeMinutes: awake, ExerciseMinutes: exercise, ExerciseSessions: sessions, ExerciseHeartRateAvg: parseFloatPtr(r[16]), ExerciseHeartRateMax: parseFloatPtr(r[17]), RestingHeartRate: parseFloatPtr(r[18]), DataTypes: types, CapturedAt: captured, ImportedAt: imported}, nil
}

func (p *Postgres) GetHealthDailySnapshot(ctx context.Context, userID, localDate string) (HealthDailySnapshot, error) {
	rows, err := p.query(ctx, healthSnapshotSelect+` WHERE user_id=$1::uuid AND local_date=$2::date
		ORDER BY CASE WHEN source_package='com.xiaomi.wearable' THEN 0 ELSE 10 END, imported_at DESC LIMIT 1`, sp(userID), sp(localDate))
	if err != nil {
		return HealthDailySnapshot{}, err
	}
	if len(rows) == 0 {
		return HealthDailySnapshot{}, ErrNotFound
	}
	return scanHealthSnapshot(rows[0])
}

func (p *Postgres) ListHealthDailySnapshotsForDate(ctx context.Context, userID, localDate string) ([]HealthDailySnapshot, error) {
	rows, err := p.query(ctx, healthSnapshotSelect+` WHERE user_id=$1::uuid AND local_date=$2::date
		ORDER BY CASE WHEN source_package='com.xiaomi.wearable' THEN 0 ELSE 10 END, imported_at DESC`, sp(userID), sp(localDate))
	if err != nil {
		return nil, err
	}
	out := make([]HealthDailySnapshot, 0, len(rows))
	for _, r := range rows {
		v, e := scanHealthSnapshot(r)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, nil
}

func (p *Postgres) ListHealthDailySnapshots(ctx context.Context, userID string, limit int) ([]HealthDailySnapshot, error) {
	if limit <= 0 || limit > 90 {
		limit = 30
	}
	rows, err := p.query(ctx, `SELECT id::text,user_id::text,local_date::text,provider,source_package,source_label,steps::text,distance_m::text,active_calories_kcal::text,sleep_minutes::text,deep_sleep_minutes::text,light_sleep_minutes::text,rem_sleep_minutes::text,awake_minutes::text,exercise_minutes::text,exercise_sessions::text,exercise_heart_rate_avg::text,exercise_heart_rate_max::text,resting_heart_rate::text,data_types::text,captured_at::text,imported_at::text FROM (
		SELECT DISTINCT ON (local_date) * FROM health_daily_snapshots WHERE user_id=$1::uuid
		ORDER BY local_date DESC, CASE WHEN source_package='com.xiaomi.wearable' THEN 0 ELSE 10 END, imported_at DESC
	) resolved ORDER BY local_date DESC LIMIT $2::int`, sp(userID), sp(strconv.Itoa(limit)))
	if err != nil {
		return nil, err
	}
	out := make([]HealthDailySnapshot, 0, len(rows))
	for _, r := range rows {
		v, e := scanHealthSnapshot(r)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, nil
}
