.PHONY: test coverage coverage-html generate lint build clean help

# Default target
.DEFAULT_GOAL := help

## help: Show this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## test: Run all tests
test:
	@echo "Running tests..."
	@go test ./... -v

## coverage: Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out | tail -1

## coverage-html: Generate HTML coverage report
coverage-html: coverage
	@echo "Generating HTML coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## generate: Generate mocks
generate:
	@echo "Generating mocks..."
	@go generate ./...

## lint: Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run --timeout=5m

## build: Build the binary
build:
	@echo "Building..."
	@go build -o promener

## clean: Clean generated files
clean:
	@echo "Cleaning..."
	@rm -f promener coverage.out coverage.html
	@find . -name "mock_*.go" -type f -delete

## install-tools: Install development tools
install-tools:
	@echo "Installing development tools..."
	@go install go.uber.org/mock/mockgen@latest
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin

## ci: Run CI checks (generate, test)
ci: generate test
	@echo "CI checks passed!"
