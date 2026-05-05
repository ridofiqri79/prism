# PLAN DASH 04 - Lender & Financing Mix

## Objective

Menganalisis lender, tingkat kepastian pembiayaan, komposisi loan/grant/local, currency exposure, dan conversion dari indication hingga Loan Agreement.

## Endpoint

```http
GET /api/v1/dashboard/lender-financing-mix
```

## Query Params

| Param | Type |
|---|---|
| `lender_type` | Bilateral/Multilateral/KSA optional |
| `lender_id` | UUID optional |
| `currency` | ISO code optional |
| `period_id` | UUID optional |
| `publish_year` | int optional |

## Lender Certainty Ladder

| Stage | Data Source | Meaning |
|---|---|---|
| `LENDER_INDICATION` | `lender_indication` + BB project cost | Minat awal |
| `LOI` | `loi` + BB project cost | Minat konkret |
| `GB_FUNDING_SOURCE` | `gb_funding_source` | Sumber pendanaan GB |
| `DK_FINANCING` | `dk_financing_detail` | Daftar Kegiatan |
| `LA` | `loan_agreement` | Legal binding |

## Frontend Scope

- `src/pages/dashboard/LenderFinancingMixDashboardPage.vue`
- `src/components/dashboard/LenderCertaintyChart.vue`
- `src/components/dashboard/CurrencyExposureChart.vue`
- `src/components/dashboard/CofinancingNetworkTable.vue`
- `src/components/dashboard/LenderConversionTable.vue`

## Acceptance Criteria

- Conversion rate by lender tersedia.
- Currency exposure tidak melakukan kalkulasi kurs otomatis.
- Lender type mengikuti master lender.
- Cofinancing dihitung per proyek, bukan per baris lender saja.
- Filter tidak memakai budget year.
