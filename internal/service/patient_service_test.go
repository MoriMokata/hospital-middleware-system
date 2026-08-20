package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/MoriMokata/hospital-middleware-system/internal/domain"
	"github.com/MoriMokata/hospital-middleware-system/internal/his"
	"github.com/MoriMokata/hospital-middleware-system/internal/repository"
)

type fakePatientRepo struct {
	patients []domain.Patient
}

func (f *fakePatientRepo) Upsert(_ context.Context, p domain.Patient) (domain.Patient, error) {
	for i, existing := range f.patients {
		if existing.HospitalID != p.HospitalID {
			continue
		}
		if p.NationalID != nil && existing.NationalID != nil && *existing.NationalID == *p.NationalID {
			p.ID = existing.ID
			f.patients[i] = p
			return p, nil
		}
		if p.PassportID != nil && existing.PassportID != nil && *existing.PassportID == *p.PassportID {
			p.ID = existing.ID
			f.patients[i] = p
			return p, nil
		}
	}
	p.ID = uuid.New()
	f.patients = append(f.patients, p)
	return p, nil
}

func (f *fakePatientRepo) Search(_ context.Context, hospitalID uuid.UUID, filter repository.PatientFilter) ([]domain.Patient, error) {
	var out []domain.Patient
	for _, p := range f.patients {
		if p.HospitalID != hospitalID {
			continue
		}
		if !matchesLike(filter.NationalID, p.NationalID) ||
			!matchesLike(filter.PassportID, p.PassportID) ||
			!matchesLike(filter.FirstName, p.FirstName) ||
			!matchesLike(filter.LastName, p.LastName) ||
			!matchesLike(filter.PhoneNumber, p.PhoneNumber) ||
			!matchesLike(filter.Email, p.Email) {
			continue
		}
		if filter.DateOfBirth != nil {
			if p.DateOfBirth == nil || !p.DateOfBirth.Equal(*filter.DateOfBirth) {
				continue
			}
		}
		out = append(out, p)
	}
	return out, nil
}

func matchesLike(filterValue, fieldValue *string) bool {
	if filterValue == nil || *filterValue == "" {
		return true
	}
	if fieldValue == nil {
		return false
	}
	return strings.Contains(strings.ToLower(*fieldValue), strings.ToLower(*filterValue))
}

type fakeHISClient struct {
	result domain.Patient
	err    error
}

func (f *fakeHISClient) Search(_ context.Context, _ string) (domain.Patient, error) {
	return f.result, f.err
}

func newTestPatientService(hisClient his.HISClient) (*PatientService, domain.Hospital, domain.Hospital) {
	hospitalA := domain.Hospital{ID: uuid.New(), Slug: "hospital-a", HISAdapterType: "fake_his", HISBaseURL: "http://fake"}
	hospitalB := domain.Hospital{ID: uuid.New(), Slug: "hospital-b", HISAdapterType: "fake_his", HISBaseURL: "http://fake"}
	hospitals := &fakeHospitalRepo{bySlug: map[string]domain.Hospital{
		"hospital-a": hospitalA,
		"hospital-b": hospitalB,
	}}
	patients := &fakePatientRepo{}
	factories := map[string]HISClientFactory{
		"fake_his": func(string) his.HISClient { return hisClient },
	}
	return NewPatientService(hospitals, patients, factories), hospitalA, hospitalB
}

func strp(s string) *string { return &s }

func TestPatientService_Search_NeverReturnsAnotherHospitalsPatients(t *testing.T) {
	svc, hospitalA, hospitalB := newTestPatientService(&fakeHISClient{err: his.ErrNotFound})

	nationalID := "1234567890123"
	svc.Patients.Upsert(context.Background(), domain.Patient{HospitalID: hospitalA.ID, NationalID: &nationalID, FirstName: strp("Somsri A")})
	svc.Patients.Upsert(context.Background(), domain.Patient{HospitalID: hospitalB.ID, NationalID: &nationalID, FirstName: strp("Somsri B")})

	results, err := svc.Search(context.Background(), hospitalB.ID.String(), SearchInput{NationalID: &nationalID})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Search() for hospital B returned %d results, want 1 (only hospital B's own record)", len(results))
	}

	resultsA, err := svc.Search(context.Background(), hospitalA.ID.String(), SearchInput{NationalID: &nationalID})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(resultsA) != 1 {
		t.Fatalf("Search() for hospital A returned %d results, want 1", len(resultsA))
	}
}

func TestPatientService_Search_SyncsFromHISOnIDLookup(t *testing.T) {
	nationalID := "1234567890123"
	hisPatient := domain.Patient{NationalID: &nationalID, FirstNameEN: strp("Somsri"), LastNameEN: strp("Jaidee")}
	svc, hospitalA, _ := newTestPatientService(&fakeHISClient{result: hisPatient})

	results, err := svc.Search(context.Background(), hospitalA.ID.String(), SearchInput{NationalID: &nationalID})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("Search() returned %d results, want 1 synced from HIS", len(results))
	}
	if results[0].FirstNameEN == nil || *results[0].FirstNameEN != "Somsri" {
		t.Errorf("Search() result = %+v, want synced HIS data", results[0])
	}
}

func TestPatientService_Search_HISErrorFallsBackToLocalDB(t *testing.T) {
	svc, hospitalA, _ := newTestPatientService(&fakeHISClient{err: errors.New("upstream timeout")})

	nationalID := "1234567890123"
	svc.Patients.Upsert(context.Background(), domain.Patient{HospitalID: hospitalA.ID, NationalID: &nationalID, FirstName: strp("Cached Somsri")})

	results, err := svc.Search(context.Background(), hospitalA.ID.String(), SearchInput{NationalID: &nationalID})
	if err != nil {
		t.Fatalf("Search() error = %v, want no error (HIS failure must not crash the request)", err)
	}
	if len(results) != 1 {
		t.Fatalf("Search() returned %d results, want 1 (fallback to local DB)", len(results))
	}
}

func TestPatientService_Search_InvalidDateOfBirth(t *testing.T) {
	svc, hospitalA, _ := newTestPatientService(&fakeHISClient{err: his.ErrNotFound})

	badDate := "not-a-date"
	_, err := svc.Search(context.Background(), hospitalA.ID.String(), SearchInput{DateOfBirth: &badDate})
	var verr *ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("Search() error = %v, want *ValidationError", err)
	}
}
