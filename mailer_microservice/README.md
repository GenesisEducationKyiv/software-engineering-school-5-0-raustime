# 📬 Mailer Microservice

Мікросервіс для надсилання email-повідомлень з підтримкою шаблонів, SMTP і gRPC streaming.

---

## 🚀 Запуск сервісу

### 🐳 Через Docker

```bash
make docker-build
make docker-run
make docker-run-compose
```

або вручну:

```bash
docker build -t mailer_service .
docker run -p 8089:8089 \
  -e PORT=8089 \
  -e APP_BASE_URL=http://localhost:8089 \
  -e SMTP_HOST=smtp.example.com \
  -e SMTP_PORT=587 \
  -e SMTP_USER=test@example.com \
  -e SMTP_PASSWORD=pass123 \
  mailer_service
```

---

## 🧪 Запуск тестів через Docker

> Тести виконуються у чистому середовищі на основі офіційного образу `golang:1.23`

### 🔹 Юніт-тести
```bash
make docker-test-unit
```

### 🔹 Інтеграційні тести
```bash
make docker-test-integration
```

### 🔹 E2E тести
```bash
make docker-test-e2e
```

### 🔹 Всі тести
```bash
make docker-test
```

---

## 📁 Шаблони
Шаблони HTML для email знаходяться в `internal/templates/`:

- `confirmation_email.html`
- `weather_email.html`

---

## ⚙️ ENV змінні

| Змінна          | Приклад значення             | Обов'язково |
|------------------|-------------------------------|-------------|
| `PORT`           | `8089`                        | ✅          |
| `APP_BASE_URL`   | `http://localhost:8089`       | ✅          |
| `SMTP_HOST`      | `smtp.example.com`            | ✅          |
| `SMTP_PORT`      | `587`                         | ✅          |
| `SMTP_USER`      | `user@example.com`            | ✅          |
| `SMTP_PASSWORD`  | `secretpassword`              | ✅          |
| `TEMPLATE_DIR`   | `internal/templates`          | ❌ (default)

---

## 🧰 Команди Make

```bash
make docker-build          # збірка Docker-образу
make docker-run            # запуск сервісу в Docker
make docker-test           # всі тести у golang:1.23
make docker-test-unit      # тільки юніт-тести
make docker-test-integration # тільки інтеграційні тести
```

---

## 📦 Залежності

- Docker
- Make
- SMTP-сервер (локальний або зовнішній)
