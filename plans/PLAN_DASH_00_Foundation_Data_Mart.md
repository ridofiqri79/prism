# PRISM Dashboard Implementation Plan — Codex Phase Guide

Target: implementasi 7 dashboard inti PRISM secara bertahap, aman terhadap relasi BB → GB → DK → LA → Monitoring, dan konsisten dengan stack backend Go/Echo/sqlc/pgx serta frontend Vue 3/Pinia/PrimeVue/ECharts.

> Cara pakai dengan Codex: jalankan satu phase per sesi. Jangan meminta Codex mengerjakan semua phase sekaligus. Setelah satu phase selesai, lakukan test, commit, lalu lanjut ke phase berikutnya.

## Dokumen yang wajib dibaca Codex sebelum coding

- `docs/PRISM_Project_Idea.md`
- `docs/PRISM_Business_Rules.md`
- `docs/PRISM_API_Contract.md`
- `docs/PRISM_Backend_Structure.md`
- `docs/PRISM_Frontend_Structure.md`
- `sql/schema/prism_ddl.sql`

## Prinsip umum dashboard

1. Dashboard bersifat read-only. Tidak boleh ada mutation data dari endpoint dashboard.
2. Default dashboard memakai snapshot latest untuk BB/GB agar revisi tidak double count.
3. Downstream DK/LA tetap memakai concrete snapshot yang tersimpan saat DK dibuat.
4. Nilai USD tidak dikonversi otomatis. Gunakan nilai USD yang sudah tersimpan di database.
5. Untuk monitoring, gunakan `realized_usd`, `planned_usd`, dan `amount_usd` dari LA sebagai basis utama.
6. Untuk wilayah, jangan menggandakan nilai proyek nasional/provinsi ke seluruh kota/kabupaten.
7. Semua endpoint mengikuti response `{ data, meta }` dan error contract yang sudah ada.
8. Semua query backend ditulis di `sql/queries/dashboard.sql`, lalu jalankan `make generate`.
9. Frontend wajib melalui `dashboard.service.ts`, `dashboard.store.ts`, dan types di `dashboard.types.ts`.
10. UI memakai PrimeVue untuk komponen, Tailwind hanya untuk layout/spacing, dan ECharts untuk chart.

## 7 Dashboard Inti

1. Executive Portfolio Dashboard
2. Pipeline & Bottleneck Dashboard
3. Green Book Readiness Dashboard
4. Lender & Financing Mix Dashboard
5. K/L Portfolio Performance Dashboard
6. Loan Agreement & Disbursement Dashboard
7. Data Quality & Governance Dashboard


---

# Phase 0 — Dashboard Foundation & Data Mart

## Objective

Membangun fondasi backend agar 7 dashboard tidak mengambil data mentah secara berulang, tidak double count akibat revisi BB/GB, dan memiliki definisi metrik yang konsisten.

## Scope

- Backend only.
- Tidak membuat UI dashboard dulu.
- Membuat query aggregate dasar dan response model bersama.
- Menyiapkan contract endpoint dashboard.

## Output yang Diharapkan

- `sql/queries/dashboard.sql`
- Generated sqlc files di `internal/database/queries/`
- `internal/model/dashboard.go`
- `internal/service/dashboard_service.go`
- `internal/handler/dashboard_handler.go`
- Route group `/api/v1/dashboard/*`
- Basic tests untuk query/service dashboard.

## Metrik Dasar yang Harus Distandarkan

| Metrik | Definisi |
|---|---|
| `bb_pipeline_usd` | SUM `bb_project_cost.amount_usd` untuk `funding_type = 'Foreign'` |
| `gb_pipeline_usd` | SUM `gb_funding_source.loan_usd + gb_funding_source.grant_usd` |
| `gb_local_usd` | SUM `gb_funding_source.local_usd` |
| `dk_financing_usd` | SUM `dk_financing_detail.amount_usd + dk_financing_detail.grant_usd` |
| `dk_counterpart_usd` | SUM `dk_financing_detail.counterpart_usd` |
| `la_commitment_usd` | SUM `loan_agreement.amount_usd` |
| `planned_disbursement_usd` | SUM `monitoring_disbursement.planned_usd` |
| `realized_disbursement_usd` | SUM `monitoring_disbursement.realized_usd` |
| `absorption_pct` | `realized_usd / planned_usd * 100`, jika planned = 0 maka 0 |
| `la_absorption_pct` | `cumulative_realized_usd / loan_agreement.amount_usd * 100` |
| `undisbursed_usd` | `loan_agreement.amount_usd - cumulative_realized_usd` |

## Status Derivation

Codex harus membuat helper query untuk menurunkan status proyek dari relasi, bukan dari field status eksplisit.

| Stage | Definisi Minimal |
|---|---|
| `BB_ONLY` | BB Project belum punya LoI dan belum punya GB relation |
| `BB_WITH_LENDER_INDICATION` | Ada `lender_indication`, belum ada LoI |
| `BB_WITH_LOI` | Ada LoI, belum masuk GB |
| `GB` | Ada relasi `gb_project_bb_project`, belum masuk DK |
| `DK` | Ada relasi `dk_project_gb_project`, belum ada LA |
| `LA_SIGNED_NOT_EFFECTIVE` | Ada LA tetapi `effective_date > CURRENT_DATE` |
| `LA_EFFECTIVE_NO_MONITORING` | Ada LA efektif tetapi belum ada monitoring |
| `MONITORING_ACTIVE` | Ada LA efektif dan minimal satu monitoring |

## Backend Tasks

1. Buat file `sql/queries/dashboard.sql`.
2. Tambahkan query dasar:
   - `GetDashboardSummary`
   - `GetDashboardStageCounts`
   - `GetDashboardStageAmounts`
   - `GetDashboardMonitoringRollup`
   - `GetDashboardLAExposureRollup`
   - `GetDashboardLenderRollup`
   - `GetDashboardInstitutionRollup`
3. Semua query harus menerima filter opsional:
   - `period_id`
   - `publish_year`
   - `budget_year`
   - `quarter`
   - `lender_id`
   - `institution_id`
   - `include_history`
4. Gunakan latest snapshot default untuk BB/GB:
   - BB latest per `project_identity_id`
   - GB latest per `gb_project_identity_id`
5. Jangan membuat materialized view dulu kecuali query terbukti lambat.
6. Tambahkan DTO umum di `internal/model/dashboard.go`:
   - `DashboardFilterRequest`
   - `MetricCard`
   - `StageMetric`
   - `TimeSeriesPoint`
   - `BreakdownItem`
   - `RiskItem`
7. Tambahkan service read-only.
8. Tambahkan handler dan route.
9. Jalankan:
   - `make generate`
   - `go test ./...`

## Suggested API Contract

```http
GET /api/v1/dashboard/summary
GET /api/v1/dashboard/stage-funnel
GET /api/v1/dashboard/monitoring-rollup
GET /api/v1/dashboard/filter-options
```

## Acceptance Criteria

- Query tidak double count BB/GB akibat revisi.
- Semua endpoint dashboard read-only.
- Semua response memakai format `{ "data": ... }`.
- Jika filter kosong, dashboard memakai data latest/current secara default.
- `go test ./...` lulus.

## Prompt Codex

```text
You are working on PRISM dashboard Phase 0 only.
Read docs/PRISM_Project_Idea.md, docs/PRISM_Business_Rules.md, docs/PRISM_API_Contract.md, docs/PRISM_Backend_Structure.md, and sql/schema/prism_ddl.sql first.
Implement dashboard backend foundation only: dashboard.sql sqlc queries, dashboard DTOs, dashboard service, dashboard handler, and routes under /api/v1/dashboard.
Do not build frontend yet. Do not mutate existing business modules. Do not write raw SQL in Go files. All SQL must be in sql/queries/dashboard.sql and then run make generate.
Use latest BB/GB snapshots by default and avoid double counting revision snapshots. Add tests where feasible and run go test ./...
```
