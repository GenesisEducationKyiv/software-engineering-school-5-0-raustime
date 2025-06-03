# syntax=docker/dockerfile:1
FROM golang:1.21

# Встановлюємо bash, netcat, postgresql-client
RUN apt-get update && apt-get install -y bash netcat-openbsd postgresql-client && rm -rf /var/lib/apt/lists/*

# Робоча директорія всередині контейнера
WORKDIR /app

# Копіюємо go.mod та go.sum і завантажуємо залежності
COPY go.mod go.sum ./
RUN go mod download

# Копіюємо весь код у контейнер
COPY . .

# Збираємо бінарник під ім'ям app
RUN go build -o app main.go

# Даємо права на виконання скрипту очікування Postgres
RUN chmod +x /app/wait-for-postgres.sh

# Переконуємося, що app має права на виконання
RUN chmod +x ./app

# Вказуємо скрипт як entrypoint
ENTRYPOINT ["/app/wait-for-postgres.sh"]

# Запускаємо бінарник
CMD ["./app"]
