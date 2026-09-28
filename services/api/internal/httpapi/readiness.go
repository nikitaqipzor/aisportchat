package httpapi

import (
	"context"
	"net/http"
	"time"
)

// ready checks the backing database when the configured store exposes a ping.
// The in-memory development store has no external dependency to check.
func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	if p, ok := s.store.(interface{ Ping(context.Context) error }); ok {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := p.Ping(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
