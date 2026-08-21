package pkg

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ErrInvalidToken covers any token that fails to parse, fails signature
// verification, or is expired.
var ErrInvalidToken = errors.New("pkg: invalid or expired token")

// Claims are the JWT claims issued on /staff/login (see docs/api-spec.md):
// staff_id, hospital_id, username, iat, exp.
type Claims struct {
	StaffID    string
	HospitalID string
	Username   string
}

type tokenClaims struct {
	StaffID    string `json:"staff_id"`
	HospitalID string `json:"hospital_id"`
	Username   string `json:"username"`
	jwt.RegisteredClaims
}

// GenerateToken signs a JWT for the given claims, expiring after expiry.
func GenerateToken(secret string, expiry time.Duration, claims Claims) (string, error) {
	now := time.Now()
	tc := tokenClaims{
		StaffID:    claims.StaffID,
		HospitalID: claims.HospitalID,
		Username:   claims.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tc)
	return token.SignedString([]byte(secret))
}

// ParseToken verifies the token's signature and expiry and returns its
// claims. Any failure (bad signature, expired, malformed) maps to
// ErrInvalidToken.
func ParseToken(secret, tokenString string) (Claims, error) {
	var tc tokenClaims
	token, err := jwt.ParseWithClaims(tokenString, &tc, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return Claims{}, ErrInvalidToken
	}
	return Claims{StaffID: tc.StaffID, HospitalID: tc.HospitalID, Username: tc.Username}, nil
}
