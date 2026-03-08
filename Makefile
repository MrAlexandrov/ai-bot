.PHONY: setup up down build pull logs restart shell-ollama

# Подготовка: копируем .env и скачиваем зависимости Go
setup:
	@if [ ! -f .env ]; then cp .env.example .env && echo "Создан .env — заполни TELEGRAM_TOKEN"; fi
	cd backend && go mod tidy
	cd bot && go mod tidy

# Запуск всех сервисов
up:
	docker compose up -d
	@echo ""
	@echo "Сервисы запущены:"
	@echo "  OpenClaw Web UI → http://localhost:3000"
	@echo "  Backend API     → http://localhost:8080"
	@echo "  Ollama API      → http://localhost:11434"

# Остановка
down:
	docker compose down

# Сборка образов
build:
	docker compose build --no-cache

# Скачать модель в Ollama (обязательно после первого запуска)
pull-model:
	@MODEL=$$(grep OLLAMA_MODEL .env 2>/dev/null | cut -d= -f2 | tr -d ' ' | grep -v '^#'); \
	MODEL=$${MODEL:-qwen2.5:7b}; \
	echo "Скачиваем модель: $$MODEL"; \
	docker exec ollama ollama pull $$MODEL

# Логи всех сервисов
logs:
	docker compose logs -f

# Логи отдельного сервиса (make logs-backend)
logs-%:
	docker compose logs -f $*

# Перезапуск сервиса (make restart-bot)
restart-%:
	docker compose restart $*

# Открыть shell внутри Ollama
shell-ollama:
	docker exec -it ollama sh

# Список скачанных моделей
models:
	docker exec ollama ollama list

# Статус всех контейнеров
status:
	docker compose ps
