package pkg

import (
	"testing"
	"time"
)

func TestGenerateToken_ParseToken_RoundTrip(t *testing.T) {
	claims := Claims{StaffID: "staff-1", HospitalID: "hospital-1", Username: "somchai.p"}
	token, err := GenerateToken("secret", time.Hour, claims)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	got, err := ParseToken("secret", token)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if got != claims {
		t.Errorf("ParseToken() = %+v, want %+v", got, claims)
	}
}

func TestParseToken_Expired(t *testing.T) {
	claims := Claims{StaffID: "staff-1", HospitalID: "hospital-1", Username: "somchai.p"}
	token, err := GenerateToken("secret", -time.Hour, claims)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if _, err := ParseToken("secret", token); err != ErrInvalidToken {
		t.Fatalf("ParseToken() error = %v, want ErrInvalidToken", err)
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	claims := Claims{StaffID: "staff-1", HospitalID: "hospital-1", Username: "somchai.p"}
	token, err := GenerateToken("secret", time.Hour, claims)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if _, err := ParseToken("a-different-secret", token); err != ErrInvalidToken {
		t.Fatalf("ParseToken() error = %v, want ErrInvalidToken", err)
	}
}

func TestParseToken_MalformedToken(t *testing.T) {
	if _, err := ParseToken("secret", "not.a.jwt"); err != ErrInvalidToken {
		t.Fatalf("ParseToken() error = %v, want ErrInvalidToken", err)
	}
}
