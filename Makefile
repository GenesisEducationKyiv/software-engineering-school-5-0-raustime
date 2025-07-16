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
