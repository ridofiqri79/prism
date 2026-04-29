# PLAN BE-04 — Green Book Endpoints

> **Scope:** CRUD GB Project dengan Activities, Funding Source, Disbursement Plan, Funding Allocation.
> **Deliverable:** GB Project tersimpan lengkap dengan semua sub-tabel dalam satu transaksi.
> **Referensi:** docs/PRISM_API_Contract.md (Green Book), docs/PRISM_Business_Rules.md (bagian 4)
> **Revision update:** Ikuti `docs/PRISM_BB_GB_Revision_Versioning_Plan.md` untuk logical identity, clone revisi, latest BB resolver, dan perubahan uniqueness `gb_code` menjadi per `green_book_id`.

---

## Task 1 — sql/queries/gb_project.sql

Query mencakup:
- CRUD `green_book` (header)
- CRUD `gb_project`
- Junction: `gb_project_bb_project`, `gb_project_bappenas_partner`, `gb_project_institution`, `gb_project_location`
- CRUD `gb_activity` (dengan sort_order)
- CRUD `gb_funding_source`
- CRUD `gb_disbursement_plan` (unique constraint year per project)
- CRUD `gb_funding_allocation` (linked ke `gb_activity`)

Query penting:
```sql
-- name: GetGBProjectWithRelations :one
-- Join gb_project dengan semua relasi untuk satu kali fetch

-- name: ListGBActivitiesByProject :many
SELECT * FROM gb_activity WHERE gb_project_id = $1 ORDER BY sort_order ASC;

-- name: UpsertGBDisbursementPlan :one
INSERT INTO gb_disbursement_plan (gb_project_id, year, amount_usd)
VALUES ($1, $2, $3)
ON CONFLICT (gb_project_id, year) DO UPDATE SET amount_usd = $3, updated_at = NOW()
RETURNING *;

-- name: CreateGBFundingAllocation :one
INSERT INTO gb_funding_allocation (gb_activity_id, services, constructions, goods, trainings, other)
VALUES ($1, $2, $3, $4, $5, $6) RETURNING *;
```

Jalankan `make generate`.

---

## Task 2 — internal/service/green_book_service.go

```go
func (s *GBProjectService) CreateGBProject(ctx context.Context, gbID pgtype.UUID, req model.CreateGBProjectRequest) (*model.GBProjectResponse, error) {
    // Validasi bisnis:
    // 1. Minimal 1 BB Project
    if len(req.BBProjectIDs) == 0 {
        return nil, errors.Validation(errors.FieldError{Field: "bb_project_ids", Message: "Minimal 1 BB Project"})
    }
    // 2. EA dan IA boleh overlap bila sesuai data proyek
    // 3. gb_code unik dalam Green Book yang sama; boleh sama pada revisi lain via gb_project_identity_id
    // 4. BB Project yang dipilih di-resolve ke versi latest sebelum disimpan
    // 5. Jika lebih dari satu BB Project dipilih, semuanya harus resolve ke header Blue Book yang sama
    // 6. Mitra Kerja Bappenas opsional, multi-value, dan hanya boleh Eselon II
    // 7. Tahun disbursement tidak duplikat
    // 8. Duration, jika diisi, adalah integer jumlah bulan > 0
    yearSet := map[int]bool{}
    for _, d := range req.DisbursementPlan {
        if yearSet[d.Year] {
            return nil, errors.BusinessRule(fmt.Sprintf("Tahun %d duplikat di disbursement plan", d.Year))
        }
        yearSet[d.Year] = true
    }

    tx, err := s.db.Begin(ctx)
    defer tx.Rollback(ctx)
    qtx := s.queries.WithTx(tx)

    // Insert project
    project, err := qtx.CreateGBProject(ctx, ...)

    // Insert BB relations, Mitra Kerja Bappenas, institutions, locations
    for _, bbID := range req.BBProjectIDs { qtx.AddGBProjectBBProject(ctx, ...) }

    // Insert activities (sorted)
    activityIDs := []pgtype.UUID{}
    for i, act := range req.Activities {
        a, _ := qtx.CreateGBActivity(ctx, queries.CreateGBActivityParams{
            GbProjectID: project.ID, ActivityName: act.ActivityName,
            ImplementationLocation: act.ImplementationLocation, Piu: act.Piu, SortOrder: int32(i),
        })
        activityIDs = append(activityIDs, a.ID)
    }

    // Insert funding sources
    for _, fs := range req.FundingSources { qtx.CreateGBFundingSource(ctx, ...) }

    // Insert disbursement plan
    for _, dp := range req.DisbursementPlan { qtx.UpsertGBDisbursementPlan(ctx, ...) }

    // Insert funding allocation (activity_index → activityIDs[i])
    for _, fa := range req.FundingAllocations {
        if fa.ActivityIndex >= len(activityIDs) { continue }
        qtx.CreateGBFundingAllocation(ctx, queries.CreateGBFundingAllocationParams{
            GbActivityID: activityIDs[fa.ActivityIndex],
            Services: fa.Services, Constructions: fa.Constructions,
            Goods: fa.Goods, Trainings: fa.Trainings, Other: fa.Other,
        })
    }

    tx.Commit(ctx)
    return s.buildResponse(ctx, project), nil
}
```

---

## Task 3 — Handler & Routes

```go
gb := api.Group("/green-books")
gb.GET("", gbHandler.ListGreenBooks, permission.Require("green_book", "read"))
gb.POST("", gbHandler.CreateGreenBook, permission.Require("green_book", "create"))
gb.GET("/:id", gbHandler.GetGreenBook, permission.Require("green_book", "read"))
gb.PUT("/:id", gbHandler.UpdateGreenBook, permission.Require("green_book", "update"))
gb.DELETE("/:id", gbHandler.DeleteGreenBook, permission.Require("green_book", "delete"))

gb.GET("/:gbId/projects", gbHandler.ListGBProjects, permission.Require("gb_project", "read"))
gb.POST("/:gbId/projects", gbHandler.CreateGBProject, permission.Require("gb_project", "create"))
gb.GET("/:gbId/projects/:id", gbHandler.GetGBProject, permission.Require("gb_project", "read"))
gb.PUT("/:gbId/projects/:id", gbHandler.UpdateGBProject, permission.Require("gb_project", "update"))
gb.DELETE("/:gbId/projects/:id", gbHandler.DeleteGBProject, permission.Require("gb_project", "delete"))
```

---

## Checklist

- [x] `sql/queries/gb_project.sql` — semua query GB
- [x] `make generate`
- [x] `internal/model/green_book.go`
- [x] `internal/service/green_book_service.go` — validasi + transaksi dengan activity_index mapping
- [x] `internal/handler/green_book_handler.go`
- [x] Routes terdaftar
- [x] `POST /gb-projects` dengan duplikat tahun disbursement → 422
- [x] `POST /gb-projects` sukses → Activities + FundingAllocation tersimpan berurutan
