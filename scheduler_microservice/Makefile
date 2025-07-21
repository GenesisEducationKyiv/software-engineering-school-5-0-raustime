# Makefile for scheduler_microservice

SHELL := /bin/bash

BINARY_NAME=scheduler
PORT=8092

.PHONY: docker-build docker-run docker-test docker-test-unit docker-test-integration clean lint

docker-build:
	docker build -t scheduler_service .

docker-run:
	docker compose -f docker-compose.scheduler.yml up --build scheduler-service
	
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
