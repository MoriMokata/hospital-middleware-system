package pkg

import "testing"

func TestHashPassword_VerifyPassword_CorrectPassword(t *testing.T) {
	hash, err := HashPassword("P@ssw0rd123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if hash == "P@ssw0rd123" {
		t.Fatal("HashPassword() returned the plaintext password unhashed")
	}
	if !VerifyPassword(hash, "P@ssw0rd123") {
		t.Error("VerifyPassword() = false for the correct password, want true")
	}
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	hash, err := HashPassword("P@ssw0rd123")
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}
	if VerifyPassword(hash, "wrong-password") {
		t.Error("VerifyPassword() = true for an incorrect password, want false")
	}
}
