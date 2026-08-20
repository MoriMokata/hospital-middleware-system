package his

import (
	"context"
	"errors"

	"github.com/MoriMokata/hospital-middleware-system/internal/domain"
)

// ErrNotFound is returned when the HIS has no patient matching the given id.
var ErrNotFound = errors.New("his: patient not found")

// HISClient looks up a single patient by national_id or passport_id from a
// hospital's own Hospital Information System, returning it mapped into the
// normalized domain.Patient shape. The returned Patient's ID, HospitalID,
// CreatedAt, and UpdatedAt are left zero — the caller (the patient service)
// fills those in before persisting. One implementation per hospital; add a
// new hospital by adding a new HISClient implementation, not by changing
// callers (see docs/architecture.md).
type HISClient interface {
	Search(ctx context.Context, id string) (domain.Patient, error)
}
