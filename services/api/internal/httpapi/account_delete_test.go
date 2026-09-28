package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/aifitness"
	"github.com/example/ai-fitness-os/services/api/internal/auth"
	"github.com/example/ai-fitness-os/services/api/internal/media"
	"github.com/example/ai-fitness-os/services/api/internal/store"
)

type failingAccountMedia struct{ media.Store }
func (f failingAccountMedia) DeleteUserMedia(context.Context, string) error { return errors.New("disk unavailable") }

func TestAccountDeletionRemovesPrivateDataAndRevokesTokens(t *testing.T) {
	ctx := context.Background()
	st, blobs := store.NewMemory(), media.NewMemoryStore()
	h := NewServerWithAIAndMedia(st, auth.NewTokenManager("delete-test-secret", time.Hour, time.Hour), aifitness.NewLocalProvider(), blobs)
	register := func(email string) (store.User, auth.Tokens) {
		t.Helper()
		resp := doJSON(t, h, http.MethodPost, "/api/v1/auth/register", map[string]any{"email": email, "password": "strong-pass-123"}, "")
		if resp.Code != http.StatusCreated { t.Fatalf("register=%d %s", resp.Code, resp.Body.String()) }
		var out struct { User store.User `json:"user"`; Tokens auth.Tokens `json:"tokens"` }
		if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil { t.Fatal(err) }
		return out.User, out.Tokens
	}
	u1, a1 := register("delete-owner@example.com")
	u2, a2 := register("delete-other@example.com")
	for _, u := range []store.User{u1,u2} {
		key := "body-scans/"+u.ID+"/orphan/photo.jpg"
		if err := blobs.Put(ctx, key, "image/jpeg", []byte("private")); err != nil { t.Fatal(err) }
	}
	_, err := st.UpsertProfile(ctx, store.Profile{UserID: u1.ID, Injuries: []string{"shoulder"}})
	if err != nil { t.Fatal(err) }
	if rr := doJSON(t, h, http.MethodDelete, "/api/v1/auth/account", nil, a1.AccessToken); rr.Code != http.StatusNoContent { t.Fatalf("delete=%d %s", rr.Code, rr.Body.String()) }
	if _, err := st.FindUserByID(ctx, u1.ID); !errors.Is(err, store.ErrNotFound) { t.Fatalf("deleted account exists: %v", err) }
	if _, err := st.GetRefreshSession(ctx, auth.HashRefreshToken(a1.RefreshToken)); !errors.Is(err, store.ErrNotFound) { t.Fatalf("refresh persists: %v", err) }
	if _, err := blobs.Get(ctx, "body-scans/"+u1.ID+"/orphan/photo.jpg"); !errors.Is(err, media.ErrNotFound) { t.Fatalf("photo persists: %v", err) }
	if _, err := blobs.Get(ctx, "body-scans/"+u2.ID+"/orphan/photo.jpg"); err != nil { t.Fatalf("other user's photo removed: %v", err) }
	if rr := doJSON(t, h, http.MethodGet, "/api/v1/profile", nil, a1.AccessToken); rr.Code != http.StatusUnauthorized { t.Fatalf("old access token=%d", rr.Code) }
	if rr := doJSON(t, h, http.MethodPost, "/api/v1/auth/refresh", map[string]any{"refresh_token":a1.RefreshToken}, ""); rr.Code != http.StatusUnauthorized { t.Fatalf("old refresh token=%d", rr.Code) }
	if rr := doJSON(t, h, http.MethodGet, "/api/v1/profile", nil, a2.AccessToken); rr.Code != http.StatusOK { t.Fatalf("other user=%d", rr.Code) }
}

func TestAccountDeletionMediaFailureQueuesRetry(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemory()
	blobs := media.NewMemoryStore()
	h := NewServerWithAIAndMedia(st, auth.NewTokenManager("delete-failure-secret", time.Hour, time.Hour), aifitness.NewLocalProvider(), failingAccountMedia{blobs})
	resp := doJSON(t, h, http.MethodPost, "/api/v1/auth/register", map[string]any{"email":"delete-failure@example.com", "password":"strong-pass-123"}, "")
	var out struct { User store.User `json:"user"`; Tokens auth.Tokens `json:"tokens"` }
	if err := json.Unmarshal(resp.Body.Bytes(), &out); err != nil { t.Fatal(err) }
	if rr := doJSON(t, h, http.MethodDelete, "/api/v1/auth/account", nil, out.Tokens.AccessToken); rr.Code != http.StatusAccepted { t.Fatalf("delete queued=%d body=%s", rr.Code, rr.Body.String()) }
	if _, err := st.FindUserByID(ctx, out.User.ID); !errors.Is(err, store.ErrNotFound) { t.Fatalf("account remained after queued deletion: %v", err) }
	if _, err := st.GetRefreshSession(ctx, auth.HashRefreshToken(out.Tokens.RefreshToken)); !errors.Is(err, store.ErrNotFound) { t.Fatalf("session remained after queued deletion: %v", err) }
	key := "body-scans/"+out.User.ID+"/orphan/photo.jpg"
	if err := blobs.Put(ctx, key, "image/jpeg", []byte("private")); err != nil { t.Fatal(err) }
	if err := st.RetryPendingMedia(ctx, blobs.DeleteUserMedia); err != nil { t.Fatalf("retry: %v", err) }
	if _, err := blobs.Get(ctx, key); !errors.Is(err, media.ErrNotFound) { t.Fatalf("retry left media: %v", err) }
}
