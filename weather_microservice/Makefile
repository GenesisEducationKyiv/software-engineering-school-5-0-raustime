APP_NAME=weather_microservice
DOCKER_IMAGE=weather_microservice:latest

build:
	go build -o $(APP_NAME) ./cmd/main.go

run:
	go run ./cmd/main.go

test:
	go test -v ./...

docker-build:
	docker build -t $(DOCKER_IMAGE) .

docker-run:
	docker run --rm -p 8080:8080 --env-file .env $(DOCKER_IMAGE)

compose-up:
	docker-compose -f docker-compose.weather.yml up --build

compose-down:
	docker-compose -f docker-compose.weather.yml down --volumes --remove-orphans
	docker container rm -f weather_redis || true
