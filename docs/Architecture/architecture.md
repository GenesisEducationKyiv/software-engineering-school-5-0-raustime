# Архітектура Weather-API

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
              v
+------------------------------------------------+
|             Infrastructure Layer               |
|  - Weather Adapters (OpenWeather, WeatherAPI)  |
|  - Redis Cache                                 |
|  - Mailer (SMTP + Templates)                   |
|  - Configuration Loader                        |
|  - Database Access (Bun ORM, Repositories)     |
+------------------------------------------------+

Розподіл шарів

1. Presentation Layer

Відповідає за прийом HTTP-запитів та їх попередню обробку

router.go, middleware.go, weather_handler.go, subscription_handler.go

1. Application Layer

Реалізує бізнес-логіку, оркеструє виклики між компонентами

weather_service.go, subscription_service.go, mailer_service.go, scheduler.go

1. Domain Layer

Містить основні сутності та контракти: WeatherData, Subscription, apierrors

contracts.go, models/subscription.go

1. Infrastructure Layer

Доступ до БД через bun

subscription_repo.go, migration.go

Інтеграція з Redis, зовнішніми API, SMTP, логування

redis_cache.go, openweather_adapter.go, weather_adapter.go, weather_logger.go, mailer_service.go
