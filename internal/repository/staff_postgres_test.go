package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/MoriMokata/hospital-middleware-system/internal/domain"
)

func TestPostgresStaffRepository_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	hospitalID := uuid.New()
	staffID := uuid.New()
	now := time.Now()

	mock.ExpectQuery("INSERT INTO staff").
		WithArgs(hospitalID, "somchai.p", "hashed-password").
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).AddRow(staffID.String(), now, now))

	repo := NewPostgresStaffRepository(db)
	got, err := repo.Create(context.Background(), domain.Staff{
		HospitalID:   hospitalID,
		Username:     "somchai.p",
		PasswordHash: "hashed-password",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if got.ID != staffID {
		t.Errorf("Create() ID = %v, want %v", got.ID, staffID)
	}
}

func TestPostgresStaffRepository_Create_DuplicateUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	hospitalID := uuid.New()
	mock.ExpectQuery("INSERT INTO staff").
		WithArgs(hospitalID, "somchai.p", "hashed-password").
		WillReturnError(&pgconn.PgError{Code: pgUniqueViolation})

	repo := NewPostgresStaffRepository(db)
	_, err = repo.Create(context.Background(), domain.Staff{
		HospitalID:   hospitalID,
		Username:     "somchai.p",
		PasswordHash: "hashed-password",
	})
	if err != ErrConflict {
		t.Fatalf("Create() error = %v, want ErrConflict", err)
	}
}

func TestPostgresStaffRepository_FindByHospitalAndUsername_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	hospitalID := uuid.New()
	mock.ExpectQuery("SELECT id, hospital_id, username, password_hash").
		WithArgs(hospitalID, "nobody").
		WillReturnRows(sqlmock.NewRows([]string{"id", "hospital_id", "username", "password_hash", "created_at", "updated_at"}))

	repo := NewPostgresStaffRepository(db)
	_, err = repo.FindByHospitalAndUsername(context.Background(), hospitalID, "nobody")
	if err != ErrNotFound {
		t.Fatalf("FindByHospitalAndUsername() error = %v, want ErrNotFound", err)
	}
}
