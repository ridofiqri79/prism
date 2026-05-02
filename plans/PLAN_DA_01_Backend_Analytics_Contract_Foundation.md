# PLAN DA-01 - Backend Analytics Contract Foundation

> **Scope:** Menyiapkan fondasi backend Dashboard Analytics: API contract, DTO, filter parser, route, dan query base.
> **Deliverable:** Endpoint foundation siap dipakai phase agregasi berikutnya.
> **Dependencies:** `PLAN_BE_06_Monitoring.md`, `PLAN_BE_11_Journey_Import_Project_List_Versioning.md`, `PLAN_DA_00_Dashboard_Analytics_Roadmap.md`.

---

## Instruksi Untuk Agent

Baca sebelum mulai:

- `AGENTS.md`
- `docs/PRISM_Business_Rules.md`
- `docs/prism_ddl.sql`
- `docs/PRISM_API_Contract.md`
- `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`
- `docs/PRISM_Backend_Structure.md`
- `plans/PLAN_DA_00_Dashboard_Analytics_Roadmap.md`

Aturan kerja:

- Kerjakan backend query-first.
- Tulis SQL source di `prism-backend/sql/queries/`, bukan di file Go.
- Setelah SQL berubah, jalankan `make generate` dari `prism-backend`.
- Jangan edit `internal/database/queries/*.go` manual.
- Jangan tambah dependency baru.
- Jangan ubah semantics endpoint dashboard lama kecuali memang diperlukan.

---

## Task 1 - Update API Contract

Update `docs/PRISM_API_Contract.md` bagian `Dashboard & Aggregasi`.

Tambahkan subsection:

- `GET /dashboard/analytics/overview`
- `GET /dashboard/analytics/institutions`
- `GET /dashboard/analytics/lenders`
- `GET /dashboard/analytics/absorption`
- `GET /dashboard/analytics/yearly`
- `GET /dashboard/analytics/lender-proportion`
- `GET /dashboard/analytics/risks`

Tambahkan filter umum:

```text
budget_year
quarter
lender_ids
lender_types
institution_ids
pipeline_statuses
project_statuses
region_ids
program_title_ids
foreign_loan_min
foreign_loan_max
include_history
```

Catatan contract:

- Default `include_history=false`.
- Semua hitungan project portfolio default latest snapshot.
- `absorption_pct` computed server-side.
- Stage lender harus eksplisit: indication, funding source, agreement, monitoring.
- Drilldown response memakai query object yang bisa diterjemahkan frontend ke Project Master atau Monitoring.

---

## Task 2 - Buat SQL Foundation

Buat file:

```text
prism-backend/sql/queries/dashboard_analytics.sql
```

Isi minimal:

1. CTE `latest_bb_project_rows`
   - satu row latest per `project_identity_id`,
   - mengikuti ordering Project Master yang sudah ada,
   - hormati `include_history`.

2. CTE `project_portfolio_rows`
   - boleh mengadaptasi logic dari `project.sql`,
   - membawa `pipeline_status`,
   - membawa `project_status`,
   - membawa `loan_types`,
   - membawa `indication_lender_ids`,
   - membawa `fixed_lender_ids`,
   - membawa `executing_agency_ids`,
   - membawa `region_ids`,
   - membawa nilai funding USD.

3. CTE `monitoring_fact_rows`
   - source: `monitoring_disbursement -> loan_agreement -> dk_project -> lender`,
   - membawa `budget_year`, `quarter`, `planned_usd`, `realized_usd`, `agreement_amount_usd`,
   - membawa `lender_id`, `lender_type`, `dk_project_id`, `institution_id`.

4. CTE `institution_rollup`
   - recursive CTE untuk resolve institution ke root ancestor,
   - expose `source_institution_id`, `root_institution_id`, `root_institution_name`, `root_institution_level`.

Catatan:

- Jika sqlc tidak nyaman dengan reusable CTE lintas query, ulangi CTE di query masing-masing. Utamakan correctness dan type safety.
- Jangan membuat view DB dulu kecuali benar-benar diperlukan dan disetujui user.

---

## Task 3 - Model DTO Foundation

Update atau buat file:

```text
prism-backend/internal/model/dashboard_analytics.go
```

DTO minimal:

```go
type DashboardAnalyticsFilter struct {
    BudgetYear       *int32
    Quarter          *string
    LenderIDs        []string
    LenderTypes      []string
    InstitutionIDs   []string
    PipelineStatuses []string
    ProjectStatuses  []string
    RegionIDs        []string
    ProgramTitleIDs  []string
    ForeignLoanMin   *float64
    ForeignLoanMax   *float64
    IncludeHistory   bool
}

type DashboardDrilldownQuery struct {
    Target string `json:"target"`
    Query  map[string][]string `json:"query"`
}
```

Gunakan strongly typed struct untuk response konkret. Drilldown query memakai `map[string][]string` karena targetnya adalah query parameter URL.

---

## Task 4 - Filter Parser

Update atau buat helper di handler:

```text
prism-backend/internal/handler/dashboard_analytics_handler.go
```

Parser harus mendukung:

- repeated query param,
- comma-separated,
- array suffix `[]`,
- UUID validation untuk ID filters,
- enum validation untuk `quarter`, `lender_types`, `pipeline_statuses`, `project_statuses`.

Jika filter invalid, return validation error dengan field yang jelas.

---

## Task 5 - Service Foundation

Buat:

```text
prism-backend/internal/service/dashboard_analytics_service.go
```

Service bertanggung jawab untuk:

- convert filter model ke sqlc params,
- menghitung `absorption_pct`,
- membangun drilldown query object,
- menjaga div-by-zero safe,
- menjaga label stage lender tidak tertukar.

Jangan taruh parsing query param di service.

---

## Task 6 - Handler dan Route Skeleton

Buat handler method skeleton:

- `Overview`
- `Institutions`
- `Lenders`
- `Absorption`
- `Yearly`
- `LenderProportion`
- `Risks`

Route:

```go
analytics := api.Group("/dashboard/analytics")
analytics.GET("/overview", dashboardAnalyticsHandler.Overview)
analytics.GET("/institutions", dashboardAnalyticsHandler.Institutions)
analytics.GET("/lenders", dashboardAnalyticsHandler.Lenders)
analytics.GET("/absorption", dashboardAnalyticsHandler.Absorption)
analytics.GET("/yearly", dashboardAnalyticsHandler.Yearly)
analytics.GET("/lender-proportion", dashboardAnalyticsHandler.LenderProportion)
analytics.GET("/risks", dashboardAnalyticsHandler.Risks)
```

Permission awal:

- Ikuti dashboard lama: authenticated.
- Jangan tambah permission module baru kecuali user menyetujui eksplisit.

---

## Acceptance Criteria

- API contract memuat semua endpoint analytics target.
- `dashboard_analytics.sql` ada dan bisa digenerate sqlc.
- `make generate` berhasil.
- Backend compile.
- Route skeleton mengembalikan response valid, walaupun endpoint yang agregasinya belum dikerjakan boleh return struktur kosong yang contract-safe.
- Tidak ada raw SQL di Go.
- Tidak ada perubahan frontend di phase ini.

---

## Verification

Jalankan dari `prism-backend`:

```powershell
make generate
go test ./...
```

Jika `make` tidak tersedia di Windows, gunakan perintah repo yang setara dan catat di hasil kerja.

---

## Checklist

- [x] `docs/PRISM_API_Contract.md` update bagian Dashboard Analytics
- [x] `sql/queries/dashboard_analytics.sql` dibuat
- [x] Query foundation latest snapshot dibuat
- [x] Query foundation monitoring fact dibuat
- [x] Query foundation institution rollup dibuat
- [x] `make generate` berhasil
- [x] `internal/model/dashboard_analytics.go` dibuat
- [x] `internal/service/dashboard_analytics_service.go` dibuat
- [x] `internal/handler/dashboard_analytics_handler.go` dibuat
- [x] Routes `/dashboard/analytics/*` terdaftar
- [x] `go test ./...` berhasil atau blocker dicatat
