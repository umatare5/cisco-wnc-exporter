.PHONY: help build lint test-unit test-unit-coverage clean

# Binary name and paths
BINARY_NAME := cisco-wnc-exporter
BUILD_DIR := ./tmp
BINARY_PATH := $(BUILD_DIR)/$(BINARY_NAME)
COVERAGE_DIR := ./coverage

# Go build flags
LDFLAGS := -X github.com/umatare5/cisco-wnc-exporter/internal/cli.version=$(shell cat VERSION)
BUILD_FLAGS := -ldflags "$(LDFLAGS)"

# Default target
.DEFAULT_GOAL := help

# Show available targets
help:
	@echo "Available targets:"
	@echo "  build              - Build the binary"
	@echo "  lint               - Run linters (golangci-lint)"
	@echo "  test-unit          - Run unit tests with colored output"
	@echo "  test-unit-coverage      - Generate HTML coverage report"
	@echo "  clean              - Remove build artifacts and backup files"
	@echo "  image              - Build Docker image"
	@echo ""
	@echo "Requirements:"
	@echo "  - gotestsum: go install gotest.tools/gotestsum@latest"
	@echo "  - golangci-lint: https://golangci-lint.run/usage/install/"

build: $(BINARY_PATH)

# Build the binary
$(BINARY_PATH):
	mkdir -p $(BUILD_DIR)
	go build $(BUILD_FLAGS) -o $(BINARY_PATH) ./cmd

# Lint the code
lint:
	golangci-lint run
	go mod tidy

# Run unit tests with gotestsum (shows individual test results with color)
test-unit:
	@command -v gotestsum >/dev/null 2>&1 || { echo "Error: gotestsum is not installed. Run: go install gotest.tools/gotestsum@latest"; exit 1; }
	mkdir -p $(COVERAGE_DIR)
	gotestsum --format testname -- -coverprofile=$(COVERAGE_DIR)/report.out ./...

# Generate coverage report (HTML)
test-unit-coverage: test-unit
	go tool cover -html=$(COVERAGE_DIR)/report.out -o $(COVERAGE_DIR)/report.html
	@echo "Coverage report generated: $(COVERAGE_DIR)/report.html"

# Clean build artifacts and backup files
clean:
	rm -rf $(BUILD_DIR) $(COVERAGE_DIR)
	find . -name "*.bak*" -type f -delete 2>/dev/null || true

# Docker targets
image:
	docker build -t ${USER}/cisco-wnc-exporter .
