.PHONY: test-unit test-integration test-e2e test-all

test-unit:
	go test -v $$(go list ./... | grep -vE '/tests/(integration|e2e)')

test-integration:
	@if [ -d "./tests/integration" ]; then \
		go test -v ./tests/integration/...; \
	else \
		echo "No integration tests directory found"; \
	fi

test-e2e:
	@if [ -d "./tests/e2e" ]; then \
		go test -v ./tests/e2e/...; \
	else \
		echo "No e2e tests directory found"; \
	fi

test-all:
	go test -v ./...