.PHONY: build clean test install help

# Build the aidu-manager binary
build:
	@echo "Building aidu-manager..."
	@mkdir -p bin
	@go build -o bin/aidu-manager ./cmd/aidu-manager
	@echo "Build complete: bin/aidu-manager"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run ./... || echo "Note: golangci-lint not installed"

# Build and install to PATH
install: build
	@echo "Installing to /usr/local/bin..."
	@sudo cp bin/aidu-manager /usr/local/bin/
	@echo "Installed: aidu-manager is now in your PATH"

# Quick orchestration example
example:
	@./aidu-session orchestrate "what is 2+2"

# Show help
help:
	@echo "AIDU Manager Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build      - Build the aidu-manager binary"
	@echo "  make clean      - Remove build artifacts"
	@echo "  make test       - Run tests"
	@echo "  make deps       - Install/update dependencies"
	@echo "  make fmt        - Format Go code"
	@echo "  make lint       - Run linter"
	@echo "  make install    - Build and install to /usr/local/bin"
	@echo "  make example    - Run a simple orchestration example"
	@echo "  make help       - Show this help message"
	@echo ""
	@echo "Quick Start:"
	@echo "  1. make build"
	@echo "  2. ./aidu-session orchestrate \"your task here\""
	@echo ""
	@echo "For more information, see QUICKSTART.md"
