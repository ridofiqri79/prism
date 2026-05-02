# PLAN DA-03 - Backend Risk, Data Quality, and Drilldown Analytics

> **Scope:** Tambah insight lanjutan dashboard: risk watchlist, data quality, monitoring compliance, closing risk, bottleneck, dan drilldown query.
> **Deliverable:** Endpoint `/dashboard/analytics/risks` dan drilldown metadata yang bisa dipakai frontend untuk navigasi filter.
> **Dependencies:** `PLAN_DA_01_Backend_Analytics_Contract_Foundation.md`, `PLAN_DA_02_Backend_KL_Lender_Absorption.md`.

---

## Instruksi Untuk Agent

Baca sebelum mulai:

- `AGENTS.md`
- `docs/PRISM_Business_Rules.md`
- `docs/prism_ddl.sql`
- `docs/PRISM_API_Contract.md`
- `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`
- `plans/PLAN_DA_00_Dashboard_Analytics_Roadmap.md`
- `plans/PLAN_DA_01_Backend_Analytics_Contract_Foundation.md`
- `plans/PLAN_DA_02_Backend_KL_Lender_Absorption.md`

Aturan:

- Endpoint risk harus menjelaskan basis hitungannya.
- Jangan membuat data quality issue yang tidak bisa diklik/drilldown.
- Jangan memalsukan alokasi wilayah ke kabupaten/kota.
- Jangan auto-convert currency.

---

## Task 1 - Risk Endpoint Contract

Endpoint:

```text
GET /dashboard/analytics/risks
```

Query tambahan:

| Param | Default | Keterangan |
|-------|---------|------------|
| `low_absorption_threshold` | `50` | Ambang penyerapan rendah |
| `closing_months_threshold` | `12` | Ambang closing risk |
| `stale_monitoring_quarters` | `1` | Jumlah triwulan toleransi monitoring |

Response target:

```json
{
  "data": {
    "summary": {
      "low_absorption_count": 0,
      "effective_without_monitoring_count": 0,
      "closing_risk_count": 0,
      "extended_loan_count": 0,
      "data_quality_issue_count": 0,
      "bottleneck_project_count": 0
    },
    "risk_cards": [],
    "watchlists": {
      "low_absorption_projects": [],
      "effective_without_monitoring": [],
      "closing_risks": [],
      "extended_loans": [],
      "pipeline_bottlenecks": []
    },
    "data_quality": []
  }
}
```

---

## Task 2 - Low Absorption Watchlist

Metric:

- Project atau Loan Agreement dengan `planned_usd > 0`.
- `realized_usd / planned_usd * 100 < low_absorption_threshold`.

Item fields:

```json
{
  "project_id": "uuid",
  "project_name": "Nama Project",
  "loan_agreement_id": "uuid",
  "loan_code": "IP-001",
  "lender": {},
  "institution": {},
  "budget_year": 2025,
  "quarter": "TW1",
  "planned_usd": 0,
  "realized_usd": 0,
  "absorption_pct": 0,
  "drilldown": {}
}
```

Sort:

1. absorption ascending,
2. planned_usd descending,
3. project name ascending.

---

## Task 3 - Effective Without Monitoring

Metric:

- Loan Agreement dengan `effective_date <= CURRENT_DATE`.
- Tidak punya monitoring pada filter `budget_year`/`quarter`, atau tidak punya monitoring sama sekali jika filter kosong.

Item fields:

- Loan Agreement,
- project,
- lender,
- Kementerian/Lembaga,
- effective date,
- days since effective,
- drilldown ke Monitoring List atau Loan Agreement detail.

Catatan:

- Jangan menolak data lama hanya karena budget year kosong.
- Jika filter periode aktif, issue harus periode-specific.

---

## Task 4 - Closing Risk

Metric default:

- `closing_date <= CURRENT_DATE + interval '12 months'`
- absorption `< 80`
- Loan Agreement masih efektif.

Item fields:

- loan code,
- project name,
- lender,
- Kementerian/Lembaga,
- closing date,
- months to closing,
- agreement amount USD,
- realized USD,
- absorption percent.

Jadikan threshold 80 sebagai service constant dan dokumentasikan di contract.

---

## Task 5 - Extended Loan Insight

Metric:

- `closing_date > original_closing_date`.

Response:

- count,
- amount USD,
- average extension days,
- breakdown by lender,
- breakdown by Kementerian/Lembaga,
- watchlist top extended loan.

Gunakan view helper `loan_agreement_extended` jika sudah tersedia dan cocok.

---

## Task 6 - Pipeline Bottleneck

Metric:

- Project yang berhenti di stage tertentu tanpa downstream berikutnya.

Stage examples:

- `BB` tanpa Green Book,
- `GB` tanpa Daftar Kegiatan,
- `DK` tanpa Loan Agreement,
- `LA` efektif tanpa Monitoring.

Response fields:

- stage,
- project count,
- total loan USD,
- oldest updated/created date jika tersedia,
- drilldown ke Project Master dengan `pipeline_statuses`.

Catatan:

- Untuk `BB` dan `GB`, gunakan latest snapshot default.
- Jangan menyebut sebagai "terlambat" jika tidak ada SLA bisnis. Pakai label "Belum berlanjut".

---

## Task 7 - Data Quality Cards

Cards minimal:

| Code | Makna | Drilldown target |
|------|-------|------------------|
| `NO_EXECUTING_AGENCY` | Project tanpa Executing Agency | Project Master |
| `NO_LENDER` | Project tanpa lender indication/funding/LA sesuai stage | Project Master |
| `NO_REGION` | Project tanpa location | Project Master |
| `NO_FUNDING_AMOUNT` | Funding amount USD kosong/0 | Project Master |
| `EFFECTIVE_NO_MONITORING` | LA efektif belum dimonitor | Monitoring |
| `PLANNED_ZERO_REALIZED_POSITIVE` | Realisasi ada tetapi planned 0 | Monitoring |

Response item:

```json
{
  "code": "NO_EXECUTING_AGENCY",
  "label": "Project tanpa Executing Agency",
  "count": 0,
  "severity": "warning",
  "drilldown": {
    "target": "projects",
    "query": { "missing_fields": ["NO_EXECUTING_AGENCY"] }
  }
}
```

Jika Project Master belum support `missing_fields`, tambahkan plan subtask untuk backend/frontend Project Master filter. Jangan hardcode route yang tidak bisa membaca query.

---

## Task 8 - Drilldown Query Builder

Service harus punya helper untuk menghasilkan drilldown:

```go
func projectDrilldown(query model.ProjectDrilldownFilter) model.DashboardDrilldownQuery
func monitoringDrilldown(query model.MonitoringDrilldownFilter) model.DashboardDrilldownQuery
```

Target yang boleh:

- `projects`
- `monitoring`
- `loan_agreements`
- `spatial_distribution`

Rules:

- Query harus memakai nama param yang benar-benar didukung frontend target.
- Jika target filter belum ada, tulis task eksplisit untuk menambah filter target.
- Jangan membuat URL string mentah di backend. Return structured query object.

---

## Task 9 - Tests

Tambahkan test untuk:

- low absorption div-by-zero aman,
- effective LA tanpa monitoring terdeteksi,
- non-effective LA tidak masuk effective-without-monitoring,
- closing risk threshold,
- extended loan dihitung dari tanggal,
- data quality issue menghasilkan drilldown query,
- bottleneck stage tidak double count revisi.

---

## Acceptance Criteria

- `/dashboard/analytics/risks` mengembalikan summary, cards, watchlists, data quality.
- Semua risk/data quality item punya drilldown.
- Tidak ada issue yang basis datanya ambigu.
- Test backend mencakup threshold dan drilldown.
- Tidak ada perubahan frontend di phase ini kecuali contract docs jika perlu.

---

## Verification

```powershell
cd prism-backend
make generate
go test ./...
```

---

## Checklist

- [x] API contract `/dashboard/analytics/risks` update
- [x] Low absorption watchlist query/service
- [x] Effective without monitoring query/service
- [x] Closing risk query/service
- [x] Extended loan insight query/service
- [x] Pipeline bottleneck query/service
- [x] Data quality cards query/service
- [x] Drilldown query builder
- [x] Backend tests
- [x] `make generate` berhasil
- [x] `go test ./...` berhasil atau blocker dicatat
