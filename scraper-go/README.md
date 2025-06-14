# Cloud Instance Scraper (Go)

A high-performance, parallel cloud instance data scraper written in Go. This scraper collects instance type information and pricing data from various cloud providers.

## Features

- **Parallel Execution**: Leverages Go's concurrency model for fast data collection
- **Multi-Service Support**: Scrapes EC2, RDS, ElastiCache, Redshift, and OpenSearch
- **Intelligent Caching**: Two-tier caching system:
  - Service-level cache (configurable TTL, default 1h)
  - Instance-level cache (24h TTL) for incremental updates
- **Pricing Data**: Fetches on-demand, reserved, and spot pricing
- **Comprehensive Testing**: Unit tests with coverage reporting
- **AWS SDK v2**: Uses the latest AWS SDK for Go

## Installation

```bash
go mod download
make build
```

## Usage

### Basic Usage

```bash
# Run all scrapers
AWS_PROFILE=your-profile ./bin/scraper

# Run specific services
AWS_PROFILE=your-profile ./bin/scraper -s ec2,rds

# Run with caching
AWS_PROFILE=your-profile ./bin/scraper --use-cache

# Run for specific regions
AWS_PROFILE=your-profile ./bin/scraper -r us-east-1,us-west-2
```

### Command Line Options

- `-s, --services`: Services to scrape (default: all)
- `-r, --regions`: Regions to scrape (default: all available)
- `-o, --output`: Output directory (default: "www")
- `-c, --concurrency`: Max concurrent operations (default: 10)
- `--use-cache`: Enable caching
- `--cache-dir`: Cache directory (default: ".cache/scraper")
- `--cache-ttl`: Cache TTL (default: "1h")

## Development

### Running Tests

```bash
# Run all tests
make test

# Run tests without AWS integration
make test-short

# Run with coverage
make test-coverage

# Run with race detector
make test-race
```

### Building

```bash
# Build binary
make build

# Clean build artifacts
make clean

# Update dependencies
make deps
```

### Project Structure

```
scraper-go/
├── cmd/scraper/          # Main application entry point
├── pkg/
│   ├── aws/              # AWS service scrapers
│   │   ├── ec2_scraper.go
│   │   ├── rds_scraper.go
│   │   ├── elasticache_scraper.go
│   │   ├── redshift_scraper.go
│   │   └── opensearch_scraper.go
│   └── models/           # Data models
├── internal/
│   └── cache/            # Caching implementation
└── tests/                # Test files
```

## Caching Strategy

The scraper implements a two-tier caching system:

1. **Service-level cache**: Caches complete scraping results per service/region combination
2. **Instance-level cache**: Caches individual instance pricing data with 24h expiration

This allows for:
- Fast subsequent runs when data hasn't changed
- Incremental updates when only some data needs refreshing
- Resilience to API throttling by reusing cached pricing data

## Performance

- Parallel region processing with configurable concurrency
- Efficient batch processing of AWS Pricing API results
- Automatic retry with exponential backoff for throttled requests
- Progress reporting for long-running operations

## Error Handling

- Graceful handling of AWS API throttling
- Automatic retries with exponential backoff
- Context-aware cancellation support
- Detailed error reporting

## Contributing

1. Ensure all tests pass: `make test`
2. Run linting: `make lint`
3. Update tests for new features
4. Follow Go best practices and conventions