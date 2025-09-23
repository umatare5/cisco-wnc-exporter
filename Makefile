.PHONY: build lint test test-unit clean

# Binary name and paths
BINARY_NAME := cisco-wnc-exporter
BUILD_DIR := ./tmp
BINARY_PATH := $(BUILD_DIR)/$(BINARY_NAME)

# Go build flags
LDFLAGS := -X github.com/umatare5/cisco-wnc-exporter/internal/cli.version=$(shell cat VERSION)
BUILD_FLAGS := -ldflags "$(LDFLAGS)"

# Default target
build: $(BINARY_PATH)

# Build the binary
$(BINARY_PATH):
	mkdir -p $(BUILD_DIR)
	go build $(BUILD_FLAGS) -o $(BINARY_PATH) ./cmd

# Lint the code
lint:
	golangci-lint run
	go mod tidy

# Run unit tests
test-unit:
	go test -v ./...

# Run all tests (alias)
test: test-unit

# Clean build artifacts and backup files
clean:
	rm -rf $(BUILD_DIR)
	find . -name "*.bak*" -type f -delete 2>/dev/null || true

# Docker targets
image:
	docker build -t ${USER}/cisco-wnc-exporter .

force-image:
	docker build --no-cache -t ${USER}/cisco-wnc-exporter .
