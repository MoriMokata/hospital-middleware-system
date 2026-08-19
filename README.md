# hospital-middleware-system

Hospital Middleware API for the Agnos back-end candidate assignment. Lets authenticated hospital staff search patient information, backed by each hospital's own Hospital Information System (HIS) — normalized behind a common data model so multiple hospitals with different HIS response shapes can be supported.

## Tech stack

Go · Gin · Docker · Nginx · Postgres

## Status

🚧 In progress. Currently: **Task 01 — initial project skeleton** (no HTTP server or business logic wired up yet; see `cmd/api/main.go` and the placeholder packages under `internal/`).

## Running locally

Docker Compose setup (nginx + api + postgres) is added in Task 02. Until then:

```sh
go build ./...
go run ./cmd/api
```
