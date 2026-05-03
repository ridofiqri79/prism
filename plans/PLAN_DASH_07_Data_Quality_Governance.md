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

# Phase 7 — Data Quality & Governance Dashboard

## Objective

Membangun dashboard untuk memastikan PRISM dapat menjadi single source of truth: data lengkap, konsisten, tidak melanggar aturan bisnis, dan perubahan penting dapat diaudit oleh ADMIN.

## Dashboard Questions

1. Data apa yang kosong atau tidak lengkap?
2. Relasi mana yang putus atau tidak logis?
3. Apakah ada nilai USD/original/currency yang tidak konsisten?
4. Apakah ada LA efektif yang belum dimonitor?
5. Siapa user yang paling banyak melakukan perubahan data?
6. Tabel/modul mana yang paling sering berubah?

## Backend Scope

Endpoint:

```http
GET /api/v1/dashboard/data-quality-governance
```

Query params:

| Param | Type |
|---|---|
| `severity` | info/warning/error optional |
| `module` | string optional |
| `issue_type` | string optional |
| `only_unresolved` | bool optional |
| `audit_days` | int default 30 |

Response:

```json
{
  "data": {
    "summary": {
      "total_issues": 120,
      "error_count": 10,
      "warning_count": 80,
      "info_count": 30,
      "audit_events_30d": 400
    },
    "issues": [
      {
        "severity": "warning",
        "module": "green_book",
        "issue_type": "GB_WITHOUT_DISBURSEMENT_PLAN",
        "record_id": "uuid",
        "record_label": "GB-2026-001 - Project Name",
        "message": "Green Book project belum memiliki disbursement plan",
        "recommended_action": "Lengkapi disbursement plan per tahun"
      }
    ],
    "audit_summary": {
      "by_user": [],
      "by_table": [],
      "recent_activity": []
    }
  }
}
```

## Issue Types

| Issue Type | Severity | Rule |
|---|---|---|
| `BB_WITHOUT_BAPPENAS_PARTNER` | warning | BB Project tanpa mitra kerja Bappenas |
| `BB_INDICATION_WITHOUT_LOI` | info | Ada indication, belum LoI |
| `LOI_WITHOUT_GB` | warning | Ada LoI, belum GB |
| `GB_WITHOUT_BB_REFERENCE` | error | GB Project tanpa BB reference |
| `GB_WITHOUT_FUNDING_SOURCE` | warning | GB Project tanpa funding source |
| `GB_WITHOUT_DISBURSEMENT_PLAN` | warning | GB Project tanpa disbursement plan |
| `GB_WITHOUT_ACTIVITY` | warning | GB Project tanpa activities |
| `DK_WITHOUT_FINANCING_DETAIL` | error | DK Project tanpa financing detail |
| `DK_WITHOUT_ACTIVITY_DETAIL` | warning | DK Project tanpa activity details |
| `DK_WITHOUT_LA` | warning | DK Project belum LA |
| `LA_NOT_EFFECTIVE` | info | LA belum efektif |
| `EFFECTIVE_LA_WITHOUT_MONITORING` | error | LA efektif tanpa monitoring |
| `MONITORING_PLANNED_ZERO_REALIZED_POSITIVE` | error | planned 0 tapi realized > 0 |
| `MONITORING_COMPONENT_NAME_EMPTY` | warning | komponen monitoring kosong |
| `CURRENCY_USD_MISMATCH` | warning | currency USD tapi original != USD |
| `CLOSING_DATE_SOON_NO_RECENT_MONITORING` | error | closing dekat tapi monitoring belum update |

## Audit Rules

- Field audit hanya dikirim untuk ADMIN.
- STAFF boleh melihat data quality issues, tetapi tidak melihat raw audit trail.
- Gunakan `v_audit_recent_activity` jika tersedia.
- Jangan expose `old_data`/`new_data` mentah untuk STAFF.

## Frontend Scope

Files:

- `src/pages/dashboard/DataQualityGovernanceDashboardPage.vue`
- `src/components/dashboard/DataQualityIssueTable.vue`
- `src/components/dashboard/AuditSummaryTable.vue`
- `src/components/dashboard/IssueSeverityBadge.vue`

UI:

- KPI cards: total issue, error, warning, info, audit events.
- Issue table with filters severity/module/issue type.
- Audit tab visible only for ADMIN.
- Click record to open related module detail if route exists.

## Acceptance Criteria

- Data quality issues dihitung backend, bukan hardcoded frontend.
- ADMIN melihat audit summary; STAFF tidak melihat detail audit.
- No raw `password_hash`, `old_data`, or `new_data` exposed to STAFF.
- Issue types konsisten dan bisa difilter.
- Test minimal untuk 5 issue type utama.

## Prompt Codex

```text
Implement PRISM dashboard Phase 7: Data Quality & Governance Dashboard.
Add GET /api/v1/dashboard/data-quality-governance. Compute data quality issues from PRISM tables and business rules. Include audit summary only for ADMIN; never expose raw audit old_data/new_data or password_hash to STAFF.
Add Vue page /dashboard/data-quality-governance with issue table, severity filters, and ADMIN-only audit tab. Use service/store/types and permission helper. Run tests and build.
```
