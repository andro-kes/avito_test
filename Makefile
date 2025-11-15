.PHONY: build run test lint clean docker-build docker-up docker-down help

# Переменные
BINARY_NAME=server
DOCKER_IMAGE=avito-api
GO_CMD=go
GO_BUILD=$(GO_CMD) build
GO_TEST=$(GO_CMD) test
GO_LINT=golangci-lint

# Цвета для вывода
GREEN=\033[0;32m
YELLOW=\033[0;33m
NC=\033[0m # No Color

help: ## Показать справку
	@echo "$(GREEN)Доступные команды:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-15s$(NC) %s\n", $$1, $$2}'

build: ## Собрать приложение
	@echo "$(GREEN)Сборка приложения...$(NC)"
	$(GO_BUILD) -o $(BINARY_NAME) ./cmd/server
	@echo "$(GREEN)Готово!$(NC)"

run: ## Запустить приложение локально
	@echo "$(GREEN)Запуск приложения...$(NC)"
	$(GO_CMD) run ./cmd/server

test: ## Запустить тесты
	@echo "$(GREEN)Запуск тестов...$(NC)"
	$(GO_TEST) -v ./...

test-coverage: ## Запустить тесты с покрытием
	@echo "$(GREEN)Запуск тестов с покрытием...$(NC)"
	$(GO_TEST) -v -coverprofile=coverage.out ./...
	$(GO_CMD) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Отчет сохранен в coverage.html$(NC)"

lint: ## Запустить линтер
	@echo "$(GREEN)Проверка кода линтером...$(NC)"
	@if ! command -v $(GO_LINT) > /dev/null; then \
		echo "$(YELLOW)Установка golangci-lint...$(NC)"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2; \
	fi
	$(GO_LINT) run ./...
	@echo "$(GREEN)Линтинг завершен!$(NC)"

lint-fix: ## Исправить ошибки линтера автоматически
	@echo "$(GREEN)Исправление ошибок линтера...$(NC)"
	@if ! command -v $(GO_LINT) > /dev/null; then \
		echo "$(YELLOW)Установка golangci-lint...$(NC)"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.55.2; \
	fi
	$(GO_LINT) run --fix ./...
	@echo "$(GREEN)Исправления применены!$(NC)"

clean: ## Очистить артефакты сборки
	@echo "$(GREEN)Очистка...$(NC)"
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	$(GO_CMD) clean ./...
	@echo "$(GREEN)Готово!$(NC)"

docker-build: ## Собрать Docker образ
	@echo "$(GREEN)Сборка Docker образа...$(NC)"
	docker build -t $(DOCKER_IMAGE) .
	@echo "$(GREEN)Готово!$(NC)"

docker-up: ## Запустить через docker-compose
	@echo "$(GREEN)Запуск docker-compose...$(NC)"
	docker compose up -d
	@echo "$(GREEN)Сервисы запущены!$(NC)"
	@echo "$(YELLOW)API доступен на http://localhost:8080$(NC)"

docker-down: ## Остановить docker-compose
	@echo "$(GREEN)Остановка docker-compose...$(NC)"
	docker compose down
	@echo "$(GREEN)Готово!$(NC)"

docker-logs: ## Показать логи docker-compose
	docker compose logs -f

docker-restart: ## Перезапустить docker-compose
	@echo "$(GREEN)Перезапуск docker-compose...$(NC)"
	docker compose restart
	@echo "$(GREEN)Готово!$(NC)"

deps: ## Установить зависимости
	@echo "$(GREEN)Установка зависимостей...$(NC)"
	$(GO_CMD) mod download
	$(GO_CMD) mod tidy
	@echo "$(GREEN)Готово!$(NC)"

fmt: ## Форматировать код
	@echo "$(GREEN)Форматирование кода...$(NC)"
	$(GO_CMD) fmt ./...
	@echo "$(GREEN)Готово!$(NC)"

vet: ## Проверить код через go vet
	@echo "$(GREEN)Проверка кода...$(NC)"
	$(GO_CMD) vet ./...
	@echo "$(GREEN)Готово!$(NC)"