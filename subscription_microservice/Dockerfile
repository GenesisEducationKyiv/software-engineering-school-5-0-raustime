# ---- build stage ----
FROM golang:1.23 AS builder

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
COPY migrations ./migrations

# 🔧 Назва бінарника — subscription
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o subscription ./cmd/main.go

# ---- run stage ----
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

# Бінарник і, за потреби, міграції або шаблони
COPY --from=builder /app/subscription /app/subscription
COPY --from=builder /app/migrations /app/migrations

ENV GRPC_PORT=8090
ENV HTTP_PORT=8091

EXPOSE 8090
EXPOSE 8091

ENTRYPOINT ["/app/subscription"]