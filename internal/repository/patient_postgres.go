package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"github.com/MoriMokata/hospital-middleware-system/internal/domain"
)

const patientColumns = `
	id, hospital_id, patient_hn, national_id, passport_id,
	first_name, middle_name, last_name,
	first_name_th, middle_name_th, last_name_th,
	first_name_en, middle_name_en, last_name_en,
	date_of_birth, phone_number, email, gender,
	raw_source, synced_at, created_at, updated_at`

// PostgresPatientRepository implements PatientRepository against Postgres.
type PostgresPatientRepository struct {
	DB *sql.DB
}

func NewPostgresPatientRepository(db *sql.DB) *PostgresPatientRepository {
	return &PostgresPatientRepository{DB: db}
}

func scanPatient(row *sql.Row) (domain.Patient, error) {
	var p domain.Patient
	var raw []byte
	err := row.Scan(
		&p.ID, &p.HospitalID, &p.PatientHN, &p.NationalID, &p.PassportID,
		&p.FirstName, &p.MiddleName, &p.LastName,
		&p.FirstNameTH, &p.MiddleNameTH, &p.LastNameTH,
		&p.FirstNameEN, &p.MiddleNameEN, &p.LastNameEN,
		&p.DateOfBirth, &p.PhoneNumber, &p.Email, &p.Gender,
		&raw, &p.SyncedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if raw != nil {
		p.RawSource = json.RawMessage(raw)
	}
	return p, err
}

// Upsert matches an existing patient by hospitalID + whichever of
// NationalID/PassportID is set on the given patient, updating it if found
// or inserting a new row otherwise. At least one of NationalID/PassportID
// must be set — Upsert is only meaningful for records synced from a HIS
// id lookup.
func (r *PostgresPatientRepository) Upsert(ctx context.Context, patient domain.Patient) (domain.Patient, error) {
	var (
		matchCol string
		matchVal string
	)
	switch {
	case patient.NationalID != nil && *patient.NationalID != "":
		matchCol, matchVal = "national_id", *patient.NationalID
	case patient.PassportID != nil && *patient.PassportID != "":
		matchCol, matchVal = "passport_id", *patient.PassportID
	default:
		return domain.Patient{}, fmt.Errorf("upsert patient: national_id or passport_id is required")
	}

	existingID, err := r.findIDByHospitalAndColumn(ctx, patient.HospitalID, matchCol, matchVal)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return domain.Patient{}, err
	}

	if errors.Is(err, ErrNotFound) {
		return r.insert(ctx, patient)
	}
	patient.ID = existingID
	return r.update(ctx, patient)
}

func (r *PostgresPatientRepository) findIDByHospitalAndColumn(ctx context.Context, hospitalID uuid.UUID, column, value string) (uuid.UUID, error) {
	q := fmt.Sprintf(`SELECT id FROM patients WHERE hospital_id = $1 AND %s = $2`, column)
	var id uuid.UUID
	err := r.DB.QueryRowContext(ctx, q, hospitalID, value).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.UUID{}, ErrNotFound
	}
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("find patient by %s: %w", column, err)
	}
	return id, nil
}

func (r *PostgresPatientRepository) insert(ctx context.Context, p domain.Patient) (domain.Patient, error) {
	q := `
		INSERT INTO patients (
			hospital_id, patient_hn, national_id, passport_id,
			first_name, middle_name, last_name,
			first_name_th, middle_name_th, last_name_th,
			first_name_en, middle_name_en, last_name_en,
			date_of_birth, phone_number, email, gender,
			raw_source, synced_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, now()
		) RETURNING ` + patientColumns

	row := r.DB.QueryRowContext(ctx, q,
		p.HospitalID, p.PatientHN, p.NationalID, p.PassportID,
		p.FirstName, p.MiddleName, p.LastName,
		p.FirstNameTH, p.MiddleNameTH, p.LastNameTH,
		p.FirstNameEN, p.MiddleNameEN, p.LastNameEN,
		p.DateOfBirth, p.PhoneNumber, p.Email, p.Gender,
		p.RawSource,
	)
	patient, err := scanPatient(row)
	if err != nil {
		return domain.Patient{}, fmt.Errorf("insert patient: %w", err)
	}
	return patient, nil
}

func (r *PostgresPatientRepository) update(ctx context.Context, p domain.Patient) (domain.Patient, error) {
	q := `
		UPDATE patients SET
			patient_hn = $2, national_id = $3, passport_id = $4,
			first_name = $5, middle_name = $6, last_name = $7,
			first_name_th = $8, middle_name_th = $9, last_name_th = $10,
			first_name_en = $11, middle_name_en = $12, last_name_en = $13,
			date_of_birth = $14, phone_number = $15, email = $16, gender = $17,
			raw_source = $18, synced_at = now(), updated_at = now()
		WHERE id = $1
		RETURNING ` + patientColumns

	row := r.DB.QueryRowContext(ctx, q,
		p.ID, p.PatientHN, p.NationalID, p.PassportID,
		p.FirstName, p.MiddleName, p.LastName,
		p.FirstNameTH, p.MiddleNameTH, p.LastNameTH,
		p.FirstNameEN, p.MiddleNameEN, p.LastNameEN,
		p.DateOfBirth, p.PhoneNumber, p.Email, p.Gender,
		p.RawSource,
	)
	patient, err := scanPatient(row)
	if err != nil {
		return domain.Patient{}, fmt.Errorf("update patient: %w", err)
	}
	return patient, nil
}

// Search scopes results to hospitalID (the isolation boundary — always
// sourced from the caller's JWT, never client input) and ANDs in whatever
// filter fields are set. String fields match partial/case-insensitive.
func (r *PostgresPatientRepository) Search(ctx context.Context, hospitalID uuid.UUID, filter PatientFilter) ([]domain.Patient, error) {
	q := strings.Builder{}
	q.WriteString("SELECT " + patientColumns + " FROM patients WHERE hospital_id = $1")

	args := []any{hospitalID}
	addLike := func(column string, value *string) {
		if value == nil || *value == "" {
			return
		}
		args = append(args, "%"+*value+"%")
		fmt.Fprintf(&q, " AND %s ILIKE $%d", column, len(args))
	}

	addLike("national_id", filter.NationalID)
	addLike("passport_id", filter.PassportID)
	addLike("first_name", filter.FirstName)
	addLike("middle_name", filter.MiddleName)
	addLike("last_name", filter.LastName)
	addLike("phone_number", filter.PhoneNumber)
	addLike("email", filter.Email)

	if filter.DateOfBirth != nil {
		args = append(args, *filter.DateOfBirth)
		fmt.Fprintf(&q, " AND date_of_birth = $%d", len(args))
	}

	q.WriteString(" ORDER BY created_at DESC")

	rows, err := r.DB.QueryContext(ctx, q.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("search patients: %w", err)
	}
	defer rows.Close()

	var results []domain.Patient
	for rows.Next() {
		var p domain.Patient
		var raw []byte
		if err := rows.Scan(
			&p.ID, &p.HospitalID, &p.PatientHN, &p.NationalID, &p.PassportID,
			&p.FirstName, &p.MiddleName, &p.LastName,
			&p.FirstNameTH, &p.MiddleNameTH, &p.LastNameTH,
			&p.FirstNameEN, &p.MiddleNameEN, &p.LastNameEN,
			&p.DateOfBirth, &p.PhoneNumber, &p.Email, &p.Gender,
			&raw, &p.SyncedAt, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan patient row: %w", err)
		}
		if raw != nil {
			p.RawSource = json.RawMessage(raw)
		}
		results = append(results, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("search patients: %w", err)
	}
	return results, nil
}
