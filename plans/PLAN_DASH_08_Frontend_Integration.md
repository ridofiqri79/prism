# PLAN DASH 08 - Frontend Dashboard Integration

## Objective

Menyatukan dashboard analytics dalam UI PRISM: sidebar, route, filter, komponen chart/tabel, loading/error state, dan navigasi ke detail proyek/journey.

## Files

- `src/router/routes/dashboard.routes.ts`
- `src/pages/dashboard/DashboardHomePage.vue`
- `src/pages/dashboard/ExecutivePortfolioDashboardPage.vue`
- `src/pages/dashboard/PipelineBottleneckDashboardPage.vue`
- `src/pages/dashboard/GreenBookReadinessDashboardPage.vue`
- `src/pages/dashboard/LenderFinancingMixDashboardPage.vue`
- `src/pages/dashboard/KLPortfolioPerformanceDashboardPage.vue`
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
| `/dashboard/data-quality-governance` | DataQualityGovernanceDashboardPage |

## Reusable Components

| Component | Purpose |
|---|---|
| `DashboardFilterBar.vue` | Period, year, lender, K/L filters |
| `MetricCard.vue` | KPI card standardized |
| `RiskBadge.vue` | Low/medium/high risk |
| `AmountDisplay.vue` | USD formatting |
| `InsightCallout.vue` | Narasi insight otomatis |

## Acceptance Criteria

- Semua dashboard route aktif bisa dibuka dari satu halaman Dashboard.
- Loading, error, dan empty state konsisten.
- Filter bar tidak menampilkan quarter atau budget year.
- Semua data fetch melalui `DashboardService`.
- Frontend type-check/build lulus.
