package httpapi

import "net/http"

// Header-less requests retain their original endpoint semantics.
func workoutOperationID(w http.ResponseWriter, r *http.Request) (string, bool) {
	values := r.Header.Values("Idempotency-Key")
	if len(values) == 0 { return "", true }
	if len(values) != 1 || len(values[0]) < 8 || len(values[0]) > 128 {
		writeError(w, http.StatusUnprocessableEntity, "invalid Idempotency-Key")
		return "", false
	}
	for _, c := range values[0] {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == '.' || c == ':' { continue }
		writeError(w, http.StatusUnprocessableEntity, "invalid Idempotency-Key")
		return "", false
	}
	return values[0], true
}
