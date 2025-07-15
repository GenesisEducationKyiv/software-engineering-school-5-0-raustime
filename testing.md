# Testing Guide

## Юніт тести

docker-compose --profile testing up test-runner-unit

## Інтеграційні тести

docker-compose --profile testing up test-runner-integration

## E2E тести

docker-compose --profile testing up test-runner-e2e

## З детальними логами

docker-compose --profile testing up --no-deps test-runner-unit

## Без фонового режиму (бачити вивід в реальному часі)

docker-compose --profile testing up --no-deps --no-detach test-runner-unit

## Якщо змінили код і хочете перебілдити

docker-compose --profile testing up --build test-runner-unit

## Запуск тестів один раз і вихід

docker-compose --profile testing run --rm test-runner-unit
docker-compose --profile testing run --rm test-runner-integration
docker-compose --profile testing run --rm test-runner-e2e

## Спочатку піднімаємо основні сервіси

docker-compose up -d

## Потім запускаємо тести

docker-compose --profile testing run --rm test-runner-unit
docker-compose --profile testing run --rm test-runner-integration
docker-compose --profile testing run --rm test-runner-e2e

## Зупиняємо все

docker-compose down

## Тестування архітектури

go test -v -timeout 2m ./docs/Architecture/tests/

