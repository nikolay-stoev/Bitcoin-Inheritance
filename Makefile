# Bitcoin Inheritance Protocol Makefile

.PHONY: build test clean run-generate run-owner run-inheritor help

# Default target
.DEFAULT_GOAL := help

# Variables
BINARY_NAME=bitcoin-inheritance
BUILD_DIR=build
SOURCE_FILES=$(shell find . -type f -name '*.go' | grep -v vendor)

# Build the binary
build: $(BUILD_DIR)/$(BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME): $(SOURCE_FILES)
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) .

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	go clean

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy
	go mod download

# Generate a new contract (testnet)
run-generate: build
	@echo "Generating new inheritance contract..."
	./$(BUILD_DIR)/$(BINARY_NAME) generate --testnet --timelock-days 180

# Run owner withdrawal (placeholder)
run-owner: build
	@echo "Running owner withdrawal..."
	./$(BUILD_DIR)/$(BINARY_NAME) owner-withdraw --testnet

# Run inheritor withdrawal (placeholder)
run-inheritor: build
	@echo "Running inheritor withdrawal..."
	./$(BUILD_DIR)/$(BINARY_NAME) inheritor-withdraw --testnet

# Run with custom timelock
run-custom-timelock: build
	@echo "Generating contract with 30-day timelock..."
	./$(BUILD_DIR)/$(BINARY_NAME) generate --testnet --timelock-days 30

# Check code formatting
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run static analysis
vet:
	@echo "Running go vet..."
	go vet ./...

# Run all checks
check: fmt vet test

# Development setup
dev-setup:
	@echo "Setting up development environment..."
	go install github.com/btcsuite/btcd/cmd/btcd@latest
	go install github.com/btcsuite/btcd/cmd/btcctl@latest

# Help
help:
	@echo "Bitcoin Inheritance Protocol - Available Commands:"
	@echo ""
	@echo "  build            - Build the binary"
	@echo "  test             - Run tests"
	@echo "  clean            - Clean build artifacts"
	@echo "  deps             - Install dependencies"
	@echo "  run-generate     - Generate new contract (testnet)"
	@echo "  run-owner        - Run owner withdrawal (placeholder)"
	@echo "  run-inheritor    - Run inheritor withdrawal (placeholder)"
	@echo "  run-custom-timelock - Generate contract with 30-day timelock"
	@echo "  fmt              - Format code"
	@echo "  vet              - Run static analysis"
	@echo "  check            - Run all checks (fmt, vet, test)"
	@echo "  dev-setup        - Install btcd and btcctl"
	@echo "  help             - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make run-generate"
	@echo "  make test"
