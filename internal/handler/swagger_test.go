package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSwaggerUI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/swagger", SwaggerUI)

	req := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "swagger-ui") {
		t.Error("response body does not look like a Swagger UI page")
	}
}

func TestOpenAPISpec(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	spec := []byte("openapi: 3.0.3\ninfo:\n  title: test\n")
	router.GET("/openapi.yaml", OpenAPISpec(spec))

	req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.String() != string(spec) {
		t.Errorf("body = %q, want %q", rec.Body.String(), spec)
	}
}
