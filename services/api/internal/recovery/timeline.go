package recovery

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/healthdata"
)

// FactorImpact explains the deterministic contribution of one readiness factor.
// WeightedDeltaPoints follows the same weights used by Summary, so it is an
// explanation of the formula rather than model-generated attribution.
type FactorImpact struct {
	Key                 string   `json:"key"`
	Label               string   `json:"label"`
	TodayScore          int      `json:"today_score"`
	PreviousScore       *int     `json:"previous_score,omitempty"`
	Weight              float64  `json:"weight"`
	WeightedPoints      float64  `json:"weighted_points"`
	WeightedDeltaPoints *float64 `json:"weighted_delta_points,omitempty"`
	Direction           string   `json:"direction"`
	Message             string   `json:"message"`
}

type AdaptationImpact struct {
	VolumeMultiplier       float64 `json:"volume_multiplier"`
	IntensityMultiplier    float64 `json:"intensity_multiplier"`
	VolumeChangePercent    int     `json:"volume_change_percent"`
	IntensityChangePercent int     `json:"intensity_change_percent"`
}

type HealthTrend struct {
	Key              string   `json:"key"`
	Label            string   `json:"label"`
	Unit             string   `json:"unit"`
	Current          *float64 `json:"current,omitempty"`
	Previous         *float64 `json:"previous,omitempty"`
	Baseline7D       *float64 `json:"baseline_7d,omitempty"`
	Baseline28D      *float64 `json:"baseline_28d,omitempty"`
	DeltaVs7D        *float64 `json:"delta_vs_7d,omitempty"`
	DeltaVs28D       *float64 `json:"delta_vs_28d,omitempty"`
	AffectsReadiness bool     `json:"affects_readiness"`
}

type Timeline struct {
	Date                string           `json:"date"`
	Score               int              `json:"score"`
	Status              string           `json:"status"`
	PreviousDate        string           `json:"previous_date"`
	PreviousScore       *int             `json:"previous_score,omitempty"`
	ScoreDelta          *int             `json:"score_delta,omitempty"`
	ComparisonAvailable bool             `json:"comparison_available"`
	Factors             []FactorImpact   `json:"factors"`
	HealthTrends        []HealthTrend    `json:"health_trends"`
	Adaptation          AdaptationImpact `json:"adaptation"`
	Summary             []string         `json:"summary"`
}

type factorDefinition struct {
	key    string
	label  string
	weight float64
	value  func(FactorScores) int
}

var factorDefinitions = []factorDefinition{
	{key: "sleep", label: "Сон", weight: .30, value: func(v FactorScores) int { return v.Sleep }},
	{key: "energy", label: "Энергия", weight: .20, value: func(v FactorScores) int { return v.Energy }},
	{key: "stress", label: "Стресс", weight: .15, value: func(v FactorScores) int { return v.Stress }},
	{key: "soreness", label: "Soreness", weight: .20, value: func(v FactorScores) int { return v.Soreness }},
	{key: "training_load", label: "Нагрузка 48 ч", weight: .15, value: func(v FactorScores) int { return v.Training }},
}

func (s *Service) Timeline(ctx context.Context, userID, date string) (Timeline, error) {
	day, err := time.Parse("2006-01-02", date)
	if err != nil {
		return Timeline{}, err
	}
	current, err := s.Summary(ctx, userID, date)
	if err != nil {
		return Timeline{}, err
	}
	previousDate := day.AddDate(0, 0, -1).Format("2006-01-02")
	previous, previousErr := s.Summary(ctx, userID, previousDate)
	comparison := previousErr == nil && current.CheckInCompleted && previous.CheckInCompleted

	out := Timeline{
		Date:                date,
		Score:               current.Score,
		Status:              current.Status,
		PreviousDate:        previousDate,
		ComparisonAvailable: comparison,
		Adaptation: AdaptationImpact{
			VolumeMultiplier:       current.VolumeMultiplier,
			IntensityMultiplier:    current.IntensityMultiplier,
			VolumeChangePercent:    int(math.Round((current.VolumeMultiplier - 1) * 100)),
			IntensityChangePercent: int(math.Round((current.IntensityMultiplier - 1) * 100)),
		},
	}
	if comparison {
		p := previous.Score
		d := current.Score - previous.Score
		out.PreviousScore = &p
		out.ScoreDelta = &d
	}

	for _, def := range factorDefinitions {
		today := def.value(current.Factors)
		impact := FactorImpact{
			Key:            def.key,
			Label:          def.label,
			TodayScore:     today,
			Weight:         def.weight,
			WeightedPoints: roundTimeline(float64(today) * def.weight),
			Direction:      "stable",
		}
		if comparison {
			prev := def.value(previous.Factors)
			delta := roundTimeline(float64(today-prev) * def.weight)
			impact.PreviousScore = &prev
			impact.WeightedDeltaPoints = &delta
			switch {
			case delta >= .5:
				impact.Direction = "up"
			case delta <= -.5:
				impact.Direction = "down"
			}
			impact.Message = factorDeltaMessage(def.label, delta)
		} else {
			impact.Message = def.label + " сейчас даёт " + formatPoints(impact.WeightedPoints) + " пункта в readiness."
		}
		out.Factors = append(out.Factors, impact)
	}

	out.HealthTrends = buildHealthTrends(current, previous, comparison)

	if comparison {
		sort.SliceStable(out.Factors, func(i, j int) bool {
			ai, aj := 0.0, 0.0
			if out.Factors[i].WeightedDeltaPoints != nil {
				ai = math.Abs(*out.Factors[i].WeightedDeltaPoints)
			}
			if out.Factors[j].WeightedDeltaPoints != nil {
				aj = math.Abs(*out.Factors[j].WeightedDeltaPoints)
			}
			return ai > aj
		})
		if out.ScoreDelta != nil {
			if *out.ScoreDelta > 0 {
				out.Summary = append(out.Summary, "Readiness вырос на "+formatInt(*out.ScoreDelta)+" п. относительно вчера.")
			} else if *out.ScoreDelta < 0 {
				out.Summary = append(out.Summary, "Readiness снизился на "+formatInt(-*out.ScoreDelta)+" п. относительно вчера.")
			} else {
				out.Summary = append(out.Summary, "Readiness не изменился относительно вчера.")
			}
		}
	}
	if out.Adaptation.VolumeChangePercent < 0 || out.Adaptation.IntensityChangePercent < 0 {
		out.Summary = append(out.Summary, "Сегодня Workout Engine ограничивает нагрузку: объём "+formatSignedPercent(out.Adaptation.VolumeChangePercent)+", интенсивность "+formatSignedPercent(out.Adaptation.IntensityChangePercent)+".")
	} else {
		out.Summary = append(out.Summary, "Сегодня Recovery Engine не требует снижения базовой нагрузки.")
	}
	if current.HealthInsights != nil {
		if current.HealthInsights.Freshness.Status == "stale" {
			out.Summary = append(out.Summary, "Wearable-данные устарели; сон часов не используется как свежий источник.")
		} else if d, ok := current.HealthInsights.Deviations["sleep_minutes"]; ok && math.Abs(d.Delta) >= 30 {
			out.Summary = append(out.Summary, "Сон отличается от 28-дневной нормы на "+formatSignedMinutes(int(math.Round(d.Delta)))+".")
		}
	}
	return out, nil
}

func buildHealthTrends(current, previous Readiness, comparison bool) []HealthTrend {
	if current.HealthInsights == nil {
		return nil
	}
	type def struct {
		key, label, unit string
		current          func(*healthdata.Insights) (*float64, *healthdata.BaselineMetric, *healthdata.BaselineMetric)
		previous         func(Readiness) *float64
		affects          bool
	}
	defs := []def{
		{key: "sleep_minutes", label: "Сон", unit: "мин", affects: current.SleepSource == "wearable", current: func(i *healthdata.Insights) (*float64, *healthdata.BaselineMetric, *healthdata.BaselineMetric) {
			if i.Snapshot == nil || i.Snapshot.SleepMinutes <= 0 {
				return nil, i.Baseline7D.SleepMinutes, i.Baseline28D.SleepMinutes
			}
			v := float64(i.Snapshot.SleepMinutes)
			return &v, i.Baseline7D.SleepMinutes, i.Baseline28D.SleepMinutes
		}, previous: func(r Readiness) *float64 {
			if r.Wearable == nil || r.Wearable.SleepMinutes <= 0 {
				return nil
			}
			v := float64(r.Wearable.SleepMinutes)
			return &v
		}},
		{key: "steps", label: "Шаги", unit: "шагов", current: func(i *healthdata.Insights) (*float64, *healthdata.BaselineMetric, *healthdata.BaselineMetric) {
			if i.Snapshot == nil {
				return nil, i.Baseline7D.Steps, i.Baseline28D.Steps
			}
			v := float64(i.Snapshot.Steps)
			return &v, i.Baseline7D.Steps, i.Baseline28D.Steps
		}, previous: func(r Readiness) *float64 {
			if r.Wearable == nil {
				return nil
			}
			v := float64(r.Wearable.Steps)
			return &v
		}},
		{key: "active_calories_kcal", label: "Активные ккал", unit: "ккал", current: func(i *healthdata.Insights) (*float64, *healthdata.BaselineMetric, *healthdata.BaselineMetric) {
			if i.Snapshot == nil {
				return nil, i.Baseline7D.ActiveCalories, i.Baseline28D.ActiveCalories
			}
			v := i.Snapshot.ActiveCaloriesKcal
			return &v, i.Baseline7D.ActiveCalories, i.Baseline28D.ActiveCalories
		}, previous: func(r Readiness) *float64 {
			if r.Wearable == nil {
				return nil
			}
			v := r.Wearable.ActiveCaloriesKcal
			return &v
		}},
		{key: "exercise_minutes", label: "Активность", unit: "мин", current: func(i *healthdata.Insights) (*float64, *healthdata.BaselineMetric, *healthdata.BaselineMetric) {
			if i.Snapshot == nil {
				return nil, i.Baseline7D.ExerciseMinutes, i.Baseline28D.ExerciseMinutes
			}
			v := float64(i.Snapshot.ExerciseMinutes)
			return &v, i.Baseline7D.ExerciseMinutes, i.Baseline28D.ExerciseMinutes
		}, previous: func(r Readiness) *float64 {
			if r.Wearable == nil {
				return nil
			}
			v := float64(r.Wearable.ExerciseMinutes)
			return &v
		}},
	}
	out := make([]HealthTrend, 0, len(defs))
	for _, d := range defs {
		cur, b7, b28 := d.current(current.HealthInsights)
		if cur == nil && b7 == nil && b28 == nil {
			continue
		}
		t := HealthTrend{Key: d.key, Label: d.label, Unit: d.unit, Current: cur, AffectsReadiness: d.affects}
		if comparison {
			t.Previous = d.previous(previous)
		}
		if b7 != nil {
			v := b7.Average
			t.Baseline7D = &v
			if cur != nil {
				delta := roundTimeline(*cur - v)
				t.DeltaVs7D = &delta
			}
		}
		if b28 != nil {
			v := b28.Average
			t.Baseline28D = &v
			if cur != nil {
				delta := roundTimeline(*cur - v)
				t.DeltaVs28D = &delta
			}
		}
		out = append(out, t)
	}
	return out
}

func factorDeltaMessage(label string, delta float64) string {
	switch {
	case delta >= .5:
		return label + " добавил " + formatPoints(delta) + " п. к readiness относительно вчера."
	case delta <= -.5:
		return label + " убрал " + formatPoints(-delta) + " п. из readiness относительно вчера."
	default:
		return label + " почти не изменил readiness относительно вчера."
	}
}

func roundTimeline(v float64) float64 { return math.Round(v*10) / 10 }
func formatPoints(v float64) string   { return fmtFloat1(v) }
func formatInt(v int) string          { return fmt.Sprintf("%d", v) }
func formatSignedPercent(v int) string {
	if v > 0 {
		return fmt.Sprintf("+%d%%", v)
	}
	return fmt.Sprintf("%d%%", v)
}
func formatSignedMinutes(v int) string {
	if v > 0 {
		return fmt.Sprintf("+%d мин", v)
	}
	return fmt.Sprintf("%d мин", v)
}
func fmtFloat1(v float64) string { return strconv.FormatFloat(v, 'f', 1, 64) }
