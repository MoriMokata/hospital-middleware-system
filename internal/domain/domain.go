package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Hospital is a row in the hospitals table — the registry of every hospital
// the middleware integrates with.
type Hospital struct {
	ID             uuid.UUID
	Name           string
	Slug           string
	HISAdapterType string
	HISBaseURL     string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Staff is a row in the staff table. Username is only unique within a
// hospital (see the (hospital_id, username) unique index).
type Staff struct {
	ID           uuid.UUID
	HospitalID   uuid.UUID
	Username     string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Patient is a row in the patients table — a superset schema covering both
// the generic search fields used by /patient/search and the *_th/*_en split
// names returned by hospitals like Hospital A. Pointer fields are nullable
// in the DB. HospitalID is the data-isolation boundary: every repository
// query against patients must filter on it.
type Patient struct {
	ID            uuid.UUID
	HospitalID    uuid.UUID
	PatientHN     *string
	NationalID    *string
	PassportID    *string
	FirstName     *string
	MiddleName    *string
	LastName      *string
	FirstNameTH   *string
	MiddleNameTH  *string
	LastNameTH    *string
	FirstNameEN   *string
	MiddleNameEN  *string
	LastNameEN    *string
	DateOfBirth   *time.Time
	PhoneNumber   *string
	Email         *string
	Gender        *string
	RawSource     json.RawMessage
	SyncedAt      *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
