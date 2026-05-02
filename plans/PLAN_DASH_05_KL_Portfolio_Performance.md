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

# Phase 5 — K/L Portfolio Performance Dashboard

## Objective

Membangun dashboard untuk membandingkan portofolio dan kinerja K/L/Badan sebagai Executing Agency/Implementing Agency pada BB, GB, DK, LA, dan Monitoring.

## Dashboard Questions

1. K/L mana yang memiliki nilai pipeline dan LA terbesar?
2. K/L mana yang paling banyak proyek stuck di GB, DK, atau LA?
3. Bagaimana serapan per K/L?
4. K/L mana yang memiliki data belum lengkap?
5. K/L mana yang memiliki LA closing date dekat atau under-disbursement tinggi?

## Backend Scope

Endpoint:

```http
GET /api/v1/dashboard/kl-portfolio-performance
```

Query params:

| Param | Type |
|---|---|
| `institution_id` | UUID optional |
| `institution_role` | Executing Agency/Implementing Agency optional |
| `period_id` | UUID optional |
| `publish_year` | int optional |
| `budget_year` | int optional |
| `quarter` | enum optional |
| `sort_by` | pipeline_usd/la_commitment_usd/absorption_pct/risk_count |

Response:

```json
{
  "data": {
    "summary": {
      "total_institutions": 32,
      "highest_exposure_institution": "Kementerian ...",
      "lowest_absorption_institution": "Kementerian ..."
    },
    "items": [
      {
        "institution_id": "uuid",
        "institution_name": "Kementerian ...",
        "bb_project_count": 10,
        "gb_project_count": 8,
        "dk_project_count": 5,
        "la_count": 4,
        "pipeline_usd": 1000000000,
        "la_commitment_usd": 500000000,
        "planned_usd": 100000000,
        "realized_usd": 60000000,
        "absorption_pct": 60,
        "risk_count": 3
      }
    ]
  }
}
```

## Backend Rules

- BB/GB institutions memakai table role `bb_project_institution` dan `gb_project_institution`.
- DK executing agency memakai `dk_project.institution_id`.
- LA dan monitoring diturunkan dari DK project.
- Jika satu proyek punya multi EA/IA, gunakan count distinct project per institution, bukan sum global tanpa distinct.
- Untuk nilai proyek multi-institution, tampilkan sebagai exposure per institution, tetapi jangan gunakan total antar institution sebagai total nasional karena akan double count.

## Frontend Scope

Files:

- `src/pages/dashboard/KLPortfolioPerformanceDashboardPage.vue`
- `src/components/dashboard/KLPerformanceTable.vue`
- `src/components/dashboard/KLPerformanceRadar.vue` optional
- `src/components/dashboard/KLRiskBadge.vue`

UI:

- Ranking table K/L.
- KPI cards: total K/L, top exposure, lowest absorption, highest risk.
- Scatter chart: LA commitment vs absorption.
- Filter institution role and budget year.

## Performance Score

Optional score per K/L:

```text
performance_score =
  40% absorption_pct normalized +
  25% pipeline_progress_score +
  20% data_completeness_score +
  15% closing_risk_score_inverse
```

Kategori:

- `Good`: >= 80
- `Watch`: 60–79
- `High Risk`: < 60

## Acceptance Criteria

- K/L ranking muncul dengan project count per stage.
- Absorption per K/L benar dari monitoring LA terkait DK Project.
- Multi-institution project tidak merusak total nasional.
- Tabel dapat sort by amount, absorption, risk.
- Route frontend `/dashboard/kl-portfolio-performance` tersedia.

## Prompt Codex

```text
Implement PRISM dashboard Phase 5: K/L Portfolio Performance Dashboard.
Add GET /api/v1/dashboard/kl-portfolio-performance. Aggregate project count and amounts by institution across BB, GB, DK, LA, and monitoring. Respect institution role and use distinct project counting for many-to-many institution relations. For multi-institution exposure, do not use the institution-level total as national total.
Add Vue page /dashboard/kl-portfolio-performance with ranking table, filters, KPI cards, and scatter/bar chart. Run tests and build.
```
