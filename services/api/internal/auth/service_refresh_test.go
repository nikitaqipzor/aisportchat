package auth

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func TestRefreshTokenConsumedOnceUnderConcurrentRequests(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	user, err := st.CreateUser(ctx, "rotation@example.com", "hash")
	if err != nil { t.Fatal(err) }
	svc := NewService(st, NewTokenManager("rotation-test-secret", time.Minute, time.Hour))
	initial, err := svc.issueTokens(ctx, user.ID)
	if err != nil { t.Fatal(err) }
	const attempts = 32
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make(chan error, attempts)
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := svc.Refresh(ctx, initial.RefreshToken)
			results <- err
		}()
	}
	close(start)
	wg.Wait()
	close(results)
	success := 0
	for err := range results { if err == nil { success++ } }
	if success != 1 { t.Fatalf("one rotation should succeed, got %d", success) }
	if _, err := svc.Refresh(ctx, initial.RefreshToken); err == nil { t.Fatal("consumed token reused") }
}

type brokenConsumeStore struct {
	store.Store
}

func (brokenConsumeStore) ConsumeRefreshSession(context.Context, string) (store.RefreshSession, error) {
	return store.RefreshSession{}, context.DeadlineExceeded
}

func TestRefreshDoesNotIssueTokensWhenConsumptionFails(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	user, err := st.CreateUser(ctx, "failed-rotation@example.com", "hash")
	if err != nil { t.Fatal(err) }
	svc := NewService(st, NewTokenManager("rotation-test-secret", time.Minute, time.Hour))
	initial, err := svc.issueTokens(ctx, user.ID)
	if err != nil { t.Fatal(err) }
	svc.store = brokenConsumeStore{st}
	if tokens, err := svc.Refresh(ctx, initial.RefreshToken); err == nil || tokens.AccessToken != "" || tokens.RefreshToken != "" {
		t.Fatalf("issued tokens after failed consume: %+v, %v", tokens, err)
	}
}
