// Command migrate applies pending SQL migrations to the database at
// DB_DSN. Usage: go run ./cmd/migrate
package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/MoriMokata/hospital-middleware-system/migrations"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN is required")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := migrations.Up(context.Background(), db); err != nil {
		log.Fatalf("migrate up: %v", err)
	}
	log.Println("migrations applied")
}
