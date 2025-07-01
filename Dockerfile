# syntax=docker/dockerfile:1
FROM golang:1.23-alpine as builder

# Install build dependencies
RUN apk add --no-cache bash netcat-openbsd postgresql-client make

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Build the binary with static linking for Alpine
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o app ./cmd

# Copy and make executable the wait script (assuming it exists in your project)
# If the script doesn't exist, create it with proper line endings
COPY wait-for-postgres.sh /app/wait-for-postgres.sh
RUN chmod +x /app/wait-for-postgres.sh

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add \
    bash \
    netcat-openbsd \
    postgresql-client \
    ca-certificates

# Create app directory
WORKDIR /app

# Copy binary and script from builder stage
COPY --from=builder /app/app .
COPY --from=builder /app/wait-for-postgres.sh .

# Ensure script is executable
RUN chmod +x ./wait-for-postgres.sh ./app

# Set entrypoint and default command
ENTRYPOINT ["/app/wait-for-postgres.sh"]
CMD ["/app/app"]