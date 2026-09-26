package httpapi

import (
	"errors"
	"net/http"

	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func (s *Server) deleteAccount(w http.ResponseWriter, r *http.Request) {
	userID := currentUserID(r.Context())
	err := s.store.DeleteAccount(r.Context(), userID, s.bodyScanService.DeleteUserMedia)
	if errors.Is(err, store.ErrMediaPending) {
		writeJSON(w, http.StatusAccepted, map[string]string{"status": "media_cleanup_pending"})
		return
	}
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "account no longer exists")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "account deletion failed; retry later")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
