# 🌦️ WeatherAPI Service

Цей сервіс надає погодні дані через HTTP API, використовуючи адаптери до зовнішніх сервісів — OpenWeather та WeatherAPI. Також підтримується логіка підписок, надсилання листів та кешування.

## 🔧 Технології

- **Go 1.23**
- **Docker + Docker Compose**
- **Prometheus metrics** (`/metrics`)
- Адаптери до:
  - OpenWeather API
  - WeatherAPI
- Кеш: Noop або Redis (налаштовується)
- Email-повідомлення (через абстрактний `MailerService`)
- Підтримка `graceful shutdown`

---

## 🚀 Швидкий старт

Використовуйте `make` для запуску:

```bash
make build                      # Побудувати образи
make up                         # Запустити контейнери
make down                       # Зупинити і видалити контейнери
make restart                    # Перезапуск (down + up)
make logs                       # Переглянути логи
make logs-weather_service       # Логи тільки weather-сервісу
make logs-mailer_service        # Логи тільки mailer-сервісу
make logs-subscription_service  # Логи тільки subscription-сервісу
make logs-scheduler_service     # Логи тільки scheduler-сервісу

make up-bench                   # Порівняльний бенчмарк gRPC та REST
```

## 🔧 Генереація коду з proto

- install the Buf CLI
  Windows:
    scoop install buf
  macOS or Linux:
    brew install bufbuild/buf/buf
  NPM:
    npm install @bufbuild/buf

- buf generate для генерації


## 📊 Порівняльний тест HTTP vs ConnectRPC (gRPC)

| Показник              | **HTTP REST (wrk)** | **gRPC ConnectRPC (ghz)** |
| --------------------- | ------------------- | ------------------------- |
| **Кількість запитів** | 70,529              | 20,000                    |
| **Час тесту**         | 15s                 | 5.75s                     |
| **RPS (запитів/сек)** | **4,671**           | **3,477**                 |
| **Середня затримка**  | 31.14 ms            | 110.23 ms                 |
| **Найшвидший запит**  | —                   | 13.87 ms                  |
| **Найповільніший**    | —                   | 551.25 ms                 |
| **Медіана (p50)**     | —                   | 100.93 ms                 |
| **p95**               | —                   | 194.29 ms                 |
| **Коди відповіді**    | 200 OK              | 200 OK                    |
| **Протокол**          | HTTP/1.1            | HTTP/2 (plaintext, h2c)   |
| **Формат даних**      | JSON                | JSON                      |
