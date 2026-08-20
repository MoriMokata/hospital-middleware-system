package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/MoriMokata/hospital-middleware-system/internal/domain"
	"github.com/MoriMokata/hospital-middleware-system/internal/pkg"
	"github.com/MoriMokata/hospital-middleware-system/internal/repository"
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
	byHospitalAndUsername map[string]domain.Staff
}

func staffKey(hospitalID uuid.UUID, username string) string {
	return hospitalID.String() + "/" + username
}

func (f *fakeStaffRepo) Create(_ context.Context, staff domain.Staff) (domain.Staff, error) {
	key := staffKey(staff.HospitalID, staff.Username)
	if _, exists := f.byHospitalAndUsername[key]; exists {
		return domain.Staff{}, repository.ErrConflict
	}
	staff.ID = uuid.New()
	staff.CreatedAt = time.Now()
	staff.UpdatedAt = staff.CreatedAt
	f.byHospitalAndUsername[key] = staff
	return staff, nil
}

func (f *fakeStaffRepo) FindByHospitalAndUsername(_ context.Context, hospitalID uuid.UUID, username string) (domain.Staff, error) {
	staff, ok := f.byHospitalAndUsername[staffKey(hospitalID, username)]
	if !ok {
		return domain.Staff{}, repository.ErrNotFound
	}
	return staff, nil
}

func newTestStaffService() (*StaffService, domain.Hospital) {
	hospital := domain.Hospital{ID: uuid.New(), Slug: "hospital-a", Name: "Hospital A"}
	hospitals := &fakeHospitalRepo{bySlug: map[string]domain.Hospital{"hospital-a": hospital}}
	staff := &fakeStaffRepo{byHospitalAndUsername: map[string]domain.Staff{}}
	return NewStaffService(hospitals, staff, "test-secret", time.Hour), hospital
}

func TestStaffService_CreateStaff_Success(t *testing.T) {
	svc, _ := newTestStaffService()

	out, err := svc.CreateStaff(context.Background(), CreateStaffInput{
		Username: "somchai.p", Password: "P@ssw0rd123", Hospital: "hospital-a",
	})
	if err != nil {
		t.Fatalf("CreateStaff() error = %v", err)
	}
	if out.Username != "somchai.p" || out.Hospital != "hospital-a" {
		t.Errorf("CreateStaff() = %+v", out)
	}
	if out.ID == uuid.Nil {
		t.Error("CreateStaff() did not assign an ID")
	}
}

func TestStaffService_CreateStaff_DuplicateUsername(t *testing.T) {
	svc, _ := newTestStaffService()
	ctx := context.Background()
	in := CreateStaffInput{Username: "somchai.p", Password: "P@ssw0rd123", Hospital: "hospital-a"}

	if _, err := svc.CreateStaff(ctx, in); err != nil {
		t.Fatalf("first CreateStaff() error = %v", err)
	}
	if _, err := svc.CreateStaff(ctx, in); err != ErrUsernameTaken {
		t.Fatalf("second CreateStaff() error = %v, want ErrUsernameTaken", err)
	}
}

func TestStaffService_CreateStaff_HospitalNotFound(t *testing.T) {
	svc, _ := newTestStaffService()

	_, err := svc.CreateStaff(context.Background(), CreateStaffInput{
		Username: "somchai.p", Password: "P@ssw0rd123", Hospital: "no-such-hospital",
	})
	if err != ErrHospitalNotFound {
		t.Fatalf("CreateStaff() error = %v, want ErrHospitalNotFound", err)
	}
}

func TestStaffService_CreateStaff_ValidationErrors(t *testing.T) {
	svc, _ := newTestStaffService()

	tests := []struct {
		name string
		in   CreateStaffInput
	}{
		{"missing username", CreateStaffInput{Username: "", Password: "P@ssw0rd123", Hospital: "hospital-a"}},
		{"short password", CreateStaffInput{Username: "somchai.p", Password: "short", Hospital: "hospital-a"}},
		{"missing hospital", CreateStaffInput{Username: "somchai.p", Password: "P@ssw0rd123", Hospital: ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateStaff(context.Background(), tt.in)
			var verr *ValidationError
			if err == nil {
				t.Fatal("CreateStaff() expected a validation error, got nil")
			}
			if !errors.As(err, &verr) {
				t.Fatalf("CreateStaff() error = %v, want *ValidationError", err)
			}
		})
	}
}

func TestStaffService_Login_Success(t *testing.T) {
	svc, _ := newTestStaffService()
	ctx := context.Background()
	if _, err := svc.CreateStaff(ctx, CreateStaffInput{Username: "somchai.p", Password: "P@ssw0rd123", Hospital: "hospital-a"}); err != nil {
		t.Fatalf("CreateStaff() error = %v", err)
	}

	out, err := svc.Login(ctx, LoginInput{Username: "somchai.p", Password: "P@ssw0rd123", Hospital: "hospital-a"})
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if out.AccessToken == "" {
		t.Error("Login() did not return an access_token")
	}
	if out.ExpiresIn != int(time.Hour.Seconds()) {
		t.Errorf("Login() ExpiresIn = %d, want %d", out.ExpiresIn, int(time.Hour.Seconds()))
	}

	claims, err := pkg.ParseToken(svc.JWTSecret, out.AccessToken)
	if err != nil {
		t.Fatalf("ParseToken() error = %v", err)
	}
	if claims.Username != "somchai.p" {
		t.Errorf("claims.Username = %q, want somchai.p", claims.Username)
	}
}

func TestStaffService_Login_WrongPassword(t *testing.T) {
	svc, _ := newTestStaffService()
	ctx := context.Background()
	if _, err := svc.CreateStaff(ctx, CreateStaffInput{Username: "somchai.p", Password: "P@ssw0rd123", Hospital: "hospital-a"}); err != nil {
		t.Fatalf("CreateStaff() error = %v", err)
	}

	_, err := svc.Login(ctx, LoginInput{Username: "somchai.p", Password: "wrong-password", Hospital: "hospital-a"})
	if err != ErrInvalidCredentials {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestStaffService_Login_UnknownUsername(t *testing.T) {
	svc, _ := newTestStaffService()

	_, err := svc.Login(context.Background(), LoginInput{Username: "nobody", Password: "P@ssw0rd123", Hospital: "hospital-a"})
	if err != ErrInvalidCredentials {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestStaffService_Login_HospitalNotFound(t *testing.T) {
	svc, _ := newTestStaffService()

	_, err := svc.Login(context.Background(), LoginInput{Username: "somchai.p", Password: "P@ssw0rd123", Hospital: "no-such-hospital"})
	if err != ErrHospitalNotFound {
		t.Fatalf("Login() error = %v, want ErrHospitalNotFound", err)
	}
}
