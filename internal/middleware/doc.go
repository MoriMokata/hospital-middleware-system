// Package middleware contains Gin middleware — currently JWT auth, which
// extracts staff_id and hospital_id from the token into the request
// context (request logging and panic recovery are covered by gin.Default()
// in cmd/api/main.go).
package middleware
