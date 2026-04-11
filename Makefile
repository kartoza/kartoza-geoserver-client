.PHONY: all build build-web build-tui build-frontend clean clean-all dev dev-web dev-tui \
        install test lint format shell migrate kill-server redeploy help \
        docs docs-build

# Version from git tag or commit
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Python settings
PYTHON := python
PIP := pip
DJANGO_SETTINGS := cloudbench.settings.development

# Default target
all: build

# === Python/Django Web Server ===

# Build the web application (frontend + collect static)
build-web: build-frontend
	@echo "Collecting static files..."
	$(PYTHON) manage.py collectstatic --noinput

# Build the React frontend
build-frontend:
	@echo "Building React frontend..."
	cd web && npm install && npm run build
	@echo "Copying frontend to static directory..."
	rm -rf static/*
	cp -r internal/webserver/static/* static/ 2>/dev/null || cp -r web/dist/* static/ 2>/dev/null || true

# Run Django development server
dev-web:
	@echo "Starting Django development server on http://localhost:8080..."
	DJANGO_SETTINGS_MODULE=$(DJANGO_SETTINGS) $(PYTHON) manage.py runserver 0.0.0.0:8080

# Run Django with uvicorn (ASGI)
run-web:
	@echo "Starting Django with uvicorn on http://localhost:8080..."
	uvicorn cloudbench.asgi:application --host 0.0.0.0 --port 8080 --reload

# === TUI Application ===

# Build the TUI application
build-tui:
	@echo "TUI is a Python application, no build step required."
	@echo "Run 'make dev-tui' or 'python -m tui' to start."

# Run TUI in development mode
dev-tui:
	@echo "Starting CloudBench TUI..."
	$(PYTHON) -m tui

# Run TUI
run-tui: dev-tui

# === Combined Build ===

build: build-web build-tui
	@echo "Build complete!"

# === Database Migrations ===

# Create and apply migrations
migrate:
	$(PYTHON) manage.py migrate

# Make migrations
makemigrations:
	$(PYTHON) manage.py makemigrations

# === Django Shell ===

shell:
	$(PYTHON) manage.py shell

# === Testing ===

test:
	@echo "Running pytest..."
	pytest

test-coverage:
	@echo "Running tests with coverage..."
	pytest --cov=apps --cov=cloudbench --cov=tui --cov-report=html

# === Linting and Formatting ===

lint:
	@echo "Running ruff..."
	ruff check apps cloudbench tui
	@echo "Running mypy..."
	mypy apps cloudbench tui

format:
	@echo "Formatting with black..."
	black apps cloudbench tui
	@echo "Sorting imports with ruff..."
	ruff check --fix --select I apps cloudbench tui

# === Documentation ===

docs:
	@echo "Serving documentation on http://localhost:8000..."
	mkdocs serve

docs-build:
	@echo "Building documentation..."
	mkdocs build

# === Cleaning ===

clean:
	@echo "Cleaning..."
	rm -rf static/*
	rm -rf staticfiles/*
	rm -rf web/node_modules
	rm -rf .pytest_cache
	rm -rf .mypy_cache
	rm -rf .ruff_cache
	rm -rf htmlcov
	rm -rf *.egg-info
	rm -rf dist
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete 2>/dev/null || true

clean-all: clean
	@echo "Deep cleaning..."
	rm -rf .venv
	rm -rf node_modules
	rm -rf web/node_modules

# === Server Management ===

kill-server:
	@echo "Killing any running server instances..."
	@fuser -k 8080/tcp 2>/dev/null || true
	@sleep 1

redeploy: kill-server clean build-web
	@echo ""
	@echo "================================================"
	@echo "Build complete! Starting server..."
	@echo "================================================"
	@echo ""
	$(PYTHON) manage.py runserver 0.0.0.0:8080 &
	@sleep 2
	@echo ""
	@echo "Server started on http://localhost:8080"

# === Installation ===

install:
	@echo "Installing dependencies..."
	$(PIP) install -e ".[dev]"

# === Legacy Go Targets (for backward compatibility during migration) ===

build-go-tui:
	@echo "Building Go TUI..."
	go build -ldflags "-X main.version=$(VERSION)" -o bin/kartoza-cloudbench-go .

build-go-web:
	@echo "Building Go Web server..."
	go build -ldflags "-X main.version=$(VERSION)" -o bin/kartoza-cloudbench-web-go ./cmd/web

# === Help ===

help:
	@echo "Kartoza CloudBench - Build Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Django/Python Targets:"
	@echo "  dev-web          Run Django development server (port 8080)"
	@echo "  run-web          Run Django with uvicorn (ASGI)"
	@echo "  dev-tui          Run the Textual TUI"
	@echo "  build-web        Build frontend and collect static files"
	@echo "  build-frontend   Build React frontend only"
	@echo "  migrate          Run database migrations"
	@echo "  shell            Open Django shell"
	@echo ""
	@echo "Testing & Quality:"
	@echo "  test             Run pytest"
	@echo "  test-coverage    Run tests with coverage report"
	@echo "  lint             Run ruff and mypy"
	@echo "  format           Format code with black"
	@echo ""
	@echo "Documentation:"
	@echo "  docs             Serve documentation locally"
	@echo "  docs-build       Build documentation site"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean            Remove build artifacts"
	@echo "  clean-all        Deep clean including venv"
	@echo "  install          Install package with dev dependencies"
	@echo "  kill-server      Kill running server instances"
	@echo "  redeploy         Kill, clean, rebuild, restart"
	@echo ""
	@echo "Legacy Go Targets:"
	@echo "  build-go-tui     Build Go TUI binary"
	@echo "  build-go-web     Build Go web server binary"
	@echo ""
	@echo "  help             Show this help message"
