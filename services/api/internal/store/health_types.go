package store

import "time"

// HealthDailySnapshot is a normalized, device-origin daily aggregate imported from
// an on-device health provider (Health Connect on Android). The server never trusts
// the client to calculate readiness; it only stores bounded factual metrics.
type HealthDailySnapshot struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"-"`
	LocalDate            string    `json:"date"`
	Provider             string    `json:"provider"`
	SourcePackage        string    `json:"source_package"`
	SourceLabel          string    `json:"source_label"`
	Steps                int64     `json:"steps"`
	DistanceM            float64   `json:"distance_m"`
	ActiveCaloriesKcal   float64   `json:"active_calories_kcal"`
	SleepMinutes         int       `json:"sleep_minutes"`
	DeepSleepMinutes     int       `json:"deep_sleep_minutes"`
	LightSleepMinutes    int       `json:"light_sleep_minutes"`
	REMSleepMinutes      int       `json:"rem_sleep_minutes"`
	AwakeMinutes         int       `json:"awake_minutes"`
	ExerciseMinutes      int       `json:"exercise_minutes"`
	ExerciseSessions     int       `json:"exercise_sessions"`
	ExerciseHeartRateAvg *float64  `json:"exercise_heart_rate_avg,omitempty"`
	ExerciseHeartRateMax *float64  `json:"exercise_heart_rate_max,omitempty"`
	RestingHeartRate     *float64  `json:"resting_heart_rate,omitempty"`
	DataTypes            []string  `json:"data_types"`
	CapturedAt           time.Time `json:"captured_at"`
	ImportedAt           time.Time `json:"imported_at"`
}
