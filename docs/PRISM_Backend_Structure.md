# PRISM — Backend Project Structure (Go)

> Stack: **Go** · **Echo** · **sqlc** · **pgx** · **golang-migrate** · **SSE**

---

## Stack Overview

| Komponen | Library | Alasan |
|----------|---------|--------|
| HTTP Framework | `echo` | Ringan, middleware-friendly, routing bersih |
| PostgreSQL Driver | `pgx/v5` | Native PostgreSQL, performa terbaik |
| Query Generator | `sqlc` | Generate type-safe Go dari SQL murni |
| Migration | `golang-migrate` | Berbasis file SQL, mudah di-version control |
| Auth / JWT | `golang-jwt/jwt` | Standard, minimal dependency |
| Config | `viper` | Support env file + env var |
| Realtime | SSE (built-in) | Cukup untuk notifikasi — lebih simpel dari WebSocket |
| Logging | `zerolog` | Structured logging, sangat hemat alokasi memori |

---

## Project Structure

```
prism-backend/
├── cmd/
│   └── api/
│       └── main.go                  # Entry point — init config, DB, server
│
├── internal/
│   ├── config/
│   │   └── config.go                # Load env, validasi config
│   │
│   ├── database/
│   │   ├── db.go                    # Init pgx connection pool
│   │   └── queries/                 # Generated oleh sqlc — jangan edit manual
│   │       ├── bb_project.sql.go
│   │       ├── gb_project.sql.go
│   │       ├── dk_project.sql.go
│   │       ├── loan_agreement.sql.go
│   │       ├── monitoring.sql.go
│   │       ├── lender.sql.go
│   │       ├── institution.sql.go
│   │       ├── region.sql.go
│   │       ├── user.sql.go
│   │       └── ...
│   │
│   ├── handler/                     # HTTP handler per modul
│   │   ├── auth_handler.go
│   │   ├── blue_book_handler.go
│   │   ├── green_book_handler.go
│   │   ├── daftar_kegiatan_handler.go
│   │   ├── loan_agreement_handler.go
│   │   ├── monitoring_handler.go
│   │   ├── master_handler.go        # lender, institution, wilayah, dll.
│   │   ├── user_handler.go
│   │   └── sse_handler.go           # Server-Sent Events endpoint
│   │
│   ├── middleware/
│   │   ├── auth.go                  # Validasi JWT, inject user ke context
│   │   ├── permission.go            # Cek user_permission per modul & operasi
│   │   ├── audit.go                 # Set app.current_user_id di DB session
│   │   └── logger.go                # Request logging
│   │
│   ├── service/                     # Business logic — di sinilah aturan bisnis PRISM
│   │   ├── blue_book_service.go
│   │   ├── green_book_service.go
│   │   ├── daftar_kegiatan_service.go
│   │   ├── loan_agreement_service.go
│   │   ├── monitoring_service.go
│   │   └── notification_service.go  # Trigger SSE event ke subscriber
│   │
│   ├── model/                       # Request/Response struct (DTO)
│   │   ├── auth.go
│   │   ├── blue_book.go
│   │   ├── green_book.go
│   │   ├── daftar_kegiatan.go
│   │   ├── loan_agreement.go
│   │   ├── monitoring.go
│   │   └── common.go                # Pagination, error response, dll.
│   │
│   └── sse/
│       └── broker.go                # SSE broker — manage subscriber & broadcast event
│
├── migrations/                      # File SQL untuk golang-migrate
│   ├── 000001_init_master.up.sql
│   ├── 000001_init_master.down.sql
│   ├── 000002_blue_book.up.sql
│   ├── 000002_blue_book.down.sql
│   ├── 000003_green_book.up.sql
│   ├── 000003_green_book.down.sql
│   ├── 000004_daftar_kegiatan.up.sql
│   ├── 000004_daftar_kegiatan.down.sql
│   ├── 000005_loan_agreement.up.sql
│   ├── 000005_loan_agreement.down.sql
│   ├── 000006_monitoring.up.sql
│   ├── 000006_monitoring.down.sql
│   ├── 000007_users.up.sql
│   ├── 000007_users.down.sql
│   ├── 000008_audit_trail.up.sql
│   └── 000008_audit_trail.down.sql
│
├── sql/                             # Source SQL untuk sqlc
│   ├── queries/
│   │   ├── bb_project.sql
│   │   ├── gb_project.sql
│   │   ├── dk_project.sql
│   │   ├── loan_agreement.sql
│   │   ├── monitoring.sql
│   │   ├── lender.sql
│   │   ├── institution.sql
│   │   ├── region.sql
│   │   └── user.sql
│   └── schema/                      # DDL — referensi sqlc untuk generate
│       └── prism_ddl.sql
│
├── .env.example
├── .env                             # Tidak di-commit ke git
├── sqlc.yaml                        # Konfigurasi sqlc
├── go.mod
├── go.sum
├── Makefile                         # Shortcut: migrate, generate, run, build
└── Dockerfile
```

---

## Alur Request

```
Client
  │
  ▼
Echo Router
  │
  ├── middleware/logger.go       — log request
  ├── middleware/auth.go         — validasi JWT, inject user ke ctx
  ├── middleware/permission.go   — cek can_read/can_create/dll per modul
  └── middleware/audit.go        — SET LOCAL app.current_user_id
  │
  ▼
Handler (parsing request, validasi input)
  │
  ▼
Service (business logic, transaksi DB)
  │
  ▼
sqlc Queries (type-safe SQL ke PostgreSQL via pgx)
  │
  ▼
Response JSON ke Client
```

---

## Permission Middleware

Setiap route didaftarkan dengan modul dan operasi yang dibutuhkan:

```go
// Contoh registrasi route dengan permission check
api.GET("/bb-projects",      h.ListBBProject,   permission.Require("bb_project", "read"))
api.POST("/bb-projects",     h.CreateBBProject, permission.Require("bb_project", "create"))
api.PUT("/bb-projects/:id",  h.UpdateBBProject, permission.Require("bb_project", "update"))
api.DELETE("/bb-projects/:id", h.DeleteBBProject, permission.Require("bb_project", "delete"))
```

Middleware `permission.Require` mengecek tabel `user_permission` berdasarkan user yang ada di JWT context. ADMIN selalu dilewatkan tanpa cek.

---

## Audit Middleware

Setiap transaksi DB di-set user aktif agar trigger audit berjalan:

```go
// middleware/audit.go
func SetAuditUser(db *pgxpool.Pool) echo.MiddlewareFunc {
    return func(next echo.HandlerFunc) echo.HandlerFunc {
        return func(c echo.Context) error {
            user := c.Get("user").(*model.AuthUser)
            conn, _ := db.Acquire(c.Request().Context())
            defer conn.Release()
            conn.Exec(c.Request().Context(),
                "SET LOCAL app.current_user_id = $1", user.ID)
            return next(c)
        }
    }
}
```

---

## SSE (Server-Sent Events)

Digunakan untuk notifikasi realtime ringan — misalnya ketika Staff lain menginput data baru atau ada update monitoring.

```
Client subscribe ke /events
  │
  ▼
sse/broker.go — simpan channel per client
  │
  ▼
notification_service.go — broadcast event setelah operasi DB berhasil
  │
  ▼
Client menerima event tanpa perlu polling
```

---

## sqlc.yaml

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

## Makefile

```makefile
.PHONY: run build migrate generate

run:
	go run ./cmd/api

build:
	go build -o bin/prism ./cmd/api

migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1

generate:
	sqlc generate

test:
	go test ./...
```

---

## Catatan Pengembangan

- `internal/database/queries/` **tidak boleh diedit manual** — selalu edit file di `sql/queries/` lalu jalankan `make generate`.
- Migration bersifat **incremental** — satu file per perubahan skema, tidak mengedit file lama.
- `.env` tidak di-commit — gunakan `.env.example` sebagai template untuk tim.
- SSE broker di `sse/broker.go` menggunakan channel Go — tidak butuh library eksternal seperti Redis Pub/Sub untuk skala PRISM yang user-nya terbatas.
- Jika di masa depan skala bertambah besar dan butuh multi-instance, SSE broker bisa diganti dengan Redis Pub/Sub tanpa mengubah interface `notification_service.go`.

---

## Panduan untuk Coding Agent

Bagian ini ditujukan khusus untuk coding agent (Claude Code, Copilot, Cursor, dll.) agar output yang dihasilkan konsisten dengan arsitektur PRISM.

---

### Konteks Proyek

Ini adalah sistem backend untuk **PRISM (Project Loan Integrated Monitoring System)** — sistem monitoring pinjaman luar negeri milik Bappenas. Backend ditulis dalam **Go** dengan stack: Echo · sqlc · pgx/v5 · golang-migrate.

Database: **PostgreSQL**. Skema lengkap ada di `sql/schema/prism_ddl.sql`. Selalu jadikan file ini sebagai referensi utama sebelum menulis query atau model apapun.

---

### Aturan Wajib

**1. Jangan pernah menulis SQL di luar folder `sql/queries/`**

Semua query ditulis di `sql/queries/<modul>.sql` dengan format sqlc annotation, lalu di-generate dengan `make generate`. Jangan menulis raw SQL string di dalam file `.go`.

```sql
-- sql/queries/bb_project.sql

-- name: GetBBProject :one
SELECT * FROM bb_project
WHERE id = $1 AND status = 'active';

-- name: ListBBProjectByBlueBook :many
SELECT * FROM bb_project
WHERE blue_book_id = $1 AND status = 'active'
ORDER BY bb_code ASC;

-- name: CreateBBProject :one
INSERT INTO bb_project (
    blue_book_id, program_title_id, bappenas_partner_id,
    bb_code, project_name, duration, objective,
    scope_of_work, outputs, outcomes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;
```

**2. Jangan edit `internal/database/queries/` secara manual**

File di folder ini adalah hasil generate sqlc. Setiap kali ada perubahan query, edit file `.sql` di `sql/queries/` lalu jalankan:
```bash
make generate
```

**3. Handler hanya boleh berisi parsing dan response — tidak ada logic bisnis**

```go
// BENAR ✓
func (h *BBProjectHandler) Create(c echo.Context) error {
    var req model.CreateBBProjectRequest
    if err := c.Bind(&req); err != nil {
        return echo.ErrBadRequest
    }
    result, err := h.service.CreateBBProject(c.Request().Context(), req)
    if err != nil {
        return err
    }
    return c.JSON(http.StatusCreated, result)
}

// SALAH ✗ — logic bisnis di handler
func (h *BBProjectHandler) Create(c echo.Context) error {
    var req model.CreateBBProjectRequest
    c.Bind(&req)
    // validasi relasi, cek duplikat bb_code, dll. — ini seharusnya di service
    if req.BBCode == "" { ... }
    h.db.Exec("INSERT INTO bb_project ...")  // raw SQL di handler — JANGAN
    ...
}
```

**4. Semua operasi DB yang mengubah data wajib dalam transaksi**

```go
// service/blue_book_service.go
func (s *BBProjectService) CreateBBProject(ctx context.Context, req model.CreateBBProjectRequest) (*model.BBProjectResponse, error) {
    tx, err := s.db.Begin(ctx)
    if err != nil {
        return nil, err
    }
    defer tx.Rollback(ctx)

    qtx := s.queries.WithTx(tx)

    project, err := qtx.CreateBBProject(ctx, queries.CreateBBProjectParams{...})
    if err != nil {
        return nil, err
    }

    // Insert EA/IA, location, national priority dalam transaksi yang sama
    for _, instID := range req.ExecutingAgencyIDs {
        err = qtx.AddBBProjectInstitution(ctx, queries.AddBBProjectInstitutionParams{
            BbProjectID:   project.ID,
            InstitutionID: instID,
            Role:          "Executing Agency",
        })
        if err != nil {
            return nil, err
        }
    }

    tx.Commit(ctx)
    return toResponse(project), nil
}
```

**5. Selalu gunakan struct dari `model/` sebagai request/response — bukan struct sqlc langsung**

Struct sqlc adalah representasi database. Struct `model/` adalah kontrak API. Keduanya berbeda dan tidak boleh dicampur.

```go
// model/blue_book.go

type CreateBBProjectRequest struct {
    BlueBookID         string   `json:"blue_book_id" validate:"required,uuid"`
    ProgramTitleID     string   `json:"program_title_id" validate:"required,uuid"`
    BBCode             string   `json:"bb_code" validate:"required"`
    ProjectName        string   `json:"project_name" validate:"required"`
    ExecutingAgencyIDs []string `json:"executing_agency_ids" validate:"required,min=1"`
    LocationIDs        []string `json:"location_ids" validate:"required,min=1"`
    NationalPriorityIDs []string `json:"national_priority_ids"`
    // ...
}

type BBProjectResponse struct {
    ID          string `json:"id"`
    BBCode      string `json:"bb_code"`
    ProjectName string `json:"project_name"`
    // ...
}
```

**6. Error handling — jangan return error mentah ke client**

```go
// BENAR ✓
if err != nil {
    log.Error().Err(err).Str("bb_code", req.BBCode).Msg("failed to create bb_project")
    return echo.NewHTTPError(http.StatusInternalServerError, "gagal menyimpan data")
}

// SALAH ✗ — expose internal error ke client
if err != nil {
    return c.JSON(500, err.Error())
}
```

**7. Permission check dilakukan di layer router — bukan di service atau handler**

```go
// BENAR ✓ — di router
api.POST("/bb-projects", h.Create, permission.Require("bb_project", "create"))

// SALAH ✗ — cek permission di dalam service
func (s *Service) CreateBBProject(ctx context.Context, ...) {
    user := ctx.Value("user").(model.AuthUser)
    if !user.CanCreate("bb_project") { ... } // jangan di sini
}
```

---

### Pola Penamaan

| Konteks | Konvensi | Contoh |
|---------|----------|--------|
| File handler | `<modul>_handler.go` | `blue_book_handler.go` |
| File service | `<modul>_service.go` | `blue_book_service.go` |
| File query SQL | `<modul>.sql` | `bb_project.sql` |
| sqlc query name | `VerbNoun` | `GetBBProject`, `ListGBProjectByGreenBook` |
| Route URL | `kebab-case` | `/bb-projects`, `/gb-projects/:id/activities` |
| Struct request | `<Verb><Noun>Request` | `CreateBBProjectRequest` |
| Struct response | `<Noun>Response` | `BBProjectResponse` |
| Konstanta modul | `snake_case` | `"bb_project"`, `"loan_agreement"` |

---

### Urutan Langkah Saat Menambah Fitur Baru

Ikuti urutan ini setiap kali agent diminta menambah endpoint atau fitur baru:

```
1. Cek skema di sql/schema/prism_ddl.sql
         ↓
2. Tulis query di sql/queries/<modul>.sql
         ↓
3. Jalankan: make generate
         ↓
4. Buat/update struct di internal/model/<modul>.go
         ↓
5. Implementasi logic di internal/service/<modul>_service.go
         ↓
6. Implementasi handler di internal/handler/<modul>_handler.go
         ↓
7. Daftarkan route di cmd/api/main.go dengan permission middleware
```

---

### Hal yang Tidak Boleh Dilakukan Agent

| Larangan | Alasan |
|----------|--------|
| Edit `internal/database/queries/*.go` | File ini di-generate, akan tertimpa saat `make generate` |
| Tulis SQL string di file `.go` | Semua SQL harus via sqlc |
| Gunakan `interface{}` atau `any` untuk data DB | Gunakan struct yang strongly-typed dari sqlc |
| Tambah dependency baru tanpa konfirmasi | Jaga dependency tetap minimal |
| Langsung return `error.Error()` ke client | Selalu wrap dengan pesan yang aman |
| Bypass middleware permission di route | Permission check harus konsisten di layer router |
| Buat migration baru yang mengedit tabel yang sudah ada tanpa `ALTER TABLE` | Selalu incremental, jangan drop-recreate |
