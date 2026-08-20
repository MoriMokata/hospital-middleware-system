package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/MoriMokata/hospital-middleware-system/internal/domain"
)

// ErrNotFound is returned when a lookup finds no matching row.
var ErrNotFound = errors.New("repository: not found")

// ErrConflict is returned when a write would violate a uniqueness
// constraint (e.g. a duplicate username within a hospital).
var ErrConflict = errors.New("repository: conflict")

// HospitalRepository looks up hospitals by their public slug (e.g.
// "hospital-a"), the identifier clients pass as `hospital` on
// /staff/create and /staff/login, or by id (used to resolve the caller's
// own hospital — and its HIS adapter type/base URL — from the JWT's
// hospital_id claim during /patient/search).
type HospitalRepository interface {
	FindBySlug(ctx context.Context, slug string) (domain.Hospital, error)
	FindByID(ctx context.Context, id uuid.UUID) (domain.Hospital, error)
}

// StaffRepository creates and looks up staff. Username uniqueness is
// scoped to a hospital, matching the (hospital_id, username) index.
type StaffRepository interface {
	Create(ctx context.Context, staff domain.Staff) (domain.Staff, error)
	FindByHospitalAndUsername(ctx context.Context, hospitalID uuid.UUID, username string) (domain.Staff, error)
}

// PatientFilter holds the optional /patient/search fields. String fields
// match partial/case-insensitive; DateOfBirth matches exactly. Every field
// is combined with AND. HospitalID is supplied separately to Search (from
// the caller's JWT), never as part of this filter, so it can never be
// bypassed by a client-supplied value.
type PatientFilter struct {
	NationalID  *string
	PassportID  *string
	FirstName   *string
	MiddleName  *string
	LastName    *string
	DateOfBirth *time.Time
	PhoneNumber *string
	Email       *string
}

// PatientRepository stores and searches the local normalized patients
// table. Every query must be scoped by hospitalID — this is the data
// isolation boundary described in docs/er-diagram.md.
type PatientRepository interface {
	// Upsert inserts or updates a patient synced from a HIS lookup,
	// matched by hospitalID + whichever of NationalID/PassportID is set.
	Upsert(ctx context.Context, patient domain.Patient) (domain.Patient, error)
	Search(ctx context.Context, hospitalID uuid.UUID, filter PatientFilter) ([]domain.Patient, error)
}
