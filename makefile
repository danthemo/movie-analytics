.PHONY: help setup run run-go run-python run-db stop clean

help:
	@echo "Available commands:"
	@echo "  make setup      - Setup everything (venv + deps)"
	@echo "  make run        - Run all services in Docker"
	@echo "  make run-db     - Run Postgres"
	@echo "  make run-python - Run Python FastAPI (with venv)"
	@echo "  make run-go     - Run Go server"
	@echo "  make stop       - Stop Postgres"

setup:
	@echo "Setting up Python venv..."
	cd internal/pythonclient/pythonParser && python3 -m venv venv
	cd internal/pythonclient/pythonParser && . venv/bin/activate && pip install -r requirements.txt && playwright install chromium
	@echo "Setting up Go modules..."
	go mod download
	@echo "Setup complete!"

run-db:
	docker compose up postgres -d
	@echo "Postgres running on :5432"

run-python:
	cd internal/pythonclient/pythonParser && . venv/bin/activate && OKKO_HEADLESS=false uvicorn service:app --host 0.0.0.0 --port 8000

run-go:
	go run cmd/server/main.go

run:
	docker compose up --build

stop:
	docker compose down
	@echo "Stopped!"

clean:
	docker compose down -v
	rm -rf internal/pythonclient/pythonParser/venv
