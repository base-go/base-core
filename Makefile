# Makefile for Base Go Application - Rails-style test commands

.PHONY: test test-unit test-integration test-features test-models test-services test-coverage clean

# Default test command (runs all tests)
test:
	@echo "Running all tests..."
	go test ./test/... -v

# Model tests
test-models:
	@echo "Running model tests..."
	go test ./test/test_authentication/models_test.go ./test/test_profile/models_test.go ./test/test_helper.go -v

# Integration tests
test-integration:
	@echo "Running integration tests..."
	go test ./test/integration/... -v

# Simple tests
test-simple:
	@echo "Running simple tests..."
	go test ./test/simple_test.go ./test/test_helper.go -v

# Authentication module tests
test-authentication:
	@echo "Running authentication module tests..."
	go test ./test/test_authentication/... -v

# Profile module tests
test-profile:
	@echo "Running profile module tests..."
	go test ./test/test_profile/... -v

# Service tests
test-services:
	@echo "Running service tests..."
	go test ./test/authentication/service_test.go ./test/profile/service_test.go ./test/test_helper.go -v

# Test with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test ./test/... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Test specific package
test-package:
	@echo "Usage: make test-package PACKAGE=./test/features"
	go test $(PACKAGE) -v

# Clean test artifacts
clean:
	rm -f coverage.out coverage.html

# Setup test database (if needed)
test-setup:
	@echo "Setting up test environment..."
	go mod tidy

# Watch tests (requires entr or similar)
test-watch:
	find . -name "*.go" | entr -c make test

# Rails-style test commands
test-all: test
test-fast: test-unit
test-slow: test-features

# Help
help:
	@echo "Available test commands:"
	@echo "  make test          - Run all tests"
	@echo "  make test-unit     - Run unit tests (models, services)"
	@echo "  make test-features - Run feature/integration tests"
	@echo "  make test-models   - Run model tests only"
	@echo "  make test-services - Run service tests only"
	@echo "  make test-coverage - Run tests with coverage report"
	@echo "  make clean         - Clean test artifacts"
	@echo "  make test-setup    - Setup test environment"
