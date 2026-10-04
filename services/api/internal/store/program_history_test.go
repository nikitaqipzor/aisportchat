package store

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestProgramHistoryKeysetAtTiedTimestamp(t *testing.T) {
	m := NewMemory()
	stamp := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 121; i++ {
		id := fmt.Sprintf("00000000-0000-4000-8000-%012x", i)
		m.programs[id] = Program{ID: id, UserID: "owner", UpdatedAt: stamp}
	}
	m.programs["other"] = Program{ID: "other", UserID: "other", UpdatedAt: stamp.Add(time.Hour)}
	seen := map[string]bool{}
	var before *ProgramHistoryCursor
	for page := 0; page < 3; page++ {
		items, err := m.ListPrograms(context.Background(), "owner", 50, before)
		if err != nil { t.Fatal(err) }
		want := 50
		if page == 2 { want = 21 }
		if len(items) != want { t.Fatalf("page %d length %d, want %d", page, len(items), want) }
		for _, item := range items {
			id := item.Program.ID
			if seen[id] { t.Fatalf("duplicate program %s", id) }
			seen[id] = true
		}
		last := items[len(items)-1].Program
		before = &ProgramHistoryCursor{UpdatedAt: last.UpdatedAt, ID: last.ID}
	}
	if len(seen) != 121 { t.Fatalf("saw %d of 121 programs", len(seen)) }
	// Newer inserts do not change which older programs follow an existing cursor.
	m.programs["new"] = Program{ID: "new", UserID: "owner", UpdatedAt: stamp.Add(time.Hour)}
	items, err := m.ListPrograms(context.Background(), "owner", 50, before)
	if err != nil || len(items) != 0 { t.Fatalf("page after end: %d, %v", len(items), err) }
}
