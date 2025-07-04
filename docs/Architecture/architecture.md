Архітектура Weather-API

Цей документ описує архітектуру погодного API-додатку, який реалізований на Go і надає функціональність отримання поточних погодних умов та систему підписок з email-розсилкою.

Загальна структура

Додаток побудований на основі багатошарової архітектури з елементами чистої архітектури (Clean Architecture) і шаблоном Chain of Responsibility для роботи з провайдерами погоди.

+-----------------------------+
|      Frontend (Vue.js)     |
| - Weather Dashboard        |
| - Subscription Mgmt        |
+-------------+--------------+
              |
              v
+-----------------------------+
|  Presentation Layer         |
|  - WeatherHandler           |
|  - SubscriptionHandler      |
|  - Middleware (CORS, etc.) |
+-------------+--------------+
              |
              v
+-----------------------------+
|  Application Layer          |
|  - WeatherService           |
|  - SubscriptionService      |
|  - Scheduler (jobs)         |
+-------------+--------------+
              |
              v
+-----------------------------+
|  Domain Layer               |
|  - WeatherData              |
|  - Subscription             |
|  - Business Rules           |
+-------------+--------------+
              |
     +--------+--------+
     |                 |
     v                 v
+-----------+   +------------------------+
|  Database |   | Infrastructure         |
| (Postgres)|   | - Weather Adapters     |
| - Repos   |   | - Redis Cache          |
| - Models  |   | - Mailer (SMTP)        |
+-----------+   | - Config Loader        |
                +------------------------+

Розподіл шарів

1. Presentation Layer

Відповідає за прийом HTTP-запитів та їх попередню обробку

router.go, middleware.go, weather_handler.go, subscription_handler.go

2. Application Layer

Реалізує бізнес-логіку, оркеструє виклики між компонентами

weather_service.go, subscription_service.go, mailer_service.go, scheduler.go

3. Domain Layer

Містить основні сутності та контракти: WeatherData, Subscription, apierrors

contracts.go, models/subscription.go

4. Infrastructure Layer

Інтеграція з Redis, зовнішніми API, SMTP, логування

redis_cache.go, openweather_adapter.go, weather_adapter.go, weather_logger.go, mailer_service.go

5. Data Layer

Доступ до БД через bun

subscription_repo.go, migration.go

