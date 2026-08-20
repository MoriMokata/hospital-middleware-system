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
	"github.com/MoriMokata/hospital-middleware-system/internal/repository"
	"github.com/MoriMokata/hospital-middleware-system/internal/service"
)

type fakeHospitalRepo struct {
	bySlug map[string]domain.Hospital
}

func (f *fakeHospitalRepo) FindBySlug(_ context.Context, slug string) (domain.Hospital, error) {
	h, ok := f.bySlug[slug]
	if !ok {
		return domain.Hospital{}, repository.ErrNotFound
	}
	return h, nil
}

func (f *fakeHospitalRepo) FindByID(_ context.Context, id uuid.UUID) (domain.Hospital, error) {
	for _, h := range f.bySlug {
		if h.ID == id {
			return h, nil
		}
	}
	return domain.Hospital{}, repository.ErrNotFound
}

type fakeStaffRepo struct {
	byKey map[string]domain.Staff
}

func (f *fakeStaffRepo) Create(_ context.Context, staff domain.Staff) (domain.Staff, error) {
	key := staff.HospitalID.String() + "/" + staff.Username
	if _, exists := f.byKey[key]; exists {
		return domain.Staff{}, repository.ErrConflict
	}
	staff.ID = uuid.New()
	staff.CreatedAt = time.Now()
	f.byKey[key] = staff
	return staff, nil
}

func (f *fakeStaffRepo) FindByHospitalAndUsername(_ context.Context, hospitalID uuid.UUID, username string) (domain.Staff, error) {
	staff, ok := f.byKey[hospitalID.String()+"/"+username]
	if !ok {
		return domain.Staff{}, repository.ErrNotFound
	}
	return staff, nil
}

func newTestStaffHandler() *StaffHandler {
	hospital := domain.Hospital{ID: uuid.New(), Slug: "hospital-a", Name: "Hospital A"}
	hospitals := &fakeHospitalRepo{bySlug: map[string]domain.Hospital{"hospital-a": hospital}}
	staff := &fakeStaffRepo{byKey: map[string]domain.Staff{}}
	return NewStaffHandler(service.NewStaffService(hospitals, staff, "test-secret", time.Hour))
}

func performCreateStaff(h *StaffHandler, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/staff/create", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/staff/create", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func performLogin(h *StaffHandler, body string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/staff/login", h.Login)

	req := httptest.NewRequest(http.MethodPost, "/staff/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestStaffHandler_Create_Success(t *testing.T) {
	h := newTestStaffHandler()
	rec := performCreateStaff(h, `{"username":"somchai.p","password":"P@ssw0rd123","hospital":"hospital-a"}`)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if _, hasPassword := body["password"]; hasPassword {
		t.Error("response must never include the password")
	}
	if _, hasHash := body["password_hash"]; hasHash {
		t.Error("response must never include the password hash")
	}
	if body["username"] != "somchai.p" || body["hospital"] != "hospital-a" {
		t.Errorf("response body = %v", body)
	}
}

func TestStaffHandler_Create_DuplicateUsername(t *testing.T) {
	h := newTestStaffHandler()
	body := `{"username":"somchai.p","password":"P@ssw0rd123","hospital":"hospital-a"}`

	performCreateStaff(h, body)
	rec := performCreateStaff(h, body)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusConflict, rec.Body.String())
	}
	assertErrorCode(t, rec, "USERNAME_TAKEN")
}

func TestStaffHandler_Create_HospitalNotFound(t *testing.T) {
	h := newTestStaffHandler()
	rec := performCreateStaff(h, `{"username":"somchai.p","password":"P@ssw0rd123","hospital":"no-such-hospital"}`)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	assertErrorCode(t, rec, "HOSPITAL_NOT_FOUND")
}

func TestStaffHandler_Create_MissingFields(t *testing.T) {
	h := newTestStaffHandler()
	rec := performCreateStaff(h, `{"hospital":"hospital-a"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	assertErrorCode(t, rec, "VALIDATION_ERROR")
}

func TestStaffHandler_Create_PasswordTooShort(t *testing.T) {
	h := newTestStaffHandler()
	rec := performCreateStaff(h, `{"username":"somchai.p","password":"short","hospital":"hospital-a"}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	assertErrorCode(t, rec, "VALIDATION_ERROR")
}

func TestStaffHandler_Login_Success(t *testing.T) {
	h := newTestStaffHandler()
	performCreateStaff(h, `{"username":"somchai.p","password":"P@ssw0rd123","hospital":"hospital-a"}`)

	rec := performLogin(h, `{"username":"somchai.p","password":"P@ssw0rd123","hospital":"hospital-a"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var body struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if body.AccessToken == "" {
		t.Error("response missing access_token")
	}
	if body.TokenType != "Bearer" {
		t.Errorf("token_type = %q, want Bearer", body.TokenType)
	}
	if body.ExpiresIn != int(time.Hour.Seconds()) {
		t.Errorf("expires_in = %d, want %d", body.ExpiresIn, int(time.Hour.Seconds()))
	}
}

func TestStaffHandler_Login_WrongPassword(t *testing.T) {
	h := newTestStaffHandler()
	performCreateStaff(h, `{"username":"somchai.p","password":"P@ssw0rd123","hospital":"hospital-a"}`)

	rec := performLogin(h, `{"username":"somchai.p","password":"wrong-password","hospital":"hospital-a"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	assertErrorCode(t, rec, "INVALID_CREDENTIALS")
}

func TestStaffHandler_Login_UnknownUsername(t *testing.T) {
	h := newTestStaffHandler()

	rec := performLogin(h, `{"username":"nobody","password":"P@ssw0rd123","hospital":"hospital-a"}`)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
	assertErrorCode(t, rec, "INVALID_CREDENTIALS")
}

func TestStaffHandler_Login_HospitalNotFound(t *testing.T) {
	h := newTestStaffHandler()

	rec := performLogin(h, `{"username":"somchai.p","password":"P@ssw0rd123","hospital":"no-such-hospital"}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d, body = %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
	assertErrorCode(t, rec, "HOSPITAL_NOT_FOUND")
}

func assertErrorCode(t *testing.T, rec *httptest.ResponseRecorder, wantCode string) {
	t.Helper()
	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON response: %v", err)
	}
	if body.Error.Code != wantCode {
		t.Errorf("error.code = %q, want %q", body.Error.Code, wantCode)
	}
}
