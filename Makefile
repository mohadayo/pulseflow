.PHONY: up down test test-python test-go test-ts lint build

up:
	docker compose up --build -d

down:
	docker compose down

build:
	docker compose build

test: test-python test-go test-ts

test-python:
	cd services/analytics-api && pip install -q -r requirements.txt && pytest -v

test-go:
	cd services/event-processor && go test -v ./...

test-ts:
	cd services/api-gateway && npm install --silent && npm test

lint: lint-python lint-go lint-ts

lint-python:
	cd services/analytics-api && pip install -q flake8 && flake8 --max-line-length=120 .

lint-go:
	cd services/event-processor && go vet ./...

lint-ts:
	cd services/api-gateway && npx --yes eslint src/ 2>/dev/null || true

logs:
	docker compose logs -f

health:
	@echo "Gateway:   " && curl -s http://localhost:5000/health | head -1
	@echo "Analytics: " && curl -s http://localhost:5001/health | head -1
	@echo "Processor: " && curl -s http://localhost:5002/health | head -1
