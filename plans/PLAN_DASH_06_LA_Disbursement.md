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

# Phase 6 — Loan Agreement & Disbursement Dashboard

## Objective

Membangun dashboard monitoring legal commitment dan serapan: LA signed/effective, closing date, extension, undisbursed balance, planned vs realized, dan early warning under-disbursement.

## Dashboard Questions

1. Berapa total LA signed, efektif, belum efektif, diperpanjang?
2. Berapa nilai komitmen LA, realisasi kumulatif, dan sisa undisbursed?
3. LA mana yang closing dalam 12/6/3 bulan?
4. LA mana yang time elapsed tinggi tetapi serapan rendah?
5. Bagaimana planned vs realized per TA/triwulan?
6. Komponen apa yang paling tertinggal?

## Backend Scope

Endpoint:

```http
GET /api/v1/dashboard/la-disbursement
```

Query params:

| Param | Type |
|---|---|
| `budget_year` | int optional |
| `quarter` | enum optional |
| `lender_id` | UUID optional |
| `institution_id` | UUID optional |
| `is_extended` | bool optional |
| `closing_months` | int optional: 3/6/12 |
| `risk_level` | low/medium/high optional |

Response:

```json
{
  "data": {
    "summary": {
      "la_count": 42,
      "effective_count": 38,
      "not_effective_count": 4,
      "extended_count": 5,
      "commitment_usd": 1500000000,
      "planned_usd": 800000000,
      "realized_usd": 500000000,
      "absorption_pct": 62.5,
      "undisbursed_usd": 1000000000
    },
    "quarterly_trend": [],
    "closing_risks": [],
    "under_disbursement_risks": [],
    "component_breakdown": []
  }
}
```

## Risk Logic

| Risk | Rule |
|---|---|
| `CLOSING_12_MONTHS` | `closing_date <= today + interval '12 months'` |
| `CLOSING_6_MONTHS` | `closing_date <= today + interval '6 months'` |
| `CLOSING_3_MONTHS` | `closing_date <= today + interval '3 months'` |
| `UNDER_DISBURSEMENT_MEDIUM` | `time_elapsed_pct - la_absorption_pct >= 20` |
| `UNDER_DISBURSEMENT_HIGH` | `time_elapsed_pct - la_absorption_pct >= 40` |
| `EFFECTIVE_NO_MONITORING` | `effective_date <= today` and no monitoring |
| `EXTENDED` | `closing_date != original_closing_date` |

## Required Formulas

```text
cumulative_realized_usd = SUM(monitoring_disbursement.realized_usd)
la_absorption_pct = cumulative_realized_usd / loan_agreement.amount_usd * 100
undisbursed_usd = loan_agreement.amount_usd - cumulative_realized_usd
time_elapsed_pct = (CURRENT_DATE - effective_date) / (closing_date - effective_date) * 100
required_monthly_disbursement_usd = undisbursed_usd / remaining_months
```

If denominator <= 0, return 0 or null safely.

## Frontend Scope

Files:

- `src/pages/dashboard/LADisbursementDashboardPage.vue`
- `src/components/dashboard/DisbursementTrendChart.vue`
- `src/components/dashboard/ClosingRiskTable.vue`
- `src/components/dashboard/UnderDisbursementTable.vue`
- `src/components/dashboard/ComponentBreakdownChart.vue`

UI:

- KPI cards: LA count, commitment, realized, absorption, undisbursed, extended.
- Line/bar chart planned vs realized per quarter.
- Table closing risk.
- Table under-disbursement risk.
- Component breakdown chart if `monitoring_komponen` exists.

## Acceptance Criteria

- Monitoring hanya memakai LA efektif dan data monitoring tersimpan.
- Extension memakai computed comparison `closing_date != original_closing_date`.
- Risk items clickable to LA detail.
- `absorption_pct` aman ketika planned/amount = 0.
- Dashboard dapat difilter budget year/quarter/lender/KL.

## Prompt Codex

```text
Implement PRISM dashboard Phase 6: Loan Agreement & Disbursement Dashboard.
Add GET /api/v1/dashboard/la-disbursement. Compute LA commitment, effective/not effective, extended LA, planned vs realized, cumulative realized, undisbursed, closing risk, and under-disbursement risk. Use monitoring_disbursement and monitoring_komponen. Guard all division by zero.
Add Vue page /dashboard/la-disbursement with KPI cards, trend chart, closing risk table, under-disbursement table, and component breakdown. Run tests and build.
```
