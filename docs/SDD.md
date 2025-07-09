# Сервіс Підписки на Погоду

## Software Design Document (SDD)

**Версія:** 1.01  
**Дата:** 09/06/2025  
**Автор:** Roman Ustymenko  

[шаблон, на основі якого створено документ](https://wildart.github.io/MISG5020/standards/SDD_Template.pdf)

---

## ЗМІСТ

1. [ВСТУП](#1-вступ)
   - 1.1 [Призначення](#11-призначення)
   - 1.2 [Область застосування](#12-область-застосування)
   - 1.3 [Огляд документу](#13-огляд-документу)
   - 1.4 [Довідкові матеріали](#14-довідкові-матеріали)
   - 1.5 [Визначення та абревіатури](#15-визначення-та-абревіатури)

2. [ОГЛЯД СИСТЕМИ](#2-огляд-системи)

3. [АРХІТЕКТУРА СИСТЕМИ](#3-архітектура-системи)
   - 3.1 [Архітектурний дизайн](#31-архітектурний-дизайн)
   - 3.2 [Опис декомпозиції](#32-опис-декомпозиції)
   - 3.3 [Обґрунтування дизайну](#33-обґрунтування-дизайну)

4. [ПРОЕКТУВАННЯ ДАНИХ](#4-проектування-даних)
   - 4.1 [Опис даних](#41-опис-даних)
   - 4.2 [Словник даних](#42-словник-даних)

5. [ПРОЕКТУВАННЯ КОМПОНЕНТІВ](#5-проектування-компонентів)

6. [ПРОЕКТУВАННЯ КОРИСТУВАЦЬКОГО ІНТЕРФЕЙСУ](#6-проектування-користувацького-інтерфейсу)
   - 6.1 [Огляд користувацького інтерфейсу](#61-огляд-користувацького-інтерфейсу)
   - 6.2 [Зображення екранів](#62-зображення-екранів)
   - 6.3 [Об'єкти екрану та дії](#63-обєкти-екрану-та-дії)

7. [МАТРИЦЯ ВИМОГ](#7-матриця-вимог)

8. [ДОДАТКИ](#8-додатки)

---

## 1. ВСТУП

### 1.1 Призначення

Цей документ системного дизайну (SDD) описує архітектуру та системний дизайн сервісу підписки на погоду. Документ призначений для розробників, архітекторів програмного забезпечення та інших технічних спеціалістів, які відповідають за реалізацію, тестування та підтримку системи.

SDD служить основним довідковим документом для розробки коду та містить всю інформацію, необхідну програмісту для написання коду.

### 1.2 Область застосування

Сервіс підписки на погоду - це веб-застосунок, який дозволяє користувачам:

- Отримувати поточну інформацію про погоду для обраного міста
- Підписуватися на регулярні оновлення погоди через електронну пошту
- Керувати частотою отримання повідомлень (щогодини або щодня)
- Підтверджувати підписку через електронну пошту
- Скасовувати підписку

Метою проекту є створення простого, надійного та масштабованого сервісу для автоматичного надсилання погодних оновлень користувачам.

### 1.3 Огляд документу

Цей документ організований відповідно до стандарту IEEE STD 1016 та містить:

- Огляд системи та її контексту
- Детальний опис архітектури системи
- Проектування даних та структур
- Специфікації компонентів
- Опис користувацького інтерфейсу
- Зв'язок з функціональними вимогами

### 1.4 Довідкові матеріали

- IEEE STD 1016-2009: Recommended Practice for Software Design Descriptions
- WeatherAPI.com Documentation
- PostgreSQL Documentation
- Go Programming Language Specification
- Vue.js Framework Documentation

### 1.5 Визначення та абревіатури

| Термін | Визначення |
|--------|------------|
| API | Application Programming Interface |
| SDD | Software Design Document |
| SPA | Single Page Application |
| REST | Representational State Transfer |
| SMTP | Simple Mail Transfer Protocol |
| TTL | Time To Live |
| ORM | Object-Relational Mapping |
| JWT | JSON Web Token |
| HTTP | HyperText Transfer Protocol |
| JSON | JavaScript Object Notation |
| SQL | Structured Query Language |

---

## 2. ОГЛЯД СИСТЕМИ

Сервіс підписки на погоду являє собою розподілену веб-систему, що складається з клієнтської частини (Vue.js SPA), серверної частини (Go backend), бази даних (PostgreSQL) та інтеграції з зовнішніми сервісами (WeatherAPI.com, SMTP сервіс, (для прототипу обрано Google SMTP)).

Система призначена для автоматизації процесу надання актуальної погодної інформації користувачам через електронну пошту. Користувачі можуть підписатися на оновлення, вказавши свою електронну адресу, місто та бажану частоту отримання повідомлень.

Ключові характеристики системи:

- **Реактивність**: Веб-інтерфейс реагує на дії користувача в реальному часі
- **Надійність**: Система забезпечує стабільну роботу та обробку помилок
- **Масштабованість**: Архітектура дозволяє збільшення кількості користувачів
- **Безпека**: Підтвердження підписки через токени, валідація даних
- **Ефективність**: Кешування погодних даних для зменшення навантаження на зовнішні API

Система працює в режимі 24/7, автоматично надсилаючи погодні оновлення згідно з налаштуваннями користувачів.

---

## 3. АРХІТЕКТУРА СИСТЕМИ

### 3.1 Архітектурний дизайн

Архітектура системи — це чітко виражена багатошарова архітектура (**Layered Architecture**), із використанням принципів чистої архітектури (**Clean Architecture**) та шаблонів розробки.

```text
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
```

**Основні підсистеми:**

1. **Frontend Subsystem (Vue.js)**
   - Відповідальність: Надання користувацького інтерфейсу
   - Взаємодія: HTTP запити до Backend API

2. **Backend API Subsystem (Go)**
   - Відповідальність: Обробка бізнес-логіки, API endpoints
   - Взаємодія: SQL запити до БД, HTTP запити до зовнішніх API

3. **Database Subsystem (PostgreSQL)**
   - Відповідальність: Зберігання даних підписок
   - Взаємодія: SQL запити від Backend

4. **Email Subsystem (SMTP/SendGrid)**
   - Відповідальність: Надсилання електронних листів
   - Взаємодія: SMTP протокол від Backend

5. **Weather API Integration Subsystem**
   - Відповідальність: Отримання погодних даних
   - Взаємодія: HTTP API calls до WeatherAPI.com

6. **Caching Subsystem (In-Memory)**
   - Відповідальність: Кешування погодних даних
   - Взаємодія: Внутрішня взаємодія з Backend

7. **Scheduling Subsystem (Go Goroutines)**
   - Відповідальність: Планування та надсилання регулярних email
   - Взаємодія: Внутрішня взаємодія з Email та Database підсистемами

### 3.2 Опис декомпозиції

**Sequence Diagram - Subscription Flow:**

```text
User -> Frontend: Submit subscription form
Frontend -> Backend: POST /subscribe
Backend -> Database: Create subscription (unconfirmed)
Backend -> EmailService: Send confirmation email
EmailService -> User: Confirmation email with token
User -> Backend: GET /confirm/{token}
Backend -> Database: Update subscription (confirmed=true)
Backend -> User: Confirmation success page
```

### 3.3 Обґрунтування дизайну

**Вибір архітектурних рішень:**

1. **Багаторівнева архітектура**
   - Переваги: Чітке розділення відповідальності, легкість тестування, можливість незалежної розробки шарів
   - Альтернативи: Монолітна архітектура була б простішою, але менш гнучкою для майбутнього розвитку

2. **RESTful API дизайн**
   - Переваги: Стандартизований підхід, легкість інтеграції, кешування HTTP відповідей
   - Альтернативи: GraphQL надав би більше гнучкості, але додав би складності

3. **Redis кеширування**
   - Переваги: Швидкість доступу, простота реалізації, персистентність та розподіленість
   - Альтернативи: Швидкість доступу, простота реалізації

4. **Go Goroutines для планування**
   - Переваги: Вбудована конкурентність, ефективність, відсутність зовнішніх залежностей
   - Альтернативи: Cron jobs були б надійнішими, але менш інтегрованими з основним застосунком

**Критичні компроміси:**

- Простота vs Масштабованість: Обрано простоту для швидкої розробки MVP
- Консистентність vs Доступність: Пріоритет надано доступності через асинхронну email розсилку
- Безпека vs Зручність: Баланс між безпечними токенами та зручністю користування

---

## 4. ПРОЕКТУВАННЯ ДАНИХ

### 4.1 Опис даних

Система використовує реляційну базу даних PostgreSQL для зберігання основних даних та in-memory структури для кешування.

**Основні сутності даних:**

1. **Subscriptions** - центральна сутність системи, що зберігає інформацію про підписки користувачів
2. **Weather Cache** - тимчасове зберігання погодних даних в пам'яті
3. **Configuration** - системні налаштування

**Трансформація даних:**

- Вхідні дані від користувача валідуються та нормалізуються
- Погодні дані від зовнішнього API адаптуються до внутрішнього формату
- Email шаблони генеруються з погодних даних та налаштувань підписки

**Організація зберігання:**

- PostgreSQL: Персистентне зберігання підписок
- In-Memory Map: Швидке кешування погодних даних
- File System: Логи та конфігураційні файли

### 4.2 Словник даних

**Database Entities:**

| Entity | Type | Description |
|--------|------|-------------|
| **Subscription** | Table | Основна таблиця підписок |
| ID | int64 | Унікальний ідентифікатор підписки (Primary Key) |
| Email | string(255) | Email адреса користувача (Unique, Not Null) |
| City | string(100) | Назва міста для погодних оновлень (Not Null) |
| Frequency | string(10) | Частота розсилки: "hourly" або "daily" (Not Null) |
| Confirmed | boolean | Статус підтвердження підписки (Not Null, Default: false) |
| Token | string(64) | Унікальний токен для підтвердження/відписки (Not Null) |
| CreatedAt | timestamp | Дата створення підписки (Not Null, Default: current_timestamp) |
| ConfirmedAt | timestamp | Дата підтвердження підписки (Nullable) |

**In-Memory Structures:**

| Structure | Type | Description |
|-----------|------|-------------|
| **WeatherCache** | Map | Кеш погодних даних |
| CacheKey | string | Ключ кешу (назва міста) |
| CacheValue | struct | Закешовані погодні дані |
| Temperature | float64 | Температура в градусах Цельсія |
| Humidity | int | Вологість у відсотках |
| Description | string | Опис погодних умов |
| CachedAt | time.Time | Час кешування |
| TTL | time.Duration | Час життя кешу (5 хвилин) |

**API Data Structures:**

| Structure | Type | Description |
|-----------|------|-------------|
| **SubscriptionRequest** | JSON | Запит на створення підписки |
| email | string | Email адреса |
| city | string | Назва міста |
| frequency | string | Частота розсилки |
| **WeatherResponse** | JSON | Відповідь з погодними даними |
| temperature | float64 | Температура |
| humidity | int | Вологість |
| description | string | Опис погоди |
| **ErrorResponse** | JSON | Відповідь з помилкою |
| error | string | Повідомлення про помилку |
| code | int | Код помилки |

**Service Functions:**

| Function | Parameters | Return Type | Description |
|----------|------------|-------------|-------------|
| **GetWeather** | city: string | (*WeatherData, error) | Отримання погодних даних |
| **CreateSubscription** | req: SubscriptionRequest | (*Subscription, error) | Створення підписки |
| **ConfirmSubscription** | token: string | error | Підтвердження підписки |
| **UnsubscribeUser** | token: string | error | Відписка користувача |
| **SendEmail** | to: string, subject: string, body: string | error | Надсилання email |
| **CacheWeather** | city: string, data: *WeatherData | void | Кешування погодних даних |
| **GetCachedWeather** | city: string | (*WeatherData, bool) | Отримання з кешу |

---

## 5. ПРОЕКТУВАННЯ КОМПОНЕНТІВ

**WeatherService Component:**

```pseudocode
FUNCTION GetWeather(city)
  INPUT: city (string)
  OUTPUT: WeatherData, error
  
  // Check cache first
  cachedData, found = cache.Get(city)
  IF found AND not expired THEN
    RETURN cachedData, nil
  END IF
  
  // Fetch from external API
  url = buildWeatherAPIURL(city)
  response = httpClient.Get(url)
  IF response.error THEN
    RETURN nil, response.error
  END IF
  
  // Parse and transform data
  weatherData = parseWeatherResponse(response.body)
  
  // Cache the result
  cache.Set(city, weatherData, TTL)
  
  RETURN weatherData, nil
END FUNCTION
```

**SubscriptionService Component:**

```pseudocode
FUNCTION CreateSubscription(request)
  INPUT: SubscriptionRequest
  OUTPUT: Subscription, error
  
  // Validate input
  IF not isValidEmail(request.email) THEN
    RETURN nil, "Invalid email format"
  END IF
  
  // Check if subscription exists
  existing = database.FindByEmail(request.email)
  IF existing != nil THEN
    RETURN nil, "Subscription already exists"
  END IF
  
  // Create new subscription
  subscription = Subscription{
    Email: request.email,
    City: request.city,
    Frequency: request.frequency,
    Confirmed: false,
    Token: generateSecureToken(),
    CreatedAt: now()
  }
  
  // Save to database
  database.Create(subscription)
  
  // Send confirmation email
  emailService.SendConfirmation(subscription.Email, subscription.Token)
  
  RETURN subscription, nil
END FUNCTION
```

**EmailScheduler Component:**

```pseudocode
FUNCTION RunScheduler()
  ticker = createTicker(1 hour)
  
  FOR EACH tick IN ticker DO
    currentHour = getCurrentHour()
    
    // Get hourly subscriptions
    IF true THEN
      hourlySubscriptions = database.FindConfirmedByFrequency("hourly")
      processSubscriptions(hourlySubscriptions)
    END IF
    
    // Get daily subscriptions (only at specific hour)
    IF currentHour == 8 THEN // 8 AM
      dailySubscriptions = database.FindConfirmedByFrequency("daily")
      processSubscriptions(dailySubscriptions)
    END IF
  END FOR
END FUNCTION

FUNCTION processSubscriptions(subscriptions)
  FOR EACH subscription IN subscriptions DO
    weather = weatherService.GetWeather(subscription.City)
    IF weather != nil THEN
      emailService.SendWeatherUpdate(subscription, weather)
      logSuccess(subscription.Email, subscription.City)
    ELSE
      logError(subscription.Email, "Failed to fetch weather")
    END IF
  END FOR
END FUNCTION
```

**CacheManager Component:**

```pseudocode
FUNCTION CacheManager()
  cache = make(map[string]CachedWeather)
  mutex = RWMutex
  
  FUNCTION Set(key, value, ttl)
    mutex.Lock()
    cache[key] = CachedWeather{
      Data: value,
      ExpiresAt: now().Add(ttl)
    }
    mutex.Unlock()
  END FUNCTION
  
  FUNCTION Get(key)
    mutex.RLock()
    cached, exists = cache[key]
    mutex.RUnlock()
    
    IF not exists THEN
      RETURN nil, false
    END IF
    
    IF cached.ExpiresAt.Before(now()) THEN
      delete(cache, key)
      RETURN nil, false
    END IF
    
    RETURN cached.Data, true
  END FUNCTION
END FUNCTION
```

**API Handler Components:**

```pseudocode
FUNCTION HandleSubscribe(request)
  // Parse request body
  subscriptionReq = parseJSON(request.body)
  
  // Create subscription
  subscription, error = subscriptionService.Create(subscriptionReq)
  IF error THEN
    RETURN errorResponse(400, error.message)
  END IF
  
  RETURN successResponse(201, "Subscription created. Check email for confirmation.")
END FUNCTION

FUNCTION HandleWeather(request)
  city = request.query.Get("city")
  IF city == "" THEN
    RETURN errorResponse(400, "City parameter required")
  END IF
  
  weather, error = weatherService.GetWeather(city)
  IF error THEN
    RETURN errorResponse(500, "Failed to fetch weather")
  END IF
  
  RETURN jsonResponse(200, weather)
END FUNCTION
```

---

## 6. ПРОЕКТУВАННЯ КОРИСТУВАЦЬКОГО ІНТЕРФЕЙСУ

### 6.1 Огляд користувацького інтерфейсу

Користувацький інтерфейс реалізований як одностороннє веб-додаток (SPA) з використанням Vue.js фреймворку. Інтерфейс призначений для забезпечення простого та інтуїтивно зрозумілого досвіду користування.

**Основні функціональності для користувача:**

1. **Перегляд погоди**: Користувач може ввести назву міста та отримати поточну погодну інформацію
2. **Створення підписки**: Форма для введення email, міста та частоти отримання оновлень
3. **Підтвердження підписки**: Автоматичне перенаправлення з email для активації підписки
4. **Скасування підписки**: Можливість відписатися через посилання в email

**Зворотний зв'язок системи:**

- Миттєве відображення погодних даних
- Статусні повідомлення про успішну підписку
- Повідомлення про помилки з чіткими інструкціями
- Підтвердження операцій (підписка, відписка)

### 6.2 Зображення екранів

**Головна сторінка:**

```text
┌────────────────────────────────────────────┐
│           Weather Subscription             │
├────────────────────────────────────────────┤
│                                            │
│  ┌─────── Get Current Weather ──────┐      │
│  │                                  │      │
│  │  City:                           │      │
│  │        [________________]        │      │
│  │          [Get Weather]           │      │
│  │       Weather data received!     │      │
│  │  Temperature: 21°C               │      │
│  │  Humidity: 65%                   │      │
│  │  Condition: Sunny                │      │
│  └──────────────────────────────────┘      │
│                                            │
│  ┌───── Subscribe to Weather Forecast──────┐       
│  │                                  │      │
│  │  Email:                          │      │
│  │         [________________]       │      │
│  │  City:                           │      │
│  │         [________________]       │      │
│  │ Forecast Type:[Daily ▼]          │      │
│  │            [Subscribe]           │      │
│  │                                  │      │
│  └──────────────────────────────────┘      │
│                                            │
└────────────────────────────────────────────┘
```

**Сторінка підтвердження:**

```text
┌─────────────────────────────────────────────┐
│           Subscription Confirmed            │
├─────────────────────────────────────────────┤
│                                             │
│            ✓ Success!                       │
│                                             │
│    Your subscription has been confirmed.    │
│    You will receive weather updates for     │
│    [City] [Frequency].                      │
│                                             │
│         [Return to Homepage]                │
│                                             │
└─────────────────────────────────────────────┘
```

**Сторінка відписки:**

```
┌─────────────────────────────────────────────┐
│              Unsubscribed                   │
├─────────────────────────────────────────────┤
│                                             │
│            ✓ Unsubscribed                   │
│                                             │
│    You have been successfully unsubscribed  │
│    from weather updates.                    │
│                                             │
│         [Return to Homepage]                │
│                                             │
└─────────────────────────────────────────────┘
```

### 6.3 Об'єкти екрану та дії

**Головна сторінка - Weather Section:**

| Object | Type | Action | Result |
|--------|------|---------|---------|
| City Input Field | TextInput | onInput | Real-time validation |
| Get Weather Button | Button | onClick | API call to /weather |
| Weather Display | Container | - | Shows temperature, humidity, description |
| Error Message | Alert | - | Displays API errors |

**Головна сторінка - Subscription Section:**

| Object | Type | Action | Result |
|--------|------|---------|---------|
| Email Input | EmailInput | onInput | Email format validation |
| City Input | TextInput | onInput | Required field validation |
| Frequency Dropdown | Select | onChange | Update selected frequency |
| Subscribe Button | Button | onClick | API call to /subscribe |
| Success Message | Alert | - | Confirmation message |
| Error Message | Alert | - | Validation/API errors |

**Підтвердження підписки:**

| Object | Type | Action | Result |
|--------|------|---------|---------|
| Confirmation Message | Container | onLoad | Display confirmation status |
| Return Button | Button | onClick | Navigate to homepage |

**Станографіка взаємодій:**

```text
User Input → Frontend Validation → API Request → Backend Processing → Response → UI Update
```

**Обробка помилок на UI:**

- Валідація форм в реальному часі
- Відображення повідомлень про помилки API
- Graceful fallback для недоступності сервісів

---

## 7. МАТРИЦЯ ВИМОГ

| Функціональна вимога | Компонент системи | Модуль/Функція | Опис реалізації |
|---------------------|-------------------|----------------|-----------------|
| **FR-001**: Отримання поточної погоди | WeatherService | GetWeather() | API інтеграція з WeatherAPI.com + кешування |
| **FR-002**: Підписка на оновлення погоди | SubscriptionService | CreateSubscription() | Валідація даних + збереження в БД |
| **FR-003**: Підтвердження підписки через email | EmailService + API Handler | SendConfirmation() + HandleConfirm() | Генерація токенів + email надсилання |
| **FR-004**: Відписка через email | SubscriptionService | UnsubscribeUser() | Перевірка токену + видалення з БД |
| **FR-005**: Регулярна автоматична розсилка | EmailScheduler | RunScheduler() + processSubscriptions() | Go Goroutines + Timer |
| **FR-006**: Підтримка частоти надсилання | Database Model | Frequency field | Поле в таблиці subscriptions |
| **FR-007**: Актуальні погодні дані в email | WeatherService + EmailService | GetWeather() + SendWeatherUpdate() | Інтеграція сервісів |

| Нефункціональна вимога | Компонент системи | Реалізація | Метрика |
|------------------------|-------------------|------------|---------|
| **NFR-001**: Час відповіді API ≤ 500ms | CacheManager + WeatherService | In-memory кешування (TTL 5 min) | Response time monitoring |
| **NFR-002**: Висока доступність | Backend Architecture | Error handling + retry logic | 99.9% uptime target |
| **NFR-003**: Безпечне зберігання даних | Database + Token Management | PostgreSQL + secure token generation | Token entropy + DB security |
| **NFR-004**: Масштабованість | System Architecture | Модульна архітектура + Docker | Horizontal scaling capability |
| **NFR-005**: Логування дій | Logging System | Structured logging (JSON) | Log aggregation |

**Trace Matrix - Requirements to Components:**

```text
Requirements ──┬── WeatherService
               ├── SubscriptionService  
               ├── EmailService
               ├── CacheManager
               ├── EmailScheduler
               ├── Database Layer
               ├── API Handlers
               └── Frontend Components
```

**Компоненти та їх відповідальність за вимоги:**

| Компонент | Основні вимоги | Додаткові вимоги |
|-----------|----------------|------------------|
| **WeatherService** | FR-001, FR-007 | NFR-001, NFR-002 |
| **SubscriptionService** | FR-002, FR-004 | NFR-003, NFR-004 |
| **EmailService** | FR-003, FR-007 | NFR-002, NFR-005 |
| **EmailScheduler** | FR-005, FR-006 | NFR-002, NFR-005 |
| **CacheManager** | FR-001 | NFR-001, NFR-004 |
| **Database Layer** | FR-002, FR-004, FR-006 | NFR-003, NFR-004 |
| **API Handlers** | FR-001, FR-002, FR-003, FR-004 | NFR-001, NFR-002 |
| **Frontend** | FR-001, FR-002 | NFR-001, User Experience |

---

## 8. ДОДАТКИ

### Додаток A: Діаграми архітектури

**A.1 Діаграма розгортання системи:**

```text
┌─────────────────────────────────────────────────┐
│                   Client Tier                   │
│  ┌─────────────────────────────────────────┐   │
│  │           Web Browser                   │   │
│  │         (Vue.js SPA)                    │   │
│  └─────────────────────────────────────────┘   │
└─────────────────┬───────────────────────────────┘
                  │ HTTPS/REST
                  ▼
┌─────────────────────────────────────────────────┐
│                Application Tier                 │
│  ┌─────────────────────────────────────────┐   │
│  │         Go Backend Server               │   │
│  │  ┌─────────────────────────────────┐   │   │
│  │  │      API Gateway            │   │   │
│  │  ├─────────────────────────────────┤   │   │
│  │  │    Weather Service              │   │   │
│  │  │    Subscription Service         │   │   │
│  │  │    Email Service                │   │   │
│  │  │    Cache Manager                │   │   │
│  │  │    Email Scheduler              │   │   │
│  │  └─────────────────────────────────┘   │   │
│  └─────────────────────────────────────────┘   │
└─────────────────┬───────────────────────────────┘
                  │
        ┌─────────┼─────────┐
        │         │         │
        ▼         ▼         ▼
┌─────────────┐ ┌───────┐ ┌─────────────┐
│    Data     │ │ Email │ │  External   │
│    Tier     │ │ Tier  │ │ Services    │
│ ┌─────────┐ │ │┌─────┐│ │ ┌─────────┐ │
│ │PostgreSQL│ │ ││SMTP/││ │ │Weather- │ │
│ │Database │ │ ││Send-││ │ │API.com  │ │
│ │         │ │ ││Grid ││ │ │         │ │
│ └─────────┘ │ │└─────┘│ │ └─────────┘ │
└─────────────┘ └───────┘ └─────────────┘
```

**A.2 Діаграма потоку даних (DFD Level 1):**

```text
                    ┌─────────────┐
                    │    User     │
                    └──────┬──────┘
                           │
                    ┌──────▼──────┐
                    │ 1.0 Manage  │
                    │Subscription │
                    └──────┬──────┘
                           │
                  ┌────────▼────────┐
                  │   Subscription  │
                  │     Store       │
                  └────────┬────────┘
                           │
                    ┌──────▼──────┐
                    │ 2.0 Get     │
                    │ Weather     │
                    └──────┬──────┘
                           │
                  ┌────────▼────────┐
                  │  Weather Cache  │
                  └────────┬────────┘
                           │
                    ┌──────▼──────┐
                    │ 3.0 Send    │
                    │ Emails      │
                    └──────┬──────┘
                           │
                  ┌────────▼────────┐
                  │  Email Service  │
                  └─────────────────┘
```

### Додаток B: Специфікації API

**B.1 REST API Endpoints:**

| Method | Endpoint | Request Body | Response | Description |
|--------|----------|--------------|----------|-------------|
| GET | `/weather?city={city}` | None | WeatherData JSON | Отримання поточної погоди |
| POST | `/subscribe` | SubscriptionRequest JSON | Success/Error JSON | Створення підписки |
| GET | `/confirm/{token}` | None | HTML Page | Підтвердження підписки |
| GET | `/unsubscribe/{token}` | None | HTML Page | Скасування підписки |
| GET | `/health` | None | Status JSON | Перевірка стану системи |

**B.2 JSON Schema Definitions:**

```json
{
  "SubscriptionRequest": {
    "type": "object",
    "required": ["email", "city", "frequency"],
    "properties": {
      "email": {
        "type": "string",
        "format": "email",
        "maxLength": 255
      },
      "city": {
        "type": "string",
        "minLength": 2,
        "maxLength": 100
      },
      "frequency": {
        "type": "string",
        "enum": ["hourly", "daily"]
      }
    }
  },
  "WeatherData": {
    "type": "object",
    "properties": {
      "temperature": {
        "type": "number",
        "description": "Temperature in Celsius"
      },
      "humidity": {
        "type": "integer",
        "minimum": 0,
        "maximum": 100
      },
      "description": {
        "type": "string",
        "maxLength": 200
      }
    }
  }
}
```

### Додаток C: Database Schema

**C.1 SQL DDL для створення таблиць:**

```sql
-- Створення таблиці підписок
CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR NOT NULL UNIQUE,
    city VARCHAR NOT NULL,
    frequency VARCHAR NOT NULL,
    confirmed BOOLEAN NOT NULL DEFAULT false,
    token VARCHAR NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT current_timestamp,
    confirmed_at TIMESTAMPTZ
);

-- Create index for better performance
CREATE INDEX IF NOT EXISTS idx_subscriptions_email ON subscriptions(email);
CREATE INDEX IF NOT EXISTS idx_subscriptions_confirmed ON subscriptions(confirmed);
CREATE INDEX IF NOT EXISTS idx_subscriptions_city ON subscriptions(city);
```

### Додаток D: План тестування

**D.1 Типи тестів:**

1. **Unit Tests** - тестування окремих компонентів
   - WeatherService.GetWeather()
   - SubscriptionService.CreateSubscription()
   - EmailService.SendEmail()
   - CacheManager.Get/Set()

2. **Integration Tests** - тестування взаємодії компонентів
   - API endpoints тестування
   - Database операції
   - Email надсилання
   - Cache invalidation

3. **End-to-End Tests** - тестування повного workflow
   - Підписка → Підтвердження → Отримання email
   - Отримання погоди → Кешування → Відповідь

**D.2 Test Coverage Goals:**

- Unit Tests: ≥ 90%
- Integration Tests: ≥ 80%  
- E2E Tests: ≥ 70%

---

## Підпис документу

**Підготовлено:** Ustymenko Roman  
**Переглянуто:** System Architect  
**Затверджено:** Project Manager  

**Дата останнього оновлення:** 09/06/2025  
**Версія документу:** 1.01  

---

*Цей документ є живим документом і буде оновлюватися в процесі розвитку проекту. Всі зміни повинні бути задокументовані та затверджені.*
