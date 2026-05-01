.PHONY: test test-verbose test-cover clean

PACKAGES := $(shell go list ./...)

test:
	@echo "Running tests in all packages..."
	@go test ./...

test-cover:
	@echo "Running tests with coverage..."
	@go test -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

test-race:
	@echo "Running tests with race detector..."
	@go test -race ./...

clean:
	@rm -f coverage.out coverage.html
	@echo "Cleaned up coverage files"