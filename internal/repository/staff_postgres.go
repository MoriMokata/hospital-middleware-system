package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/MoriMokata/hospital-middleware-system/internal/domain"
)

const pgUniqueViolation = "23505"

// PostgresStaffRepository implements StaffRepository against Postgres.
type PostgresStaffRepository struct {
	DB *sql.DB
}

func NewPostgresStaffRepository(db *sql.DB) *PostgresStaffRepository {
	return &PostgresStaffRepository{DB: db}
}

func (r *PostgresStaffRepository) Create(ctx context.Context, staff domain.Staff) (domain.Staff, error) {
	const q = `
		INSERT INTO staff (hospital_id, username, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at`

	err := r.DB.QueryRowContext(ctx, q, staff.HospitalID, staff.Username, staff.PasswordHash).
		Scan(&staff.ID, &staff.CreatedAt, &staff.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.Staff{}, ErrConflict
		}
		return domain.Staff{}, fmt.Errorf("create staff: %w", err)
	}
	return staff, nil
}

func (r *PostgresStaffRepository) FindByHospitalAndUsername(ctx context.Context, hospitalID uuid.UUID, username string) (domain.Staff, error) {
	const q = `
		SELECT id, hospital_id, username, password_hash, created_at, updated_at
		FROM staff
		WHERE hospital_id = $1 AND username = $2`

	var s domain.Staff
	err := r.DB.QueryRowContext(ctx, q, hospitalID, username).
		Scan(&s.ID, &s.HospitalID, &s.Username, &s.PasswordHash, &s.CreatedAt, &s.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Staff{}, ErrNotFound
	}
	if err != nil {
		return domain.Staff{}, fmt.Errorf("find staff by hospital and username: %w", err)
	}
	return s, nil
}
