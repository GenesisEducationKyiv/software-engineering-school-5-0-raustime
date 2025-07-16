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
make build     # Побудувати образи
make up        # Запустити контейнери
make down      # Зупинити і видалити контейнери
make restart   # Перезапуск (down + up)
make logs      # Переглянути логи
make logs-weather_service  # Логи тільки weather-сервісу
make logs-mailer_service # Логи тільки mailer-сервісу
make logs-subscription_service # Логи тільки subscription-сервісу
make logs-scheduler_service # Логи тільки scheduler-сервісу