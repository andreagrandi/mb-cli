.PHONY: build test test-verbose fmt vet lint clean deps build-all help

BINARY_NAME := mb-cli
BUILD_DIR := bin
MAIN_PKG := ./cmd/mb-cli
MODULE := github.com/andreagrandi/mb-cli

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PKG)

test:
	@go test ./tests/ ./internal/...

test-verbose:
	@go test -v ./tests/ ./internal/...

fmt:
	@gofmt -s -w .

vet:
	@go vet ./...

lint:
	@golangci-lint run

clean:
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html

deps:
	@go mod download
	@go mod tidy

build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PKG)
	@GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PKG)
	@GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PKG)
	@GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PKG)
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PKG)

help:
	@echo "Available targets:"
	@echo "  build        - Build the binary to bin/"
	@echo "  test         - Run all tests"
	@echo "  test-verbose - Run tests with verbose output"
	@echo "  fmt          - Format code with gofmt"
	@echo "  vet          - Static analysis with go vet"
	@echo "  lint         - Run golangci-lint"
	@echo "  clean        - Remove build artifacts"
	@echo "  deps         - Download and tidy dependencies"
	@echo "  build-all    - Cross-platform builds"
	@echo "  help         - Show this help"
