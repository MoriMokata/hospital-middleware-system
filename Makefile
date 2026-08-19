.PHONY: migrate test build

migrate:
	go run ./cmd/migrate

test:
	go test ./...

build:
	go build ./...
