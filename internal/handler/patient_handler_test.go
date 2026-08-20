package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/MoriMokata/hospital-middleware-system/internal/domain"
	"github.com/MoriMokata/hospital-middleware-system/internal/his"
	"github.com/MoriMokata/hospital-middleware-system/internal/middleware"
	"github.com/MoriMokata/hospital-middleware-system/internal/pkg"
	"github.com/MoriMokata/hospital-middleware-system/internal/repository"
	"github.com/MoriMokata/hospital-middleware-system/internal/service"
)

type fakePatientRepo struct {
	patients []domain.Patient
}

func (f *fakePatientRepo) Upsert(_ context.Context, p domain.Patient) (domain.Patient, error) {
	p.ID = uuid.New()
	f.patients = append(f.patients, p)
	return p, nil
}

func (f *fakePatientRepo) Search(_ context.Context, hospitalID uuid.UUID, filter repository.PatientFilter) ([]domain.Patient, error) {
	var out []domain.Patient
	for _, p := range f.patients {
		if p.HospitalID == hospitalID {
			out = append(out, p)
		}
	}
	return out, nil
}

type noopHISClient struct{}

func (noopHISClient) Search(context.Context, string) (domain.Patient, error) {
	return domain.Patient{}, his.ErrNotFound
}

const testJWTSecret = "test-secret"

func strp(s string) *string { return &s }

func newTestHospitals() (hospitalA, hospitalB domain.Hospital) {
	hospitalA = domain.Hospital{ID: uuid.New(), Slug: "hospital-a", HISAdapterType: "noop"}
	hospitalB = domain.Hospital{ID: uuid.New(), Slug: "hospital-b", HISAdapterType: "noop"}
	return hospitalA, hospitalB
}

func newTestPatientHandlerRouter(hospitalA, hospitalB domain.Hospital, seed []domain.Patient) *gin.Engine {
	hospitals := &fakeHospitalRepo{bySlug: map[string]domain.Hospital{
		"hospital-a": hospitalA,
		"hospital-b": hospitalB,
	}}
	patients := &fakePatientRepo{patients: seed}
	factories := map[string]service.HISClientFactory{
		"noop": func(string) his.HISClient { return noopHISClient{} },
	}
	patientService := service.NewPatientService(hospitals, patients, factories)
	patientHandler := NewPatientHandler(patientService)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/patient/search", middleware.Auth(testJWTSecret), patientHandler.Search)
	return router
}

func tokenFor(hospitalID uuid.UUID) string {
	token, _ := pkg.GenerateToken(testJWTSecret, time.Hour, pkg.Claims{StaffID: "staff-1", HospitalID: hospitalID.String(), Username: "somchai.p"})
	return token
}

func performPatientSearch(router *gin.Engine, authHeader, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/patient/search", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestPatientHandler_Search_ReturnsOwnHospitalMatches(t *testing.T) {
	nationalID := "1234567890123"
	hospitalA, hospitalB := newTestHospitals()
	router := newTestPatientHandlerRouter(hospitalA, hospitalB, []domain.Patient{
		{HospitalID: hospitalA.ID, NationalID: &nationalID},
		{HospitalID: hospitalB.ID, NationalID: &nationalID},
	})

	rec := performPatientSearch(router, "Bearer "+tokenFor(hospitalA.ID), `{}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body struct {
		Results []map[string]any `json:"results"`
		Count   int              `json:"count"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if body.Count != 1 {
		t.Fatalf("count = %d, want 1 (must not include hospital B's patient)", body.Count)
	}
}

func TestPatientHandler_Search_MissingToken(t *testing.T) {
	hospitalA, hospitalB := newTestHospitals()
	router := newTestPatientHandlerRouter(hospitalA, hospitalB, nil)

	rec := performPatientSearch(router, "", `{}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestPatientHandler_Search_InvalidToken(t *testing.T) {
	hospitalA, hospitalB := newTestHospitals()
	router := newTestPatientHandlerRouter(hospitalA, hospitalB, nil)

	rec := performPatientSearch(router, "Bearer not-a-real-token", `{}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestPatientHandler_Search_EmptyBodyReturnsAllOwnHospitalPatients(t *testing.T) {
	hospitalA, hospitalB := newTestHospitals()
	router := newTestPatientHandlerRouter(hospitalA, hospitalB, []domain.Patient{
		{HospitalID: hospitalA.ID, FirstName: strp("Somsri")},
		{HospitalID: hospitalA.ID, FirstName: strp("Somchai")},
	})

	rec := performPatientSearch(router, "Bearer "+tokenFor(hospitalA.ID), `{}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body struct {
		Count int `json:"count"`
	}
	json.Unmarshal(rec.Body.Bytes(), &body)
	if body.Count != 2 {
		t.Fatalf("count = %d, want 2", body.Count)
	}
}

func TestPatientHandler_Search_InvalidDateOfBirth(t *testing.T) {
	hospitalA, hospitalB := newTestHospitals()
	router := newTestPatientHandlerRouter(hospitalA, hospitalB, nil)

	rec := performPatientSearch(router, "Bearer "+tokenFor(hospitalA.ID), `{"date_of_birth":"not-a-date"}`)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	assertErrorCode(t, rec, "VALIDATION_ERROR")
}
