# Makefile for mailer_microservice

SHELL := /bin/bash

BINARY_NAME=mailer
PORT=8089

.PHONY: docker-build docker-run docker-test docker-test-unit docker-test-integration clean lint

docker-build:
	docker build -t mailer_service .

docker-run:
	docker compose -f docker-compose.mailer.yml up --build

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
