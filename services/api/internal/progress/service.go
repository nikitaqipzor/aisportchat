package progress

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

type Service struct{ store store.Store }

func NewService(st store.Store) *Service { return &Service{store: st} }

type MeasurementInput struct {
	LoggedAt   *time.Time `json:"logged_at,omitempty"`
	WeightKG   *float64   `json:"weight_kg,omitempty"`
	WaistCM    *float64   `json:"waist_cm,omitempty"`
	ChestCM    *float64   `json:"chest_cm,omitempty"`
	ShoulderCM *float64   `json:"shoulder_cm,omitempty"`
	ArmCM      *float64   `json:"arm_cm,omitempty"`
	ForearmCM  *float64   `json:"forearm_cm,omitempty"`
	HipCM      *float64   `json:"hip_cm,omitempty"`
	ThighCM    *float64   `json:"thigh_cm,omitempty"`
	CalfCM     *float64   `json:"calf_cm,omitempty"`
	NeckCM     *float64   `json:"neck_cm,omitempty"`
}

type Summary struct {
	Days            int                    `json:"days"`
	Points          int                    `json:"points"`
	Latest          *store.BodyMeasurement `json:"latest,omitempty"`
	WeightStartKG   *float64               `json:"weight_start_kg,omitempty"`
	WeightCurrentKG *float64               `json:"weight_current_kg,omitempty"`
	WeightDeltaKG   *float64               `json:"weight_delta_kg,omitempty"`
	WaistStartCM    *float64               `json:"waist_start_cm,omitempty"`
	WaistCurrentCM  *float64               `json:"waist_current_cm,omitempty"`
	WaistDeltaCM    *float64               `json:"waist_delta_cm,omitempty"`
}

func (s *Service) Log(ctx context.Context, userID string, in MeasurementInput) (store.BodyMeasurement, error) {
	vals := []*float64{in.WeightKG, in.WaistCM, in.ChestCM, in.ShoulderCM, in.ArmCM, in.ForearmCM, in.HipCM, in.ThighCM, in.CalfCM, in.NeckCM}
	found := false
	for _, v := range vals {
		if v != nil {
			found = true
			if *v <= 0 || *v > 400 {
				return store.BodyMeasurement{}, errors.New("measurement values must be between 0 and 400")
			}
		}
	}
	if !found {
		return store.BodyMeasurement{}, errors.New("at least one measurement is required")
	}
	when := time.Now().UTC()
	if in.LoggedAt != nil {
		when = in.LoggedAt.UTC()
	}
	out, err := s.store.CreateBodyMeasurement(ctx, store.BodyMeasurement{UserID: userID, LoggedAt: when, WeightKG: clone(in.WeightKG), WaistCM: clone(in.WaistCM), ChestCM: clone(in.ChestCM), ShoulderCM: clone(in.ShoulderCM), ArmCM: clone(in.ArmCM), ForearmCM: clone(in.ForearmCM), HipCM: clone(in.HipCM), ThighCM: clone(in.ThighCM), CalfCM: clone(in.CalfCM), NeckCM: clone(in.NeckCM)})
	if err != nil {
		return store.BodyMeasurement{}, err
	}
	if in.WeightKG != nil {
		_, _ = s.store.UpsertProfile(ctx, store.Profile{UserID: userID, WeightKG: clone(in.WeightKG)})
	}
	return out, nil
}

func (s *Service) History(ctx context.Context, userID string, days int, now time.Time) ([]store.BodyMeasurement, error) {
	if days <= 0 {
		days = 30
	}
	if days > 730 {
		days = 730
	}
	to := now.UTC()
	from := to.AddDate(0, 0, -days)
	return s.store.ListBodyMeasurements(ctx, userID, from, to)
}

func (s *Service) Summary(ctx context.Context, userID string, days int, now time.Time) (Summary, error) {
	items, err := s.History(ctx, userID, days, now)
	if err != nil {
		return Summary{}, err
	}
	if days <= 0 {
		days = 30
	}
	out := Summary{Days: days, Points: len(items)}
	if len(items) == 0 {
		return out, nil
	}
	latest := items[len(items)-1]
	out.Latest = &latest
	var firstW, lastW, firstWaist, lastWaist *float64
	for i := range items {
		if items[i].WeightKG != nil {
			if firstW == nil {
				firstW = clone(items[i].WeightKG)
			}
			lastW = clone(items[i].WeightKG)
		}
		if items[i].WaistCM != nil {
			if firstWaist == nil {
				firstWaist = clone(items[i].WaistCM)
			}
			lastWaist = clone(items[i].WaistCM)
		}
	}
	out.WeightStartKG = firstW
	out.WeightCurrentKG = lastW
	if firstW != nil && lastW != nil {
		v := round(*lastW - *firstW)
		out.WeightDeltaKG = &v
	}
	out.WaistStartCM = firstWaist
	out.WaistCurrentCM = lastWaist
	if firstWaist != nil && lastWaist != nil {
		v := round(*lastWaist - *firstWaist)
		out.WaistDeltaCM = &v
	}
	return out, nil
}

func clone(v *float64) *float64 {
	if v == nil {
		return nil
	}
	x := *v
	return &x
}
func round(v float64) float64 { return math.Round(v*10) / 10 }
