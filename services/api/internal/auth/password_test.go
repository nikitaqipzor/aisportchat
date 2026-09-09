package auth

import "testing"

func TestPasswordHashRoundTrip(t *testing.T) {
	hash, err := HashPassword("strong-pass-123")
	if err != nil {
		t.Fatal(err)
	}
	if !VerifyPassword(hash, "strong-pass-123") {
		t.Fatal("expected password to verify")
	}
	if VerifyPassword(hash, "wrong-password") {
		t.Fatal("wrong password must not verify")
	}
}
