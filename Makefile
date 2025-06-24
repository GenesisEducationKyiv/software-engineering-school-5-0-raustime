.PHONY: test-unit test-integration test-e2e test-all

UNIT_PACKAGES := $(shell go list ./... | grep -vE '/tests/(integration|e2e)')
INTEGRATION_PACKAGES := $(shell go list ./tests/integration/...)
E2E_PACKAGES := $(shell go list ./tests/e2e/...)

test-unit:
	go test -v $(UNIT_PACKAGES)

test-integration:
	go test -v $(INTEGRATION_PACKAGES)

test-e2e:
	go test -v $(E2E_PACKAGES)

test-all:
	go test -v ./...
