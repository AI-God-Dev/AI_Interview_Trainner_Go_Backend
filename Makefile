.PHONY: build run test lint clean docker-build docker-run

build:
	@go build -o bin/server .

run:
	@go run .

test:
	@go test -v -race -coverprofile=coverage.out ./pkg/... ./app/services/...

test-integration:
	@go test -v -tags=integration -coverprofile=integration.out ./...

test-all: test test-integration

lint:
	@golangci-lint run

format:
	@go fmt ./...
	@goimports -w .

clean:
	@rm -rf bin/
	@go clean

docker-build:
	@docker build -t ai-interview-trainer:latest .

docker-run:
	@docker-compose up

deps:
	@go mod download
	@go mod tidy

swagger:
	@swag init

dev:
	@air

