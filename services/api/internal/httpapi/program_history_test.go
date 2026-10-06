package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/auth"
	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func TestProgramHistoryPagesBeyondFifty(t *testing.T) {
	testProgramHistoryPagesBeyondFifty(t, store.NewMemory())
}

func testProgramHistoryPagesBeyondFifty(t *testing.T, st store.Store) {
	ctx := context.Background()
	suffix := time.Now().UnixNano()
	owner, err := st.CreateUser(ctx, fmt.Sprintf("history-owner-%d@example.com", suffix), "hash")
	if err != nil { t.Fatal(err) }
	other, err := st.CreateUser(ctx, fmt.Sprintf("history-other-%d@example.com", suffix), "hash")
	if err != nil { t.Fatal(err) }
	tm := auth.NewTokenManager("test-secret", 15*time.Minute, 24*time.Hour)
	h := NewServerWithDependencies(st, tm)
	access, _, err := tm.NewAccessToken(owner.ID)
	if err != nil { t.Fatal(err) }
	for i := 0; i < 121; i++ {
		_, err = st.CreateProgram(ctx, store.Program{UserID: owner.ID, Title: fmt.Sprintf("Program %d", i), GoalType: "strength", Weeks: 4, WorkoutsPerWeek: 3, Environment: "gym", Status: "archived", StartDate: time.Now().UTC()}, nil)
		if err != nil { t.Fatal(err) }
	}
	_, err = st.CreateProgram(ctx, store.Program{UserID: other.ID, Title: "private", GoalType: "strength", Weeks: 4, WorkoutsPerWeek: 3, Environment: "gym", Status: "archived", StartDate: time.Now().UTC()}, nil)
	if err != nil { t.Fatal(err) }
	seen := map[string]bool{}
	cursor := ""
	for page := 0; page < 3; page++ {
		path := "/api/v1/programs/history?limit=50"
		if cursor != "" { path += "&cursor=" + url.QueryEscape(cursor) }
		res := doJSON(t, h, http.MethodGet, path, nil, access)
		if res.Code != http.StatusOK { t.Fatalf("page %d: %d %s", page, res.Code, res.Body.String()) }
		var body struct {
			Items []store.ProgramWithSessions `json:"items"`
			HasMore bool `json:"has_more"`
			NextCursor string `json:"next_cursor"`
		}
		if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil { t.Fatal(err) }
		want := 50
		if page == 2 { want = 21 }
		if len(body.Items) != want || body.HasMore != (page < 2) { t.Fatalf("page %d: count=%d has_more=%v", page, len(body.Items), body.HasMore) }
		for _, item := range body.Items {
			if item.Program.Title == "private" || item.Program.Title == "inserted after first page" || seen[item.Program.ID] { t.Fatalf("private/new/duplicate program: %s", item.Program.ID) }
			seen[item.Program.ID] = true
		}
		cursor = body.NextCursor
		if page == 0 {
			if _, err := st.CreateProgram(ctx, store.Program{UserID: owner.ID, Title: "inserted after first page", GoalType: "strength", Weeks: 4, WorkoutsPerWeek: 3, Environment: "gym", Status: "archived", StartDate: time.Now().UTC()}, nil); err != nil { t.Fatal(err) }
		}
	}
	if len(seen) != 121 || cursor != "" { t.Fatalf("seen=%d final cursor=%q", len(seen), cursor) }
	for _, path := range []string{"?limit=0", "?limit=51", "?limit=garbage", "?cursor=not-base64", "?cursor=e30"} {
		res := doJSON(t, h, http.MethodGet, "/api/v1/programs/history"+path, nil, access)
		if res.Code != http.StatusBadRequest { t.Fatalf("%s: %d", path, res.Code) }
	}
}
