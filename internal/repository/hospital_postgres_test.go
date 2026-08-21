package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestPostgresHospitalRepository_FindBySlug_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	id := uuid.New()
	now := time.Now()
	mock.ExpectQuery("SELECT id, name, slug, his_adapter_type, his_base_url").
		WithArgs("hospital-a").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug", "his_adapter_type", "his_base_url", "created_at", "updated_at"}).
			AddRow(id.String(), "Hospital A", "hospital-a", "hospital_a", "https://hospital-a.api.co.th", now, now))

	repo := NewPostgresHospitalRepository(db)
	h, err := repo.FindBySlug(context.Background(), "hospital-a")
	if err != nil {
		t.Fatalf("FindBySlug() error = %v", err)
	}
	if h.ID != id || h.Slug != "hospital-a" || h.HISAdapterType != "hospital_a" {
		t.Errorf("FindBySlug() = %+v", h)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostgresHospitalRepository_FindBySlug_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, name, slug, his_adapter_type, his_base_url").
		WithArgs("no-such-hospital").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug", "his_adapter_type", "his_base_url", "created_at", "updated_at"}))

	repo := NewPostgresHospitalRepository(db)
	_, err = repo.FindBySlug(context.Background(), "no-such-hospital")
	if err != ErrNotFound {
		t.Fatalf("FindBySlug() error = %v, want ErrNotFound", err)
	}
}

func TestPostgresHospitalRepository_FindBySlug_NullHISBaseURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	id := uuid.New()
	now := time.Now()
	mock.ExpectQuery("SELECT id, name, slug, his_adapter_type, his_base_url").
		WithArgs("hospital-b").
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug", "his_adapter_type", "his_base_url", "created_at", "updated_at"}).
			AddRow(id.String(), "Hospital B", "hospital-b", "hospital_b", nil, now, now))

	repo := NewPostgresHospitalRepository(db)
	h, err := repo.FindBySlug(context.Background(), "hospital-b")
	if err != nil {
		t.Fatalf("FindBySlug() error = %v, want no error for a NULL his_base_url", err)
	}
	if h.HISBaseURL != "" {
		t.Errorf("HISBaseURL = %q, want empty string for NULL", h.HISBaseURL)
	}
}

func TestPostgresHospitalRepository_FindByID_Found(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	id := uuid.New()
	now := time.Now()
	mock.ExpectQuery("SELECT id, name, slug, his_adapter_type, his_base_url").
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug", "his_adapter_type", "his_base_url", "created_at", "updated_at"}).
			AddRow(id.String(), "Hospital A", "hospital-a", "hospital_a", "https://hospital-a.api.co.th", now, now))

	repo := NewPostgresHospitalRepository(db)
	h, err := repo.FindByID(context.Background(), id)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if h.ID != id || h.Slug != "hospital-a" {
		t.Errorf("FindByID() = %+v", h)
	}
}

func TestPostgresHospitalRepository_FindByID_NotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	defer db.Close()

	id := uuid.New()
	mock.ExpectQuery("SELECT id, name, slug, his_adapter_type, his_base_url").
		WithArgs(id).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "slug", "his_adapter_type", "his_base_url", "created_at", "updated_at"}))

	repo := NewPostgresHospitalRepository(db)
	_, err = repo.FindByID(context.Background(), id)
	if err != ErrNotFound {
		t.Fatalf("FindByID() error = %v, want ErrNotFound", err)
	}
}
