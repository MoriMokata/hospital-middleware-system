package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/MoriMokata/hospital-middleware-system/internal/pkg"
)

func newAuthTestRouter(secret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/protected", Auth(secret), func(c *gin.Context) {
		hospitalID, _ := HospitalIDFromContext(c)
		staffID, _ := StaffIDFromContext(c)
		c.JSON(http.StatusOK, gin.H{"hospital_id": hospitalID, "staff_id": staffID})
	})
	return router
}

func doProtectedRequest(router *gin.Engine, authHeader string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestAuth_ValidToken_Passes(t *testing.T) {
	secret := "test-secret"
	router := newAuthTestRouter(secret)
	token, err := pkg.GenerateToken(secret, time.Hour, pkg.Claims{StaffID: "staff-1", HospitalID: "hospital-1", Username: "somchai.p"})
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	rec := doProtectedRequest(router, "Bearer "+token)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}
}

func TestAuth_MissingToken(t *testing.T) {
	router := newAuthTestRouter("test-secret")

	rec := doProtectedRequest(router, "")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuth_MalformedHeader(t *testing.T) {
	router := newAuthTestRouter("test-secret")

	rec := doProtectedRequest(router, "not-a-bearer-token")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuth_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	router := newAuthTestRouter(secret)
	token, err := pkg.GenerateToken(secret, -time.Hour, pkg.Claims{StaffID: "staff-1", HospitalID: "hospital-1", Username: "somchai.p"})
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	rec := doProtectedRequest(router, "Bearer "+token)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuth_TamperedSignature(t *testing.T) {
	secret := "test-secret"
	router := newAuthTestRouter(secret)
	token, err := pkg.GenerateToken("a-different-secret", time.Hour, pkg.Claims{StaffID: "staff-1", HospitalID: "hospital-1", Username: "somchai.p"})
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	rec := doProtectedRequest(router, "Bearer "+token)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
