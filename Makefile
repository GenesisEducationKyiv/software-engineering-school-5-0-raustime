.PHONY: build up down restart logs install-tools bench bench-http bench-grpc up-bench check-ghz

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

install-tools:
	@echo "🔧 Installing wrk, hey, and ghz if missing..."

	@if ! command -v wrk >/dev/null 2>&1; then \
		echo 'Installing wrk...'; \
		sleep 2; \
		if [ "$(uname)" = "Darwin" ]; then brew install wrk; \
		else sudo apt-get update && sudo apt-get install -y wrk; fi; \
	else echo '✔ wrk already installed'; fi

	@if ! command -v hey >/dev/null 2>&1; then \
		echo 'Installing hey...'; \
		sleep 2; \
		if [ "$(uname)" = "Darwin" ]; then brew install hey; \
		else \
			curl -L https://github.com/rakyll/hey/releases/download/v0.1.4/hey_linux_amd64 -o /tmp/hey && \
			sudo mv /tmp/hey /usr/local/bin/hey && \
			sudo chmod +x /usr/local/bin/hey; \
		fi; \
	else echo '✔ hey already installed'; fi

	@if ! command -v ghz >/dev/null 2>&1; then \
		echo 'Installing ghz...'; \
		sleep 2; \
		GOBIN=$HOME/go/bin go install github.com/bojand/ghz/cmd/ghz@latest; \
		echo '⚠️  Додайте до PATH якщо потрібно: export PATH=$HOME/go/bin:$PATH'; \
	else echo '✔ ghz already installed'; fi

	@$(MAKE) check-ghz

check-ghz:
	@echo "🔍 Перевірка ghz..."
	@if command -v ghz >/dev/null 2>&1; then \
		echo "✔ ghz знайдено у $$PATH: $$(which ghz)"; \
	else \
		echo "❌ ghz не знайдено в \\$$PATH."; \
		echo "👉 Спробуйте вручну додати до PATH:"; \
		echo "   export PATH=\$$PATH:\$$HOME/go/bin"; \
		exit 1; \
	fi

bench: install-tools bench-http bench-grpc

bench-http:
	@echo "\n🔵 Benchmarking HTTP (REST)..."
	@wrk -t4 -c800 -d15s http://localhost:8080/api/weather?city=Kyiv

bench-grpc:
	@echo "\n🟣 Benchmarking ConnectRPC over HTTP/2 using ghz..."
	@ghz \
	  --insecure \
	  --proto weather_microservice/proto/weather/v1/weather.proto \
	  --call weather.WeatherService.GetWeather \
	  --data '{"city":"Kyiv"}' \
	  --format summary \
	  --concurrency 400 \
	  --total 20000 \
	  localhost:8081
	  
up-bench: up
	@echo "\n⏳ Waiting for services to start..." && sleep 5
	@$(MAKE) bench