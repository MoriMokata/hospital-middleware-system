package migrations

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}

func TestUp_AppliesPendingMigrationsInOrder(t *testing.T) {
	db, mock := newMockDB(t)

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").WillReturnResult(sqlmock.NewResult(0, 0))

	for _, name := range []string{"0001_init_schema.sql", "0002_seed_hospital_a.sql"} {
		mock.ExpectQuery("SELECT EXISTS").WithArgs(name).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
		mock.ExpectBegin()
		mock.ExpectExec(".*").WillReturnResult(sqlmock.NewResult(0, 0))
		mock.ExpectExec("INSERT INTO schema_migrations").WithArgs(name).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()
	}

	if err := Up(context.Background(), db); err != nil {
		t.Fatalf("Up() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUp_SkipsAlreadyAppliedMigrations(t *testing.T) {
	db, mock := newMockDB(t)

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").WillReturnResult(sqlmock.NewResult(0, 0))

	for _, name := range []string{"0001_init_schema.sql", "0002_seed_hospital_a.sql"} {
		mock.ExpectQuery("SELECT EXISTS").WithArgs(name).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))
	}

	if err := Up(context.Background(), db); err != nil {
		t.Fatalf("Up() error = %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations (no migration should have been applied): %v", err)
	}
}

func TestUp_RollsBackAndReturnsErrorOnApplyFailure(t *testing.T) {
	db, mock := newMockDB(t)

	mock.ExpectExec("CREATE TABLE IF NOT EXISTS schema_migrations").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectQuery("SELECT EXISTS").WithArgs("0001_init_schema.sql").
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))
	mock.ExpectBegin()
	mock.ExpectExec(".*").WillReturnError(errors.New("syntax error"))
	mock.ExpectRollback()

	err := Up(context.Background(), db)
	if err == nil {
		t.Fatal("Up() expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
