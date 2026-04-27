# PLAN BE-00 — Backend Foundation

> **Scope:** Setup fondasi backend: config, database pool, error types, middleware stack, server entry point.
> **Deliverable:** Server Echo berjalan di Docker, `/health` endpoint OK, middleware chain terpasang.
> **Referensi:** docs/PRISM_Backend_Structure.md, docs/PRISM_Error_Handling.md

---

## Instruksi untuk Codex

Baca dulu:
- `docs/PRISM_Backend_Structure.md` — struktur folder, aturan sqlc, layer architecture
- `docs/PRISM_Error_Handling.md` — custom error types dan pg error mapping

Semua file dibuat di dalam `prism-backend/`.

---

## Task 1 — sqlc.yaml

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "sql/queries/"
    schema: "sql/schema/"
    gen:
      go:
        package: "queries"
        out: "internal/database/queries"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
```

---

## Task 2 — Makefile

```makefile
.PHONY: run build generate migrate-up migrate-down test test-short fmt lint

run:
	go run ./cmd/api

build:
	go build -o bin/prism ./cmd/api

generate:
	sqlc generate

migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1

test:
	go test ./...

test-short:
	go test ./... -short

fmt:
	gofmt -w .
```

---

## Task 3 — .air.toml

```toml
root = "."
tmp_dir = "tmp"

[build]
  cmd = "go build -o ./tmp/main ./cmd/api"
  bin = "./tmp/main"
  include_ext = ["go"]
  exclude_dir = ["tmp", "vendor", "internal/database/queries"]
  delay = 500

[log]
  time = true
```

---

## Task 4 — internal/config/config.go

Load semua environment variables via `viper.AutomaticEnv()`:

```go
type Config struct {
    Port          string
    Env           string
    DatabaseURL   string
    JWTSecret     string
    JWTExpiresIn  int
}

func Load() (*Config, error) {
    viper.AutomaticEnv()
    viper.SetDefault("PORT", "8080")
    viper.SetDefault("ENV", "development")
    viper.SetDefault("JWT_EXPIRES_IN", 86400)

    cfg := &Config{
        Port:         viper.GetString("PORT"),
        Env:          viper.GetString("ENV"),
        DatabaseURL:  viper.GetString("DATABASE_URL"),
        JWTSecret:    viper.GetString("JWT_SECRET"),
        JWTExpiresIn: viper.GetInt("JWT_EXPIRES_IN"),
    }

    if cfg.DatabaseURL == "" {
        return nil, fmt.Errorf("DATABASE_URL is required")
    }
    if cfg.JWTSecret == "" {
        return nil, fmt.Errorf("JWT_SECRET is required")
    }
    return cfg, nil
}
```

---

## Task 5 — internal/database/db.go

Init pgxpool connection pool:

```go
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
    cfg, err := pgxpool.ParseConfig(databaseURL)
    if err != nil {
        return nil, fmt.Errorf("parse db config: %w", err)
    }

    cfg.MaxConns = 20
    cfg.MinConns = 2
    cfg.MaxConnLifetime = 1 * time.Hour
    cfg.MaxConnIdleTime = 30 * time.Minute

    pool, err := pgxpool.NewWithConfig(ctx, cfg)
    if err != nil {
        return nil, fmt.Errorf("create pool: %w", err)
    }

    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("ping db: %w", err)
    }

    return pool, nil
}
```

---

## Task 6 — internal/errors/errors.go

Custom error types dan pg error mapping:

```go
package errors

import (
    "errors"
    "net/http"
    "github.com/jackc/pgx/v5/pgconn"
)

type AppError struct {
    Code       string      `json:"code"`
    Message    string      `json:"message"`
    StatusCode int         `json:"-"`
    Details    []FieldError `json:"details,omitempty"`
}

type FieldError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}

func (e *AppError) Error() string { return e.Message }

func NotFound(msg string) *AppError {
    return &AppError{Code: "NOT_FOUND", Message: msg, StatusCode: http.StatusNotFound}
}

func Conflict(msg string) *AppError {
    return &AppError{Code: "CONFLICT", Message: msg, StatusCode: http.StatusConflict}
}

func Validation(fields ...FieldError) *AppError {
    return &AppError{Code: "VALIDATION_ERROR", Message: "Input tidak valid", StatusCode: http.StatusBadRequest, Details: fields}
}

func BusinessRule(msg string) *AppError {
    return &AppError{Code: "BUSINESS_RULE_ERROR", Message: msg, StatusCode: http.StatusUnprocessableEntity}
}

func Unauthorized(msg string) *AppError {
    return &AppError{Code: "UNAUTHORIZED", Message: msg, StatusCode: http.StatusUnauthorized}
}

func Forbidden(msg string) *AppError {
    return &AppError{Code: "FORBIDDEN", Message: msg, StatusCode: http.StatusForbidden}
}

func FromPgError(err error) *AppError {
    var pgErr *pgconn.PgError
    if !errors.As(err, &pgErr) {
        return &AppError{Code: "INTERNAL_ERROR", Message: "Terjadi kesalahan database", StatusCode: 500}
    }
    switch pgErr.Code {
    case "23505":
        return &AppError{Code: "CONFLICT", Message: "Data sudah ada", StatusCode: 409}
    case "23503":
        return &AppError{Code: "VALIDATION_ERROR", Message: "Referensi data tidak valid", StatusCode: 400}
    case "23514":
        return &AppError{Code: "VALIDATION_ERROR", Message: "Data tidak memenuhi aturan", StatusCode: 400}
    default:
        return &AppError{Code: "INTERNAL_ERROR", Message: "Terjadi kesalahan database", StatusCode: 500}
    }
}
```

---

## Task 7 — internal/middleware/

**`logger.go`** — zerolog request logging:
- Log method, path, status, latency, request_id
- Skip `/health` endpoint

**`auth.go`** — JWT validation:
- Parse `Authorization: Bearer <token>`
- Validate signature dan expiry via `golang-jwt`
- Inject `AuthUser` ke echo context: `c.Set("user", user)`
- Return 401 jika token tidak valid/expired

**`permission.go`** — CRUD permission check:
```go
func Require(module, action string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            user := c.Get("user").(*model.AuthUser)
            // ADMIN selalu lolos
            if user.Role == "ADMIN" {
                return next(c)
            }
            // Cek user_permission di DB
            // Jika tidak ada entri → 403 FORBIDDEN
            // Jika ada tapi action = false → 403 FORBIDDEN
        }
    }
}
```

**`audit.go`** — set user aktif untuk trigger audit:
```go
// Set LOCAL hanya berlaku per transaksi
conn.Exec(ctx, "SET LOCAL app.current_user_id = $1", user.ID.String())
```

---

## Task 8 — internal/model/common.go

Shared request/response types:

```go
type PaginationParams struct {
    Page  int    `query:"page"`
    Limit int    `query:"limit"`
    Sort  string `query:"sort"`
    Order string `query:"order"`
}

type PaginationMeta struct {
    Page       int `json:"page"`
    Limit      int `json:"limit"`
    Total      int `json:"total"`
    TotalPages int `json:"total_pages"`
}

type ListResponse[T any] struct {
    Data []T            `json:"data"`
    Meta PaginationMeta `json:"meta"`
}

type DataResponse[T any] struct {
    Data T `json:"data"`
}

type ErrorDetail struct {
    Field   string `json:"field"`
    Message string `json:"message"`
}
```

---

## Task 9 — internal/sse/broker.go

SSE broker sederhana berbasis channel Go:

```go
type Broker struct {
    clients    map[string]chan []byte
    register   chan client
    unregister chan string
    broadcast  chan []byte
    mu         sync.RWMutex
}

func (b *Broker) Subscribe(userID string) (<-chan []byte, func()) { ... }
func (b *Broker) Publish(event string, data any) { ... }  // marshal ke JSON, broadcast ke semua client
func (b *Broker) Run() { ... }  // select loop
```

---

## Task 10 — cmd/api/main.go

Entry point — wiring semua komponen:

```go
func main() {
    // 1. Load config
    cfg, err := config.Load()

    // 2. Init DB pool
    pool, err := database.NewPool(ctx, cfg.DatabaseURL)

    // 3. Init sqlc queries
    q := queries.New(pool)

    // 4. Init SSE broker
    broker := sse.NewBroker()
    go broker.Run()

    // 5. Init Echo
    e := echo.New()
    e.HTTPErrorHandler = middleware.ErrorHandler  // custom error handler

    // 6. Global middleware
    e.Use(middleware.Logger())
    e.Use(echomiddleware.Recover())
    e.Use(echomiddleware.CORS())

    // 7. Health endpoint (no auth)
    e.GET("/health", func(c echo.Context) error {
        return c.JSON(200, map[string]string{"status": "ok"})
    })

    // 8. API v1 group
    api := e.Group("/api/v1")
    api.Use(middleware.Auth(cfg.JWTSecret))
    api.Use(middleware.SetAuditUser(pool))

    // 9. Register routes (akan ditambah di plan berikutnya)
    // handler.RegisterAuthRoutes(api, q, cfg)
    // handler.RegisterMasterRoutes(api, q)
    // ...

    // 10. SSE endpoint
    e.GET("/events", handler.SSEHandler(broker))

    // 11. Start server
    e.Logger.Fatal(e.Start(":" + cfg.Port))
}
```

---

## Task 11 — Custom Error Handler

```go
// internal/middleware/error_handler.go
func ErrorHandler(err error, c echo.Context) {
    var appErr *apperrors.AppError
    var httpErr *echo.HTTPError

    switch {
    case errors.As(err, &appErr):
        c.JSON(appErr.StatusCode, map[string]any{"error": appErr})
    case errors.As(err, &httpErr):
        c.JSON(httpErr.Code, map[string]any{
            "error": map[string]any{"code": "HTTP_ERROR", "message": fmt.Sprintf("%v", httpErr.Message)},
        })
    default:
        log.Error().Err(err).Msg("unhandled error")
        c.JSON(500, map[string]any{
            "error": map[string]any{"code": "INTERNAL_ERROR", "message": "Terjadi kesalahan, silakan coba lagi"},
        })
    }
}
```

---

## Verifikasi

```bash
docker compose -f docker-compose.dev.yml up --build
curl http://localhost:8080/health
# → {"status":"ok"}
```

---

## Checklist

- [x] `sqlc.yaml`
- [x] `Makefile`
- [x] `.air.toml`
- [x] `internal/config/config.go`
- [x] `internal/database/db.go`
- [x] `internal/errors/errors.go` — AppError + pg mapping
- [x] `internal/middleware/logger.go`
- [x] `internal/middleware/auth.go`
- [x] `internal/middleware/permission.go`
- [x] `internal/middleware/audit.go`
- [x] `internal/middleware/error_handler.go`
- [x] `internal/model/common.go` — generic types
- [x] `internal/sse/broker.go`
- [x] `cmd/api/main.go` — wiring + /health endpoint
- [x] Docker dev berjalan, `/health` OK
