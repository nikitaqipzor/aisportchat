package recovery

import (
	"context"
	"errors"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/catalog"
	"github.com/example/ai-fitness-os/services/api/internal/healthdata"
	"github.com/example/ai-fitness-os/services/api/internal/store"
)

type CheckInInput struct {
	Date           string         `json:"date"`
	SleepHours     float64        `json:"sleep_hours"`
	SleepQuality   int            `json:"sleep_quality"`
	Energy         int            `json:"energy"`
	Stress         int            `json:"stress"`
	MuscleSoreness map[string]int `json:"muscle_soreness"`
	Notes          string         `json:"notes,omitempty"`
}

type FactorScores struct {
	Sleep    int `json:"sleep"`
	Energy   int `json:"energy"`
	Stress   int `json:"stress"`
	Soreness int `json:"soreness"`
	Training int `json:"training_load"`
}

type MuscleStatus struct {
	Muscle                string  `json:"muscle"`
	Score                 int     `json:"score"`
	Status                string  `json:"status"`
	Soreness              int     `json:"soreness"`
	RecentSets7D          float64 `json:"recent_sets_7d"`
	RecentSets48H         float64 `json:"recent_sets_48h"`
	HoursSinceLastWorkout *int    `json:"hours_since_last_workout,omitempty"`
}

type Readiness struct {
	Date                string                     `json:"date"`
	CheckInCompleted    bool                       `json:"check_in_completed"`
	Score               int                        `json:"score"`
	Status              string                     `json:"status"`
	Factors             FactorScores               `json:"factors"`
	VolumeMultiplier    float64                    `json:"volume_multiplier"`
	IntensityMultiplier float64                    `json:"intensity_multiplier"`
	Reasons             []string                   `json:"reasons"`
	Muscles             []MuscleStatus             `json:"muscles"`
	CheckIn             *store.RecoveryCheckIn     `json:"check_in,omitempty"`
	Wearable            *store.HealthDailySnapshot `json:"wearable,omitempty"`
	SleepSource         string                     `json:"sleep_source"`
	HealthInsights      *healthdata.Insights       `json:"health_insights,omitempty"`
}

type Service struct{ store store.Store }

func NewService(st store.Store) *Service { return &Service{store: st} }

func (s *Service) SaveCheckIn(ctx context.Context, userID string, in CheckInInput) (Readiness, error) {
	if _, err := time.Parse("2006-01-02", in.Date); err != nil {
		return Readiness{}, errors.New("date must be YYYY-MM-DD")
	}
	if in.SleepHours < 0 || in.SleepHours > 16 {
		return Readiness{}, errors.New("sleep_hours must be between 0 and 16")
	}
	if in.SleepQuality < 1 || in.SleepQuality > 5 || in.Energy < 1 || in.Energy > 5 || in.Stress < 1 || in.Stress > 5 {
		return Readiness{}, errors.New("sleep_quality, energy and stress must be between 1 and 5")
	}
	valid := map[string]bool{}
	for _, m := range catalog.Muscles {
		valid[m.ID] = true
	}
	clean := map[string]int{}
	for muscle, value := range in.MuscleSoreness {
		if !valid[muscle] {
			return Readiness{}, errors.New("unknown muscle in muscle_soreness")
		}
		if value < 1 || value > 5 {
			return Readiness{}, errors.New("muscle soreness must be between 1 and 5")
		}
		clean[muscle] = value
	}
	_, err := s.store.UpsertRecoveryCheckIn(ctx, store.RecoveryCheckIn{
		UserID: userID, LocalDate: in.Date, SleepHours: in.SleepHours, SleepQuality: in.SleepQuality,
		Energy: in.Energy, Stress: in.Stress, MuscleSoreness: clean, Notes: strings.TrimSpace(in.Notes),
	})
	if err != nil {
		return Readiness{}, err
	}
	return s.Summary(ctx, userID, in.Date)
}

func (s *Service) Summary(ctx context.Context, userID, date string) (Readiness, error) {
	at, err := time.Parse("2006-01-02", date)
	if err != nil {
		return Readiness{}, errors.New("date must be YYYY-MM-DD")
	}
	// For today's score use the current instant; historical dates use end-of-day.
	ref := at.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	if date == time.Now().UTC().Format("2006-01-02") {
		ref = time.Now().UTC()
	}
	check, checkErr := s.store.GetRecoveryCheckIn(ctx, userID, date)
	health, healthErr := s.store.GetHealthDailySnapshot(ctx, userID, date)
	workouts, err := s.store.ListWorkouts(ctx, userID, 100)
	if err != nil {
		return Readiness{}, err
	}
	muscles := muscleStatuses(workouts, ref, sorenessMap(check, checkErr == nil))
	trainingScore := trainingLoadScore(workouts, ref)
	var wearable *store.HealthDailySnapshot
	if healthErr == nil {
		copy := health
		wearable = &copy
	}
	healthHistory, historyErr := s.store.ListHealthDailySnapshots(ctx, userID, 60)
	if historyErr != nil {
		return Readiness{}, historyErr
	}
	healthSources, sourceErr := s.store.ListHealthDailySnapshotsForDate(ctx, userID, date)
	if sourceErr != nil {
		return Readiness{}, sourceErr
	}
	insights := healthdata.BuildInsightsWithSources(date, wearable, healthHistory, healthSources, time.Now().UTC())
	if checkErr != nil {
		reasons := []string{"Заполни короткий check-in, чтобы адаптация тренировки учитывала энергию, стресс и soreness."}
		if wearable != nil && wearable.SleepMinutes > 0 {
			reasons = append(reasons, "Сон уже синхронизирован из "+wearable.SourceLabel+"; субъективный check-in дополнит данные часов.")
		}
		return Readiness{
			Date: date, CheckInCompleted: false, Score: trainingScore, Status: readinessStatus(trainingScore),
			Factors: FactorScores{Training: trainingScore}, VolumeMultiplier: 1, IntensityMultiplier: 1,
			Reasons: reasons, Muscles: muscles, Wearable: wearable, SleepSource: sleepSource(wearable), HealthInsights: &insights,
		}, nil
	}
	effectiveSleepHours := check.SleepHours
	source := "check_in"
	wearableSleepUsable := wearable != nil && wearable.SleepMinutes > 0
	if date == time.Now().UTC().Format("2006-01-02") && insights.Freshness.Status == "stale" {
		wearableSleepUsable = false
	}
	if wearableSleepUsable {
		effectiveSleepHours = float64(wearable.SleepMinutes) / 60.0
		source = "wearable"
	}
	factors := FactorScores{
		Sleep: sleepScorePersonal(effectiveSleepHours, check.SleepQuality, insights.Baseline28D.SleepMinutes), Energy: check.Energy * 20,
		Stress: (6 - check.Stress) * 20, Soreness: sorenessFactor(check.MuscleSoreness), Training: trainingScore,
	}
	score := clampInt(int(math.Round(float64(factors.Sleep)*.30+float64(factors.Energy)*.20+float64(factors.Stress)*.15+float64(factors.Soreness)*.20+float64(factors.Training)*.15)), 0, 100)
	vol, intensity := multipliers(score)
	reasons := reasonsFor(factors, score)
	if source == "wearable" {
		reasons = append(reasons, "Длительность сна взята из Health Connect ("+wearable.SourceLabel+"); качество сна остаётся субъективной оценкой check-in.")
	} else if wearable != nil && wearable.SleepMinutes > 0 && insights.Freshness.Status == "stale" {
		reasons = append(reasons, "Синхронизация часов устарела, поэтому для сна временно использован ручной check-in.")
	}
	if baseline := insights.Baseline28D.SleepMinutes; baseline != nil && baseline.SampleDays >= 5 {
		deltaMinutes := int(math.Round(effectiveSleepHours*60 - baseline.Average))
		if deltaMinutes <= -30 {
			reasons = append(reasons, "Сегодня сна меньше твоей 28-дневной нормы.")
		} else if deltaMinutes >= 45 {
			reasons = append(reasons, "Сегодня сна больше твоей 28-дневной нормы.")
		}
	}
	return Readiness{Date: date, CheckInCompleted: true, Score: score, Status: readinessStatus(score), Factors: factors, VolumeMultiplier: vol, IntensityMultiplier: intensity, Reasons: reasons, Muscles: muscles, CheckIn: &check, Wearable: wearable, SleepSource: source, HealthInsights: &insights}, nil
}

func sleepSource(w *store.HealthDailySnapshot) string {
	if w != nil && w.SleepMinutes > 0 {
		return "wearable"
	}
	return "none"
}

func (s *Service) AdaptationForMuscle(ctx context.Context, userID, date, muscle string) (float64, float64, []string, bool) {
	r, err := s.Summary(ctx, userID, date)
	if err != nil || !r.CheckInCompleted {
		return 1, 1, nil, false
	}
	vol, intensity := r.VolumeMultiplier, r.IntensityMultiplier
	reasons := append([]string(nil), r.Reasons...)
	for _, m := range r.Muscles {
		if m.Muscle != muscle {
			continue
		}
		if m.Score < 40 {
			vol = math.Min(vol, .55)
			intensity = math.Min(intensity, .80)
			reasons = append(reasons, "Выбранная мышца имеет низкий recovery score — объём дополнительно снижен.")
		} else if m.Score < 60 {
			vol = math.Min(vol, .70)
			intensity = math.Min(intensity, .88)
			reasons = append(reasons, "Выбранная мышца восстановилась частично — объём снижен.")
		}
		break
	}
	return vol, intensity, reasons, true
}

func sorenessMap(c store.RecoveryCheckIn, ok bool) map[string]int {
	if !ok {
		return map[string]int{}
	}
	return c.MuscleSoreness
}

func sleepScore(hours float64, quality int) int {
	duration := 100.0
	switch {
	case hours <= 4:
		duration = 20
	case hours < 8:
		duration = 20 + (hours-4)*20
	case hours <= 9:
		duration = 100
	case hours <= 10:
		duration = 95
	default:
		duration = 85
	}
	qualityScore := float64(quality * 20)
	return clampInt(int(math.Round(duration*.65+qualityScore*.35)), 0, 100)
}

func sleepScorePersonal(hours float64, quality int, baseline *healthdata.BaselineMetric) int {
	absolute := sleepScore(hours, quality)
	if baseline == nil || baseline.SampleDays < 5 || baseline.Average <= 0 {
		return absolute
	}
	// Personal history influences the trend component, but a chronically short
	// baseline is not treated as an ideal target. This is fitness guidance, not diagnosis.
	targetHours := baseline.Average / 60.0
	if targetHours < 7 {
		targetHours = 7
	}
	if targetHours > 9 {
		targetHours = 9
	}
	shortage := math.Max(0, targetHours-hours)
	relative := clampInt(int(math.Round(100-shortage*18)), 20, 100)
	return clampInt(int(math.Round(float64(absolute)*0.7+float64(relative)*0.3)), 0, 100)
}

func sorenessFactor(values map[string]int) int {
	if len(values) == 0 {
		return 100
	}
	total := 0.0
	for _, v := range values {
		total += float64((6 - v) * 20)
	}
	return clampInt(int(math.Round(total/float64(len(values)))), 0, 100)
}

func trainingLoadScore(items []store.WorkoutDetails, ref time.Time) int {
	sets48 := 0
	for _, w := range items {
		if w.Workout.Status != "completed" || w.Workout.CompletedAt == nil {
			continue
		}
		d := ref.Sub(*w.Workout.CompletedAt)
		if d < 0 || d > 48*time.Hour {
			continue
		}
		sets48 += len(w.Sets)
	}
	switch {
	case sets48 == 0:
		return 100
	case sets48 <= 8:
		return 92
	case sets48 <= 16:
		return 82
	case sets48 <= 24:
		return 68
	default:
		return 55
	}
}

func muscleStatuses(items []store.WorkoutDetails, ref time.Time, soreness map[string]int) []MuscleStatus {
	type agg struct {
		sets7, sets48 float64
		last          *time.Time
	}
	m := map[string]*agg{}
	for _, muscle := range catalog.Muscles {
		m[muscle.ID] = &agg{}
	}
	for _, w := range items {
		if w.Workout.Status != "completed" || w.Workout.CompletedAt == nil {
			continue
		}
		age := ref.Sub(*w.Workout.CompletedAt)
		if age < 0 || age > 7*24*time.Hour {
			continue
		}
		setsByExercise := map[string]int{}
		for _, set := range w.Sets {
			setsByExercise[set.WorkoutExerciseID]++
		}
		for _, row := range w.Exercises {
			ex, ok := catalog.ExerciseByID(row.ExerciseID)
			if !ok {
				continue
			}
			sets := float64(setsByExercise[row.ID])
			if sets == 0 {
				continue
			}
			apply := func(id string, weighted float64) {
				a := m[id]
				if a == nil {
					return
				}
				a.sets7 += weighted
				if age <= 48*time.Hour {
					a.sets48 += weighted
				}
				if a.last == nil || w.Workout.CompletedAt.After(*a.last) {
					t := *w.Workout.CompletedAt
					a.last = &t
				}
			}
			apply(ex.PrimaryMuscle, sets)
			for _, sec := range ex.SecondaryMuscles {
				apply(sec, sets*.5)
			}
		}
	}
	out := make([]MuscleStatus, 0, len(catalog.Muscles))
	for _, muscle := range catalog.Muscles {
		a := m[muscle.ID]
		sore := soreness[muscle.ID]
		if sore == 0 {
			sore = 1
		}
		score := 100 - (sore-1)*16
		var hours *int
		if a.last != nil {
			h := int(math.Max(0, ref.Sub(*a.last).Hours()))
			hours = &h
			switch {
			case h < 24:
				score -= 26
			case h < 36:
				score -= 18
			case h < 48:
				score -= 10
			case h < 72:
				score -= 4
			}
		}
		score -= int(math.Min(18, a.sets48*1.2))
		score = clampInt(score, 0, 100)
		out = append(out, MuscleStatus{Muscle: muscle.ID, Score: score, Status: muscleStatus(score), Soreness: sore, RecentSets7D: round1(a.sets7), RecentSets48H: round1(a.sets48), HoursSinceLastWorkout: hours})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Score < out[j].Score })
	return out
}

func multipliers(score int) (float64, float64) {
	switch {
	case score >= 80:
		return 1, 1
	case score >= 65:
		return .90, .95
	case score >= 50:
		return .80, .90
	default:
		return .65, .85
	}
}
func readinessStatus(score int) string {
	switch {
	case score >= 80:
		return "ready"
	case score >= 65:
		return "good"
	case score >= 50:
		return "moderate"
	default:
		return "low"
	}
}
func muscleStatus(score int) string {
	switch {
	case score >= 80:
		return "ready"
	case score >= 60:
		return "moderate"
	default:
		return "fatigued"
	}
}
func reasonsFor(f FactorScores, score int) []string {
	out := []string{}
	if f.Sleep < 65 {
		out = append(out, "Сон ниже твоего оптимального диапазона.")
	}
	if f.Energy < 60 {
		out = append(out, "Субъективная энергия сегодня снижена.")
	}
	if f.Stress < 60 {
		out = append(out, "Высокий стресс уменьшает рекомендуемую нагрузку.")
	}
	if f.Soreness < 65 {
		out = append(out, "Выраженная мышечная болезненность снижает готовность.")
	}
	if f.Training < 70 {
		out = append(out, "За последние 48 часов накопилась высокая тренировочная нагрузка.")
	}
	if len(out) == 0 {
		out = append(out, "Основные показатели восстановления сегодня стабильны.")
	}
	if score < 50 {
		out = append(out, "Сегодня лучше избегать максимальных усилий и оставить запас по повторениям.")
	}
	return out
}
func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
func round1(v float64) float64 { return math.Round(v*10) / 10 }
