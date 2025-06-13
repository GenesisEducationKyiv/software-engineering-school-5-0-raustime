# Database migration commands
.PHONY: migrate-up migrate-down migrate-status migrate-create

# Run all pending migrations
migrate-up:
	go run cmd/migrate/main.go -action=up

# Rollback the last migration
migrate-down:
	go run cmd/migrate/main.go -action=down

# Show migration status
migrate-status:
	go run cmd/migrate/main.go -action=status

# Create a new migration file
migrate-create:
	@read -p "Enter migration name: " name; \
	timestamp=$$(date +%03d); \
	mkdir -p migrations; \
	touch migrations/$${timestamp}_$${name}.up.sql; \
	touch migrations/$${timestamp}_$${name}.down.sql; \
	echo "Created migrations/$${timestamp}_$${name}.up.sql"; \
	echo "Created migrations/$${timestamp}_$${name}.down.sql"

# Build the application
build:
	go build -o bin/app cmd/main.go

# Run the application
run:
	go run cmd/main.go

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/