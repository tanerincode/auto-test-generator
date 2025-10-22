.PHONY: build test fmt lint clean help

# Build the autotest binary
build:
	@echo "Building autotest..."
	go build -o bin/autotest ./cmd/autotest

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	go vet ./...
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed; skipping"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run all checks
check: fmt lint test

# Help
help:
	@echo "Available targets:"
	@echo "  build   - Build the autotest binary"
	@echo "  test    - Run tests"
	@echo "  fmt     - Format code"
	@echo "  lint    - Run linter"
	@echo "  clean   - Clean build artifacts"
	@echo "  deps    - Install dependencies"
	@echo "  check   - Run all checks (fmt, lint, test)"
	@echo "  help    - Show this help message"

