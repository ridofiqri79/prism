# PLAN DASH 07 - Data Quality & Governance

## Objective

Memastikan PRISM menjadi single source of truth: data lengkap, konsisten, tidak melanggar aturan bisnis, dan perubahan penting dapat diaudit oleh ADMIN.

## Endpoint

```http
GET /api/v1/dashboard/data-quality-governance
```

## Query Params

| Param | Type |
|---|---|
| `severity` | info/warning/error optional |
| `module` | string optional |
| `issue_type` | string optional |
| `only_unresolved` | bool optional |
| `audit_days` | int default 30 |

## Issue Types

| Issue Type | Severity | Rule |
|---|---|---|
| `BB_WITHOUT_BAPPENAS_PARTNER` | warning | BB Project tanpa Mitra Kerja Bappenas |
| `BB_INDICATION_WITHOUT_LOI` | info | Ada indication, belum Letter of Intent |
| `LOI_WITHOUT_GB` | warning | Ada Letter of Intent, belum Green Book |
| `GB_WITHOUT_BB_REFERENCE` | error | Green Book Project tanpa Blue Book reference |
| `GB_WITHOUT_FUNDING_SOURCE` | warning | Green Book Project tanpa funding source |
| `GB_WITHOUT_DISBURSEMENT_PLAN` | warning | Green Book Project tanpa disbursement plan |
| `GB_WITHOUT_ACTIVITY` | warning | Green Book Project tanpa activities |
| `DK_WITHOUT_FINANCING_DETAIL` | error | Daftar Kegiatan Project tanpa financing detail |
| `DK_WITHOUT_ACTIVITY_DETAIL` | warning | Daftar Kegiatan Project tanpa activity details |
| `DK_WITHOUT_LA` | warning | Daftar Kegiatan Project belum memiliki Loan Agreement |
| `LA_NOT_EFFECTIVE` | info | Loan Agreement belum efektif |
| `CURRENCY_USD_MISMATCH` | warning | Currency USD tapi original tidak sama dengan USD |

## Audit Rules

- Field audit hanya dikirim untuk ADMIN.
- STAFF boleh melihat data quality issues, tetapi tidak melihat raw audit trail.
- Jangan expose `old_data`/`new_data` mentah untuk STAFF.

## Frontend Scope

- `src/pages/dashboard/DataQualityGovernanceDashboardPage.vue`
- `src/components/dashboard/DataQualityIssueTable.vue`
- `src/components/dashboard/AuditSummaryTable.vue`
- `src/components/dashboard/IssueSeverityBadge.vue`

## Acceptance Criteria

- Data quality issues dihitung backend, bukan hardcoded frontend.
- ADMIN melihat audit summary; STAFF tidak melihat detail audit.
- Tidak ada raw `password_hash`, `old_data`, atau `new_data` untuk STAFF.
- Issue types konsisten dan bisa difilter.
