.PHONY: build clean test fmt lint install docker-build help

REGISTRY_PORT ?= 5050

build:
	@echo "Building aidu..."
	@mkdir -p bin
	@go build -o bin/aidu ./cmd/aidu
	@echo "Build complete: bin/aidu"

clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@echo "Clean complete"

test:
	@echo "Running tests..."
	@go test -v ./...

fmt:
	@echo "Formatting code..."
	@go fmt ./...

lint:
	@echo "Running linter..."
	@golangci-lint run ./... || echo "Note: golangci-lint not installed"

install: build
	@echo "Installing to /usr/local/bin..."
	@sudo cp bin/aidu /usr/local/bin/
	@echo "Installed: aidu is now in your PATH"

docker-build:
	@echo "Building Docker image..."
	@docker build -t localhost:$(REGISTRY_PORT)/aidu-claude:latest .
	@echo "Pushing to local registry..."
	@docker push localhost:$(REGISTRY_PORT)/aidu-claude:latest
	@echo "Docker image ready"

help:
	@echo "AIDU Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build        - Build the aidu binary"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make test         - Run tests"
	@echo "  make fmt          - Format Go code"
	@echo "  make lint         - Run linter"
	@echo "  make install      - Build and install to /usr/local/bin"
	@echo "  make docker-build - Build and push Docker image to local registry"
	@echo "  make help         - Show this help message"
	@echo ""
	@echo "Quick Start:"
	@echo "  1. ./setup-k3d.sh"
	@echo "  2. make build"
	@echo "  3. bin/aidu setup"
	@echo "  4. bin/aidu run \"your task here\""
