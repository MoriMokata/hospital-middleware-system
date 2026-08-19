# Migrations

Plain SQL files applied in filename order by the embedded runner in
[`migrations.go`](./migrations.go) (tracked in a `schema_migrations` table,
each file applied at most once). No external migration tool/CLI is
required — see the note in `docs/design.md`.

- `0001_init_schema.sql` — `hospitals`, `staff`, `patients` tables + indexes (matches `docs/er-diagram.md`)
- `0002_seed_hospital_a.sql` — seeds the `hospital-a` row used by the Hospital A HIS adapter

## Running

```sh
DB_DSN="postgres://postgres:postgres@localhost:5432/hospital_middleware?sslmode=disable" make migrate
```

Inside docker-compose, `DB_DSN` is already set on the `api` service; the API applies pending migrations automatically on startup (wired into `cmd/api/main.go` once the DB connection is introduced in Task 05).
