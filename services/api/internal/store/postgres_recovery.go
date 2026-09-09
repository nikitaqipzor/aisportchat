//go:build cgo

package store

import (
	"context"
	"encoding/json"
	"strconv"
	"time"
)

func (p *Postgres) UpsertRecoveryCheckIn(ctx context.Context, in RecoveryCheckIn) (RecoveryCheckIn, error) {
	if in.ID == "" {
		in.ID = newID()
	}
	payload, err := json.Marshal(in.MuscleSoreness)
	if err != nil {
		return RecoveryCheckIn{}, err
	}
	rows, err := p.query(ctx, `
		INSERT INTO recovery_checkins(id,user_id,local_date,sleep_hours,sleep_quality,energy,stress,muscle_soreness,notes,created_at,updated_at)
		VALUES($1::uuid,$2::uuid,$3::date,$4::numeric,$5::smallint,$6::smallint,$7::smallint,$8::jsonb,$9,now(),now())
		ON CONFLICT(user_id,local_date) DO UPDATE SET
		 sleep_hours=EXCLUDED.sleep_hours,sleep_quality=EXCLUDED.sleep_quality,energy=EXCLUDED.energy,stress=EXCLUDED.stress,
		 muscle_soreness=EXCLUDED.muscle_soreness,notes=EXCLUDED.notes,updated_at=now()
		RETURNING id::text,user_id::text,local_date::text,sleep_hours::text,sleep_quality::text,energy::text,stress::text,muscle_soreness::text,notes,created_at::text,updated_at::text`,
		sp(in.ID), sp(in.UserID), sp(in.LocalDate), sp(strconv.FormatFloat(in.SleepHours, 'f', 1, 64)), sp(strconv.Itoa(in.SleepQuality)), sp(strconv.Itoa(in.Energy)), sp(strconv.Itoa(in.Stress)), sp(string(payload)), sp(in.Notes))
	if err != nil {
		return RecoveryCheckIn{}, err
	}
	if len(rows) == 0 {
		return RecoveryCheckIn{}, ErrNotFound
	}
	return scanRecovery(rows[0])
}

func scanRecovery(r []*string) (RecoveryCheckIn, error) {
	sleep, _ := strconv.ParseFloat(val(r, 3), 64)
	quality, _ := strconv.Atoi(val(r, 4))
	energy, _ := strconv.Atoi(val(r, 5))
	stress, _ := strconv.Atoi(val(r, 6))
	created, err := parseTime(val(r, 9))
	if err != nil {
		return RecoveryCheckIn{}, err
	}
	updated, err := parseTime(val(r, 10))
	if err != nil {
		return RecoveryCheckIn{}, err
	}
	sore := map[string]int{}
	if raw := val(r, 7); raw != "" {
		_ = json.Unmarshal([]byte(raw), &sore)
	}
	return RecoveryCheckIn{ID: val(r, 0), UserID: val(r, 1), LocalDate: val(r, 2), SleepHours: sleep, SleepQuality: quality, Energy: energy, Stress: stress, MuscleSoreness: sore, Notes: val(r, 8), CreatedAt: created, UpdatedAt: updated}, nil
}

const recoverySelect = `SELECT id::text,user_id::text,local_date::text,sleep_hours::text,sleep_quality::text,energy::text,stress::text,muscle_soreness::text,notes,created_at::text,updated_at::text FROM recovery_checkins`

func (p *Postgres) GetRecoveryCheckIn(ctx context.Context, userID, localDate string) (RecoveryCheckIn, error) {
	rows, err := p.query(ctx, recoverySelect+` WHERE user_id=$1::uuid AND local_date=$2::date LIMIT 1`, sp(userID), sp(localDate))
	if err != nil {
		return RecoveryCheckIn{}, err
	}
	if len(rows) == 0 {
		return RecoveryCheckIn{}, ErrNotFound
	}
	return scanRecovery(rows[0])
}
func (p *Postgres) ListRecoveryCheckIns(ctx context.Context, userID string, limit int) ([]RecoveryCheckIn, error) {
	if limit <= 0 || limit > 90 {
		limit = 30
	}
	rows, err := p.query(ctx, recoverySelect+` WHERE user_id=$1::uuid ORDER BY local_date DESC LIMIT $2::int`, sp(userID), sp(strconv.Itoa(limit)))
	if err != nil {
		return nil, err
	}
	out := make([]RecoveryCheckIn, 0, len(rows))
	for _, r := range rows {
		v, e := scanRecovery(r)
		if e != nil {
			return nil, e
		}
		out = append(out, v)
	}
	return out, nil
}

var _ = time.RFC3339
