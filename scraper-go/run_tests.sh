#!/bin/bash

# Run tests with coverage

echo "Running unit tests with coverage..."

# Set up test environment
export AWS_PROFILE=framed

# Run tests with coverage
go test -v -coverprofile=coverage.out -covermode=atomic ./...

# Generate coverage report
echo -e "\nGenerating coverage report..."
go tool cover -html=coverage.out -o coverage.html

# Show coverage summary
echo -e "\nCoverage summary:"
go tool cover -func=coverage.out | grep total

# Run race detector tests
echo -e "\nRunning tests with race detector..."
go test -race ./...

# Run short tests (without AWS integration)
echo -e "\nRunning short tests (no AWS integration)..."
go test -short ./...

echo -e "\nTest run complete!"