# Install linter: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
.PHONY: lint build api manager migrate

lint:
	golangci-lint run ./...

build:
	go build ./...

api:
	go run ./cmd/api/cmd/main.go

manager:
	go run ./cmd/manager/cmd/main.go

migrate:
	go run ./cmd/database/main.go