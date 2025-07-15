# ---- build stage ----
FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o mailer ./cmd/main.go

# ---- run stage ----
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Копіюємо з білд-стейджу
COPY --from=builder /app/mailer /app/mailer
COPY --from=builder /app/internal/templates /app/internal/templates

# Змінні оточення можуть задаватися через docker-compose
ENV PORT=8089

EXPOSE 8089

ENTRYPOINT ["/app/mailer"]
