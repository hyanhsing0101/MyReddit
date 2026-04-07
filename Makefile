.PHONY: help deps-up deps-down deps-logs db backend frontend dev test lint fe-lint \
	mobile-pub-get mobile-clean mobile-analyze mobile-build-apk mobile-run

BACKEND_DIR := backend
FRONTEND_DIR := frontend
MOBILE_DIR := mobile
DOCKER_COMPOSE_FILE := infra/docker/docker-compose.yaml
SQL_DIR := $(BACKEND_DIR)/sql
RESET_SQL := $(SQL_DIR)/reset_schema.sql
SEED_SQL := $(SQL_DIR)/seed_default.sql

# 与 docker-compose / backend/conf/config.yaml 默认一致；仅 make db 时用（容器内 psql）
PGUSER ?= hyan
PGPASSWORD ?= root
PGDATABASE ?= myreddit

help:
	@echo "Available targets:"
	@echo "  make deps-up       - Start Postgres/Redis via Docker Compose"
	@echo "  make deps-down     - Stop Postgres/Redis"
	@echo "  make deps-logs     - View Docker service logs"
	@echo "  make db            - 执行 reset_schema.sql + seed_default.sql（经 Docker 容器 postgres，需先 deps-up）"
	@echo "  make backend       - Run backend API server"
	@echo "  make frontend      - Run Next.js dev server"
	@echo "  make dev           - Start backend + frontend together"
	@echo "  make test          - Run backend tests"
	@echo "  make lint          - Run backend go vet"
	@echo "  make fe-lint       - Run frontend ESLint"
	@echo ""
	@echo "Flutter (mobile/):"
	@echo "  make mobile-pub-get   - flutter pub get"
	@echo "  make mobile-clean     - flutter clean"
	@echo "  make mobile-analyze   - flutter analyze"
	@echo "  make mobile-build-apk - flutter build apk (debug)"
	@echo "  make mobile-run       - flutter run (需已开模拟器；交互式，支持 r/R/q)"
	@echo ""
	@echo "提示: 模拟器访问本机后端用 http://10.0.2.2:<端口>"

deps-up:
	docker compose -f $(DOCKER_COMPOSE_FILE) up -d

deps-down:
	docker compose -f $(DOCKER_COMPOSE_FILE) down

deps-logs:
	docker compose -f $(DOCKER_COMPOSE_FILE) logs -f

# 清空并重建表结构，再灌入默认种子数据（破坏性操作；不依赖本机 psql）
db:
	@echo "[db] Docker postgres：$(RESET_SQL) -> $(SEED_SQL)"
	cat $(RESET_SQL) | docker compose -f $(DOCKER_COMPOSE_FILE) exec -T postgres \
		env PGPASSWORD="$(PGPASSWORD)" psql -U "$(PGUSER)" -d "$(PGDATABASE)" -v ON_ERROR_STOP=1
	cat $(SEED_SQL) | docker compose -f $(DOCKER_COMPOSE_FILE) exec -T postgres \
		env PGPASSWORD="$(PGPASSWORD)" psql -U "$(PGUSER)" -d "$(PGDATABASE)" -v ON_ERROR_STOP=1

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

# --- Flutter / Android ---

mobile-pub-get:
	cd $(MOBILE_DIR) && flutter pub get

mobile-clean:
	cd $(MOBILE_DIR) && flutter clean

mobile-analyze:
	cd $(MOBILE_DIR) && flutter analyze

mobile-build-apk:
	cd $(MOBILE_DIR) && flutter build apk --debug

mobile-run:
	cd $(MOBILE_DIR) && flutter run
