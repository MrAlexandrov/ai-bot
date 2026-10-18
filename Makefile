DOCKER_COMPOSE_BASE = docker compose --project-directory "$(CURDIR)"
DOCKER_COMPOSE = docker compose --project-directory "$(CURDIR)" $(COMPOSE_FLAGS)
include makefiles/compose.mk
.PHONY: setup up down build pull logs restart shell-ollama configure-openclaw

# Подготовка: копируем .env и скачиваем зависимости Go
setup:
	@if [ ! -f .env ]; then cp .env.example .env && echo "Создан .env — заполни TELEGRAM_TOKEN"; fi
	cd backend && go mod tidy
	cd bot && go mod tidy

# Запуск всех сервисов
up:
	$(DOCKER_COMPOSE) up -d
	@echo ""
	@echo "Сервисы запущены:"
	@echo "  OpenClaw Web UI → http://localhost:3000"
	@echo "  Backend API     → http://localhost:8080"
	@echo "  Ollama API      → http://localhost:11434"

# Остановка
down:
	$(DOCKER_COMPOSE) down

# Сборка образов
build:
	$(DOCKER_COMPOSE) build --no-cache

# Скачать модель в Ollama (обязательно после первого запуска)
pull-model:
	@MODEL=$$(grep OLLAMA_MODEL .env 2>/dev/null | cut -d= -f2 | tr -d ' ' | grep -v '^#'); \
	MODEL=$${MODEL:-qwen2.5:7b}; \
	echo "Скачиваем модель: $$MODEL"; \
	$(DOCKER_COMPOSE) exec -T ollama ollama pull $$MODEL

# Логи всех сервисов
logs:
	$(DOCKER_COMPOSE) logs -f

# Логи отдельного сервиса (make logs-backend)
logs-%:
	$(DOCKER_COMPOSE) logs -f $*

# Перезапуск сервиса (make restart-bot)
restart-%:
	$(DOCKER_COMPOSE) restart $*

# Открыть shell внутри Ollama
shell-ollama:
	$(DOCKER_COMPOSE) exec ollama sh

# Список скачанных моделей
models:
	$(DOCKER_COMPOSE) exec -T ollama ollama list

# Статус всех контейнеров
status:
	$(DOCKER_COMPOSE) ps

# Настроить OpenClaw на использование Ollama (запускать после make up)
configure-openclaw:
	@MODEL=$$(grep OLLAMA_MODEL .env 2>/dev/null | cut -d= -f2 | tr -d ' '); \
	MODEL=$${MODEL:-qwen2.5:7b}; \
	echo "Настраиваем OpenClaw → Ollama ($$MODEL)..."; \
	$(DOCKER_COMPOSE) exec -T openclaw openclaw config patch \
	  "{\"models\":{\"providers\":{\"ollama\":{\"baseUrl\":\"http://ollama:11434/v1\",\"apiKey\":\"ollama-local\",\"api\":\"openai-responses\",\"models\":[{\"id\":\"$$MODEL\",\"name\":\"$$MODEL\",\"contextWindow\":32768,\"maxOutput\":8192}]}}}}"; \
	$(DOCKER_COMPOSE) exec -T openclaw openclaw config patch \
	  "{\"agents\":{\"default\":{\"model\":\"ollama/$$MODEL\"}},\"channels\":{\"telegram\":{\"groupPolicy\":\"allowlist\"}}}"; \
	echo "Перезапускаем OpenClaw..."; \
	$(DOCKER_COMPOSE) restart openclaw

# Read-only validation of every tracked Compose combination, without runtime secrets.
.PHONY: compose-layout-check
compose-layout-check:
	$(DOCKER_COMPOSE_BASE) --env-file /dev/null -f compose/docker-compose.yml --profile '*' config --no-interpolate --no-env-resolution --quiet

check: compose-layout-check
