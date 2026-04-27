# PLAN BE-02 — Master Data Endpoints

> **Scope:** CRUD endpoints untuk semua entitas master: country, lender, institution, region, program_title, bappenas_partner, period, national_priority.
> **Deliverable:** Semua master data bisa diakses dan dikelola via API.
> **Referensi:** docs/PRISM_API_Contract.md (Master Data), docs/PRISM_Business_Rules.md (bagian 2, 8, 9)

---

## Pola Umum

Setiap modul master mengikuti pola yang sama:

1. Tulis query di `sql/queries/<modul>.sql`
2. `make generate`
3. Model di `internal/model/master.go`
4. Service di `internal/service/master_service.go`
5. Handler di `internal/handler/master_handler.go`
6. Register route di `main.go`

Semua list endpoint support pagination: `?page=1&limit=20&sort=name&order=asc`.

---

## Task 1 — sql/queries/master.sql

Query untuk semua tabel master:

```sql
-- ===== COUNTRY =====
-- name: ListCountries :many
SELECT * FROM country ORDER BY name ASC LIMIT $1 OFFSET $2;

-- name: GetCountry :one
SELECT * FROM country WHERE id = $1;

-- name: CreateCountry :one
INSERT INTO country (name, code) VALUES ($1, $2) RETURNING *;

-- name: UpdateCountry :one
UPDATE country SET name = $2, code = $3, updated_at = NOW() WHERE id = $1 RETURNING *;

-- name: DeleteCountry :exec
DELETE FROM country WHERE id = $1;

-- ===== LENDER =====
-- name: ListLenders :many
SELECT l.*, c.name as country_name, c.code as country_code
FROM lender l LEFT JOIN country c ON c.id = l.country_id
ORDER BY l.name ASC LIMIT $1 OFFSET $2;

-- name: GetLender :one
SELECT l.*, c.name as country_name
FROM lender l LEFT JOIN country c ON c.id = l.country_id
WHERE l.id = $1;

-- name: CreateLender :one
INSERT INTO lender (country_id, name, short_name, type) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateLender :one
UPDATE lender SET country_id=$2, name=$3, short_name=$4, type=$5, updated_at=NOW() WHERE id=$1 RETURNING *;

-- name: DeleteLender :exec
DELETE FROM lender WHERE id = $1;

-- ===== INSTITUTION =====
-- name: ListInstitutions :many
SELECT * FROM institution ORDER BY level, name ASC LIMIT $1 OFFSET $2;

-- name: GetInstitution :one
SELECT i.*, p.name as parent_name
FROM institution i LEFT JOIN institution p ON p.id = i.parent_id
WHERE i.id = $1;

-- name: CreateInstitution :one
INSERT INTO institution (parent_id, name, short_name, level) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateInstitution :one
UPDATE institution SET parent_id=$2, name=$3, short_name=$4, level=$5, updated_at=NOW() WHERE id=$1 RETURNING *;

-- name: DeleteInstitution :exec
DELETE FROM institution WHERE id = $1;

-- ===== REGION =====
-- name: ListRegions :many
SELECT * FROM region ORDER BY type, name ASC LIMIT $1 OFFSET $2;

-- name: GetRegion :one
SELECT * FROM region WHERE id = $1;

-- name: CreateRegion :one
INSERT INTO region (code, name, type, parent_code) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: UpdateRegion :one
UPDATE region SET code=$2, name=$3, type=$4, parent_code=$5, updated_at=NOW() WHERE id=$1 RETURNING *;

-- name: DeleteRegion :exec
DELETE FROM region WHERE id = $1;

-- ===== PROGRAM TITLE =====
-- name: ListProgramTitles :many
SELECT * FROM program_title ORDER BY title ASC;

-- name: CreateProgramTitle :one
INSERT INTO program_title (parent_id, title) VALUES ($1, $2) RETURNING *;

-- name: UpdateProgramTitle :one
UPDATE program_title SET parent_id=$2, title=$3, updated_at=NOW() WHERE id=$1 RETURNING *;

-- name: DeleteProgramTitle :exec
DELETE FROM program_title WHERE id=$1;

-- ===== BAPPENAS PARTNER =====
-- name: ListBappenasPartners :many
SELECT * FROM bappenas_partner ORDER BY level, name ASC;

-- name: CreateBappenasPartner :one
INSERT INTO bappenas_partner (parent_id, name, level) VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateBappenasPartner :one
UPDATE bappenas_partner SET parent_id=$2, name=$3, level=$4, updated_at=NOW() WHERE id=$1 RETURNING *;

-- name: DeleteBappenasPartner :exec
DELETE FROM bappenas_partner WHERE id=$1;

-- ===== PERIOD =====
-- name: ListPeriods :many
SELECT * FROM period ORDER BY year_start DESC;

-- name: CreatePeriod :one
INSERT INTO period (name, year_start, year_end) VALUES ($1, $2, $3) RETURNING *;

-- name: UpdatePeriod :one
UPDATE period SET name=$2, year_start=$3, year_end=$4, updated_at=NOW() WHERE id=$1 RETURNING *;

-- name: DeletePeriod :exec
DELETE FROM period WHERE id=$1;

-- ===== NATIONAL PRIORITY =====
-- name: ListNationalPriorities :many
SELECT np.*, p.name as period_name
FROM national_priority np JOIN period p ON p.id = np.period_id
WHERE ($1::uuid IS NULL OR np.period_id = $1)
ORDER BY np.title ASC;

-- name: CreateNationalPriority :one
INSERT INTO national_priority (period_id, title) VALUES ($1, $2) RETURNING *;

-- name: UpdateNationalPriority :one
UPDATE national_priority SET period_id=$2, title=$3, updated_at=NOW() WHERE id=$1 RETURNING *;

-- name: DeleteNationalPriority :exec
DELETE FROM national_priority WHERE id=$1;
```

Jalankan `make generate`.

---

## Task 2 — internal/model/master.go

Request/response types untuk semua master. Ikuti shape response dari API Contract:

```go
// Contoh untuk Lender
type CreateLenderRequest struct {
    CountryID *string `json:"country_id"`
    Name      string  `json:"name" validate:"required"`
    ShortName *string `json:"short_name"`
    Type      string  `json:"type" validate:"required,oneof=Bilateral Multilateral KSA"`
}

type LenderResponse struct {
    ID        string          `json:"id"`
    Name      string          `json:"name"`
    ShortName *string         `json:"short_name"`
    Type      string          `json:"type"`
    Country   *CountryInfo    `json:"country,omitempty"`
}
```

Buat types untuk semua entitas master.

---

## Task 3 — Business Rule Validasi di Service

Implementasi validasi bisnis di `internal/service/master_service.go`:

```go
// Lender: country_id wajib untuk Bilateral dan KSA
func (s *MasterService) CreateLender(ctx context.Context, req model.CreateLenderRequest) (*model.LenderResponse, error) {
    if req.Type != "Multilateral" && req.CountryID == nil {
        return nil, errors.Validation(errors.FieldError{Field: "country_id", Message: "Wajib diisi untuk Bilateral dan KSA"})
    }
    if req.Type == "Multilateral" && req.CountryID != nil {
        return nil, errors.Validation(errors.FieldError{Field: "country_id", Message: "Harus kosong untuk Multilateral"})
    }
    // Insert ke DB
}
```

---

## Task 4 — internal/handler/master_handler.go

Satu handler file untuk semua master. Untuk setiap entitas:
- `GET /master/<plural>` — list
- `GET /master/<plural>/:id` — get by id
- `POST /master/<plural>` — create
- `PUT /master/<plural>/:id` — update
- `DELETE /master/<plural>/:id` — delete

---

## Task 5 — Register Routes di main.go

```go
master := api.Group("/master")
// Country
master.GET("/countries", masterHandler.ListCountries, permission.Require("country", "read"))
master.POST("/countries", masterHandler.CreateCountry, permission.Require("country", "create"))
master.PUT("/countries/:id", masterHandler.UpdateCountry, permission.Require("country", "update"))
master.DELETE("/countries/:id", masterHandler.DeleteCountry, permission.Require("country", "delete"))
// Lender
master.GET("/lenders", masterHandler.ListLenders, permission.Require("lender", "read"))
master.GET("/lenders/:id", masterHandler.GetLender, permission.Require("lender", "read"))
master.POST("/lenders", masterHandler.CreateLender, permission.Require("lender", "create"))
master.PUT("/lenders/:id", masterHandler.UpdateLender, permission.Require("lender", "update"))
master.DELETE("/lenders/:id", masterHandler.DeleteLender, permission.Require("lender", "delete"))
// ... dst untuk institution, region, program_title, bappenas_partner, period, national_priority
```

---

## Checklist

- [ ] `sql/queries/master.sql` — semua query master
- [ ] `make generate` berhasil
- [ ] `internal/model/master.go` — semua request/response types
- [ ] `internal/service/master_service.go` — dengan validasi bisnis lender
- [ ] `internal/handler/master_handler.go` — CRUD handler semua master
- [ ] Routes semua master terdaftar di `main.go`
- [ ] `GET /api/v1/master/lenders` → list lender
- [ ] `POST /api/v1/master/lenders` dengan type=Bilateral tanpa country_id → 400
