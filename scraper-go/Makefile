.PHONY: build test test-short test-race test-coverage clean run lint generate-data generate-gcp-data generate-oci-data

# Build the scraper
build:
	go build -o bin/scraper cmd/scraper/main.go

# Run all tests
test:
	AWS_PROFILE=framed go test -v ./...

# Run tests without AWS integration
test-short:
	go test -short -v ./...

# Run tests with race detector
test-race:
	AWS_PROFILE=framed go test -race ./...

# Run tests with coverage
test-coverage:
	AWS_PROFILE=framed go test -v -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@go tool cover -func=coverage.out | grep total

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	rm -rf .cache/

# Run the scraper
run: build
	AWS_PROFILE=framed ./bin/scraper

# Run with caching enabled
run-cached: build
	AWS_PROFILE=framed ./bin/scraper --use-cache

# Run with retry logic
run-retry:
	./run_with_retry.sh

# Run with progress tracking (resume from failures)
run-resume: build
	AWS_PROFILE=framed ./bin/scraper --use-cache --skip-completed

# Reset progress
reset-progress:
	rm -f .cache/scraper/progress.json

# Run linting
lint:
	golangci-lint run

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run EC2 scraper only
run-ec2: build
	AWS_PROFILE=framed ./bin/scraper -s ec2 -r us-east-1

# Run comparison test
compare:
	AWS_PROFILE=framed python3 ../compare_scrapers.py

# Generate frontend data
generate-data: generate-gcp-data generate-oci-data

generate-gcp-data:
	@echo "Generating GCP data for frontend..."
	@mkdir -p data
	@go run ./cmd/generate-gcp-data ./data
	@echo "✅ GCP data generation completed"

generate-oci-data:
	@echo "Generating OCI data for frontend..."
	@mkdir -p data
	@go run ./cmd/generate-oci-data ./data
	@echo "✅ OCI data generation completed"

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the scraper binary"
	@echo "  test          - Run all tests"
	@echo "  test-short    - Run tests without AWS integration"
	@echo "  test-race     - Run tests with race detector"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Remove build artifacts"
	@echo "  run           - Build and run the scraper"
	@echo "  run-cached    - Run with caching enabled"
	@echo "  lint          - Run linting"
	@echo "  deps          - Install/update dependencies"
	@echo "  run-ec2       - Run EC2 scraper only"
	@echo "  compare       - Run comparison with Vantage data"
	@echo "  generate-data - Generate GCP and OCI data for frontend"