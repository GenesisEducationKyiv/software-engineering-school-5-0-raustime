FROM golang:1.23 AS builder

WORKDIR /app
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o scheduler cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/scheduler .

CMD ["./scheduler"]
