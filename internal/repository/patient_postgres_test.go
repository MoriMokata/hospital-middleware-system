package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"

	"github.com/MoriMokata/hospital-middleware-system/internal/domain"
)

func patientRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "hospital_id", "patient_hn", "national_id", "passport_id",
		"first_name", "middle_name", "last_name",
		"first_name_th", "middle_name_th", "last_name_th",
		"first_name_en", "middle_name_en", "last_name_en",
		"date_of_birth", "phone_number", "email", "gender",
		"raw_source", "synced_at", "created_at", "updated_at",
	})
}

func TestPostgresPatientRepository_Upsert_InsertsWhenNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	hospitalID := uuid.New()
	newID := uuid.New()
	now := time.Now()
	nationalID := "1234567890123"
	firstName := "Somsri"

	mock.ExpectQuery("SELECT id FROM patients WHERE hospital_id = \\$1 AND national_id = \\$2").
		WithArgs(hospitalID, nationalID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	mock.ExpectQuery("INSERT INTO patients").
		WillReturnRows(patientRows().AddRow(
			newID.String(), hospitalID.String(), nil, nationalID, nil,
			firstName, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil,
			nil, now, now, now,
		))

	repo := NewPostgresPatientRepository(db)
	got, err := repo.Upsert(context.Background(), domain.Patient{
		HospitalID: hospitalID,
		NationalID: &nationalID,
		FirstName:  &firstName,
	})
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if got.ID != newID {
		t.Errorf("Upsert() ID = %v, want %v", got.ID, newID)
	}
}

func TestPostgresPatientRepository_Upsert_UpdatesWhenFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	hospitalID := uuid.New()
	existingID := uuid.New()
	now := time.Now()
	nationalID := "1234567890123"

	mock.ExpectQuery("SELECT id FROM patients WHERE hospital_id = \\$1 AND national_id = \\$2").
		WithArgs(hospitalID, nationalID).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(existingID.String()))

	mock.ExpectQuery("UPDATE patients SET").
		WillReturnRows(patientRows().AddRow(
			existingID.String(), hospitalID.String(), nil, nationalID, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil,
			nil, now, now, now,
		))

	repo := NewPostgresPatientRepository(db)
	got, err := repo.Upsert(context.Background(), domain.Patient{
		HospitalID: hospitalID,
		NationalID: &nationalID,
	})
	if err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}
	if got.ID != existingID {
		t.Errorf("Upsert() ID = %v, want existing %v", got.ID, existingID)
	}
}

func TestPostgresPatientRepository_Upsert_RequiresIdentifier(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	repo := NewPostgresPatientRepository(db)
	_, err = repo.Upsert(context.Background(), domain.Patient{HospitalID: uuid.New()})
	if err == nil {
		t.Fatal("Upsert() expected error when neither national_id nor passport_id is set")
	}
}

func TestPostgresPatientRepository_Search_AlwaysScopesByHospitalID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	hospitalID := uuid.New()
	mock.ExpectQuery("SELECT .* FROM patients WHERE hospital_id = \\$1").
		WithArgs(hospitalID).
		WillReturnRows(patientRows())

	repo := NewPostgresPatientRepository(db)
	results, err := repo.Search(context.Background(), hospitalID, PatientFilter{})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 0 {
		t.Errorf("Search() results = %v, want empty", results)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostgresPatientRepository_Search_AppliesNameFilterWithPartialMatch(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	hospitalID := uuid.New()
	lastName := "Jaidee"
	mock.ExpectQuery("SELECT .* FROM patients WHERE hospital_id = \\$1 AND last_name ILIKE \\$2").
		WithArgs(hospitalID, "%"+lastName+"%").
		WillReturnRows(patientRows().AddRow(
			uuid.New().String(), hospitalID.String(), nil, nil, nil,
			nil, nil, lastName,
			nil, nil, nil,
			nil, nil, nil,
			nil, nil, nil, nil,
			nil, nil, time.Now(), time.Now(),
		))

	repo := NewPostgresPatientRepository(db)
	results, err := repo.Search(context.Background(), hospitalID, PatientFilter{LastName: &lastName})
	if err != nil {
		t.Fatalf("Search() error = %v", err)
	}
	if len(results) != 1 || results[0].LastName == nil || *results[0].LastName != lastName {
		t.Errorf("Search() results = %+v", results)
	}
}
