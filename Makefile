# Install linter: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
.PHONY: lint build api manager

lint:
	golangci-lint run ./...

build:
	go build ./...

api:
	go run ./services/api/cmd/main.go

manager:
	go run ./services/manager/cmd/main.go