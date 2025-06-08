# syntax=docker/dockerfile:1
FROM golang:1.23

# Встановлюємо bash, netcat, postgresql-client
RUN apt-get update && apt-get install -y bash netcat-openbsd postgresql-client && rm -rf /var/lib/apt/lists/*

# Робоча директорія всередині контейнера
WORKDIR /app

# Створюємо скрипт очікування
RUN printf '#!/bin/bash\nset -e\n\nhost="$PGHOST"\nport="$PGPORT"\nuser="$PGUSER"\n\nuntil pg_isready -h "$host" -p "$port" -U "$user"; do\n  >&2 echo "Postgres is unavailable - sleeping"\n  sleep 1\ndone\n\n>&2 echo "Postgres is up - executing command"\nexec "$@"\n' > /app/wait-for-postgres.sh && \
    chmod +x /app/wait-for-postgres.sh

# Копіюємо go.mod та go.sum і завантажуємо залежності
COPY go.mod go.sum ./
RUN go mod download

# Копіюємо весь код у контейнер
COPY . .

# Збираємо бінарник
RUN go build -o app ./cmd && chmod +x ./app

# Вказуємо скрипт як entrypoint
ENTRYPOINT ["/app/wait-for-postgres.sh"]

# Запускаємо бінарник
CMD ["./app"]