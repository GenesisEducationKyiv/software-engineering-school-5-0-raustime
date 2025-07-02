SHELL := /bin/bash

.PHONY: test-unit test-integration test-e2e test-all

test-unit:
	@echo "Running unit tests..."
	@bash -c 'go test -v $$(go list ./... | grep -vE "/tests/(integration|e2e)") 2>/dev/null || go test -v ./cmd ./internal/...'

test-integration:
	@if [ -d "./tests/integration" ]; then \
		echo "Running integration tests..."; \
		go test -v ./tests/integration/...; \
	else \
		echo "No integration tests directory found"; \
	fi

test-e2e:
	@if [ -d "./tests/e2e" ]; then \
		echo "Running e2e tests..."; \
		go test -v ./tests/e2e/...; \
	else \
		echo "No e2e tests directory found"; \
	fi

test-all:
	@echo "Running all tests..."
	go test -v ./...