.PHONY: build test lint sqlc

build:
	go build ./...

test:
	go test ./... -v

lint:
	go vet ./...

sqlc:
	sqlc generate
