package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/auth"
	"github.com/example/ai-fitness-os/services/api/internal/store"
)

type blockedProgramWorkoutStore struct {
	store.Store
	entered chan struct{}
	release chan struct{}
}

func (s *blockedProgramWorkoutStore) CreateProgramSessionWorkout(ctx context.Context, userID, sessionID string, workout store.Workout, exercises []store.WorkoutExercise) (store.WorkoutDetails, error) {
	close(s.entered)
	<-s.release
	return s.Store.CreateProgramSessionWorkout(ctx, userID, sessionID, workout, exercises)
}

func TestArchiveDuringSessionWorkoutGenerationPreventsInsert(t *testing.T) {
	st := &blockedProgramWorkoutStore{Store: store.NewMemory(), entered: make(chan struct{}), release: make(chan struct{})}
	h := NewServerWithDependencies(st, auth.NewTokenManager("test-secret", 15*time.Minute, time.Hour))
	access := registerAndOnboard(t, h, "archive-during-generation@example.com")
	programID, sessionID := createTestProgramSession(t, h, access)
	created := make(chan int, 1)
	go func() {
		created <- doJSON(t, h, http.MethodPost, "/api/v1/programs/sessions/"+sessionID+"/workout", map[string]any{}, access).Code
	}()
	select {
	case <-st.entered:
	case <-time.After(5 * time.Second):
		t.Fatal("workout generation did not reach persistence")
	}
	archived := doJSON(t, h, http.MethodPost, "/api/v1/programs/"+programID+"/archive", map[string]any{}, access)
	if archived.Code != http.StatusOK {
		t.Fatalf("archive=%d body=%s", archived.Code, archived.Body.String())
	}
	close(st.release)
	if code := <-created; code != http.StatusConflict {
		t.Fatalf("workout after concurrent archive status=%d", code)
	}
	history := doJSON(t, h, http.MethodGet, "/api/v1/workouts/history", nil, access)
	if history.Code != http.StatusOK || !containsEmptyHistory(history.Body.Bytes()) {
		t.Fatalf("orphan workout after archive: status=%d body=%s", history.Code, history.Body.String())
	}
}

func containsEmptyHistory(raw []byte) bool {
	var history struct {
		Items []json.RawMessage `json:"items"`
	}
	return json.Unmarshal(raw, &history) == nil && len(history.Items) == 0
}

func TestProgramSessionWorkoutRejectsArchivedProgramBeforeCreatingWorkout(t *testing.T) {
	for _, archiveBy := range []string{"explicit", "replacement"} {
		t.Run(archiveBy, func(t *testing.T) {
			st := store.NewMemory()
			h := NewServerWithDependencies(st, auth.NewTokenManager("test-secret", 15*time.Minute, time.Hour))
			access := registerAndOnboard(t, h, "archived-session-"+archiveBy+"@example.com")
			programID, sessionID := createTestProgramSession(t, h, access)
			if archiveBy == "explicit" {
				response := doJSON(t, h, http.MethodPost, "/api/v1/programs/"+programID+"/archive", map[string]any{}, access)
				if response.Code != http.StatusOK {
					t.Fatalf("archive=%d body=%s", response.Code, response.Body.String())
				}
			} else {
				createTestProgramSession(t, h, access)
			}
			before := doJSON(t, h, http.MethodGet, "/api/v1/workouts/history?limit=50", nil, access)
			if before.Code != http.StatusOK {
				t.Fatalf("history before=%d body=%s", before.Code, before.Body.String())
			}
			response := doJSON(t, h, http.MethodPost, "/api/v1/programs/sessions/"+sessionID+"/workout", map[string]any{}, access)
			if response.Code != http.StatusConflict {
				t.Fatalf("archived workout creation=%d body=%s", response.Code, response.Body.String())
			}
			after := doJSON(t, h, http.MethodGet, "/api/v1/workouts/history?limit=50", nil, access)
			if after.Code != http.StatusOK || after.Body.String() != before.Body.String() {
				t.Fatalf("rejected request changed workout history: before=%s after=%s", before.Body.String(), after.Body.String())
			}
			loaded := doJSON(t, h, http.MethodGet, "/api/v1/programs/"+programID, nil, access)
			var program struct {
				Sessions []store.ProgramSession `json:"sessions"`
			}
			if loaded.Code != http.StatusOK || json.Unmarshal(loaded.Body.Bytes(), &program) != nil || len(program.Sessions) == 0 || program.Sessions[0].WorkoutID != nil {
				t.Fatalf("rejected request linked a workout: status=%d body=%s", loaded.Code, loaded.Body.String())
			}
		})
	}
}

func TestConcurrentProgramSessionWorkoutAndArchiveLeavesNoOrphan(t *testing.T) {
	st := store.NewMemory()
	testConcurrentProgramSessionWorkoutAndArchive(t, st, st)
}

func testConcurrentProgramSessionWorkoutAndArchive(t *testing.T, st, archiveStore store.Store) {
	tokens := auth.NewTokenManager("test-secret", 15*time.Minute, time.Hour)
	h := NewServerWithDependencies(st, tokens)
	archiveHandler := NewServerWithDependencies(archiveStore, tokens)
	access := registerAndOnboard(t, h, "concurrent-session-"+time.Now().UTC().Format("20060102150405.000000000")+"@example.com")
	programID, sessionID := createTestProgramSession(t, h, access)
	start := make(chan struct{})
	var wg sync.WaitGroup
	statuses := make(chan int, 2)
	archiveStatus := make(chan int, 1)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			statuses <- doJSON(t, h, http.MethodPost, "/api/v1/programs/sessions/"+sessionID+"/workout", map[string]any{}, access).Code
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		archiveStatus <- doJSON(t, archiveHandler, http.MethodPost, "/api/v1/programs/"+programID+"/archive", map[string]any{}, access).Code
	}()
	close(start)
	wg.Wait()
	close(statuses)
	if code := <-archiveStatus; code != http.StatusOK {
		t.Fatalf("archive status=%d", code)
	}
	for code := range statuses {
		switch code {
		case http.StatusOK, http.StatusCreated, http.StatusConflict:
		default:
			t.Fatalf("unexpected concurrent response status=%d", code)
		}
	}
	rejected := doJSON(t, h, http.MethodPost, "/api/v1/programs/sessions/"+sessionID+"/workout", map[string]any{}, access)
	if rejected.Code != http.StatusConflict {
		t.Fatalf("post archive status=%d body=%s", rejected.Code, rejected.Body.String())
	}
	loaded := doJSON(t, h, http.MethodGet, "/api/v1/programs/"+programID, nil, access)
	var program struct {
		Sessions []store.ProgramSession `json:"sessions"`
	}
	if loaded.Code != http.StatusOK || json.Unmarshal(loaded.Body.Bytes(), &program) != nil || len(program.Sessions) == 0 {
		t.Fatalf("program=%d body=%s", loaded.Code, loaded.Body.String())
	}
	history := doJSON(t, h, http.MethodGet, "/api/v1/workouts/history?limit=50", nil, access)
	var workouts struct {
		Items []struct {
			Workout store.Workout `json:"workout"`
		} `json:"items"`
	}
	if history.Code != http.StatusOK || json.Unmarshal(history.Body.Bytes(), &workouts) != nil {
		t.Fatalf("history=%d body=%s", history.Code, history.Body.String())
	}
	if len(workouts.Items) > 1 {
		t.Fatalf("duplicate or orphan workouts: %s", history.Body.String())
	}
	if len(workouts.Items) == 0 && program.Sessions[0].WorkoutID != nil || len(workouts.Items) == 1 && (program.Sessions[0].WorkoutID == nil || *program.Sessions[0].WorkoutID != workouts.Items[0].Workout.ID) {
		t.Fatalf("workout/session link mismatch: history=%s program=%s", history.Body.String(), loaded.Body.String())
	}
}

func TestProgramSessionWorkoutRejectsArchivedLinkedWorkoutButNewProgramStillWorks(t *testing.T) {
	st := store.NewMemory()
	h := NewServerWithDependencies(st, auth.NewTokenManager("test-secret", 15*time.Minute, time.Hour))
	access := registerAndOnboard(t, h, "archived-linked-session@example.com")
	_, oldSessionID := createTestProgramSession(t, h, access)
	endpoint := "/api/v1/programs/sessions/" + oldSessionID + "/workout"
	first := doJSON(t, h, http.MethodPost, endpoint, map[string]any{}, access)
	if first.Code != http.StatusCreated {
		t.Fatalf("first=%d body=%s", first.Code, first.Body.String())
	}
	_, newSessionID := createTestProgramSession(t, h, access)
	before := doJSON(t, h, http.MethodGet, "/api/v1/workouts/history?limit=50", nil, access)
	archived := doJSON(t, h, http.MethodPost, endpoint, map[string]any{}, access)
	if archived.Code != http.StatusConflict {
		t.Fatalf("archived linked workout=%d body=%s", archived.Code, archived.Body.String())
	}
	after := doJSON(t, h, http.MethodGet, "/api/v1/workouts/history?limit=50", nil, access)
	if after.Body.String() != before.Body.String() {
		t.Fatalf("rejected request changed workout history: before=%s after=%s", before.Body.String(), after.Body.String())
	}
	newEndpoint := "/api/v1/programs/sessions/" + newSessionID + "/workout"
	created := doJSON(t, h, http.MethodPost, newEndpoint, map[string]any{}, access)
	if created.Code != http.StatusCreated {
		t.Fatalf("active replacement program workout=%d body=%s", created.Code, created.Body.String())
	}
	linked := doJSON(t, h, http.MethodPost, newEndpoint, map[string]any{}, access)
	if linked.Code != http.StatusOK {
		t.Fatalf("active linked workout=%d body=%s", linked.Code, linked.Body.String())
	}
}

func createTestProgramSession(t *testing.T, h http.Handler, access string) (string, string) {
	t.Helper()
	response := doJSON(t, h, http.MethodPost, "/api/v1/programs/generate", map[string]any{"weeks": 4, "workouts_per_week": 3, "environment": "gym"}, access)
	if response.Code != http.StatusCreated {
		t.Fatalf("generate program=%d body=%s", response.Code, response.Body.String())
	}
	var body struct {
		Program  store.Program          `json:"program"`
		Sessions []store.ProgramSession `json:"sessions"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Program.ID == "" || len(body.Sessions) == 0 {
		t.Fatalf("missing program or sessions: %s", response.Body.String())
	}
	return body.Program.ID, body.Sessions[0].ID
}
