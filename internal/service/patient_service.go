package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/MoriMokata/hospital-middleware-system/internal/his"
	"github.com/MoriMokata/hospital-middleware-system/internal/repository"
)

// HISClientFactory builds a HISClient for a hospital given its
// his_base_url (read from the hospitals table, not hardcoded — see
// docs/architecture.md), keyed by hospitals.his_adapter_type. Add a new
// hospital by registering a new factory here, not by changing
// PatientService.
type HISClientFactory func(baseURL string) his.HISClient

// PatientService implements /patient/search (docs/api-spec.md). Every
// search is scoped to hospitalID from the caller's JWT — the data
// isolation boundary described in docs/er-diagram.md — never from
// client-supplied input.
type PatientService struct {
	Hospitals    repository.HospitalRepository
	Patients     repository.PatientRepository
	HISFactories map[string]HISClientFactory
}

func NewPatientService(hospitals repository.HospitalRepository, patients repository.PatientRepository, hisFactories map[string]HISClientFactory) *PatientService {
	return &PatientService{Hospitals: hospitals, Patients: patients, HISFactories: hisFactories}
}

// SearchInput mirrors the optional /patient/search request fields.
// DateOfBirth is the raw "YYYY-MM-DD" string from the request body.
type SearchInput struct {
	NationalID  *string
	PassportID  *string
	FirstName   *string
	MiddleName  *string
	LastName    *string
	DateOfBirth *string
	PhoneNumber *string
	Email       *string
}

type PatientOutput struct {
	ID           uuid.UUID
	PatientHN    *string
	NationalID   *string
	PassportID   *string
	FirstNameTH  *string
	MiddleNameTH *string
	LastNameTH   *string
	FirstNameEN  *string
	MiddleNameEN *string
	LastNameEN   *string
	DateOfBirth  *time.Time
	PhoneNumber  *string
	Email        *string
	Gender       *string
}

// Search runs the flow in docs/architecture.md#3-request-flow-post-patientsearch:
// if national_id/passport_id is supplied, it first looks the id up against
// the caller's hospital HIS and upserts the normalized result, then always
// queries the local patients table scoped to hospitalID. A HIS lookup
// failure (not found, timeout, 5xx) never fails the request — it just
// falls back to whatever's already in the local DB, per the assignment's
// acceptance criteria for this endpoint.
func (s *PatientService) Search(ctx context.Context, hospitalIDStr string, in SearchInput) ([]PatientOutput, error) {
	hospitalID, err := uuid.Parse(hospitalIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid hospital_id in token: %w", err)
	}

	var dob *time.Time
	if in.DateOfBirth != nil && *in.DateOfBirth != "" {
		parsed, err := time.Parse("2006-01-02", *in.DateOfBirth)
		if err != nil {
			return nil, &ValidationError{Msg: "date_of_birth must be in YYYY-MM-DD format"}
		}
		dob = &parsed
	}

	lookupID := ""
	switch {
	case in.NationalID != nil && *in.NationalID != "":
		lookupID = *in.NationalID
	case in.PassportID != nil && *in.PassportID != "":
		lookupID = *in.PassportID
	}
	if lookupID != "" {
		s.syncFromHIS(ctx, hospitalID, lookupID)
	}

	results, err := s.Patients.Search(ctx, hospitalID, repository.PatientFilter{
		NationalID:  in.NationalID,
		PassportID:  in.PassportID,
		FirstName:   in.FirstName,
		MiddleName:  in.MiddleName,
		LastName:    in.LastName,
		DateOfBirth: dob,
		PhoneNumber: in.PhoneNumber,
		Email:       in.Email,
	})
	if err != nil {
		return nil, err
	}

	out := make([]PatientOutput, len(results))
	for i, p := range results {
		out[i] = PatientOutput{
			ID:           p.ID,
			PatientHN:    p.PatientHN,
			NationalID:   p.NationalID,
			PassportID:   p.PassportID,
			FirstNameTH:  p.FirstNameTH,
			MiddleNameTH: p.MiddleNameTH,
			LastNameTH:   p.LastNameTH,
			FirstNameEN:  p.FirstNameEN,
			MiddleNameEN: p.MiddleNameEN,
			LastNameEN:   p.LastNameEN,
			DateOfBirth:  p.DateOfBirth,
			PhoneNumber:  p.PhoneNumber,
			Email:        p.Email,
			Gender:       p.Gender,
		}
	}
	return out, nil
}

// syncFromHIS best-efforts a live HIS lookup and upserts the result. Any
// failure (no adapter registered, not found upstream, timeout, 5xx) is
// swallowed — the caller falls back to the local DB search that follows.
func (s *PatientService) syncFromHIS(ctx context.Context, hospitalID uuid.UUID, id string) {
	hospital, err := s.Hospitals.FindByID(ctx, hospitalID)
	if err != nil {
		return
	}

	factory, ok := s.HISFactories[hospital.HISAdapterType]
	if !ok {
		return
	}
	client := factory(hospital.HISBaseURL)

	patient, err := client.Search(ctx, id)
	if err != nil {
		if !errors.Is(err, his.ErrNotFound) {
			// Upstream error (timeout/5xx/etc): log and fall back to local DB.
			log.Printf("his sync failed for hospital %s: %v", hospitalID, err)
		}
		return
	}

	patient.HospitalID = hospitalID
	if _, err := s.Patients.Upsert(ctx, patient); err != nil {
		log.Printf("upsert synced patient failed for hospital %s: %v", hospitalID, err)
	}
}
