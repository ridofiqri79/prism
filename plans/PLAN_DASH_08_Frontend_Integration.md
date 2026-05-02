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

# Phase 8 — Frontend Dashboard Integration & Navigation

## Objective

Menyatukan seluruh dashboard dalam UI PRISM secara konsisten: sidebar, route, filter global, reusable chart components, loading/error state, dan navigasi ke detail proyek/journey.

## Scope

Frontend only, kecuali endpoint filter options kecil bila belum tersedia.

## Files

- `src/router/routes/dashboard.routes.ts`
- `src/pages/dashboard/DashboardHomePage.vue`
- `src/pages/dashboard/ExecutivePortfolioDashboardPage.vue`
- `src/pages/dashboard/PipelineBottleneckDashboardPage.vue`
- `src/pages/dashboard/GreenBookReadinessDashboardPage.vue`
- `src/pages/dashboard/LenderFinancingMixDashboardPage.vue`
- `src/pages/dashboard/KLPortfolioPerformanceDashboardPage.vue`
- `src/pages/dashboard/LADisbursementDashboardPage.vue`
- `src/pages/dashboard/DataQualityGovernanceDashboardPage.vue`
- `src/components/dashboard/*`
- `src/services/dashboard.service.ts`
- `src/stores/dashboard.store.ts`
- `src/types/dashboard.types.ts`

## Route Plan

| Route | Page |
|---|---|
| `/dashboard` | DashboardHomePage |
| `/dashboard/executive-portfolio` | ExecutivePortfolioDashboardPage |
| `/dashboard/pipeline-bottleneck` | PipelineBottleneckDashboardPage |
| `/dashboard/green-book-readiness` | GreenBookReadinessDashboardPage |
| `/dashboard/lender-financing-mix` | LenderFinancingMixDashboardPage |
| `/dashboard/kl-portfolio-performance` | KLPortfolioPerformanceDashboardPage |
| `/dashboard/la-disbursement` | LADisbursementDashboardPage |
| `/dashboard/data-quality-governance` | DataQualityGovernanceDashboardPage |

## Reusable Components

| Component | Purpose |
|---|---|
| `DashboardFilterBar.vue` | Period, year, quarter, lender, K/L filters |
| `MetricCard.vue` | KPI card standardized |
| `DashboardChartCard.vue` | Wrapper chart with title, subtitle, loading |
| `EmptyInsightState.vue` | Empty/error state |
| `RiskBadge.vue` | Low/medium/high risk |
| `AmountDisplay.vue` | USD/IDR formatting |
| `InsightCallout.vue` | Narasi insight otomatis |

## UI Rules

1. Jangan hitung agregasi besar di frontend.
2. Frontend hanya melakukan formatting, filtering ringan untuk UI state, dan rendering.
3. Semua data fetch melalui `DashboardService`.
4. Semua state dashboard di `dashboard.store.ts` atau store per dashboard jika sudah terlalu besar.
5. Gunakan `useListControls.ts` untuk tabel paginated.
6. Gunakan ECharts untuk chart; jangan pakai chart library lain.
7. Gunakan route guard permission minimal `read` pada modul relevan.

## Dashboard Home Page

Dashboard home menampilkan card navigasi 7 dashboard:

1. Executive Portfolio
2. Pipeline & Bottleneck
3. Green Book Readiness
4. Lender & Financing Mix
5. K/L Portfolio Performance
6. Loan Agreement & Disbursement
7. Data Quality & Governance

Setiap card berisi:

- Nama dashboard
- Pertanyaan yang dijawab
- 2–3 KPI singkat jika tersedia dari `/dashboard/summary`
- Tombol `Buka Dashboard`

## Acceptance Criteria

- Sidebar/menu dashboard tersedia.
- Semua dashboard route bisa dibuka.
- Loading, error, empty state konsisten.
- Filter bar reusable dan tidak duplikasi logic.
- Frontend typecheck/build lulus.

## Prompt Codex

```text
Implement PRISM dashboard Phase 8: Frontend Dashboard Integration.
Do not change backend unless a small filter-options endpoint is missing. Create dashboard routes, dashboard home, reusable dashboard components, shared filter bar, and consistent loading/error/empty states. Integrate all 7 dashboard pages created in prior phases.
Follow PRISM frontend rules: Composition API, service/store/types, PrimeVue, Tailwind v4 rules, ECharts. No axios in components. No local aggregation of large datasets. Run typecheck/build.
```
