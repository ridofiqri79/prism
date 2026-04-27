# PLAN BE-05 — Daftar Kegiatan & Loan Agreement

> **Scope:** CRUD DK (header + project + sub-tabel multi-currency) dan LA (One-to-One dengan DK).
> **Deliverable:** DK tersimpan dengan validasi lender. LA tersimpan dengan deteksi perpanjangan.
> **Referensi:** docs/PRISM_API_Contract.md (DK & LA), docs/PRISM_Business_Rules.md (bagian 5 & 6)

---

## Task 1 — sql/queries/dk_project.sql

```sql
-- ===== DAFTAR KEGIATAN =====
-- name: ListDaftarKegiatan :many
SELECT * FROM daftar_kegiatan ORDER BY date DESC LIMIT $1 OFFSET $2;

-- name: GetDaftarKegiatan :one
SELECT * FROM daftar_kegiatan WHERE id = $1;

-- name: CreateDaftarKegiatan :one
INSERT INTO daftar_kegiatan (letter_number, subject, date) VALUES ($1, $2, $3) RETURNING *;

-- name: UpdateDaftarKegiatan :one
UPDATE daftar_kegiatan SET letter_number=$2, subject=$3, date=$4, updated_at=NOW() WHERE id=$1 RETURNING *;

-- ===== DK PROJECT =====
-- name: ListDKProjectsByDK :many
SELECT * FROM dk_project WHERE dk_id = $1;

-- name: GetDKProject :one
SELECT * FROM dk_project WHERE id = $1;

-- name: CreateDKProject :one
INSERT INTO dk_project (dk_id, program_title_id, institution_id, duration, objectives)
VALUES ($1, $2, $3, $4, $5) RETURNING *;

-- name: UpdateDKProject :one
UPDATE dk_project SET program_title_id=$2, institution_id=$3, duration=$4, objectives=$5, updated_at=NOW()
WHERE id=$1 RETURNING *;

-- name: DeleteDKProject :exec
DELETE FROM dk_project WHERE id=$1;

-- ===== DK GB JUNCTION =====
-- name: AddDKProjectGBProject :exec
INSERT INTO dk_project_gb_project (dk_project_id, gb_project_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: DeleteDKProjectGBProjects :exec
DELETE FROM dk_project_gb_project WHERE dk_project_id = $1;

-- ===== FINANCING DETAIL =====
-- name: GetDKFinancingDetails :many
SELECT df.*, l.name as lender_name FROM dk_financing_detail df
LEFT JOIN lender l ON l.id = df.lender_id WHERE df.dk_project_id = $1;

-- name: CreateDKFinancingDetail :one
INSERT INTO dk_financing_detail (dk_project_id, lender_id, currency, amount_original, grant_original, counterpart_original, amount_usd, grant_usd, counterpart_usd, remarks)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *;

-- name: DeleteDKFinancingDetails :exec
DELETE FROM dk_financing_detail WHERE dk_project_id = $1;

-- ===== LOAN ALLOCATION =====
-- name: GetDKLoanAllocations :many
SELECT dla.*, i.name as institution_name FROM dk_loan_allocation dla
LEFT JOIN institution i ON i.id = dla.institution_id WHERE dla.dk_project_id = $1;

-- name: CreateDKLoanAllocation :one
INSERT INTO dk_loan_allocation (dk_project_id, institution_id, currency, amount_original, grant_original, counterpart_original, amount_usd, grant_usd, counterpart_usd, remarks)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *;

-- name: DeleteDKLoanAllocations :exec
DELETE FROM dk_loan_allocation WHERE dk_project_id = $1;

-- ===== ACTIVITY DETAIL =====
-- name: GetDKActivityDetails :many
SELECT * FROM dk_activity_detail WHERE dk_project_id = $1 ORDER BY activity_number ASC;

-- name: CreateDKActivityDetail :one
INSERT INTO dk_activity_detail (dk_project_id, activity_number, activity_name)
VALUES ($1, $2, $3) RETURNING *;

-- name: DeleteDKActivityDetails :exec
DELETE FROM dk_activity_detail WHERE dk_project_id = $1;

-- Ambil lender ID yang diperbolehkan untuk DK project tertentu:
-- (dari lender_indication BB terkait + gb_funding_source GB terkait)
-- name: GetAllowedLenderIDsForDK :many
SELECT DISTINCT lender_id FROM (
    SELECT li.lender_id
    FROM dk_project_gb_project dkgb
    JOIN gb_project_bb_project gbbb ON gbbb.gb_project_id = dkgb.gb_project_id
    JOIN lender_indication li ON li.bb_project_id = gbbb.bb_project_id
    WHERE dkgb.dk_project_id = $1
    UNION
    SELECT gfs.lender_id
    FROM dk_project_gb_project dkgb
    JOIN gb_funding_source gfs ON gfs.gb_project_id = dkgb.gb_project_id
    WHERE dkgb.dk_project_id = $1
) allowed_lenders;
```

---

## Task 2 — sql/queries/loan_agreement.sql

```sql
-- name: ListLoanAgreements :many
SELECT la.*, l.name as lender_name, l.type as lender_type
FROM loan_agreement la JOIN lender l ON l.id = la.lender_id
ORDER BY la.created_at DESC LIMIT $1 OFFSET $2;

-- name: GetLoanAgreement :one
SELECT la.*, l.name as lender_name FROM loan_agreement la
JOIN lender l ON l.id = la.lender_id WHERE la.id = $1;

-- name: GetLoanAgreementByDKProject :one
SELECT * FROM loan_agreement WHERE dk_project_id = $1;

-- name: CreateLoanAgreement :one
INSERT INTO loan_agreement (dk_project_id, lender_id, loan_code, agreement_date, effective_date, original_closing_date, closing_date, currency, amount_original, amount_usd)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *;

-- name: UpdateLoanAgreement :one
UPDATE loan_agreement SET lender_id=$2, loan_code=$3, agreement_date=$4, effective_date=$5, original_closing_date=$6, closing_date=$7, currency=$8, amount_original=$9, amount_usd=$10, updated_at=NOW()
WHERE id=$1 RETURNING *;

-- name: DeleteLoanAgreement :exec
DELETE FROM loan_agreement WHERE id=$1;

-- name: GetAllowedLenderIDsForLA :many
-- Ambil lender dari dk_financing_detail untuk DK Project tertentu
SELECT DISTINCT lender_id FROM dk_financing_detail WHERE dk_project_id = $1 AND lender_id IS NOT NULL;
```

Jalankan `make generate`.

---

## Task 3 — Service: Validasi Lender DK

```go
func (s *DKService) CreateDKProject(ctx context.Context, dkID pgtype.UUID, req model.CreateDKProjectRequest) (*model.DKProjectResponse, error) {
    // ... insert DK project dan relasi GB dulu ...

    // Setelah GB relations tersimpan, fetch allowed lender IDs
    allowedLenders, _ := qtx.GetAllowedLenderIDsForDK(ctx, dkProject.ID)
    allowedSet := toUUIDSet(allowedLenders)

    // Validasi setiap financing detail
    for _, fd := range req.FinancingDetails {
        if fd.LenderID != nil && !allowedSet[*fd.LenderID] {
            return nil, errors.BusinessRule("Lender tidak terdaftar di GB atau BB terkait")
        }
    }
    // Insert financing details, loan allocations, activity details
}
```

---

## Task 4 — Service: Loan Agreement

```go
func (s *LAService) CreateLoanAgreement(ctx context.Context, req model.CreateLoanAgreementRequest) (*model.LAResponse, error) {
    // Cek DK sudah punya LA
    existing, _ := s.queries.GetLoanAgreementByDKProject(ctx, req.DKProjectID)
    if existing != nil {
        return nil, errors.Conflict("DK Project sudah memiliki Loan Agreement")
    }

    // Validasi lender dari DK financing detail
    allowedLenders, _ := s.queries.GetAllowedLenderIDsForLA(ctx, req.DKProjectID)
    if !inSet(req.LenderID, allowedLenders) {
        return nil, errors.BusinessRule("Lender harus berasal dari Financing Detail DK Project terkait")
    }

    la, err := s.queries.CreateLoanAgreement(ctx, ...)

    // Trigger SSE
    s.notification.Publish("loan_agreement.created", ...)

    return s.buildResponse(la), nil
}

func (s *LAService) buildResponse(la *queries.LoanAgreement) *model.LAResponse {
    return &model.LAResponse{
        // ...
        IsExtended:    la.ClosingDate != la.OriginalClosingDate,
        ExtensionDays: int(la.ClosingDate.Time.Sub(la.OriginalClosingDate.Time).Hours() / 24),
    }
}
```

---

## Task 5 — Handler & Routes

```go
// Daftar Kegiatan
dk := api.Group("/daftar-kegiatan")
dk.GET("", dkHandler.ListDK, permission.Require("daftar_kegiatan", "read"))
dk.POST("", dkHandler.CreateDK, permission.Require("daftar_kegiatan", "create"))
dk.GET("/:id", dkHandler.GetDK, permission.Require("daftar_kegiatan", "read"))
dk.PUT("/:id", dkHandler.UpdateDK, permission.Require("daftar_kegiatan", "update"))

// DK Projects
dk.GET("/:dkId/projects", dkHandler.ListDKProjects, permission.Require("daftar_kegiatan", "read"))
dk.POST("/:dkId/projects", dkHandler.CreateDKProject, permission.Require("daftar_kegiatan", "create"))
dk.GET("/:dkId/projects/:id", dkHandler.GetDKProject, permission.Require("daftar_kegiatan", "read"))
dk.PUT("/:dkId/projects/:id", dkHandler.UpdateDKProject, permission.Require("daftar_kegiatan", "update"))
dk.DELETE("/:dkId/projects/:id", dkHandler.DeleteDKProject, permission.Require("daftar_kegiatan", "delete"))

// Loan Agreement
la := api.Group("/loan-agreements")
la.GET("", laHandler.ListLA, permission.Require("loan_agreement", "read"))
la.POST("", laHandler.CreateLA, permission.Require("loan_agreement", "create"))
la.GET("/:id", laHandler.GetLA, permission.Require("loan_agreement", "read"))
la.PUT("/:id", laHandler.UpdateLA, permission.Require("loan_agreement", "update"))
la.DELETE("/:id", laHandler.DeleteLA, permission.Require("loan_agreement", "delete"))
```

---

## Checklist

- [ ] `sql/queries/dk_project.sql` — termasuk `GetAllowedLenderIDsForDK`
- [ ] `sql/queries/loan_agreement.sql` — termasuk `GetAllowedLenderIDsForLA`
- [ ] `make generate`
- [ ] `internal/model/daftar_kegiatan.go` + `internal/model/loan_agreement.go`
- [ ] `internal/service/dk_service.go` — validasi lender dari allowed set
- [ ] `internal/service/la_service.go` — cek duplicate + validasi lender + computed is_extended
- [ ] Handler DK dan LA
- [ ] Routes terdaftar
- [ ] `POST /dk-projects` dengan lender tidak dari GB/BB → 422
- [ ] `POST /loan-agreements` untuk DK yang sudah punya LA → 409
- [ ] `is_extended = true` saat `closing_date != original_closing_date`
