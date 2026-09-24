package nutrition

import (
	"context"
	"testing"
	"time"
)

func mustInstant(t *testing.T, value string) time.Time {
	t.Helper()
	when, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatal(err)
	}
	return when
}

func TestNutritionLocalCalendarDayAcrossUTCDate(t *testing.T) {
	for _, tc := range []struct {
		name, zone, instant, localDate string
	}{
		{"east-of-UTC", "Europe/Moscow", "2026-09-01T21:30:00Z", "2026-09-02"},
		{"west-of-UTC", "America/Bogota", "2026-09-02T03:30:00Z", "2026-09-01"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			loc, err := time.LoadLocation(tc.zone)
			if err != nil {
				t.Fatal(err)
			}
			weight := 80.0
			svc, _, userID := nutritionFixture(t, &weight)
			setAutoNutrition(t, svc, userID)
			ctx := context.Background()
			when := mustInstant(t, tc.instant)
			logged, err := svc.LogFoodInLocation(ctx, userID, "banana", "snack", 100, &when, loc)
			if err != nil || logged.Date != tc.localDate || len(logged.Entries) != 1 {
				t.Fatalf("logged=%+v err=%v", logged, err)
			}
			day, err := svc.DayInLocation(ctx, userID, when, loc)
			if err != nil || day.Date != tc.localDate || len(day.Entries) != 1 {
				t.Fatalf("day=%+v err=%v", day, err)
			}
			history, err := svc.HistoryInLocation(ctx, userID, 1, when, loc)
			if err != nil || len(history) != 1 || history[0].Date != tc.localDate || history[0].Calories <= 0 {
				t.Fatalf("history=%+v err=%v", history, err)
			}
			utcDay, err := svc.Day(ctx, userID, when)
			if err != nil || utcDay.Date != when.Format("2006-01-02") {
				t.Fatalf("UTC default day=%+v err=%v", utcDay, err)
			}
		})
	}
}

func TestNutritionLocalDayHandlesDSTBoundary(t *testing.T) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, date, first, last, next string
		dayHours int
	}{
		{"spring-forward", "2026-03-08", "2026-03-08T05:00:00Z", "2026-03-09T03:59:59Z", "2026-03-09T04:00:00Z", 23},
		{"fall-back", "2026-11-01", "2026-11-01T04:00:00Z", "2026-11-02T04:59:59Z", "2026-11-02T05:00:00Z", 25},
	} {
		t.Run(tc.name, func(t *testing.T) {
			weight := 80.0
			svc, _, userID := nutritionFixture(t, &weight)
			setAutoNutrition(t, svc, userID)
			ctx := context.Background()
			for _, instant := range []string{tc.first, tc.last, tc.next} {
				when := mustInstant(t, instant)
				if _, err := svc.LogFoodInLocation(ctx, userID, "banana", "snack", 100, &when, loc); err != nil {
					t.Fatal(err)
				}
			}
			start := mustInstant(t, tc.first)
			end := mustInstant(t, tc.next)
			if got := end.Sub(start); got != time.Duration(tc.dayHours)*time.Hour {
				t.Fatalf("calendar day duration=%s", got)
			}
			day, err := svc.DayInLocation(ctx, userID, start, loc)
			if err != nil || day.Date != tc.date || len(day.Entries) != 2 {
				t.Fatalf("local day=%+v err=%v", day, err)
			}
			history, err := svc.HistoryInLocation(ctx, userID, 2, end, loc)
			if err != nil || len(history) != 2 || history[0].Date != tc.date || history[0].Calories <= history[1].Calories {
				t.Fatalf("local history=%+v err=%v", history, err)
			}
		})
	}
}
