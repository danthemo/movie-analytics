.PHONY: help setup run run-go run-python run-db stop clean

help:
	@echo "Available commands:"
	@echo "  make setup      - Setup everything (venv + deps)"
	@echo "  make run        - Run all services"
	@echo "  make run-db     - Run Postgres"
	@echo "  make run-python - Run Python FastAPI (with venv)"
	@echo "  make run-go     - Run Go server"
	@echo "  make stop       - Stop Postgres"

setup:
	@echo "Setting up Python venv..."
	cd internal/pythonclient/pythonParser && python -m venv venv
	cd internal/pythonclient/pythonParser && .\venv\Scripts\activate && pip install -r requirements.txt
	@echo "Setting up Go modules..."
	go mod download
	@echo "Setup complete!"

run-db:
	docker-compose up postgres -d
	@echo "Postgres running on :5432"

run-python:
	cd internal/pythonclient/pythonParser && .\venv\Scripts\activate && uvicorn service:app --host 0.0.0.0 --port 8000

run-go:
	go run cmd/server/main.go

run: run-db run-python run-go

stop:
	docker-compose down
	@echo "Stopped!"

clean:
	docker-compose down -v
	cd internal/pythonclient/pythonParser && rmdir /s venv
