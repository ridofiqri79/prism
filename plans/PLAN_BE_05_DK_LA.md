# PLAN BE-05 — Daftar Kegiatan & Loan Agreement

> **Scope:** CRUD DK (header + project + sub-tabel multi-currency) dan LA multi DK Project dengan alokasi per project.
> **Deliverable:** DK tersimpan dengan lender dari Master Lender. LA tersimpan dengan validasi lender dari Financing Detail semua DK Project terkait, alokasi komitmen per project, dan deteksi perpanjangan.
> **Referensi:** docs/PRISM_API_Contract.md (DK & LA), docs/PRISM_Business_Rules.md (bagian 5 & 6)
> **Revision update:** Ikuti `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`. DK baru harus resolve pilihan GB Project ke versi latest, tetapi setelah DK/LA dibuat relasi downstream tetap frozen pada concrete snapshot yang tersimpan.

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

-- Catatan: duration bertipe integer jumlah bulan, nullable, dan jika diisi harus > 0.

-- name: DeleteDKProject :exec
DELETE FROM dk_project WHERE id=$1;

-- ===== DK GB JUNCTION =====
-- name: AddDKProjectGBProject :exec
INSERT INTO dk_project_gb_project (dk_project_id, gb_project_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: DeleteDKProjectGBProjects :exec
DELETE FROM dk_project_gb_project WHERE dk_project_id = $1;

-- ===== DK BAPPENAS PARTNER JUNCTION =====
-- name: AddDKProjectBappenasPartner :exec
INSERT INTO dk_project_bappenas_partner (dk_project_id, bappenas_partner_id) VALUES ($1, $2) ON CONFLICT DO NOTHING;

-- name: DeleteDKProjectBappenasPartners :exec
DELETE FROM dk_project_bappenas_partner WHERE dk_project_id = $1;

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

-- DK Financing Detail tidak membutuhkan query allowed lender dari GB/BB.
-- Validitas lender dijaga oleh foreign key ke Master Lender.
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

-- name: ListLoanAgreementsByDKProject :many
SELECT la.*, ladp.allocation_original, ladp.allocation_usd
FROM loan_agreement_dk_project ladp
JOIN loan_agreement la ON la.id = ladp.loan_agreement_id
WHERE ladp.dk_project_id = $1
ORDER BY la.created_at DESC, la.loan_code ASC;

-- name: ListLoanAgreementDKProjectsByLoanAgreements :many
SELECT ladp.*, dp.dk_id, dp.project_name, dk.subject, dk.date, dk.letter_number
FROM loan_agreement_dk_project ladp
JOIN dk_project dp ON dp.id = ladp.dk_project_id
JOIN daftar_kegiatan dk ON dk.id = dp.dk_id
WHERE ladp.loan_agreement_id = ANY($1::uuid[]);

-- name: CreateLoanAgreement :one
-- dk_project_id dipertahankan sebagai legacy primary reference; relasi resmi ada di loan_agreement_dk_project.
INSERT INTO loan_agreement (dk_project_id, lender_id, loan_code, agreement_date, effective_date, original_closing_date, closing_date, currency, amount_original, amount_usd)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING *;

-- name: UpdateLoanAgreement :one
UPDATE loan_agreement SET lender_id=$2, loan_code=$3, agreement_date=$4, effective_date=$5, original_closing_date=$6, closing_date=$7, currency=$8, amount_original=$9, amount_usd=$10, updated_at=NOW()
WHERE id=$1 RETURNING *;

-- name: DeleteLoanAgreement :exec
DELETE FROM loan_agreement WHERE id=$1;

-- name: AddLoanAgreementDKProject :exec
INSERT INTO loan_agreement_dk_project (loan_agreement_id, dk_project_id, allocation_original, allocation_usd)
VALUES ($1, $2, $3, $4)
ON CONFLICT (loan_agreement_id, dk_project_id) DO UPDATE
SET allocation_original = EXCLUDED.allocation_original,
    allocation_usd = EXCLUDED.allocation_usd,
    updated_at = NOW();

-- name: DeleteLoanAgreementDKProjects :exec
DELETE FROM loan_agreement_dk_project WHERE loan_agreement_id = $1;

-- name: GetAllowedLenderIDsForLAProjects :many
-- Ambil lender yang muncul pada financing detail semua DK Project terkait.
SELECT dfd.lender_id
FROM dk_financing_detail dfd
WHERE dfd.dk_project_id = ANY($1::uuid[]) AND dfd.lender_id IS NOT NULL
GROUP BY dfd.lender_id
HAVING COUNT(DISTINCT dfd.dk_project_id) = cardinality($1::uuid[]);
```

Jalankan `make generate`.

---

## Task 3 — Service: Simpan Financing Detail DK

```go
func (s *DKService) CreateDKProject(ctx context.Context, dkID pgtype.UUID, req model.CreateDKProjectRequest) (*model.DKProjectResponse, error) {
    // ... insert DK project, relasi GB, dan Mitra Kerja Bappenas dulu ...
    // Mitra Kerja Bappenas opsional, multi-value, dan hanya boleh Eselon II.

    // Insert financing details. Lender boleh dari Master Lender mana pun.
    // Foreign key memastikan lender_id yang dikirim valid.
    for _, fd := range req.FinancingDetails {
        // insert financing detail, currency aktif, amount non-negatif
    }
    // Insert loan allocations, activity details
}
```

---

## Task 4 — Service: Loan Agreement

```go
func (s *LAService) CreateLoanAgreement(ctx context.Context, req model.CreateLoanAgreementRequest) (*model.LAResponse, error) {
    // Validasi minimal satu DK Project, tidak duplikat, dan total allocation_original = amount_original.
    // Field dk_project_id lama hanya fallback single-project.
    // Validasi lender dari irisan DK financing detail semua project.
    allowedLenders, _ := s.queries.GetAllowedLenderIDsForLAProjects(ctx, dkProjectIDs)
    if !inSet(req.LenderID, allowedLenders) {
        return nil, errors.BusinessRule("Lender harus berasal dari Financing Detail semua DK Project terkait")
    }

    // Hitung allocation_usd per project dengan rounding adjustment agar total sama dengan amount_usd.
    // Create LA dan replace junction rows dilakukan dalam satu transaksi.
    la, err := s.queries.CreateLoanAgreement(ctx, ...)

    // Trigger SSE
    s.notification.Publish("loan_agreement.created", ...)

    return s.buildResponse(la), nil
}

func (s *LAService) buildResponse(la *queries.LoanAgreement) *model.LAResponse {
    return &model.LAResponse{
        // ...
        IsExtended:    la.OriginalClosingDate.Valid && la.ClosingDate != la.OriginalClosingDate,
        ExtensionDays: func() int { if !la.OriginalClosingDate.Valid { return 0 }; return int(la.ClosingDate.Time.Sub(la.OriginalClosingDate.Time).Hours() / 24) }(),
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

- [x] `sql/queries/dk_project.sql` — DK financing detail memakai FK lender, bukan allowed set GB/BB
- [x] `sql/queries/loan_agreement.sql` — termasuk junction `loan_agreement_dk_project`, `GetAllowedLenderIDsForLAProjects`, dan `ListLoanAgreementsByDKProject`
- [x] `make generate`
- [x] `internal/model/daftar_kegiatan.go` + `internal/model/loan_agreement.go`
- [x] `internal/service/dk_service.go` — simpan lender DK dari Master Lender
- [x] `internal/service/la_service.go` — validasi lender semua project + alokasi + loan_code unik + computed is_extended
- [x] Handler DK dan LA
- [x] Routes terdaftar
- [x] `POST /dk-projects` dengan lender valid dari Master Lender meski tidak dari GB/BB → 201
- [x] `POST /loan-agreements` mendukung satu atau banyak DK Project, termasuk lintas header DK, dengan total alokasi sama dengan `amount_original`
- [x] `is_extended = true` saat `original_closing_date` diisi dan `closing_date != original_closing_date`
