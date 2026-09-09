package httpapi

import (
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

type rateLimitEntry struct {
	count int
	reset time.Time
}

type authRateLimiter struct {
	mu      sync.Mutex
	entries map[string]rateLimitEntry
	now     func() time.Time
}

func newAuthRateLimiter() *authRateLimiter {
	return &authRateLimiter{entries: make(map[string]rateLimitEntry), now: time.Now}
}

func (l *authRateLimiter) allow(key string, limit int, window time.Duration) (bool, time.Duration) {
	if limit <= 0 || strings.TrimSpace(key) == "" {
		return true, 0
	}
	now := l.now()
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(l.entries) > 10000 {
		for candidate, candidateEntry := range l.entries {
			if !now.Before(candidateEntry.reset) {
				delete(l.entries, candidate)
			}
		}
	}
	entry, ok := l.entries[key]
	if !ok || !now.Before(entry.reset) {
		l.entries[key] = rateLimitEntry{count: 1, reset: now.Add(window)}
		return true, 0
	}
	if entry.count >= limit {
		return false, entry.reset.Sub(now).Round(time.Second)
	}
	entry.count++
	l.entries[key] = entry
	return true, 0
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

func normalizedIdentity(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func tokenFingerprint(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:8])
}

func (s *Server) enforceAuthRateLimit(w http.ResponseWriter, r *http.Request, operation, identity string, ipLimit, identityLimit int, window time.Duration) bool {
	ip := clientIP(r)
	checks := []struct {
		key   string
		limit int
	}{
		{key: "auth:" + operation + ":ip:" + ip, limit: ipLimit},
	}
	if identity != "" {
		checks = append(checks, struct {
			key   string
			limit int
		}{key: "auth:" + operation + ":identity:" + identity, limit: identityLimit})
	}
	for _, check := range checks {
		ok, retry := s.authRateLimiter.allow(check.key, check.limit, window)
		if ok {
			continue
		}
		seconds := int(retry.Seconds())
		if seconds < 1 {
			seconds = 1
		}
		w.Header().Set("Retry-After", strconv.Itoa(seconds))
		writeError(w, http.StatusTooManyRequests, "too many authentication attempts; try again later")
		return false
	}
	return true
}
