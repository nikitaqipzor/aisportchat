package recovery

import (
	"context"
	"testing"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func TestTimelineExplainsDeterministicDeltaAndAdaptation(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	uid := seedUser(t, st)
	svc := NewService(st)
	_, err := svc.SaveCheckIn(ctx, uid, CheckInInput{Date: "2026-09-01", SleepHours: 8, SleepQuality: 5, Energy: 5, Stress: 1, MuscleSoreness: map[string]int{"quads": 1}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.SaveCheckIn(ctx, uid, CheckInInput{Date: "2026-09-02", SleepHours: 5, SleepQuality: 2, Energy: 2, Stress: 4, MuscleSoreness: map[string]int{"quads": 4}})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Timeline(ctx, uid, "2026-09-02")
	if err != nil {
		t.Fatal(err)
	}
	if !out.ComparisonAvailable || out.PreviousScore == nil || out.ScoreDelta == nil || *out.ScoreDelta >= 0 {
		t.Fatalf("expected lower comparable score: %+v", out)
	}
	if out.Adaptation.VolumeMultiplier >= 1 || out.Adaptation.VolumeChangePercent >= 0 {
		t.Fatalf("expected reduced workout: %+v", out.Adaptation)
	}
	foundDown := false
	for _, f := range out.Factors {
		if f.WeightedDeltaPoints != nil && *f.WeightedDeltaPoints < -0.5 {
			foundDown = true
		}
	}
	if !foundDown {
		t.Fatalf("expected at least one negative factor contribution: %+v", out.Factors)
	}
}

func TestTimelineWithoutPreviousCheckInDoesNotInventComparison(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	uid := seedUser(t, st)
	svc := NewService(st)
	_, err := svc.SaveCheckIn(ctx, uid, CheckInInput{Date: "2026-09-02", SleepHours: 8, SleepQuality: 4, Energy: 4, Stress: 2, MuscleSoreness: map[string]int{}})
	if err != nil {
		t.Fatal(err)
	}
	out, err := svc.Timeline(ctx, uid, "2026-09-02")
	if err != nil {
		t.Fatal(err)
	}
	if out.ComparisonAvailable || out.PreviousScore != nil || out.ScoreDelta != nil {
		t.Fatalf("timeline invented previous comparison: %+v", out)
	}
}
