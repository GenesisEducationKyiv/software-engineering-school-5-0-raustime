# syntax=docker/dockerfile:1
FROM golang:1.23

# Встановлюємо bash, netcat, postgresql-client
RUN apt-get update && \
    apt-get install -y --no-install-recommends bash netcat-openbsd postgresql-client && \
    rm -rf /var/lib/apt/lists/*

# Робоча директорія всередині контейнера
WORKDIR /app

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

# Копіюємо go.mod та go.sum — кешування залежностей
COPY go.mod go.sum ./
RUN go mod download

# Копіюємо решту коду
COPY . .

# Збірка бінарника
RUN go build -o app ./cmd && chmod +x ./app

# Встановлення entrypoint та команд за замовчуванням
ENTRYPOINT ["/app/wait-for-postgres.sh"]
CMD ["./app"]
