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

# Phase 1 — Executive Portfolio Dashboard

## Objective

Membangun dashboard pimpinan untuk melihat posisi portofolio nasional: total pipeline, total komitmen legal, serapan, bottleneck utama, top K/L, top lender, dan proyek berisiko.

## Dashboard Questions

1. Berapa total proyek dan nilai pada setiap tahap BB, GB, DK, LA, dan Monitoring?
2. Berapa nilai yang masih pipeline dibanding yang sudah legal binding?
3. Tahap mana yang menjadi bottleneck terbesar?
4. K/L dan lender mana yang memiliki eksposur terbesar?
5. Proyek mana yang butuh perhatian pimpinan?

## Backend Scope

Endpoint:

```http
GET /api/v1/dashboard/executive-portfolio
```

Query params:

| Param | Type | Keterangan |
|---|---|---|
| `period_id` | UUID optional | Filter periode BB |
| `publish_year` | int optional | Filter tahun GB |
| `budget_year` | int optional | Filter TA monitoring |
| `quarter` | TW1/TW2/TW3/TW4 optional | Filter triwulan |
| `include_history` | bool default false | Hitung snapshot historis jika true |

Response shape:

```json
{
  "data": {
    "cards": [
      { "key": "bb_projects", "label": "Blue Book Projects", "value": 120, "unit": "project" },
      { "key": "la_commitment_usd", "label": "LA Commitment", "value": 1500000000, "unit": "USD" }
    ],
    "funnel": [
      { "stage": "BB", "project_count": 120, "amount_usd": 5000000000 },
      { "stage": "GB", "project_count": 80, "amount_usd": 3500000000 }
    ],
    "top_institutions": [],
    "top_lenders": [],
    "risk_items": []
  }
}
```

## Frontend Scope

Files:

- `src/types/dashboard.types.ts`
- `src/services/dashboard.service.ts`
- `src/stores/dashboard.store.ts`
- `src/pages/dashboard/ExecutivePortfolioDashboardPage.vue`
- `src/components/dashboard/MetricCard.vue`
- `src/components/dashboard/StageFunnelChart.vue`
- `src/components/dashboard/TopBreakdownTable.vue`
- `src/components/dashboard/RiskItemTable.vue`

Layout:

1. Header: `Executive Portfolio`
2. Filter bar: period, GB year, budget year, quarter
3. KPI cards: BB, GB, DK, LA, realized, absorption
4. Funnel chart BB → GB → DK → LA → Monitoring
5. Top 10 K/L by LA commitment / pipeline
6. Top 10 lenders by LA commitment / pipeline
7. Risk table:
   - LA closing <= 12 months
   - LA effective no monitoring
   - LA time elapsed high but absorption low
   - GB without DK
   - DK without LA

## Insight Rules

Generate insight strings in backend or frontend helper:

- `Bottleneck terbesar berada pada tahap {stage}, dengan {count} proyek senilai USD {amount}.`
- `Serapan kumulatif mencapai {pct}% dari rencana monitoring.`
- `{n} Loan Agreement akan closing dalam 12 bulan.`

## Acceptance Criteria

- Dashboard bisa dibuka di route `/dashboard/executive-portfolio`.
- Semua angka berasal dari endpoint dashboard, bukan hitung manual di komponen.
- Funnel tidak double count akibat relasi many-to-many.
- Risk table bisa diklik menuju detail proyek/journey bila route tersedia.
- Frontend build/typecheck lulus.

## Prompt Codex

```text
Implement PRISM dashboard Phase 1: Executive Portfolio Dashboard.
Use the dashboard foundation from Phase 0. Add GET /api/v1/dashboard/executive-portfolio and the Vue page /dashboard/executive-portfolio.
Follow PRISM frontend rules: service -> store -> page, types in src/types, no axios in components, Composition API only, PrimeVue components, ECharts for charts.
The dashboard must show KPI cards, funnel BB→GB→DK→LA→Monitoring, top institutions, top lenders, and risk items. Do not mutate data. Preserve latest snapshot default to avoid revision double counting.
Run backend tests and frontend typecheck/build.
```
