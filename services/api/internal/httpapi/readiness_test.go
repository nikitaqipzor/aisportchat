package httpapi

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/auth"
	"github.com/example/ai-fitness-os/services/api/internal/store"
)

type readinessStore struct {
	store.Store
	err error
	called bool
}

func (s *readinessStore) Ping(ctx context.Context) error {
	s.called = true
	return s.err
}

func TestReadyzChecksStoreWithoutExposingErrors(t *testing.T) {
	st := &readinessStore{Store: store.NewMemory()}
	h := NewServerWithDependencies(st, auth.NewTokenManager("test-secret", time.Minute, time.Hour))
	for _, tc := range []struct {
		err error
		want int
		body string
	}{
		{nil, http.StatusOK, `"status":"ok"`},
		{errors.New("database secret=hidden"), http.StatusServiceUnavailable, `"status":"unavailable"`},
	} {
		st.err, st.called = tc.err, false
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/readyz", nil))
		if !st.called || rr.Code != tc.want || !strings.Contains(rr.Body.String(), tc.body) || strings.Contains(rr.Body.String(), "secret") {
			t.Fatalf("called=%v status=%d body=%s", st.called, rr.Code, rr.Body.String())
		}
	}
}
