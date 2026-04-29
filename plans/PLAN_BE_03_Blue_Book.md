# PLAN BE-03 — Blue Book Endpoints

> **Scope:** CRUD Blue Book (header + BB Project + LoI + Lender Indication + Project Cost).
> **Deliverable:** Full CRUD BB Project dengan semua sub-entitas dalam satu transaksi.
> **Referensi:** docs/PRISM_API_Contract.md (Blue Book), docs/PRISM_Business_Rules.md (bagian 3)
> **Revision update:** Ikuti `docs/PRISM_BB_GB_Revision_Versioning_Plan.md` untuk logical identity, clone revisi, dan perubahan uniqueness `bb_code` menjadi per `blue_book_id`.

---

## Task 1 — sql/queries/bb_project.sql

```sql
-- ===== BLUE BOOK =====
-- name: ListBlueBooks :many
SELECT bb.*, p.name as period_name, p.year_start, p.year_end
FROM blue_book bb JOIN period p ON p.id = bb.period_id
ORDER BY bb.created_at DESC LIMIT $1 OFFSET $2;

-- name: GetBlueBook :one
SELECT bb.*, p.name as period_name
FROM blue_book bb JOIN period p ON p.id = bb.period_id
WHERE bb.id = $1;

-- name: CreateBlueBook :one
INSERT INTO blue_book (period_id, publish_date, revision_number, revision_year, status)
VALUES ($1, $2, $3, $4, 'active') RETURNING *;

-- name: SupersedeBlueBooksByPeriod :exec
UPDATE blue_book SET status = 'superseded', updated_at = NOW()
WHERE period_id = $1 AND status = 'active';

-- ===== BB PROJECT =====
-- name: ListBBProjectsByBlueBook :many
SELECT * FROM bb_project WHERE blue_book_id = $1 AND status = 'active'
ORDER BY bb_code ASC LIMIT $2 OFFSET $3;

-- name: GetBBProject :one
SELECT * FROM bb_project WHERE id = $1;

-- name: GetBBProjectByCode :one
SELECT * FROM bb_project WHERE bb_code = $1;

-- name: CreateBBProject :one
INSERT INTO bb_project (blue_book_id, program_title_id, bappenas_partner_id, bb_code, project_name, duration, objective, scope_of_work, outputs, outcomes)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *;

-- name: UpdateBBProject :one
UPDATE bb_project SET program_title_id=$2, bappenas_partner_id=$3, project_name=$4, duration=$5, objective=$6, scope_of_work=$7, outputs=$8, outcomes=$9, updated_at=NOW()
WHERE id=$1 RETURNING *;

-- name: SoftDeleteBBProject :one
UPDATE bb_project SET status='deleted', updated_at=NOW() WHERE id=$1 RETURNING *;

-- ===== BB INSTITUTIONS =====
-- name: GetBBProjectInstitutions :many
SELECT bpi.role, i.* FROM bb_project_institution bpi
JOIN institution i ON i.id = bpi.institution_id
WHERE bpi.bb_project_id = $1;

-- name: AddBBProjectInstitution :exec
INSERT INTO bb_project_institution (bb_project_id, institution_id, role) VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: DeleteBBProjectInstitutions :exec
DELETE FROM bb_project_institution WHERE bb_project_id = $1;

-- ===== BB LOCATIONS =====
-- name: GetBBProjectLocations :many
SELECT r.* FROM bb_project_location bpl JOIN region r ON r.id = bpl.region_id
WHERE bpl.bb_project_id = $1;

-- name: AddBBProjectLocation :exec
INSERT INTO bb_project_location (bb_project_id, region_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: DeleteBBProjectLocations :exec
DELETE FROM bb_project_location WHERE bb_project_id = $1;

-- ===== BB NATIONAL PRIORITIES =====
-- name: GetBBProjectNationalPriorities :many
SELECT np.* FROM bb_project_national_priority bpnp
JOIN national_priority np ON np.id = bpnp.national_priority_id
WHERE bpnp.bb_project_id = $1;

-- name: AddBBProjectNationalPriority :exec
INSERT INTO bb_project_national_priority (bb_project_id, national_priority_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: DeleteBBProjectNationalPriorities :exec
DELETE FROM bb_project_national_priority WHERE bb_project_id = $1;

-- ===== PROJECT COSTS =====
-- name: GetBBProjectCosts :many
SELECT * FROM bb_project_cost WHERE bb_project_id = $1;

-- name: CreateBBProjectCost :one
INSERT INTO bb_project_cost (bb_project_id, funding_type, funding_category, amount_usd) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: DeleteBBProjectCosts :exec
DELETE FROM bb_project_cost WHERE bb_project_id = $1;

-- ===== LENDER INDICATION =====
-- name: GetLenderIndications :many
SELECT li.*, l.name as lender_name, l.type as lender_type
FROM lender_indication li JOIN lender l ON l.id = li.lender_id
WHERE li.bb_project_id = $1;

-- name: CreateLenderIndication :one
INSERT INTO lender_indication (bb_project_id, lender_id, remarks) VALUES ($1, $2, $3) RETURNING *;

-- name: DeleteLenderIndications :exec
DELETE FROM lender_indication WHERE bb_project_id = $1;

-- ===== LoI =====
-- name: GetLoIsByBBProject :many
SELECT loi.*, l.name as lender_name
FROM loi JOIN lender l ON l.id = loi.lender_id
WHERE loi.bb_project_id = $1 ORDER BY loi.date DESC;

-- name: CreateLoI :one
INSERT INTO loi (bb_project_id, lender_id, subject, date, letter_number)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: DeleteLoI :exec
DELETE FROM loi WHERE id = $1 AND bb_project_id = $2;
```

Jalankan `make generate`.

---

## Task 2 — internal/model/blue_book.go

Request/response types. Contoh:

```go
type CreateBBProjectRequest struct {
    ProgramTitleID      *string  `json:"program_title_id"`
    BappenasPartnerID   *string  `json:"bappenas_partner_id"`
    BBCode              string   `json:"bb_code" validate:"required"`
    ProjectName         string   `json:"project_name" validate:"required"`
    Duration            *int32   `json:"duration"` // jumlah bulan
    Objective           *string  `json:"objective"`
    ScopeOfWork         *string  `json:"scope_of_work"`
    Outputs             *string  `json:"outputs"`
    Outcomes            *string  `json:"outcomes"`
    ExecutingAgencyIDs  []string `json:"executing_agency_ids" validate:"required,min=1"`
    ImplementingAgencyIDs []string `json:"implementing_agency_ids" validate:"required,min=1"`
    LocationIDs         []string `json:"location_ids" validate:"required,min=1"`
    NationalPriorityIDs []string `json:"national_priority_ids"`
    ProjectCosts        []ProjectCostItem `json:"project_costs"`
    LenderIndications   []LenderIndicationItem `json:"lender_indications"`
}
```

---

## Task 3 — internal/service/blue_book_service.go

```go
func (s *BBProjectService) CreateBBProject(ctx context.Context, bbID pgtype.UUID, req model.CreateBBProjectRequest) (*model.BBProjectResponse, error) {
    // Validasi bisnis:
    // 1. Cek bb_code belum dipakai dalam Blue Book yang sama.
    //    Kode yang sama boleh muncul pada revisi Blue Book lain melalui project_identity_id.
    existing, _ := s.queries.GetBBProjectByBlueBookAndCode(ctx, bbID, req.BBCode)
    if existing != nil {
        return nil, errors.Conflict("BB Code sudah digunakan di Blue Book ini")
    }

    // 2. EA dan IA boleh overlap bila sesuai data proyek

    // Transaksi: insert project + semua relasi
    tx, err := s.db.Begin(ctx)
    defer tx.Rollback(ctx)
    qtx := s.queries.WithTx(tx)

    project, err := qtx.CreateBBProject(ctx, ...)

    // Insert institutions (EA dan IA)
    for _, id := range req.ExecutingAgencyIDs {
        qtx.AddBBProjectInstitution(ctx, ...)
    }
    for _, id := range req.ImplementingAgencyIDs {
        qtx.AddBBProjectInstitution(ctx, ...)
    }

    // Insert locations, national priorities, costs, lender indications
    // ...

    tx.Commit(ctx)

    // Trigger SSE
    s.notification.Publish("bb_project.created", map[string]any{"id": project.ID})

    return s.buildResponse(ctx, project), nil
}

func (s *BBProjectService) UpdateBBProject(...) {
    // Update: DELETE lama + INSERT baru untuk relasi many-to-many (dalam transaksi)
}
```

---

## Task 4 — internal/handler/blue_book_handler.go

Handler untuk semua Blue Book endpoint dari API Contract.

---

## Task 5 — Register Routes

```go
bb := api.Group("/blue-books")
bb.GET("", bbHandler.ListBlueBooks, permission.Require("blue_book", "read"))
bb.POST("", bbHandler.CreateBlueBook, permission.Require("blue_book", "create"))
bb.GET("/:id", bbHandler.GetBlueBook, permission.Require("blue_book", "read"))
bb.PUT("/:id", bbHandler.UpdateBlueBook, permission.Require("blue_book", "update"))
bb.DELETE("/:id", bbHandler.DeleteBlueBook, permission.Require("blue_book", "delete"))

// BB Projects
bb.GET("/:bbId/projects", bbHandler.ListBBProjects, permission.Require("bb_project", "read"))
bb.POST("/:bbId/projects", bbHandler.CreateBBProject, permission.Require("bb_project", "create"))
bb.GET("/:bbId/projects/:id", bbHandler.GetBBProject, permission.Require("bb_project", "read"))
bb.PUT("/:bbId/projects/:id", bbHandler.UpdateBBProject, permission.Require("bb_project", "update"))
bb.DELETE("/:bbId/projects/:id", bbHandler.DeleteBBProject, permission.Require("bb_project", "delete"))

// LoI
loi := api.Group("/bb-projects/:bbProjectId/loi")
loi.GET("", bbHandler.ListLoI, permission.Require("bb_project", "read"))
loi.POST("", bbHandler.CreateLoI, permission.Require("bb_project", "update"))
loi.DELETE("/:id", bbHandler.DeleteLoI, permission.Require("bb_project", "update"))
```

---

## Checklist

- [x] `sql/queries/bb_project.sql` — semua query BB
- [x] `make generate`
- [x] `internal/model/blue_book.go`
- [x] `internal/service/blue_book_service.go` — CRUD + validasi bisnis + transaksi + SSE
- [x] `internal/handler/blue_book_handler.go`
- [x] Routes terdaftar
- [x] `POST /bb-projects` dengan bb_code duplikat → 409
- [x] `POST /bb-projects` dengan EA = IA → diterima bila payload lain valid
- [x] `POST /bb-projects` sukses → SSE event terkirim
- [x] `DELETE /blue-books/:bbId/projects/:id` → status `deleted`, record tetap ada di DB
