# hospital-middleware-system

Hospital Middleware API for the Agnos back-end candidate assignment. Lets authenticated hospital staff search patient information, backed by each hospital's own Hospital Information System (HIS) — normalized behind a common data model so multiple hospitals with different HIS response shapes can be supported.

## Tech stack

Go · Gin · Docker · Nginx · Postgres

## Status

🚧 In progress. Currently: **Task 11 — patient search api** (`/staff/create`, `/staff/login`, `/patient/search`, and `/health` are all wired up end to end behind nginx, backed by Postgres, with a Hospital A HIS adapter).

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

### Example requests

```sh
# create a staff account
curl -X POST http://localhost:8080/staff/create \
  -H "Content-Type: application/json" \
  -d '{"username":"somchai.p","password":"P@ssw0rd123","hospital":"hospital-a"}'

# log in and grab an access token
TOKEN=$(curl -s -X POST http://localhost:8080/staff/login \
  -H "Content-Type: application/json" \
  -d '{"username":"somchai.p","password":"P@ssw0rd123","hospital":"hospital-a"}' \
  | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

# search patients (always scoped to the caller's own hospital)
curl -X POST http://localhost:8080/patient/search \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"national_id":"1234567890123"}'
```
