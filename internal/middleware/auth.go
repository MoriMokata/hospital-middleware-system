package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/MoriMokata/hospital-middleware-system/internal/pkg"
)

// Context keys the Auth middleware sets after a successful JWT check.
const (
	ContextKeyStaffID    = "staff_id"
	ContextKeyHospitalID = "hospital_id"
	ContextKeyUsername   = "username"
)

// Auth validates the JWT in the Authorization: Bearer <token> header and
// injects staff_id/hospital_id/username into the request context for
// downstream handlers. hospital_id sourced this way — never from a
// client-supplied field — is what makes patient search's hospital scoping
// safe (see docs/architecture.md).
func Auth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		token, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewErrorEnvelope("UNAUTHORIZED", "missing or invalid Authorization header"))
			return
		}

		claims, err := pkg.ParseToken(jwtSecret, token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, pkg.NewErrorEnvelope("UNAUTHORIZED", "invalid or expired token"))
			return
		}

		c.Set(ContextKeyStaffID, claims.StaffID)
		c.Set(ContextKeyHospitalID, claims.HospitalID)
		c.Set(ContextKeyUsername, claims.Username)
		c.Next()
	}
}

// HospitalIDFromContext returns the authenticated staff's hospital_id, as
// set by Auth. ok is false if Auth hasn't run on this request.
func HospitalIDFromContext(c *gin.Context) (string, bool) {
	v, ok := c.Get(ContextKeyHospitalID)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// StaffIDFromContext returns the authenticated staff's id, as set by Auth.
func StaffIDFromContext(c *gin.Context) (string, bool) {
	v, ok := c.Get(ContextKeyStaffID)
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}
