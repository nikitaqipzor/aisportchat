package healthdata

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

type BaselineMetric struct {
	Average    float64 `json:"average"`
	SampleDays int     `json:"sample_days"`
	Unit       string  `json:"unit"`
}

type BaselineWindow struct {
	WindowDays       int             `json:"window_days"`
	AvailableDays    int             `json:"available_days"`
	CoveragePercent  int             `json:"coverage_percent"`
	SleepMinutes     *BaselineMetric `json:"sleep_minutes,omitempty"`
	Steps            *BaselineMetric `json:"steps,omitempty"`
	ActiveCalories   *BaselineMetric `json:"active_calories_kcal,omitempty"`
	ExerciseMinutes  *BaselineMetric `json:"exercise_minutes,omitempty"`
	RestingHeartRate *BaselineMetric `json:"resting_heart_rate,omitempty"`
}

type MetricDeviation struct {
	Current      float64 `json:"current"`
	Baseline     float64 `json:"baseline"`
	Delta        float64 `json:"delta"`
	DeltaPercent float64 `json:"delta_percent"`
	Unit         string  `json:"unit"`
}

type HealthFreshness struct {
	Status         string     `json:"status"`
	SyncAgeMinutes int        `json:"sync_age_minutes"`
	CapturedAt     *time.Time `json:"captured_at,omitempty"`
	ImportedAt     *time.Time `json:"imported_at,omitempty"`
}

type HealthSourceCandidate struct {
	SourcePackage string    `json:"source_package"`
	SourceLabel   string    `json:"source_label"`
	Selected      bool      `json:"selected"`
	ImportedAt    time.Time `json:"imported_at"`
	DataTypes     []string  `json:"data_types"`
}

type MetricProvenance struct {
	Metric        string `json:"metric"`
	Available     bool   `json:"available"`
	SourcePackage string `json:"source_package,omitempty"`
	SourceLabel   string `json:"source_label,omitempty"`
}

type Insights struct {
	Date             string                     `json:"date"`
	Snapshot         *store.HealthDailySnapshot `json:"snapshot,omitempty"`
	Baseline7D       BaselineWindow             `json:"baseline_7d"`
	Baseline28D      BaselineWindow             `json:"baseline_28d"`
	Deviations       map[string]MetricDeviation `json:"deviations"`
	Freshness        HealthFreshness            `json:"freshness"`
	Provenance       []MetricProvenance         `json:"provenance"`
	Sources          []HealthSourceCandidate    `json:"sources"`
	ConflictResolved bool                       `json:"conflict_resolved"`
	Confidence       string                     `json:"confidence"`
	ConfidencePct    int                        `json:"confidence_percent"`
	Reasons          []string                   `json:"reasons"`
}

// Insights builds a personal, non-medical trend context. Baselines exclude the
// requested day to avoid comparing a day against an average that already contains it.
func (s *Service) Insights(ctx context.Context, userID, date string) (Insights, error) {
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return Insights{}, err
	}
	history, err := s.store.ListHealthDailySnapshots(ctx, userID, 90)
	if err != nil {
		return Insights{}, err
	}
	var current *store.HealthDailySnapshot
	if snapshot, err := s.store.GetHealthDailySnapshot(ctx, userID, date); err == nil {
		copy := snapshot
		current = &copy
	} else if err != store.ErrNotFound {
		return Insights{}, err
	}
	sources, err := s.store.ListHealthDailySnapshotsForDate(ctx, userID, date)
	if err != nil {
		return Insights{}, err
	}
	return BuildInsightsWithSources(date, current, history, sources, time.Now().UTC()), nil
}

func BuildInsightsWithSources(date string, current *store.HealthDailySnapshot, history []store.HealthDailySnapshot, sources []store.HealthDailySnapshot, now time.Time) Insights {
	out := BuildInsights(date, current, history, now)
	out.Sources = make([]HealthSourceCandidate, 0, len(sources))
	for _, source := range sources {
		selected := current != nil && source.ID == current.ID
		if current != nil && current.ID == "" {
			selected = source.SourcePackage == current.SourcePackage
		}
		out.Sources = append(out.Sources, HealthSourceCandidate{SourcePackage: source.SourcePackage, SourceLabel: source.SourceLabel, Selected: selected, ImportedAt: source.ImportedAt, DataTypes: append([]string(nil), source.DataTypes...)})
	}
	out.ConflictResolved = len(sources) > 1
	if out.ConflictResolved && current != nil {
		out.Reasons = append(out.Reasons, "Найдено несколько health-источников; для этого дня выбран приоритетный источник "+current.SourceLabel+".")
	}
	return out
}

func BuildInsights(date string, current *store.HealthDailySnapshot, history []store.HealthDailySnapshot, now time.Time) Insights {
	b7 := buildBaseline(date, 7, history)
	b28 := buildBaseline(date, 28, history)
	out := Insights{
		Date:        date,
		Snapshot:    current,
		Baseline7D:  b7,
		Baseline28D: b28,
		Deviations:  map[string]MetricDeviation{},
		Freshness:   freshness(current, now),
		Provenance:  provenance(current),
	}
	if current != nil {
		addDeviation(out.Deviations, "sleep_minutes", hasType(*current, "sleep"), float64(current.SleepMinutes), b28.SleepMinutes)
		addDeviation(out.Deviations, "steps", hasType(*current, "steps"), float64(current.Steps), b28.Steps)
		addDeviation(out.Deviations, "active_calories_kcal", hasType(*current, "active_calories"), current.ActiveCaloriesKcal, b28.ActiveCalories)
		addDeviation(out.Deviations, "exercise_minutes", hasType(*current, "exercise"), float64(current.ExerciseMinutes), b28.ExerciseMinutes)
		if current.RestingHeartRate != nil {
			addDeviation(out.Deviations, "resting_heart_rate", true, *current.RestingHeartRate, b28.RestingHeartRate)
		}
	}
	currentCompleteness := 0
	if current != nil {
		for _, key := range []string{"steps", "distance", "active_calories", "sleep", "exercise", "heart_rate"} {
			if hasType(*current, key) {
				currentCompleteness++
			}
		}
	}
	coverageComponent := math.Min(1, float64(b28.AvailableDays)/14.0)
	currentComponent := float64(currentCompleteness) / 6.0
	pct := int(math.Round((coverageComponent*0.65 + currentComponent*0.35) * 100))
	if current == nil {
		pct = int(math.Round(coverageComponent * 65))
	}
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	out.ConfidencePct = pct
	switch {
	case pct >= 75:
		out.Confidence = "high"
	case pct >= 45:
		out.Confidence = "medium"
	default:
		out.Confidence = "low"
	}
	if b28.AvailableDays < 7 {
		out.Reasons = append(out.Reasons, "Нужно минимум 7 дней истории часов, чтобы персональная норма стала устойчивее.")
	} else {
		out.Reasons = append(out.Reasons, "Персональная норма рассчитана по предыдущим дням и не включает сегодняшний день.")
	}
	if out.Freshness.Status == "stale" {
		out.Reasons = append(out.Reasons, "Последняя синхронизация wearable-данных устарела; перед адаптацией тренировки лучше обновить Mi Fitness/Health Connect.")
	}
	if d, ok := out.Deviations["sleep_minutes"]; ok {
		mins := int(math.Round(d.Delta))
		if mins <= -45 {
			out.Reasons = append(out.Reasons, "Сон заметно ниже твоей 28-дневной нормы.")
		}
		if mins >= 45 {
			out.Reasons = append(out.Reasons, "Сон заметно выше твоей 28-дневной нормы.")
		}
	}
	return out
}

func buildBaseline(date string, window int, history []store.HealthDailySnapshot) BaselineWindow {
	target, _ := time.Parse("2006-01-02", date)
	start := target.AddDate(0, 0, -window)
	selected := make([]store.HealthDailySnapshot, 0, window)
	seenDays := map[string]bool{}
	for _, item := range history {
		d, err := time.Parse("2006-01-02", item.LocalDate)
		if err != nil {
			continue
		}
		if !d.Before(target) || d.Before(start) {
			continue
		}
		if seenDays[item.LocalDate] {
			continue
		}
		if !hasAnyBaselineData(item) {
			continue
		}
		seenDays[item.LocalDate] = true
		selected = append(selected, item)
	}
	sort.Slice(selected, func(i, j int) bool { return selected[i].LocalDate < selected[j].LocalDate })
	out := BaselineWindow{WindowDays: window, AvailableDays: len(selected)}
	out.CoveragePercent = int(math.Round(math.Min(1, float64(len(selected))/float64(window)) * 100))
	out.SleepMinutes = avgMetric(selected, "sleep", "min", func(s store.HealthDailySnapshot) (float64, bool) {
		return float64(s.SleepMinutes), hasType(s, "sleep") && s.SleepMinutes > 0
	})
	out.Steps = avgMetric(selected, "steps", "steps", func(s store.HealthDailySnapshot) (float64, bool) { return float64(s.Steps), hasType(s, "steps") })
	out.ActiveCalories = avgMetric(selected, "active_calories", "kcal", func(s store.HealthDailySnapshot) (float64, bool) {
		return s.ActiveCaloriesKcal, hasType(s, "active_calories")
	})
	out.ExerciseMinutes = avgMetric(selected, "exercise", "min", func(s store.HealthDailySnapshot) (float64, bool) {
		return float64(s.ExerciseMinutes), hasType(s, "exercise")
	})
	out.RestingHeartRate = avgMetric(selected, "resting_heart_rate", "bpm", func(s store.HealthDailySnapshot) (float64, bool) {
		if s.RestingHeartRate == nil {
			return 0, false
		}
		return *s.RestingHeartRate, true
	})
	return out
}

func hasAnyBaselineData(item store.HealthDailySnapshot) bool {
	for _, key := range []string{"sleep", "steps", "active_calories", "exercise", "resting_heart_rate"} {
		if hasType(item, key) {
			return true
		}
	}
	return false
}

func avgMetric(items []store.HealthDailySnapshot, _ string, unit string, value func(store.HealthDailySnapshot) (float64, bool)) *BaselineMetric {
	total := 0.0
	count := 0
	for _, item := range items {
		if v, ok := value(item); ok {
			total += v
			count++
		}
	}
	if count == 0 {
		return nil
	}
	return &BaselineMetric{Average: round2(total / float64(count)), SampleDays: count, Unit: unit}
}

func addDeviation(out map[string]MetricDeviation, key string, available bool, current float64, baseline *BaselineMetric) {
	if !available || baseline == nil || baseline.SampleDays == 0 || baseline.Average == 0 {
		return
	}
	delta := current - baseline.Average
	out[key] = MetricDeviation{Current: round2(current), Baseline: baseline.Average, Delta: round2(delta), DeltaPercent: round2(delta / baseline.Average * 100), Unit: baseline.Unit}
}

func freshness(current *store.HealthDailySnapshot, now time.Time) HealthFreshness {
	if current == nil {
		return HealthFreshness{Status: "missing"}
	}
	ref := current.ImportedAt
	if ref.IsZero() {
		ref = current.CapturedAt
	}
	age := now.Sub(ref)
	if age < 0 {
		age = 0
	}
	mins := int(math.Round(age.Minutes()))
	status := "fresh"
	if age > 36*time.Hour {
		status = "stale"
	} else if age > 12*time.Hour {
		status = "aging"
	}
	captured := current.CapturedAt
	imported := current.ImportedAt
	return HealthFreshness{Status: status, SyncAgeMinutes: mins, CapturedAt: &captured, ImportedAt: &imported}
}

func provenance(current *store.HealthDailySnapshot) []MetricProvenance {
	keys := []string{"steps", "distance", "active_calories", "sleep", "exercise", "heart_rate", "resting_heart_rate"}
	out := make([]MetricProvenance, 0, len(keys))
	for _, key := range keys {
		p := MetricProvenance{Metric: key}
		if current != nil && hasType(*current, key) {
			p.Available = true
			p.SourcePackage = current.SourcePackage
			p.SourceLabel = current.SourceLabel
		}
		out = append(out, p)
	}
	return out
}

func hasType(s store.HealthDailySnapshot, want string) bool {
	for _, v := range s.DataTypes {
		if v == want {
			return true
		}
	}
	return false
}
func round2(v float64) float64 { return math.Round(v*100) / 100 }
