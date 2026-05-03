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

# Phase 3 — Green Book Readiness Dashboard

## Objective

Membangun dashboard untuk menilai kualitas dan kesiapan proyek Green Book: struktur pendanaan, disbursement plan, activities, funding allocation, cofinancing, dan kelengkapan data.

## Dashboard Questions

1. Berapa total proyek GB per tahun/revisi?
2. Berapa nilai loan, grant, local/counterpart di GB?
3. Apakah setiap proyek GB memiliki funding source, activities, disbursement plan, dan funding allocation?
4. Bagaimana profil alokasi belanja: services, construction, goods, training, other?
5. Proyek mana yang memiliki cofinancing dan perlu koordinasi lebih kompleks?
6. Apakah rencana disbursement tahunan realistis dan lengkap?

## Backend Scope

Endpoint:

```http
GET /api/v1/dashboard/green-book-readiness
```

Query params:

| Param | Type |
|---|---|
| `publish_year` | int optional |
| `green_book_id` | UUID optional |
| `institution_id` | UUID optional |
| `lender_id` | UUID optional |
| `readiness_status` | enum optional: COMPLETE, INCOMPLETE, COFINANCING |

Response:

```json
{
  "data": {
    "summary": {
      "total_projects": 80,
      "total_loan_usd": 3000000000,
      "total_grant_usd": 200000000,
      "total_local_usd": 500000000,
      "projects_with_cofinancing": 12,
      "projects_incomplete": 9
    },
    "disbursement_plan_by_year": [
      { "year": 2026, "amount_usd": 500000000 }
    ],
    "funding_allocation": {
      "services": 10000000,
      "constructions": 200000000,
      "goods": 50000000,
      "trainings": 1000000,
      "other": 500000
    },
    "readiness_items": []
  }
}
```

## Readiness Scoring

Codex harus membuat scoring sederhana:

| Check | Score |
|---|---:|
| Has BB reference | 20 |
| Has EA/IA | 15 |
| Has location | 10 |
| Has funding source | 20 |
| Has activities | 15 |
| Has disbursement plan | 10 |
| Has funding allocation | 10 |

Kategori:

- `READY`: score >= 85
- `PARTIAL`: 60 <= score < 85
- `INCOMPLETE`: score < 60

## Frontend Scope

Files:

- `src/pages/dashboard/GreenBookReadinessDashboardPage.vue`
- `src/components/dashboard/ReadinessScoreCard.vue`
- `src/components/dashboard/DisbursementPlanChart.vue`
- `src/components/dashboard/FundingAllocationChart.vue`
- `src/components/dashboard/ReadinessWorklistTable.vue`

UI:

- KPI cards: total GB projects, loan, grant, local, cofinancing, incomplete.
- Bar chart disbursement plan by year.
- Donut/bar chart funding allocation by category.
- Readiness table with score and missing fields.

## Acceptance Criteria

- Dashboard default memakai GB active/latest per publish year.
- Disbursement plan dihitung sebagai total proyek per tahun, bukan per lender.
- Funding allocation mengikuti relasi `gb_funding_allocation -> gb_activity`.
- Cofinancing dihitung jika satu GB Project memiliki > 1 lender di funding source.
- Readiness table menampilkan missing fields.

## Prompt Codex

```text
Implement PRISM dashboard Phase 3: Green Book Readiness Dashboard.
Add GET /api/v1/dashboard/green-book-readiness. Compute GB readiness score from BB reference, EA/IA, location, funding source, activities, disbursement plan, and funding allocation. Disbursement Plan is total project per year, not per lender. Cofinancing means more than one lender in gb_funding_source for one GB project.
Add Vue page /dashboard/green-book-readiness using ECharts and PrimeVue tables. Use service/store/types only. Run tests and build.
```
