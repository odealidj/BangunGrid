package hasher_test

import (
	"testing"

	"bangungrid/services/auth-service/internal/adapters/security/hasher"
)

func TestBcryptHasher(t *testing.T) {
	h := hasher.NewBcryptHasher(10)

	rawPassword := "EnterprisePass123!"

	// Test Hash
	hashed, err := h.Hash(rawPassword)
	if err != nil {
		t.Fatalf("expected no error hashing password, got: %v", err)
	}

	if hashed == "" || hashed == rawPassword {
		t.Fatalf("expected valid bcrypt hash, got: %s", hashed)
	}

	// Test Compare valid
	if !h.Compare(hashed, rawPassword) {
		t.Fatalf("expected password to match hash")
	}

	// Test Compare invalid
	if h.Compare(hashed, "WrongPass123!") {
		t.Fatalf("expected wrong password to fail match")
	}
}
