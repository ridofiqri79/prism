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

# Phase 4 — Lender & Financing Mix Dashboard

## Objective

Membangun dashboard untuk menganalisis lender, tingkat kepastian pembiayaan, komposisi loan/grant/local, currency exposure, dan conversion dari indication hingga LA.

## Dashboard Questions

1. Lender mana yang paling besar di tahap BB indication, LoI, GB, DK, dan LA?
2. Berapa conversion rate dari lender indication ke LA?
3. Bagaimana komposisi lender berdasarkan type: Bilateral, Multilateral, KSA?
4. Bagaimana exposure berdasarkan currency?
5. Proyek mana yang cofinancing dan lender mana saja yang terlibat?

## Backend Scope

Endpoint:

```http
GET /api/v1/dashboard/lender-financing-mix
```

Query params:

| Param | Type |
|---|---|
| `lender_type` | Bilateral/Multilateral/KSA optional |
| `lender_id` | UUID optional |
| `currency` | ISO code optional |
| `period_id` | UUID optional |
| `publish_year` | int optional |
| `budget_year` | int optional |

Response:

```json
{
  "data": {
    "summary": {
      "total_lenders": 20,
      "bilateral_usd": 1000000000,
      "multilateral_usd": 3000000000,
      "ksa_usd": 500000000,
      "cofinancing_projects": 12
    },
    "certainty_ladder": [
      { "stage": "LENDER_INDICATION", "lender_id": "uuid", "lender_name": "ADB", "project_count": 20, "amount_usd": 2000000000 },
      { "stage": "LA", "lender_id": "uuid", "lender_name": "ADB", "project_count": 8, "amount_usd": 1000000000 }
    ],
    "currency_exposure": [
      { "currency": "USD", "amount_original": 1000000000, "amount_usd": 1000000000 }
    ],
    "cofinancing_items": []
  }
}
```

## Lender Certainty Ladder

| Stage | Data Source | Meaning |
|---|---|---|
| `LENDER_INDICATION` | `lender_indication` + BB project cost | Minat awal |
| `LOI` | `loi` + BB project cost | Minat konkret |
| `GB_FUNDING_SOURCE` | `gb_funding_source` | Sumber pendanaan GB |
| `DK_FINANCING` | `dk_financing_detail` | Daftar Kegiatan |
| `LA` | `loan_agreement` | Legal binding |

## Important Rules

- `lender_id` di LA harus berasal dari DK financing detail.
- Funding source DK hanya dari lender indication BB atau funding source GB terkait.
- Currency disimpan manual; jangan auto-convert.
- Jika currency USD, nilai original dan USD sama.

## Frontend Scope

Files:

- `src/pages/dashboard/LenderFinancingMixDashboardPage.vue`
- `src/components/dashboard/LenderCertaintyChart.vue`
- `src/components/dashboard/CurrencyExposureChart.vue`
- `src/components/dashboard/CofinancingNetworkTable.vue`

UI:

- KPI cards by lender type.
- Stacked bar certainty ladder by lender.
- Currency exposure chart.
- Cofinancing project table.
- Lender conversion table:
  - indication count
  - LoI count
  - GB count
  - DK count
  - LA count
  - LA conversion %

## Acceptance Criteria

- Conversion rate by lender tersedia.
- Currency exposure tidak melakukan kalkulasi kurs otomatis.
- Lender type mengikuti master lender.
- Cofinancing dihitung per proyek, bukan per baris lender saja.
- Frontend support filter lender type/lender/currency.

## Prompt Codex

```text
Implement PRISM dashboard Phase 4: Lender & Financing Mix Dashboard.
Add GET /api/v1/dashboard/lender-financing-mix. Compute lender certainty ladder from lender indication, LoI, GB funding source, DK financing detail, and LA. Compute lender conversion and currency exposure without automatic FX conversion. Respect lender type rules and concrete downstream relations.
Add Vue page /dashboard/lender-financing-mix with charts and cofinancing table. Use dashboard service/store/types. Run tests and build.
```
