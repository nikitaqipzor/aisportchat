package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTrustedProxyClientIP(t *testing.T) {
	var got string
	wrapped, err := TrustedProxyClientIPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = clientIP(r)
	}), "172.30.239.10/32")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name, remote, forwarded, want string
	}{
		{"trusted Caddy", "172.30.239.10:3456", "203.0.113.1", "203.0.113.1"},
		{"other client same proxy", "172.30.239.10:3456", "203.0.113.2", "203.0.113.2"},
		{"direct spoof ignored", "198.51.100.4:4567", "203.0.113.1", "198.51.100.4"},
		{"last hop from Caddy", "172.30.239.10:3456", "192.0.2.9, 203.0.113.3", "203.0.113.3"},
		{"malformed header falls back", "172.30.239.10:3456", "bad", "172.30.239.10"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/healthz", nil)
			r.RemoteAddr = tc.remote
			r.Header.Set("X-Forwarded-For", tc.forwarded)
			wrapped.ServeHTTP(httptest.NewRecorder(), r)
			if got != tc.want {
				t.Fatalf("client IP = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAuthLimiterSeparatesClientsBehindTrustedProxy(t *testing.T) {
	s := &Server{authRateLimiter: newAuthRateLimiter()}
	h, err := TrustedProxyClientIPHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.enforceAuthRateLimit(w, r, "login", "", 30, 0, 5*time.Minute) {
			w.WriteHeader(http.StatusNoContent)
		}
	}), "172.30.239.10/32")
	if err != nil {
		t.Fatal(err)
	}
	request := func(ip string) int {
		r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
		r.RemoteAddr = "172.30.239.10:3456"
		r.Header.Set("X-Forwarded-For", ip)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code
	}
	for i := 0; i < 30; i++ {
		if code := request("203.0.113.1"); code != http.StatusNoContent {
			t.Fatalf("first client attempt %d: %d", i, code)
		}
	}
	if code := request("203.0.113.1"); code != http.StatusTooManyRequests {
		t.Fatalf("first client exceeded limit: %d", code)
	}
	if code := request("203.0.113.2"); code != http.StatusNoContent {
		t.Fatalf("independent second client blocked: %d", code)
	}
}

func TestTrustedProxyConfigurationRejectsInvalidCIDR(t *testing.T) {
	if _, err := TrustedProxyClientIPHandler(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), "not-an-ip"); err == nil {
		t.Fatal("invalid trusted proxy CIDR accepted")
	}
}
