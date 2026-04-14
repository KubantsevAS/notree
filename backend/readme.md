# Backend README

## Architecture Overview

The project follows a standard Go application structure with layers (cmd, internal, pkg).

```text
notree/
├── backend/
│   ├── cmd/
│   │   └── main.go              # Entry point: init config, DB, router, start server
│   ├── internal/                # Internal packages (not exported)
│   │   ├── config/              # Configuration (cleanenv)
│   │   ├── db/                  # DB connection (pgx + sqlc)
│   │   │   └── sqlc/
│   │   │       └── models.go    # Generated models and queries from sqlc
│   │   ├── http/
│   │   │   ├── handlers/        # HTTP handlers (Chi router)
│   │   │   │   └── auth.go
│   │   │   ├── dto/             # DTOs for requests/responses
│   │   │   │   └── auth.go
│   │   │   └── httputil/
│   │   │       └── response.go  # HTTP response utilities
│   │   └── service/
│   ├── pkg/
│   │   └── jwt/                 # JWT utilities
│   │       └── jwt.go
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
└── frontend/                    # TypeScript + Vite
```

## Tech Stack

- **Language**: Go
- **Configuration**: cleanenv
- **Logger**: log/slog
- **Database**: PostgreSQL + pgx + sqlc (query generation)
- **Web Server**: Chi router + net/http (chi render)
- **Validation**: go-playground/validator
- **Auth**: JWT
- **Deployment**: Docker (Dockerfile in backend)
- **Development**: Air (hot reload, planned)

Server starts from `cmd/main.go`. Auth implemented in `internal/http/handlers/auth.go` using `pkg/jwt`.
