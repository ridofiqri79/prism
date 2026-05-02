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

# Phase 2 — Pipeline & Bottleneck Dashboard

## Objective

Membangun dashboard kerja untuk membaca pergerakan proyek antar tahapan dan mengidentifikasi proyek yang stuck.

## Dashboard Questions

1. Proyek mana yang masih BB tanpa lender indication?
2. Proyek mana yang punya lender indication tetapi belum LoI?
3. Proyek mana yang sudah LoI tetapi belum masuk GB?
4. Proyek mana yang sudah GB tetapi belum DK?
5. Proyek mana yang sudah DK tetapi belum LA?
6. LA mana yang signed tetapi belum efektif?
7. LA efektif mana yang belum punya monitoring?

## Backend Scope

Endpoint:

```http
GET /api/v1/dashboard/pipeline-bottleneck
```

Query params:

| Param | Type |
|---|---|
| `stage` | enum optional |
| `period_id` | UUID optional |
| `publish_year` | int optional |
| `institution_id` | UUID optional |
| `lender_id` | UUID optional |
| `min_age_days` | int optional |
| `page`, `limit`, `sort`, `order`, `search` | standard pagination |

Response:

```json
{
  "data": {
    "stage_summary": [
      { "stage": "BB_WITH_LOI", "project_count": 12, "amount_usd": 1000000000, "avg_age_days": 180 }
    ],
    "items": [
      {
        "project_id": "uuid",
        "project_name": "...",
        "current_stage": "GB_NO_DK",
        "age_days": 240,
        "amount_usd": 100000000,
        "institution_name": "Kementerian ...",
        "lender_names": ["ADB"],
        "recommended_action": "Koordinasi usulan Daftar Kegiatan"
      }
    ]
  },
  "meta": { "page": 1, "limit": 20, "total": 100, "total_pages": 5 }
}
```

## Bottleneck Stage Rules

| Stage | Rule | Recommended Action |
|---|---|---|
| `BB_NO_LENDER` | BB Project tanpa lender indication dan LoI | Market sounding / cari lender indication |
| `INDICATION_NO_LOI` | Ada indication, belum LoI | Follow up lender untuk LoI |
| `LOI_NO_GB` | Ada LoI, belum linked ke GB | Cek readiness dan usulkan GB |
| `GB_NO_DK` | GB Project belum linked ke DK | Dorong usulan/penetapan DK |
| `DK_NO_LA` | DK Project belum punya LA | Dorong negosiasi/legal agreement |
| `LA_NOT_EFFECTIVE` | LA effective_date > today | Monitor effectiveness condition |
| `EFFECTIVE_NO_MONITORING` | LA efektif tanpa monitoring | Input monitoring triwulanan |

## Frontend Scope

Files:

- `src/pages/dashboard/PipelineBottleneckDashboardPage.vue`
- `src/components/dashboard/PipelineStageTabs.vue`
- `src/components/dashboard/BottleneckWorklistTable.vue`
- Update `dashboard.service.ts`, `dashboard.store.ts`, and `dashboard.types.ts`

UI:

- Stage tabs with count and amount.
- Worklist table with search, stage filter, K/L filter, lender filter, min age.
- Row action: open project journey/detail.
- Risk badge by age:
  - Low: < 90 days
  - Medium: 90–180 days
  - High: > 180 days

## Acceptance Criteria

- User dapat filter stage dan melihat daftar proyek stuck.
- Age dihitung dari tanggal relevan jika tersedia; fallback ke `created_at` snapshot.
- Recommended action muncul untuk setiap item.
- List paginated server-side.
- Tidak ada filter lokal untuk dataset besar.

## Prompt Codex

```text
Implement PRISM dashboard Phase 2: Pipeline & Bottleneck Dashboard.
Add backend endpoint GET /api/v1/dashboard/pipeline-bottleneck with stage summary and paginated worklist. Derive current stage from BB→LoI→GB→DK→LA→Monitoring relations, not from a stored status field.
Add frontend route/page /dashboard/pipeline-bottleneck using service/store/types. Use server-side pagination and filters. Show stage tabs and worklist table with recommended action. Do not double count many-to-many relations.
Run tests and frontend typecheck/build.
```
