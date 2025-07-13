# Makefile for building and running subscription_microservice with Docker

APP_NAME := subscription-service
GO_VERSION := 1.23.5

.PHONY: build run compose up down migrate

## Build Go binary using Docker
build:
	docker run --rm \
	  -v $$(pwd):/app \
	  -w /app \
	  golang:$(GO_VERSION)-alpine \
	  sh -c "go mod download && go build -o $(APP_NAME) ./cmd/main.go"

## Run binary locally
run:
	./$(APP_NAME)

## Start Docker Compose
compose:
	docker compose up --build

up:
	docker compose up -d

down:
	docker compose down

## Apply SQL migrations with golang-migrate
migrate:
	docker run --rm -v $$(pwd)/migrations:/migrations \
	  --network host \
	  migrate/migrate \
	  -path=/migrations -database "postgres://postgres:postgres@localhost:5432/subscriptions?sslmode=disable" up
