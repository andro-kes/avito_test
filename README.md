# PR Reviewer Assignment Service

Микросервис для автоматического назначения ревьюверов на Pull Request'ы и управления командами.

## Описание

Сервис автоматически назначает до двух активных ревьюверов из команды автора при создании PR, позволяет выполнять переназначение ревьюверов и получать список PR'ов, назначенных конкретному пользователю.

## Требования

- Docker и Docker Compose
- Go 1.24+ (для локальной разработки)

## Быстрый старт

### Запуск через Docker Compose (рекомендуется)

1. Клонируйте репозиторий:
```bash
git clone <repository-url>
cd avito
```

2. Создайте файл `.env` на основе `.env.example`:
```bash
cp .env.example .env
```

Отредактируйте `.env` при необходимости (по умолчанию используются значения из `.env.example`).

3. Запустите сервис:
```bash
docker compose up
```

Или в фоновом режиме:
```bash
docker compose up -d
```

4. Проверьте, что сервис запущен:
```bash
curl http://localhost:8080/team/get?team_name=test
```

Сервис будет доступен на `http://localhost:8080`

**Важно:** Миграции применяются автоматически при старте сервиса. PostgreSQL должен быть готов перед запуском API (используется healthcheck).

### Локальная разработка

1. Установите зависимости:
```bash
make deps
# или
go mod download
```

2. Запустите PostgreSQL локально или через Docker:
```bash
docker run -d --name postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=postgres \
  -p 5432:5432 \
  postgres:16-alpine
```

Или используйте только БД из docker-compose:
```bash
docker compose up postgres -d
```

3. Установите переменные окружения:
```bash
export DB_URL="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
export SERVE_PORT=8080
export ADMIN_TOKEN=admin-token-123
```

Или создайте `.env` файл:
```bash
cp .env.example .env
# Отредактируйте .env для локальной разработки
```

4. Запустите сервис:
```bash
make run
# или
go run cmd/server/main.go
```

Сервис запустится на порту 8080, миграции применятся автоматически.

## API Endpoints

### Команды

- `POST /team/add` - Создать команду с участниками
- `GET /team/get?team_name=<name>` - Получить команду с участниками

### Пользователи

- `POST /users/setIsActive` - Установить флаг активности пользователя (требует ADMIN_TOKEN)
- `GET /users/getReview?user_id=<id>` - Получить PR'ы, где пользователь назначен ревьювером

### Pull Requests

- `POST /pullRequest/create` - Создать PR и автоматически назначить ревьюверов
- `POST /pullRequest/merge` - Пометить PR как MERGED (идемпотентная операция)
- `POST /pullRequest/reassign` - Переназначить ревьювера

Полная спецификация API доступна в `internal/api/openAPI.yml`

## Переменные окружения

| Переменная | Описание | Значение по умолчанию |
|------------|----------|----------------------|
| `DB_URL` | URL подключения к PostgreSQL | `postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable` |
| `SERVE_PORT` | Порт для HTTP сервера | `8080` |
| `SHUTDOWN_TIMEOUT` | Таймаут graceful shutdown (секунды) | `5` |
| `ADMIN_TOKEN` | Токен для админских операций (используется в заголовке `Authorization: Bearer <token>`) | - |
| `POSTGRES_USER` | Пользователь PostgreSQL (для docker-compose) | `postgres` |
| `POSTGRES_PASSWORD` | Пароль PostgreSQL (для docker-compose) | `postgres` |
| `POSTGRES_DB` | Имя базы данных (для docker-compose) | `postgres` | 

## Makefile команды

```bash
make help       # Показать справку по всем командам
make build      # Собрать приложение
make run        # Запустить приложение локально
make test       # Запустить тесты
make lint       # Запустить линтер
make lint-fix   # Исправить ошибки линтера автоматически
make clean      # Очистить артефакты сборки
make deps       # Установить зависимости
make fmt        # Форматировать код
make vet        # Проверить код через go vet
make docker-build # Собрать Docker образ
make docker-up  # Запустить через docker-compose
make docker-down # Остановить docker-compose
make docker-logs # Показать логи docker-compose
make docker-restart # Перезапустить docker-compose
```

## Структура проекта

```
.
├── cmd/
│   └── server/          # Точка входа приложения
├── internal/
│   ├── api/             # OpenAPI спецификация
│   ├── config/          # Конфигурация
│   ├── errors/          # Обработка ошибок
│   ├── http/            # HTTP handlers и middleware
│   ├── log/             # Логирование
│   ├── migrations/      # Миграции БД (встроенные)
│   ├── models/          # Доменные модели
│   ├── repo/            # Репозитории для работы с БД
│   └── service/         # Бизнес-логика
├── migrations/           # SQL миграции (исходные файлы)
├── docker-compose.yml    # Docker Compose конфигурация
├── Dockerfile           # Docker образ
└── Makefile             # Команды сборки
```

## Допущения и решения

### 2. Логика переназначения

**Вопрос:** Из какой команды выбирать нового ревьювера при переназначении?

**Решение:** Согласно требованиям, новый ревьювер выбирается из команды заменяемого ревьювера. Это позволяет сохранить контекст команды при переназначении.

### 3. Обработка отсутствия кандидатов

**Вопрос:** Что делать, если нет доступных кандидатов для переназначения?

**Решение:** Возвращается ошибка `NO_CANDIDATE` согласно спецификации. При создании PR, если доступных ревьюверов меньше двух, назначается доступное количество (0 или 1).

### 4. Идемпотентность merge

**Решение:** Операция merge использует `COALESCE(merged_at, NOW())`, что гарантирует идемпотентность - повторный вызов не изменяет состояние и возвращает актуальные данные.

### 5. Миграции

**Решение:** Миграции встроены в бинарник через `embed.FS` и применяются автоматически при старте сервиса. Это упрощает развертывание и гарантирует актуальность схемы БД.

### 6. Формат дат в ответах API

**Вопрос:** Нужно ли возвращать `createdAt` и `mergedAt` в ответах?

**Решение:** Поля `createdAt` и `mergedAt` определены в модели, но в текущей реализации не всегда возвращаются в ответах.

## Тестирование

### Примеры запросов

1. Создать команду:
```bash
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {"user_id": "u1", "username": "Alice", "is_active": true},
      {"user_id": "u2", "username": "Bob", "is_active": true}
    ]
  }'
```

2. Создать PR:
```bash
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1",
    "pull_request_name": "Add feature",
    "author_id": "u1"
  }'
```

3. Получить PR пользователя:
```bash
curl "http://localhost:8080/users/getReview?user_id=u2"
```

## Производительность

- Объём данных: до 20 команд, до 200 пользователей
- Целевой RPS: 5
- SLI времени ответа: 300 мс
- SLI успешности: 99.9%

## Разработка

### Добавление новой миграции

1. Создайте файлы в `migrations/`:
   - `000002_<name>.up.sql` - миграция вверх
   - `000002_<name>.down.sql` - миграция вниз

2. Скопируйте файлы в `internal/migrations/migrations/` (для embed)

3. Миграции применятся автоматически при следующем запуске

### Линтинг

```bash
make lint        # Проверить код
make lint-fix    # Исправить ошибки автоматически
```

Используется `golangci-lint` с конфигурацией из `.golangci.yml`.

#### Конфигурация линтера

Проект использует `golangci-lint` с настройками в файле `.golangci.yml`. 

**Включенные линтеры:**

- **Базовые:** `errcheck`, `gosimple`, `govet`, `staticcheck`, `typecheck`, `unused`
- **Стиль кода:** `gofmt`, `goimports`, `gci`, `misspell`
- **Качество:** `gocritic`, `gocyclo`, `gosec`, `dupl`
- **Документация:** `godot`, `godox`
- **Производительность:** `prealloc`

**Основные правила:**

- Максимальная цикломатическая сложность: 15
- Проверка обработки ошибок обязательна
- Проверка безопасности кода (gosec)
- Автоматическая сортировка импортов
- Проверка документации для экспортируемых функций

**Исключения:**

- Тестовые файлы (`*_test.go`) имеют более мягкие правила
- Файлы миграций исключены из проверки
- Некоторые предупреждения о неиспользуемых импортах игнорируются для библиотек логирования

Для настройки линтера отредактируйте файл `.golangci.yml`.

Подробная документация по конфигурации линтера: [docs/LINTER.md](docs/LINTER.md)

## Проверка работы

После запуска сервиса можно проверить его работу:

1. Создать команду:
```bash
curl -X POST http://localhost:8080/team/add \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {"user_id": "u1", "username": "Alice", "is_active": true},
      {"user_id": "u2", "username": "Bob", "is_active": true},
      {"user_id": "u3", "username": "Charlie", "is_active": true}
    ]
  }'
```

2. Создать PR (автоматически назначатся ревьюверы):
```bash
curl -X POST http://localhost:8080/pullRequest/create \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1",
    "pull_request_name": "Add feature",
    "author_id": "u1"
  }'
```

3. Получить PR пользователя:
```bash
curl "http://localhost:8080/users/getReview?user_id=u2"
```

4. Переназначить ревьювера:
```bash
curl -X POST http://localhost:8080/pullRequest/reassign \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1",
    "old_user_id": "u2"
  }'
```

5. Объединить PR:
```bash
curl -X POST http://localhost:8080/pullRequest/merge \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1"
  }'
```

## Остановка сервиса

Для остановки docker-compose:
```bash
docker compose down
```

Для остановки с удалением volumes (очистка БД):
```bash
docker compose down -v
```

## Лицензия

Тестовое задание для стажёра Backend (осенняя волна 2025)