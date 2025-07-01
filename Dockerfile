# syntax=docker/dockerfile:1
FROM golang:1.23 as builder

# Встановлюємо bash, netcat, postgresql-client
RUN apt-get update && \
    apt-get install -y --no-install-recommends bash netcat-openbsd postgresql-client && \
    rm -rf /var/lib/apt/lists/*

# Робоча директорія всередині контейнера
WORKDIR /app

# Копіюємо go.mod та go.sum — кешування залежностей
COPY go.mod go.sum ./
RUN go mod download

# Копіюємо решту коду
COPY . .

# Збірка бінарника
RUN go build -o app ./cmd

# Скрипт очікування PostgreSQL
RUN echo '#!/bin/bash\n\
    set -e\n\
    host="${PGHOST:-db}"\n\
    port="${PGPORT:-5432}"\n\
    user="${PGUSER:-postgres}"\n\
    \n\
    echo "Waiting for PostgreSQL at $host:$port as user $user..."\n\
    until pg_isready -h "$host" -p "$port" -U "$user"; do\n\
    sleep 1\n\
    done\n\
    echo "PostgreSQL is up. Running command: $@"\n\
    exec "$@"' > /app/wait-for-postgres.sh && chmod +x /app/wait-for-postgres.sh

# Використовуємо alpine замість scratch для меншого розміру
FROM alpine:latest

# Встановлюємо необхідні пакети для runtime
RUN apk --no-cache add \
    bash \
    netcat-openbsd \
    postgresql-client \
    ca-certificates

# Явно створюємо директорію /app
RUN mkdir -p /app
WORKDIR /app

# Копіюємо бінарник та скрипт з builder stage
COPY --from=builder /app/app ./app
COPY --from=builder /app/wait-for-postgres.sh ./wait-for-postgres.sh
COPY --from=builder /app/Makefile ./Makefile

# Переконуємося, що скрипт має права на виконання
RUN chmod +x ./wait-for-postgres.sh

# Встановлення entrypoint та команд за замовчуванням
ENTRYPOINT ["/app/wait-for-postgres.sh"]
CMD ["/app"]