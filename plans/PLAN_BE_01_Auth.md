# PLAN BE-01 — Auth & User Management

> **Scope:** Endpoint auth (login/logout/me) dan user management (CRUD + permission).
> **Deliverable:** Login menghasilkan JWT, permission ter-seed untuk ADMIN pertama.
> **Referensi:** docs/PRISM_API_Contract.md (Auth & User Management), docs/PRISM_Business_Rules.md (bagian 10)

---

## Task 1 — sql/queries/user.sql

```sql
-- name: GetUserByUsername :one
SELECT * FROM app_user WHERE username = $1 AND is_active = true;

-- name: GetUserByID :one
SELECT * FROM app_user WHERE id = $1;

-- name: ListUsers :many
SELECT * FROM app_user
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM app_user;

-- name: CreateUser :one
INSERT INTO app_user (username, email, password_hash, role)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateUser :one
UPDATE app_user
SET username = $2, email = $3, role = $4, is_active = $5, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetUserPermissions :many
SELECT * FROM user_permission WHERE user_id = $1;

-- name: DeleteUserPermissions :exec
DELETE FROM user_permission WHERE user_id = $1;

-- name: CreateUserPermission :one
INSERT INTO user_permission (user_id, module, can_create, can_read, can_update, can_delete)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserPermissionByModule :one
SELECT * FROM user_permission WHERE user_id = $1 AND module = $2;
```

Jalankan `make generate` setelah selesai.

---

## Task 2 — internal/model/auth.go

```go
type LoginRequest struct {
    Username string `json:"username" validate:"required,min=3"`
    Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
    AccessToken string   `json:"access_token"`
    ExpiresIn   int      `json:"expires_in"`
    User        UserInfo `json:"user"`
}

type UserInfo struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    Role     string `json:"role"`
}

type MeResponse struct {
    UserInfo
    Permissions []PermissionInfo `json:"permissions"`
}

type PermissionInfo struct {
    Module    string `json:"module"`
    CanCreate bool   `json:"can_create"`
    CanRead   bool   `json:"can_read"`
    CanUpdate bool   `json:"can_update"`
    CanDelete bool   `json:"can_delete"`
}

type AuthUser struct {
    ID   pgtype.UUID
    Role string
}

type CreateUserRequest struct {
    Username string `json:"username" validate:"required,min=3"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Role     string `json:"role" validate:"required,oneof=ADMIN STAFF"`
}

type UpdateUserRequest struct {
    Username string `json:"username" validate:"required,min=3"`
    Email    string `json:"email" validate:"required,email"`
    Role     string `json:"role" validate:"required,oneof=ADMIN STAFF"`
    IsActive bool   `json:"is_active"`
}

type UpdatePermissionsRequest struct {
    Permissions []PermissionItem `json:"permissions" validate:"required"`
}

type PermissionItem struct {
    Module    string `json:"module" validate:"required"`
    CanCreate bool   `json:"can_create"`
    CanRead   bool   `json:"can_read"`
    CanUpdate bool   `json:"can_update"`
    CanDelete bool   `json:"can_delete"`
}
```

---

## Task 3 — internal/service/auth_service.go

```go
type AuthService struct {
    queries   *queries.Queries
    jwtSecret string
    expiresIn int
}

func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.LoginResponse, error) {
    user, err := s.queries.GetUserByUsername(ctx, req.Username)
    if err != nil {
        return nil, errors.Unauthorized("Username atau password salah")
    }
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        return nil, errors.Unauthorized("Username atau password salah")
    }
    // Generate JWT
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "sub":  user.ID.String(),
        "role": user.Role,
        "exp":  time.Now().Add(time.Duration(s.expiresIn) * time.Second).Unix(),
    })
    tokenStr, err := token.SignedString([]byte(s.jwtSecret))
    // Return LoginResponse
}

func (s *AuthService) GetMe(ctx context.Context, userID pgtype.UUID) (*model.MeResponse, error) {
    user, err := s.queries.GetUserByID(ctx, userID)
    perms, err := s.queries.GetUserPermissions(ctx, userID)
    // Assemble MeResponse
}
```

---

## Task 4 — internal/service/user_service.go

```go
func (s *UserService) CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.UserResponse, error) {
    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    // Insert ke DB
}

func (s *UserService) UpdatePermissions(ctx context.Context, userID pgtype.UUID, req model.UpdatePermissionsRequest) error {
    tx, err := s.db.Begin(ctx)
    defer tx.Rollback(ctx)
    qtx := s.queries.WithTx(tx)

    // Delete semua permission lama
    qtx.DeleteUserPermissions(ctx, userID)

    // Insert permission baru (replace-all)
    for _, p := range req.Permissions {
        qtx.CreateUserPermission(ctx, queries.CreateUserPermissionParams{...})
    }
    return tx.Commit(ctx)
}
```

---

## Task 5 — internal/handler/auth_handler.go

Endpoint:
- `POST /auth/login` — public, call `AuthService.Login`
- `POST /auth/logout` — authenticated, return 204
- `GET /auth/me` — authenticated, call `AuthService.GetMe`

---

## Task 6 — internal/handler/user_handler.go

Semua endpoint ADMIN only:
- `GET /users` — list dengan pagination
- `GET /users/:id`
- `POST /users`
- `PUT /users/:id`
- `DELETE /users/:id` (set `is_active = false`, bukan hard delete)
- `GET /users/:id/permissions`
- `PUT /users/:id/permissions` — replace-all, transaksional

---

## Task 7 — Update permission.go Middleware

Implement pengecekan permission ke DB:

```go
func Require(module, action string) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            user := c.Get("user").(*model.AuthUser)
            if user.Role == "ADMIN" {
                return next(c)
            }
            perm, err := queries.GetUserPermissionByModule(ctx, user.ID, module)
            if err != nil {
                return apperrors.Forbidden("Tidak memiliki akses ke modul ini")
            }
            allowed := false
            switch action {
            case "create": allowed = perm.CanCreate
            case "read":   allowed = perm.CanRead
            case "update": allowed = perm.CanUpdate
            case "delete": allowed = perm.CanDelete
            }
            if !allowed {
                return apperrors.Forbidden("Tidak memiliki izin untuk aksi ini")
            }
            return next(c)
        }
    }
}
```

---

## Task 8 — Daftarkan Routes di main.go

```go
// Auth (public)
e.POST("/api/v1/auth/login", authHandler.Login)

// Auth (authenticated)
authGroup := api.Group("/auth")
authGroup.POST("/logout", authHandler.Logout)
authGroup.GET("/me", authHandler.Me)

// Users (ADMIN only)
userGroup := api.Group("/users", middleware.RequireAdmin())
userGroup.GET("", userHandler.List)
userGroup.POST("", userHandler.Create)
userGroup.GET("/:id", userHandler.Get)
userGroup.PUT("/:id", userHandler.Update)
userGroup.DELETE("/:id", userHandler.Delete)
userGroup.GET("/:id/permissions", userHandler.GetPermissions)
userGroup.PUT("/:id/permissions", userHandler.UpdatePermissions)
```

---

## Task 9 — Seed ADMIN Pertama

Buat file `migrations/000001_seed_admin.up.sql`:

```sql
-- Seed ADMIN user pertama (password: admin123 — harus diganti setelah login pertama)
-- Hash bcrypt dari 'admin123': $2a$10$... (generate via tool atau hardcode)
INSERT INTO app_user (username, email, password_hash, role)
VALUES ('admin', 'admin@prism.go.id', '$2a$10$...', 'ADMIN')
ON CONFLICT (username) DO NOTHING;
```

Jalankan `make migrate-up`.

---

## Checklist

- [ ] `sql/queries/user.sql` — semua query user + permission
- [ ] `make generate` berhasil
- [ ] `internal/model/auth.go` — semua request/response types
- [ ] `internal/service/auth_service.go` — login + JWT + getMe
- [ ] `internal/service/user_service.go` — CRUD + updatePermissions transaksional
- [ ] `internal/handler/auth_handler.go`
- [ ] `internal/handler/user_handler.go`
- [ ] `internal/middleware/permission.go` — implementasi cek ke DB
- [ ] Routes terdaftar di `main.go`
- [ ] Seed admin di migration
- [ ] `POST /auth/login` dengan admin/admin123 → JWT
- [ ] `GET /auth/me` dengan JWT → user + permissions
