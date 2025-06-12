# Makefile for Weather API tests

.PHONY: test test-unit test-integration test-coverage test-race test-bench clean help

# Default target
help:
	@echo "Available targets:"
	@echo "  test           - Run all tests"
	@echo "  test-unit      - Run unit tests only"
	@echo "  test-integration - Run integration tests (requires database)"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-race      - Run tests with race detection"
	@echo "  test-bench     - Run benchmark tests"
	@echo "  test-verbose   - Run tests with verbose output"
	@echo "  clean          - Clean test cache"
	@echo "  help           - Show this help message"

# Run all tests
test:
	@echo "Running all tests..."
	go test -v ./...

# Run unit tests only (excluding integration tests)
test-unit:
	@echo "Running unit tests..."
	go test -v -short ./...

# Run integration tests (requires database connection)
test-integration:
	@echo "Running integration tests..."
	@echo "Make sure DATABASE_URL is set for integration tests"
	go test -v -run Integration ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	go test -v -race ./...

# Run benchmark tests
test-bench:
	@echo "Running benchmark tests..."
	go test -v -bench=. -benchmem ./...

# Run tests with verbose output and show all logs
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v -count=1 ./...

# Run specific test by name
test-specific:
	@if [ -z "$(TEST)" ]; then \
		echo "Usage: make test-specific TEST=TestName"; \
		exit 1; \
	fi
	@echo "Running test: $(TEST)"
	go test -v -run $(TEST) ./...

# Run tests for specific package
test-package:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make test-package PKG=./internal/handlers"; \
		exit 1; \
	fi
	@echo "Running tests for package: $(PKG)"
	go test -v $(PKG)

# Clean test cache and generated files
clean:
	@echo "Cleaning test cache and generated files..."
	go clean -testcache
	rm -f coverage.out coverage.html

# Setup test environment (example)
setup-test-env:
	@echo "Setting up test environment..."
	@echo "Creating .env.test file..."
	@cat > .env.test << EOF
PORT=8080
ENVIRONMENT=test
DATABASE_URL=postgres://weather_user:weather_pass@localhost:5432/weather_test?sslmode=disable
BUN_DEBUG=true
EOF
	@echo "Test environment file created: .env.test"

# Run tests with test database
test-with-db:
	@echo "Running tests with test database..."
	@if [ ! -f .env.test ]; then \
		echo "Creating test environment..."; \
		make setup-test-env; \
	fi
	@export $$(cat .env.test | xargs) && go test -v ./...

# Check test coverage percentage
test-coverage-check:
	@echo "Checking test coverage..."
	go test -coverprofile=coverage.out ./... > /dev/null
	@go tool cover -func=coverage.out | tail -1 | awk '{print "Total coverage: " $$3}'
	@coverage=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	if [ $$(echo "$$coverage < 80" | bc -l) -eq 1 ]; then \
		echo "❌ Coverage is below 80% ($$coverage%)"; \
		exit 1; \
	else \
		echo "✅ Coverage is acceptable ($$coverage%)"; \
	fi

# Run tests and generate report
test-report:
	@echo "Generating test report..."
	go test -v -json ./... > test-report.json
	@echo "Test report generated: test-report.json"

# Continuous testing (watch for changes)
test-watch:
	@echo "Starting continuous testing (install fswatch if not available)..."
	@if command -v fswatch >/dev/null 2>&1; then \
		fswatch -o . | xargs -n1 -I{} make test-unit; \
	else \
		echo "fswatch not found. Install it with: brew install fswatch (macOS) or apt-get install fswatch (Linux)"; \
	fi

# Performance testing
test-performance:
	@echo "Running performance tests..."
	go test -v -bench=. -benchtime=10s -benchmem ./...

# Memory testing
test-memory:
	@echo "Running memory tests..."
	go test -v -bench=. -benchmem -memprofile=mem.prof ./...
	@echo "Memory profile generated: mem.prof"
	@echo "View with: go tool pprof mem.prof"

# CPU profiling
test-cpu:
	@echo "Running CPU profiling tests..."
	go test -v -bench=. -cpuprofile=cpu.prof ./...
	@echo "CPU profile generated: cpu.prof"
	@echo "View with: go tool pprof cpu.prof"

# Test with different Go versions (if using Docker)
test-go-versions:
	@echo "Testing with different Go versions..."
	@for version in 1.19 1.20 1.21; do \
		echo "Testing with Go $$version..."; \
		docker run --rm -v $$(pwd):/app -w /app golang:$$version go test -v ./...; \
	done

# Lint and test
lint-and-test:
	@echo "Running linter and tests..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install it from https://golangci-lint.run/"; \
	fi
	make test

# Dependencies for testing
test-deps:
	@echo "Installing test dependencies..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go mod tidy

# All quality checks
qa: lint-and-test test-coverage-check test-performance test-memory test-cpu
	@echo "All quality checks passed!"
# Run all tests with coverage and quality checks
test-all: test-coverage lint-and-test test-coverage-check test-performance test-memory test-cpu
	@echo "All tests and quality checks passed!"
# Run all tests with coverage and quality checks
test-all-verbose: test-coverage lint-and-test test-coverage-check test-performance test-memory test-cpu
	@echo "All tests and quality checks passed with verbose output!"
# Run all tests with coverage and quality
test-all-race: test-coverage lint-and-test test-coverage-check test-performance test-memory test-cpu
	@echo "All tests and quality checks passed with
			