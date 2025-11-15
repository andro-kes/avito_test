pr-reviewer-service/                # корень репозитория
├─ .github/
│  └─ workflows/
│     └─ ci.yml                     # CI: build, lint, test, (e2e)
├─ cmd/
│  └─ server/
│     └─ main.go                    # точка входа приложения
├─ build/
│  └─ docker/                       # вспомогательные скрипты для сборки/контейнеров (опционально)
├─ internal/
│  ├─ api/
│  │  └─ openapi.yaml               # оригинальная OpenAPI-спецификация
│  ├─ http/
│  │  ├─ handlers/
│  │  │  ├─ team_handlers.go
│  │  │  ├─ user_handlers.go
│  │  │  ├─ pr_handlers.go
│  │  │  └─ health.go
│  │  ├─ middleware/
│  │  │  └─ auth.go                 # ADMIN_TOKEN middleware, logging, etc.
│  │  └─ router.go                  # регистрация маршрутов
│  ├─ service/                      # usecase / бизнес-логика
│  │  ├─ pr_service.go
│  │  ├─ team_service.go
│  │  └─ user_service.go
│  ├─ repository/                    # интерфейсы и реализации доступа к БД
│  │  ├─ repo.go                     # интерфейсы (teams, users, prs)
│  │  └─ pg/                         # реализация для Postgres
│  │     ├─ pg_repo.go
│  │     └─ migrations_helper.go
│  ├─ model/                         # доменные модели / DTO
│  │  ├─ team.go
│  │  ├─ user.go
│  │  └─ pullrequest.go
│  └─ util/
│     ├─ logger.go
│     ├─ random.go                   # helper для случайного выбора ревьюверов
│     └─ tx_manager.go               # транзакции, обёртки
├─ migrations/
│  ├─ 000001_create_schema.up.sql
│  ├─ 000001_create_schema.down.sql
│  └─ ...                           # каждая миграция отдельным файлом (migrate)
├─ scripts/
│  ├─ apply_migrations.sh
│  └─ seed.sh                        # seed данных (teams, users)
├─ configs/
│  └─ config.yaml.example            # пример конфигурации (env / yaml)
├─ docker-compose.yml
├─ docker-compose.e2e.yml            # отдельный compose для e2e (db_e2e, api_e2e, tests)
├─ Dockerfile
├─ Makefile
├─ .env.example
├─ .gitignore
├─ go.mod
├─ go.sum
├─ README.md
├─ CHECKLIST.md
├─ docs/
│  ├─ ARCHITECTURE.md
│  └─ ASSUMPTIONS.md                 # допущения (например old_user_id)
├─ test/
│  ├─ unit/
│  │  └─ ...                         # unit tests (usecase level)
│  └─ integration/
│     └─ ...                         # integration / e2e tests
└─ mocks/
   └─ ...                            # сгенерированные моки (mockery)