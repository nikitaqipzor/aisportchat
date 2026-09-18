package main

import (
	"os"
	"testing"
)

func TestResolvePort(t *testing.T) {
	oldPort, hadPort := os.LookupEnv("PORT")
	oldAPIPort, hadAPIPort := os.LookupEnv("API_PORT")
	t.Cleanup(func() {
		if hadPort {
			_ = os.Setenv("PORT", oldPort)
		} else {
			_ = os.Unsetenv("PORT")
		}
		if hadAPIPort {
			_ = os.Setenv("API_PORT", oldAPIPort)
		} else {
			_ = os.Unsetenv("API_PORT")
		}
	})

	_ = os.Unsetenv("PORT")
	_ = os.Unsetenv("API_PORT")
	if got := resolvePort(); got != "8080" {
		t.Fatalf("default port = %q, want 8080", got)
	}
	_ = os.Setenv("API_PORT", "9090")
	if got := resolvePort(); got != "9090" {
		t.Fatalf("API_PORT fallback = %q, want 9090", got)
	}
	_ = os.Setenv("PORT", "4242")
	if got := resolvePort(); got != "4242" {
		t.Fatalf("Railway PORT = %q, want 4242", got)
	}
}

func TestValidateRuntimeSecurity(t *testing.T) {
	tests := []struct {
		name   string
		env    string
		secret string
		ok     bool
	}{
		{name: "dev allows local secret", env: "development", secret: "dev-only-change-me", ok: true},
		{name: "production rejects default", env: "production", secret: "dev-only-change-me", ok: false},
		{name: "production rejects example", env: "production", secret: "change-me-with-at-least-32-random-bytes", ok: false},
		{name: "production rejects short", env: "production", secret: "short-secret", ok: false},
		{name: "production accepts strong", env: "production", secret: "A-strong-random-production-secret-0123456789", ok: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRuntimeSecurity(tt.env, tt.secret)
			if tt.ok && err != nil {
				t.Fatalf("expected success: %v", err)
			}
			if !tt.ok && err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}

func TestValidateStoreBackend(t *testing.T) {
	if err := validateStoreBackend("production", "memory"); err == nil {
		t.Fatal("production memory store must be rejected")
	}
	if err := validateStoreBackend(" Production ", "postgres"); err != nil {
		t.Fatalf("production postgres should be accepted: %v", err)
	}
	if err := validateStoreBackend("development", "memory"); err != nil {
		t.Fatalf("development memory should be accepted: %v", err)
	}
}
