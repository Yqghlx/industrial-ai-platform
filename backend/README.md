# industrial-ai-platform

Industrial AI Platform Backend Service

## CI/CD Status

![Test](https://github.com/industrial-ai/platform/workflows/Test/badge.svg)
![Build](https://github.com/industrial-ai/platform/workflows/Build/badge.svg)
![Code Quality](https://github.com/industrial-ai/platform/workflows/Code%20Quality/badge.svg)

## Test Coverage

Current coverage: **72.3%**

See [TEST_COVERAGE_REPORT.md](docs/TEST_COVERAGE_REPORT.md) for detailed coverage analysis.

## Development

### Prerequisites

- Go 1.23+
- PostgreSQL 16+
- Redis 8+

### Running Tests

```bash
# Unit tests
go test ./internal/handler/... ./internal/service/... ./pkg/... -v

# Integration tests (requires local PostgreSQL)
psql -U $(whoami) -d postgres -c "CREATE DATABASE test_platform;"
go test ./tests/integration/... -v

# E2E tests
go test ./tests/e2e/... -v

# All tests with coverage
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

### Code Quality

```bash
# Run linters
golangci-lint run

# Format check
gofmt -l .
```

## CI/CD Workflows

- **test.yml**: Run all tests on push/PR
- **build.yml**: Build binary on main branch
- **quality.yml**: Run linters and security scans

## Coverage Target

- Minimum coverage: **70%**
- Current coverage: **72.3%** ✅

## License

MIT