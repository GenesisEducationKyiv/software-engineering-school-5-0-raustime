.PHONY: build up down restart logs

build:
	docker-compose build

up:
	docker-compose up -d

down:
	docker-compose down

restart: down up

logs:
	docker-compose logs -f --tail=100

logs-%:
	docker-compose logs -f --tail=100 $*

docker-test-unit:
	docker compose run --rm -e TEST_MODULE=mailer_microservice test-runner-unit
	docker compose run --rm -e TEST_MODULE=scheduler_microservice test-runner-unit
	docker compose run --rm -e TEST_MODULE=subscription_microservice test-runner-unit
	docker compose run --rm -e TEST_MODULE=weather_microservice test-runner-unit