.PHONY: run build generate migrate-up migrate-down test test-short fmt lint

run:
	go run ./cmd/api

build:
	go build -o bin/prism ./cmd/api

generate:
	sqlc generate

migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1

test:
	go test ./...

test-short:
	go test ./... -short

fmt:
	gofmt -w .

lint:
	go vet ./...
