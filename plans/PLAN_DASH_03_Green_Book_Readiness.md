# PLAN DASH 03 - Green Book Readiness

## Objective

Menilai kualitas dan kesiapan proyek Green Book: struktur pendanaan, rencana pencairan Green Book, activities, funding allocation, cofinancing, dan kelengkapan data.

## Endpoint

```http
GET /api/v1/dashboard/green-book-readiness
```

## Query Params

| Param | Type |
|---|---|
| `publish_year` | int optional |
| `green_book_id` | UUID optional |
| `institution_id` | UUID optional |
| `lender_id` | UUID optional |
| `readiness_status` | enum optional: READY, PARTIAL, INCOMPLETE, COFINANCING |

## Readiness Scoring

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

- `src/pages/dashboard/GreenBookReadinessDashboardPage.vue`
- `src/components/dashboard/DisbursementPlanChart.vue`
- `src/components/dashboard/FundingAllocationChart.vue`
- `src/components/dashboard/ReadinessWorklistTable.vue`

## Acceptance Criteria

- Dashboard default memakai Green Book active/latest per publish year.
- Disbursement plan dihitung sebagai total proyek per tahun, bukan per lender.
- Funding allocation mengikuti relasi `gb_funding_allocation -> gb_activity`.
- Cofinancing dihitung jika satu Green Book Project memiliki lebih dari satu lender di funding source.
- Readiness table menampilkan missing fields.
