# hospital-middleware-system

Hospital Middleware API for the Agnos back-end candidate assignment. Lets authenticated hospital staff search patient information, backed by each hospital's own Hospital Information System (HIS) — normalized behind a common data model so multiple hospitals with different HIS response shapes can be supported.

## Tech stack

Go · Gin · Docker · Nginx · Postgres

## Status

🚧 In progress. Currently: **Task 02 — docker compose skeleton** (nginx + api + postgres stack runs, `/health` is wired up end to end; no business logic yet).

## Running locally

### With Docker Compose (nginx + api + postgres)

```sh
docker compose up --build
curl http://localhost:8080/health
# {"status":"ok"}
```

Postgres and the Go service are only reachable through nginx — nothing else is published to the host.

### Without Docker

```sh
go mod download
go run ./cmd/api
curl http://localhost:8080/health
```

### Tests

```sh
go test ./...
```
