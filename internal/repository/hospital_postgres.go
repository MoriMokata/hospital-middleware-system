package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/MoriMokata/hospital-middleware-system/internal/domain"
)

// PostgresHospitalRepository implements HospitalRepository against Postgres.
type PostgresHospitalRepository struct {
	DB *sql.DB
}

func NewPostgresHospitalRepository(db *sql.DB) *PostgresHospitalRepository {
	return &PostgresHospitalRepository{DB: db}
}

func (r *PostgresHospitalRepository) FindBySlug(ctx context.Context, slug string) (domain.Hospital, error) {
	const q = `
		SELECT id, name, slug, his_adapter_type, his_base_url, created_at, updated_at
		FROM hospitals
		WHERE slug = $1`

	h, err := scanHospital(r.DB.QueryRowContext(ctx, q, slug))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Hospital{}, ErrNotFound
	}
	if err != nil {
		return domain.Hospital{}, fmt.Errorf("find hospital by slug: %w", err)
	}
	return h, nil
}

func (r *PostgresHospitalRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.Hospital, error) {
	const q = `
		SELECT id, name, slug, his_adapter_type, his_base_url, created_at, updated_at
		FROM hospitals
		WHERE id = $1`

	h, err := scanHospital(r.DB.QueryRowContext(ctx, q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Hospital{}, ErrNotFound
	}
	if err != nil {
		return domain.Hospital{}, fmt.Errorf("find hospital by id: %w", err)
	}
	return h, nil
}

// scanHospital handles his_base_url being NULL (a hospital not yet wired
// up to any HIS) — the column is nullable in the schema, but
// domain.Hospital.HISBaseURL is a plain string, so NULL maps to "".
func scanHospital(row *sql.Row) (domain.Hospital, error) {
	var h domain.Hospital
	var hisBaseURL sql.NullString
	err := row.Scan(
		&h.ID, &h.Name, &h.Slug, &h.HISAdapterType, &hisBaseURL, &h.CreatedAt, &h.UpdatedAt,
	)
	if err != nil {
		return domain.Hospital{}, err
	}
	h.HISBaseURL = hisBaseURL.String
	return h, nil
}
