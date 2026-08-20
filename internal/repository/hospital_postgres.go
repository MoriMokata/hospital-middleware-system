package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

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

	var h domain.Hospital
	err := r.DB.QueryRowContext(ctx, q, slug).Scan(
		&h.ID, &h.Name, &h.Slug, &h.HISAdapterType, &h.HISBaseURL, &h.CreatedAt, &h.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Hospital{}, ErrNotFound
	}
	if err != nil {
		return domain.Hospital{}, fmt.Errorf("find hospital by slug: %w", err)
	}
	return h, nil
}
