# MiFace Makefile
.PHONY: all build test clean install run fmt vet lint help

# Binary name
BINARY_NAME=miface
BUILD_DIR=bin

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt

# Build flags
LDFLAGS=-ldflags "-s -w" -buildvcs=false
VERSION?=0.1.0

all: test build

## build: Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/miface

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -timeout 30s ./...

## test-coverage: Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	rm -f $(BINARY_NAME) $(BINARY_NAME)_test

## install: Install the binary to $GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	$(GOCMD) install ./cmd/miface

## run: Build and run with default settings
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

## run-preview: Build and run with preview window
run-preview: build
	@echo "Running $(BINARY_NAME) with preview..."
	./$(BUILD_DIR)/$(BINARY_NAME) -preview

## fmt: Format all Go files
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOCMD) vet ./...

## lint: Run golangci-lint (requires golangci-lint installed)
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run

## tidy: Tidy and verify dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	$(GOMOD) verify

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download

## help: Show this help message
help:
	@echo "MiFace - Makefile commands:"
	@echo ""
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'
