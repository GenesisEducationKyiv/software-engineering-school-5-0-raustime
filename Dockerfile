# syntax=docker/dockerfile:1
FROM golang:1.23-alpine as base

# Install common dependencies
RUN apk add --no-cache bash netcat-openbsd postgresql-client make git

# Встановлюємо bash та make (якщо його немає), щоб уникнути проблем з оболонкою
RUN apk add --no-cache bash make

# Встановлюємо bash як стандартний шелл
SHELL ["/bin/bash", "-c"]

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum for dependency caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the code
COPY . .

# Test target - keeps full Go environment for running tests
FROM base as test
CMD ["make", "test"]

# Builder target - compiles the production binary
FROM base as builder
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o app ./cmd

# Production target
FROM alpine:latest as production

# Install minimal runtime dependencies
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/app .

# Copy entire project (excluding build artifacts via .dockerignore)
COPY --from=base /app .

# Ensure binary is executable
RUN chmod +x ./app

CMD ["/app/app"]