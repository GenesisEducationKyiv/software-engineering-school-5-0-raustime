# Makefile for subscription_service

SHELL := /bin/bash

BINARY_NAME=subscription
GRPC_PORT=8090
HTTP_PORT=8091

# OS-based DB_HOST (auto-switch for host.docker.internal or 172.17.0.1)
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
	DB_HOST := 172.17.0.1
else
	DB_HOST := host.docker.internal
endif

.PHONY: docker-build docker-run docker-test docker-test-unit docker-test-integration clean lint

docker-build:
	docker build -t subscription_service .

docker-run:
	docker compose -f docker-compose.subscription.yml up --build

docker-test:
	docker run --rm -v $(PWD):/app -w /app golang:1.23 go test -v ./...

docker-test-unit:
	docker run --rm -v $(PWD):/app -w /app golang:1.23 \
	  bash -c 'go test -v $$(go list ./... | grep -vE "/tests/(integration|e2e)")'
	  
docker-test-integration:
	docker run --rm -v $(PWD):/app -w /app golang:1.23 \
	  bash -c 'if [ -d ./tests/integration ]; then go test -v ./tests/integration/...; else echo "No integration tests directory found"; fi'

lint:
	golangci-lint run ./...

clean:
	rm -f $(BINARY_NAME)
