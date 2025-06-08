# syntax=docker/dockerfile:1
FROM golang:1.23

# Встановлюємо bash, netcat, postgresql-client
RUN apt-get update && apt-get install -y bash netcat-openbsd postgresql-client && rm -rf /var/lib/apt/lists/*

# Робоча директорія всередині контейнера
WORKDIR /app

# After WORKDIR /app
COPY wait-for-postgres.sh ./
RUN chmod +x ./wait-for-postgres.sh

# Копіюємо go.mod та go.sum і завантажуємо залежності
COPY go.mod go.sum ./
RUN go mod download

# Копіюємо весь код у контейнер
COPY . .

# Збираємо бінарник 
RUN go build -o app ./cmd


# Переконуємося, що app має права на виконання
RUN chmod +x ./app

# Вказуємо скрипт як entrypoint
ENTRYPOINT ["/app/wait-for-postgres.sh"]

# Запускаємо бінарник
CMD ["./app"]
