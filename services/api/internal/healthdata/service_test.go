package healthdata

import (
	"context"
	"testing"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func TestImportSnapshotUpsertsAndIsUserScoped(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	u1, _ := st.CreateUser(ctx, "hc1@example.com", "x")
	u2, _ := st.CreateUser(ctx, "hc2@example.com", "x")
	s := NewService(st)
	in := SnapshotInput{Date: "2026-09-02", Provider: "health_connect", SourcePackage: "com.xiaomi.wearable", SourceLabel: "Mi Fitness / Xiaomi Watch S3", Steps: 10123, DistanceM: 7345, ActiveCaloriesKcal: 611, SleepMinutes: 438, DeepSleepMinutes: 82, LightSleepMinutes: 251, REMSleepMinutes: 105, ExerciseMinutes: 54, ExerciseSessions: 1, DataTypes: []string{"sleep", "steps", "exercise"}}
	got, err := s.Import(ctx, u1.ID, in)
	if err != nil {
		t.Fatal(err)
	}
	if got.Steps != 10123 || got.SleepMinutes != 438 {
		t.Fatalf("unexpected snapshot: %+v", got)
	}
	in.Steps = 11000
	got, err = s.Import(ctx, u1.ID, in)
	if err != nil {
		t.Fatal(err)
	}
	if got.Steps != 11000 {
		t.Fatalf("upsert failed")
	}
	if _, err := s.Get(ctx, u2.ID, "2026-09-02"); err != store.ErrNotFound {
		t.Fatalf("cross-user snapshot leaked: %v", err)
	}
}

func TestImportRejectsImpossibleSleepStages(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	u, _ := st.CreateUser(ctx, "hc3@example.com", "x")
	s := NewService(st)
	_, err := s.Import(ctx, u.ID, SnapshotInput{Date: "2026-09-02", SourcePackage: "com.xiaomi.wearable", SleepMinutes: 300, DeepSleepMinutes: 200, LightSleepMinutes: 200, DataTypes: []string{"sleep"}})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestInsightsUsesPreviousDaysAndTracksFreshness(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	u, _ := st.CreateUser(ctx, "baseline@example.com", "x")
	svc := NewService(st)
	baseCaptured := time.Now().UTC().Add(-2 * time.Hour)
	for i := 1; i <= 10; i++ {
		date := time.Date(2026, 9, i, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		_, err := st.UpsertHealthDailySnapshot(ctx, store.HealthDailySnapshot{
			UserID: u.ID, LocalDate: date, Provider: "health_connect", SourcePackage: "com.xiaomi.wearable", SourceLabel: "Mi Fitness / Xiaomi Watch S3",
			Steps: int64(8000 + i*100), SleepMinutes: 420 + i, ExerciseMinutes: 30,
			DataTypes: []string{"steps", "sleep", "exercise"}, CapturedAt: baseCaptured,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	out, err := svc.Insights(ctx, u.ID, "2026-09-10")
	if err != nil {
		t.Fatal(err)
	}
	if out.Baseline7D.AvailableDays != 7 {
		t.Fatalf("expected 7 prior days, got %d", out.Baseline7D.AvailableDays)
	}
	if out.Baseline28D.AvailableDays != 9 {
		t.Fatalf("current day must be excluded, got %d", out.Baseline28D.AvailableDays)
	}
	if out.Baseline28D.SleepMinutes == nil || out.Baseline28D.SleepMinutes.SampleDays != 9 {
		t.Fatalf("missing sleep baseline: %+v", out.Baseline28D.SleepMinutes)
	}
	if _, ok := out.Deviations["sleep_minutes"]; !ok {
		t.Fatalf("sleep deviation missing: %+v", out.Deviations)
	}
	if out.Confidence == "low" {
		t.Fatalf("expected usable confidence, got %+v", out)
	}
}

func TestBuildInsightsMarksStaleSnapshot(t *testing.T) {
	now := time.Now().UTC()
	current := &store.HealthDailySnapshot{LocalDate: now.Format("2006-01-02"), SourcePackage: "com.xiaomi.wearable", SourceLabel: "Mi Fitness", SleepMinutes: 450, DataTypes: []string{"sleep"}, CapturedAt: now.Add(-48 * time.Hour), ImportedAt: now.Add(-48 * time.Hour)}
	out := BuildInsights(current.LocalDate, current, nil, now)
	if out.Freshness.Status != "stale" {
		t.Fatalf("expected stale, got %+v", out.Freshness)
	}
}

func TestMultiSourceResolverPrefersXiaomiAndKeepsBothSources(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	u, _ := st.CreateUser(ctx, "sources@example.com", "x")
	svc := NewService(st)
	date := "2026-09-02"
	_, err := svc.Import(ctx, u.ID, SnapshotInput{Date: date, Provider: "health_connect", SourcePackage: "com.xiaomi.wearable", SourceLabel: "Mi Fitness", Steps: 10000, SleepMinutes: 450, DataTypes: []string{"steps", "sleep"}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = svc.Import(ctx, u.ID, SnapshotInput{Date: date, Provider: "health_connect", SourcePackage: "com.example.phone", SourceLabel: "Phone Health", Steps: 13000, SleepMinutes: 0, DataTypes: []string{"steps"}})
	if err != nil {
		t.Fatal(err)
	}
	resolved, err := svc.Get(ctx, u.ID, date)
	if err != nil {
		t.Fatal(err)
	}
	if resolved.SourcePackage != "com.xiaomi.wearable" || resolved.Steps != 10000 {
		t.Fatalf("resolver chose wrong source: %+v", resolved)
	}
	insights, err := svc.Insights(ctx, u.ID, date)
	if err != nil {
		t.Fatal(err)
	}
	if !insights.ConflictResolved || len(insights.Sources) != 2 {
		t.Fatalf("source conflict not exposed: %+v", insights)
	}
	selected := 0
	for _, source := range insights.Sources {
		if source.Selected {
			selected++
		}
	}
	if selected != 1 {
		t.Fatalf("expected one selected source, got %d: %+v", selected, insights.Sources)
	}
}

func TestEmptyHealthDaysDoNotInflateCoverage(t *testing.T) {
	now := time.Now().UTC()
	history := []store.HealthDailySnapshot{
		{LocalDate: "2026-09-01", DataTypes: []string{}, ImportedAt: now},
		{LocalDate: "2026-08-31", SleepMinutes: 450, DataTypes: []string{"sleep"}, ImportedAt: now},
	}
	out := BuildInsights("2026-09-02", nil, history, now)
	if out.Baseline7D.AvailableDays != 1 {
		t.Fatalf("empty day inflated coverage: %+v", out.Baseline7D)
	}
}

func TestImportRejectsSnapshotWithoutAnyDataType(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	u, _ := st.CreateUser(ctx, "empty-health@example.com", "x")
	_, err := NewService(st).Import(ctx, u.ID, SnapshotInput{Date: "2026-09-02", Provider: "health_connect", SourcePackage: "com.xiaomi.wearable"})
	if err == nil {
		t.Fatal("expected empty health snapshot rejection")
	}
}
