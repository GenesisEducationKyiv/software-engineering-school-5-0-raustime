# --- Builder Stage ---
FROM golang:1.23.5-alpine AS builder

WORKDIR /app

# Встановлення залежностей
COPY go.mod go.sum ./
RUN go mod download

# Копіюємо вихідний код
COPY . .

# Компіляція
RUN go build -o /subscription-service ./cmd/main.go

# --- Runtime Stage ---
FROM alpine:latest

WORKDIR /app

# Додати базові системні залежності
RUN apk add --no-cache ca-certificates

# Копіюємо зібраний бінарник
COPY --from=builder /subscription-service .

# Конфігурація за замовчуванням (може бути перевизначена через docker-compose або ENV)
ENV PORT=8080 \
    GATEWAY_PORT=8081 \
    DB_HOST=db \
    DB_PORT=5432 \
    DB_USER=postgres \
    DB_PASSWORD=postgres \
    DB_NAME=subscriptions \
    MAILER_GRPC_ADDR=http://mailer:8080

# Запуск сервісу
ENTRYPOINT ["./subscription-service"]
