.PHONY: build test lint sqlc migrate

build:
	go build ./...

test:
	go test ./... -v

lint:
	go vet ./...

sqlc:
	sqlc generate

migrate:
	goose -dir migrations sqlite3 ddd-sandbox.db up

migrate-down:
	goose -dir migrations sqlite3 ddd-sandbox.db down
