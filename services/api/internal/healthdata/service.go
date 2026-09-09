package healthdata

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

type SnapshotInput struct {
	Date                 string   `json:"date"`
	Provider             string   `json:"provider"`
	SourcePackage        string   `json:"source_package"`
	SourceLabel          string   `json:"source_label"`
	Steps                int64    `json:"steps"`
	DistanceM            float64  `json:"distance_m"`
	ActiveCaloriesKcal   float64  `json:"active_calories_kcal"`
	SleepMinutes         int      `json:"sleep_minutes"`
	DeepSleepMinutes     int      `json:"deep_sleep_minutes"`
	LightSleepMinutes    int      `json:"light_sleep_minutes"`
	REMSleepMinutes      int      `json:"rem_sleep_minutes"`
	AwakeMinutes         int      `json:"awake_minutes"`
	ExerciseMinutes      int      `json:"exercise_minutes"`
	ExerciseSessions     int      `json:"exercise_sessions"`
	ExerciseHeartRateAvg *float64 `json:"exercise_heart_rate_avg,omitempty"`
	ExerciseHeartRateMax *float64 `json:"exercise_heart_rate_max,omitempty"`
	RestingHeartRate     *float64 `json:"resting_heart_rate,omitempty"`
	DataTypes            []string `json:"data_types"`
	CapturedAt           string   `json:"captured_at"`
}

type Service struct{ store store.Store }

func NewService(st store.Store) *Service { return &Service{store: st} }

var allowedTypes = map[string]bool{
	"steps": true, "distance": true, "active_calories": true, "sleep": true,
	"exercise": true, "heart_rate": true, "resting_heart_rate": true,
}

func (s *Service) Import(ctx context.Context, userID string, in SnapshotInput) (store.HealthDailySnapshot, error) {
	if _, err := time.Parse("2006-01-02", strings.TrimSpace(in.Date)); err != nil {
		return store.HealthDailySnapshot{}, errors.New("date must be YYYY-MM-DD")
	}
	provider := strings.TrimSpace(in.Provider)
	if provider == "" {
		provider = "health_connect"
	}
	if provider != "health_connect" {
		return store.HealthDailySnapshot{}, errors.New("unsupported health provider")
	}
	pkg := strings.TrimSpace(in.SourcePackage)
	if pkg == "" || len(pkg) > 200 {
		return store.HealthDailySnapshot{}, errors.New("source_package is required")
	}
	if in.Steps < 0 || in.Steps > 200000 || in.DistanceM < 0 || in.DistanceM > 500000 || in.ActiveCaloriesKcal < 0 || in.ActiveCaloriesKcal > 20000 {
		return store.HealthDailySnapshot{}, errors.New("activity metrics are outside supported bounds")
	}
	mins := []int{in.SleepMinutes, in.DeepSleepMinutes, in.LightSleepMinutes, in.REMSleepMinutes, in.AwakeMinutes, in.ExerciseMinutes}
	for _, v := range mins {
		if v < 0 || v > 1440 {
			return store.HealthDailySnapshot{}, errors.New("minute metrics must be between 0 and 1440")
		}
	}
	if in.DeepSleepMinutes+in.LightSleepMinutes+in.REMSleepMinutes > in.SleepMinutes+30 {
		return store.HealthDailySnapshot{}, errors.New("sleep stages exceed total sleep")
	}
	if in.ExerciseSessions < 0 || in.ExerciseSessions > 100 {
		return store.HealthDailySnapshot{}, errors.New("exercise_sessions outside supported bounds")
	}
	for _, p := range []*float64{in.ExerciseHeartRateAvg, in.ExerciseHeartRateMax, in.RestingHeartRate} {
		if p != nil && (*p < 20 || *p > 260) {
			return store.HealthDailySnapshot{}, errors.New("heart rate outside supported bounds")
		}
	}
	captured := time.Now().UTC()
	if strings.TrimSpace(in.CapturedAt) != "" {
		parsed, err := time.Parse(time.RFC3339, in.CapturedAt)
		if err != nil {
			return store.HealthDailySnapshot{}, errors.New("captured_at must be RFC3339")
		}
		if parsed.After(time.Now().UTC().Add(10 * time.Minute)) {
			return store.HealthDailySnapshot{}, errors.New("captured_at cannot be in the future")
		}
		captured = parsed
	}
	types := make([]string, 0, len(in.DataTypes))
	seen := map[string]bool{}
	for _, raw := range in.DataTypes {
		v := strings.TrimSpace(raw)
		if !allowedTypes[v] {
			return store.HealthDailySnapshot{}, errors.New("unsupported health data type: " + v)
		}
		if !seen[v] {
			seen[v] = true
			types = append(types, v)
		}
	}
	sort.Strings(types)
	if len(types) == 0 {
		return store.HealthDailySnapshot{}, errors.New("health snapshot has no available data types")
	}
	return s.store.UpsertHealthDailySnapshot(ctx, store.HealthDailySnapshot{
		UserID: userID, LocalDate: in.Date, Provider: provider, SourcePackage: pkg, SourceLabel: strings.TrimSpace(in.SourceLabel),
		Steps: in.Steps, DistanceM: in.DistanceM, ActiveCaloriesKcal: in.ActiveCaloriesKcal,
		SleepMinutes: in.SleepMinutes, DeepSleepMinutes: in.DeepSleepMinutes, LightSleepMinutes: in.LightSleepMinutes, REMSleepMinutes: in.REMSleepMinutes, AwakeMinutes: in.AwakeMinutes,
		ExerciseMinutes: in.ExerciseMinutes, ExerciseSessions: in.ExerciseSessions, ExerciseHeartRateAvg: in.ExerciseHeartRateAvg, ExerciseHeartRateMax: in.ExerciseHeartRateMax, RestingHeartRate: in.RestingHeartRate,
		DataTypes: types, CapturedAt: captured,
	})
}

func (s *Service) Get(ctx context.Context, userID, date string) (store.HealthDailySnapshot, error) {
	return s.store.GetHealthDailySnapshot(ctx, userID, date)
}
func (s *Service) History(ctx context.Context, userID string, limit int) ([]store.HealthDailySnapshot, error) {
	return s.store.ListHealthDailySnapshots(ctx, userID, limit)
}
