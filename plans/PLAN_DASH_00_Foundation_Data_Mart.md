# PLAN DASH 00 - Foundation Analytics

> Current scope: dashboard analytics read-only untuk pipeline Blue Book -> Green Book -> Daftar Kegiatan -> Loan Agreement.

## Prinsip Umum

1. Dashboard bersifat read-only.
2. Default memakai latest snapshot untuk Blue Book dan Green Book.
3. Downstream Daftar Kegiatan dan Loan Agreement tetap memakai concrete snapshot yang tersimpan.
4. Nilai USD memakai nilai yang sudah tersimpan, tanpa konversi otomatis.
5. Semua query backend ditulis di `prism-backend/sql/queries/dashboard.sql`, lalu jalankan `sqlc generate`.
6. Frontend wajib lewat `dashboard.service.ts`, `dashboard.store.ts`, dan `dashboard.types.ts`.

## Dashboard Aktif

1. Ringkasan Eksekutif
2. Pipeline & Bottleneck
3. Green Book Readiness
4. Lender & Financing Mix
5. K/L Portfolio Performance
6. Data Quality & Governance

## Metrik Dasar

| Metrik | Definisi |
|---|---|
| `bb_pipeline_usd` | SUM `bb_project_cost.amount_usd` untuk `funding_type = 'Foreign'` |
| `gb_pipeline_usd` | SUM `gb_funding_source.loan_usd + gb_funding_source.grant_usd` |
| `gb_local_usd` | SUM `gb_funding_source.local_usd` |
| `dk_financing_usd` | SUM `dk_financing_detail.amount_usd + dk_financing_detail.grant_usd` |
| `dk_counterpart_usd` | SUM `dk_financing_detail.counterpart_usd` |
| `la_commitment_usd` | SUM `loan_agreement.amount_usd` |

## Endpoint Dasar

```http
GET /api/v1/dashboard/summary
GET /api/v1/dashboard/stage-funnel
GET /api/v1/dashboard/filter-options
GET /api/v1/dashboard/executive-portfolio
```

## Acceptance Criteria

- Query tidak double count akibat revisi Blue Book/Green Book.
- Filter umum tidak memakai quarter atau budget year.
- Response memakai format `{ "data": ... }`.
- `sqlc generate`, backend tests, dan frontend type-check lulus.
