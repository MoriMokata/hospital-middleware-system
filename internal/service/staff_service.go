package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/MoriMokata/hospital-middleware-system/internal/domain"
	"github.com/MoriMokata/hospital-middleware-system/internal/pkg"
	"github.com/MoriMokata/hospital-middleware-system/internal/repository"
)

const minPasswordLength = 8

// ErrHospitalNotFound means the `hospital` slug doesn't match any known
// hospital.
var ErrHospitalNotFound = errors.New("service: hospital not found")

// ErrUsernameTaken means the username already exists within that hospital.
var ErrUsernameTaken = errors.New("service: username already taken")

// ErrInvalidCredentials means the username/password/hospital combination
// didn't match on login (added here since writeStaffError in the handler
// maps it alongside the create-flow errors; used starting Task 09).
var ErrInvalidCredentials = errors.New("service: invalid credentials")

// ValidationError carries a human-readable reason a request body failed
// validation, mapped to a 400 VALIDATION_ERROR response by the handler.
type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }

// StaffService implements staff creation and login (docs/api-spec.md).
type StaffService struct {
	Hospitals repository.HospitalRepository
	Staff     repository.StaffRepository
}

func NewStaffService(hospitals repository.HospitalRepository, staff repository.StaffRepository) *StaffService {
	return &StaffService{Hospitals: hospitals, Staff: staff}
}

type CreateStaffInput struct {
	Username string
	Password string
	Hospital string
}

type StaffOutput struct {
	ID        uuid.UUID
	Username  string
	Hospital  string
	CreatedAt time.Time
}

func (s *StaffService) CreateStaff(ctx context.Context, in CreateStaffInput) (StaffOutput, error) {
	username := strings.TrimSpace(in.Username)
	hospitalSlug := strings.TrimSpace(in.Hospital)

	switch {
	case username == "":
		return StaffOutput{}, &ValidationError{Msg: "username is required"}
	case len(in.Password) < minPasswordLength:
		return StaffOutput{}, &ValidationError{Msg: "password must be at least 8 characters"}
	case hospitalSlug == "":
		return StaffOutput{}, &ValidationError{Msg: "hospital is required"}
	}

	hospital, err := s.Hospitals.FindBySlug(ctx, hospitalSlug)
	if errors.Is(err, repository.ErrNotFound) {
		return StaffOutput{}, ErrHospitalNotFound
	}
	if err != nil {
		return StaffOutput{}, err
	}

	hash, err := pkg.HashPassword(in.Password)
	if err != nil {
		return StaffOutput{}, err
	}

	created, err := s.Staff.Create(ctx, domain.Staff{
		HospitalID:   hospital.ID,
		Username:     username,
		PasswordHash: hash,
	})
	if errors.Is(err, repository.ErrConflict) {
		return StaffOutput{}, ErrUsernameTaken
	}
	if err != nil {
		return StaffOutput{}, err
	}

	return StaffOutput{
		ID:        created.ID,
		Username:  created.Username,
		Hospital:  hospital.Slug,
		CreatedAt: created.CreatedAt,
	}, nil
}
