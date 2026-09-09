package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/example/ai-fitness-os/services/api/internal/aifitness"
	"github.com/example/ai-fitness-os/services/api/internal/auth"
	"github.com/example/ai-fitness-os/services/api/internal/httpapi"
	"github.com/example/ai-fitness-os/services/api/internal/media"
	"github.com/example/ai-fitness-os/services/api/internal/store"
)

func main() {
	port := env("API_PORT", "8080")
	appEnv := env("APP_ENV", "development")
	tokenSecret := env("AUTH_TOKEN_SECRET", "dev-only-change-me")
	if err := validateRuntimeSecurity(appEnv, tokenSecret); err != nil {
		log.Fatalf("security configuration invalid: %v", err)
	}
	if !isProduction(appEnv) && isDevelopmentTokenSecret(tokenSecret) {
		log.Printf("warning: using development AUTH_TOKEN_SECRET")
	}

	st, cleanup := buildStore(appEnv)
	defer cleanup()

	tm := auth.NewTokenManager(tokenSecret, 15*time.Minute, 30*24*time.Hour)
	aiProvider := buildAIProvider()
	mediaStore, err := media.NewFileStore(env("MEDIA_ROOT", "./data/media"))
	if err != nil {
		log.Fatalf("media store init failed: %v", err)
	}
	server := &http.Server{
		Addr:              ":" + port,
		Handler:           httpapi.NewServerWithAIAndMedia(st, tm, aiProvider, mediaStore),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("AI Fitness API listening on %s store=%s ai_provider=%s model=%s", server.Addr, env("STORE_BACKEND", defaultStoreBackend()), aiProvider.Name(), aiProvider.Model())
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func buildStore(appEnv string) (store.Store, func()) {
	backend := env("STORE_BACKEND", defaultStoreBackend())
	if err := validateStoreBackend(appEnv, backend); err != nil {
		log.Fatalf("storage configuration invalid: %v", err)
	}
	if backend == "postgres" {
		pgStore, cleanup, err := store.OpenPostgresStore(os.Getenv("DATABASE_URL"))
		if err != nil {
			log.Fatalf("postgres store init failed: %v", err)
		}
		return pgStore, cleanup
	}
	if backend != "memory" {
		log.Fatalf("unsupported STORE_BACKEND=%q", backend)
	}
	log.Printf("warning: using in-memory store; data will be lost on restart")
	return store.NewMemory(), func() {}
}

func defaultStoreBackend() string {
	if os.Getenv("DATABASE_URL") != "" {
		return "postgres"
	}
	return "memory"
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func buildAIProvider() aifitness.Provider {
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		return aifitness.NewOpenAIProvider(key, env("OPENAI_MODEL", "gpt-5.6-terra"))
	}
	log.Printf("warning: OPENAI_API_KEY not set; AI features use deterministic local provider")
	return aifitness.NewLocalProvider()
}

func validateRuntimeSecurity(appEnv, tokenSecret string) error {
	if !isProduction(appEnv) {
		return nil
	}
	if isDevelopmentTokenSecret(tokenSecret) {
		return errors.New("production requires a non-development AUTH_TOKEN_SECRET")
	}
	if len(tokenSecret) < 32 {
		return fmt.Errorf("production AUTH_TOKEN_SECRET must be at least 32 bytes, got %d", len(tokenSecret))
	}
	return nil
}

func isDevelopmentTokenSecret(secret string) bool {
	switch secret {
	case "", "dev-only-change-me", "change-me", "change-me-with-at-least-32-random-bytes", "local-dev-secret-change-before-production-123456":
		return true
	default:
		return false
	}
}

func validateStoreBackend(appEnv, backend string) error {
	if isProduction(appEnv) && strings.ToLower(strings.TrimSpace(backend)) != "postgres" {
		return errors.New("production requires STORE_BACKEND=postgres")
	}
	return nil
}

func isProduction(appEnv string) bool {
	return strings.EqualFold(strings.TrimSpace(appEnv), "production")
}
