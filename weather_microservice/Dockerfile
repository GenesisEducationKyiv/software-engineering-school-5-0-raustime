# ---- build stage ----
FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o weather_microservice ./cmd/main.go

# ---- final stage ----
FROM debian:bookworm-slim

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /app/weather_microservice .
COPY .env ./

EXPOSE 8080

CMD ["./weather_microservice"]
