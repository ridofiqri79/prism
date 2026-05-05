# PLAN DASH 02 - Pipeline & Bottleneck

## Objective

Membangun dashboard kerja untuk membaca pergerakan proyek antar tahapan dan mengidentifikasi proyek yang tertahan sampai Loan Agreement.

## Endpoint

```http
GET /api/v1/dashboard/pipeline-bottleneck
```

## Query Params

| Param | Type |
|---|---|
| `stage` | enum optional |
| `period_id` | UUID optional |
| `publish_year` | int optional |
| `institution_id` | UUID optional |
| `lender_id` | UUID optional |
| `min_age_days` | int optional |
| `page`, `limit`, `sort`, `order`, `search` | standard pagination |

## Bottleneck Stage Rules

| Stage | Rule | Recommended Action |
|---|---|---|
| `BB_NO_LENDER` | Blue Book Project tanpa lender indication dan Letter of Intent | Market sounding / cari lender indication |
| `INDICATION_NO_LOI` | Ada indication, belum Letter of Intent | Follow up lender untuk Letter of Intent |
| `LOI_NO_GB` | Ada Letter of Intent, belum linked ke Green Book | Cek readiness dan usulkan Green Book |
| `GB_NO_DK` | Green Book Project belum linked ke Daftar Kegiatan | Dorong usulan/penetapan Daftar Kegiatan |
| `DK_NO_LA` | Daftar Kegiatan Project belum punya Loan Agreement | Dorong negosiasi/legal agreement |
| `LA_NOT_EFFECTIVE` | Loan Agreement `effective_date > today` | Pantau pemenuhan effectiveness condition |

## Frontend Scope

- `src/pages/dashboard/PipelineBottleneckDashboardPage.vue`
- `src/components/dashboard/PipelineStageTabs.vue`
- `src/components/dashboard/BottleneckWorklistTable.vue`
- `dashboard.service.ts`, `dashboard.store.ts`, dan `dashboard.types.ts`

## Acceptance Criteria

- User dapat filter stage dan melihat daftar proyek tertahan.
- Age dihitung dari tanggal relevan jika tersedia, fallback ke `created_at`.
- Recommended action muncul untuk setiap item.
- List paginated server-side.
- Tidak ada filter quarter atau budget year.
