# Визначення потенційних мікросервісів у проєкті

На основі аналізу поточної структури додатку можна виділити кілька функціональних блоків, які логічно винести в окремі мікросервіси. Це покращить масштабованість, модульність і дозволить розподілити навантаження.

---

## 1. Mailer Service

**Найкращий кандидат на винесення.**

### Обов’язки Mailer Service

- Надсилання листів (SMTP або сторонній email API)
- Валідація шаблонів і email-адрес
- Логування та моніторинг статусу доставки

### Причини для винесення

- Ясна межа відповідальності
- Можливість незалежного масштабування
- Залежність від SMTP чи сторонніх email API
- Вже є тестові заглушки (`MockSender`) та окрема логіка

### Рекомендована комунікація (Mailer Service)

→ Message Queue (asynchronous) NATS, RabbitMQ або Kafka
→ gRPC / HTTP API (synchronous call)


---

## 2. Weather Integration Service

Інкапсулює логіку отримання погоди з зовнішніх сервісів.

### Обов’язки Weather Integration Service

- Взаємодія з OpenWeather та WeatherAPI
- Агрегація погодних даних
- Кешування (через `NoopWeatherCache` або реальний Redis)

### Причини для винесення (Weather Integration Service)

- Зовнішні залежності (API)
- Ясно виражений API-контракт
- Можна використовувати як загальний сервіс для інших проєктів
- Легко інкапсулюється (див. `weather_adapter.go`, `openweather_adapter.go`)

### Рекомендована комунікація (Weather Integration Service)

→ gRPC або REST

---

## 3. Subscription Service

Обробка логіки підписок та розсилок.

### Обов’язки Subscription Service

- Реєстрація та керування підписками
- Тригери email-розсилок
- Виклики до `MailerService`

### Причини для винесення (Subscription Service)

- Автономна логіка (напр. user opt-in/out)
- Залежить від `MailerService` (який уже буде мікросервісом)
- Добре підходить для окремого REST/Queue-based сервісу

### Рекомендована комунікація (Subscription Service)

→ gRPC або REST
→ Підписка на події через message broker (RabbitMQ, NATS, Kafka)

---

## 4. Job Scheduler

### Обов’язки Job Scheduler

- Планові задачі:
- Оновлення кешу погоди
- Розсилки на основі розкладу
- Фонова активація інших сервісів

### Причини для винесення (Job Scheduler)

- Може бути централізованим для всього проєкту
- Дозволяє асинхронно запускати сервіси через Pub/Sub чи HTTP

### Рекомендована комунікація (Job Scheduler)

→ Pub/Sub або gRPC виклики до внутрішніх мікросервісів.

---

## Загальна картина

```plaintext
[ API Gateway / BFF ]
         |
 ┌──────────────┬───────────────────┬─────────────────────┐
▼              ▼                   ▼                     ▼
WeatherService  SubscriptionService  MailerService      JobScheduler
     |                 |                   |                |
[ WeatherAPI,          |                   |                |
  OpenWeather ]        ▼                   ▼                ▼
                   [Redis/DB]         [SMTP/API]     [Pub/Sub Tasks]
