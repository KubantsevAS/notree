# Todo

- [x] init config: cleanenv
- [x] init storage: postgres
- [x] init logger: log/slog
- [x] работа с Postgres:  pgx + sqlc
- [x] init router: chi, net/http (chi render)
- [x] run server

- [ ] Валидация: go-playground/validator
- [ ] air

## Architecture - draft

repo-name/
├── docker-compose.yml
├── .env.example
├── .gitignore
│
├── backend/
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/
│   │   └── main.go
│   └── internal/
│       └── ...
│
├── frontend/
│   ├── Dockerfile
│   ├── package.json
│   ├── vite.config.ts
│   ├── index.html
│   └── src/
│       └── ...
│
└── db/
    ├── init.sql
    └── migrations/

### Backend - draft

backend/
├── cmd/
│   └── main.go           # Точка входа: инициализация БД, роутера, запуск сервера
├── internal/             # Внутренний пакет (не доступен извне)
│   ├── config/           # Чтение конфигов (порт, строка БД)
│   ├── db/               # Инициализация pgxpool, запуск миграций
│   ├── http/
│   │   ├── handlers/     # Здесь живет chi-роутер и обработчики HTTP запросов (парсинг JSON, вызов сервисов, отдача статусов)
│   │   └── middleware/    # CORS, JWT авторизация (в будущем)
│   ├── service/          # Бизнес-логика. (Например: "Перед созданием ноды проверь, что родитель существует")
│   └── repository/       # Слой данных. ТОЛЬКО здесь вызываются сгенерированные sqlc функции.
├── db/                   # Папка для sqlc
│   ├── queries/          # .sql файлы (GetChildren, CreateNode и т.д.)
│   └── sqlc.yaml         # Конфиг sqlc
├── migrations/           # Файлы .up.sql и .down.sql для golang-migrate
├── Dockerfile
└── go.mod
