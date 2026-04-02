.PHONY: help deps-up deps-down deps-logs backend frontend dev test lint fe-lint

BACKEND_DIR := backend
FRONTEND_DIR := frontend
DOCKER_COMPOSE_FILE := infra/docker/docker-compose.yaml

help:
	@echo "Available targets:"
	@echo "  make deps-up    - Start MySQL/Redis via Docker Compose"
	@echo "  make deps-down  - Stop MySQL/Redis"
	@echo "  make deps-logs  - View Docker service logs"
	@echo "  make backend    - Run backend API server"
	@echo "  make frontend   - Run frontend dev server"
	@echo "  make dev        - Start backend + frontend together"
	@echo "  make test       - Run backend tests"
	@echo "  make lint       - Run backend go vet"
	@echo "  make fe-lint    - Run frontend lint"

deps-up:
	docker compose -f $(DOCKER_COMPOSE_FILE) up -d

deps-down:
	docker compose -f $(DOCKER_COMPOSE_FILE) down

deps-logs:
	docker compose -f $(DOCKER_COMPOSE_FILE) logs -f

backend:
	cd $(BACKEND_DIR) && go run main.go -c conf/config.yaml

frontend:
	cd $(FRONTEND_DIR) && npm run dev

# Start both dev servers in one terminal session.
# Ctrl+C stops both child processes.
dev:
	@sh -c '(cd $(BACKEND_DIR) && go run main.go -c conf/config.yaml) & \
	backend_pid=$$!; \
	(cd $(FRONTEND_DIR) && npm run dev) & \
	frontend_pid=$$!; \
	trap "kill $$backend_pid $$frontend_pid" INT TERM; \
	wait $$backend_pid $$frontend_pid'

test:
	cd $(BACKEND_DIR) && go test ./...

lint:
	cd $(BACKEND_DIR) && go vet ./...

fe-lint:
	cd $(FRONTEND_DIR) && npm run lint
