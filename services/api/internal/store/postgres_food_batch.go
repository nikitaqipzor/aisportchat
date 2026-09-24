//go:build cgo

package store

import (
	"context"
	"errors"
	"strconv"
	"time"
)

func (p *Postgres) CreateFoodEntries(ctx context.Context, userID, operationKey, payloadHash string, entries []FoodEntry) (time.Time, error) {
	if len(entries) == 0 { return time.Time{}, ErrInvalidState }
	when := entries[0].LoggedAt
	if when.IsZero() { when = time.Now().UTC() }
	for _, entry := range entries {
		if entry.UserID != userID || entry.LoggedAt != entries[0].LoggedAt || entry.QuantityG <= 0 {
			return time.Time{}, ErrInvalidState
		}
	}
	result := when
	err := p.withTx(ctx, func() error {
		if operationKey != "" {
			rows, err := p.queryLocked(ctx, `INSERT INTO food_entry_operations(user_id, operation_key, payload_hash, logged_at)
				VALUES($1::uuid,$2,$3,$4::timestamptz)
				ON CONFLICT (user_id, operation_key) DO NOTHING RETURNING logged_at::text`,
				sp(userID), sp(operationKey), sp(payloadHash), sp(when.Format(time.RFC3339Nano)))
			if err != nil { return err }
			if len(rows) == 0 {
				rows, err = p.queryLocked(ctx, `SELECT payload_hash,logged_at::text FROM food_entry_operations WHERE user_id=$1::uuid AND operation_key=$2`, sp(userID), sp(operationKey))
				if err != nil { return err }
				if len(rows) != 1 { return errors.New("food operation missing after conflict") }
				if val(rows[0], 0) != payloadHash { return ErrIdempotencyConflict }
				result, err = parseTime(val(rows[0], 1))
				return err
			}
		}
		for _, entry := range entries {
			_, err := p.queryLocked(ctx, `INSERT INTO food_entries(user_id,food_id,food_name,meal_type,logged_at,quantity_g,calories,protein_g,fat_g,carbs_g,fiber_g)
				VALUES($1::uuid,$2,$3,$4,$5::timestamptz,$6::numeric,$7::numeric,$8::numeric,$9::numeric,$10::numeric,$11::numeric)`,
				sp(userID), sp(entry.FoodID), sp(entry.FoodName), sp(entry.MealType), sp(when.Format(time.RFC3339Nano)),
				sp(strconv.FormatFloat(entry.QuantityG,'f',-1,64)),sp(strconv.FormatFloat(entry.Calories,'f',-1,64)),
				sp(strconv.FormatFloat(entry.Protein,'f',-1,64)),sp(strconv.FormatFloat(entry.Fat,'f',-1,64)),
				sp(strconv.FormatFloat(entry.Carbs,'f',-1,64)),sp(strconv.FormatFloat(entry.Fiber,'f',-1,64)))
			if err != nil { return err }
		}
		return nil
	})
	return result, err
}
