# PLAN DASH 05 - K/L Portfolio Performance

## Objective

Membandingkan portofolio dan kinerja K/L/Badan sebagai Executing Agency atau Implementing Agency pada Blue Book, Green Book, Daftar Kegiatan, dan Loan Agreement.

## Endpoint

```http
GET /api/v1/dashboard/kl-portfolio-performance
```

## Query Params

| Param | Type |
|---|---|
| `institution_id` | UUID optional |
| `institution_role` | Executing Agency/Implementing Agency optional |
| `period_id` | UUID optional |
| `publish_year` | int optional |
| `sort_by` | `pipeline_usd`, `la_commitment_usd`, atau `risk_count` |

## Response Fokus

- Project count per stage: BB, GB, DK, LA.
- `pipeline_usd`.
- `la_commitment_usd`.
- `risk_count`.
- `performance_score`.
- `performance_category`.

## Backend Rules

- BB/GB institutions memakai table role `bb_project_institution` dan `gb_project_institution`.
- DK executing agency memakai `dk_project.institution_id`.
- Loan Agreement diturunkan dari DK project.
- Jika satu proyek punya multi EA/IA, gunakan count distinct project per institution.
- Untuk nilai proyek multi-institution, tampilkan sebagai exposure per institution, tetapi jangan gunakan total antar institution sebagai total nasional.

## Frontend Scope

- `src/pages/dashboard/KLPortfolioPerformanceDashboardPage.vue`
- `src/components/dashboard/KLPerformanceTable.vue`
- `src/components/dashboard/KLPerformanceRadar.vue`
- `src/components/dashboard/KLRiskBadge.vue`

## Performance Score

```text
performance_score =
  45% pipeline_progress_score +
  35% data_completeness_score +
  20% risk_score_inverse
```

Kategori:

- `Good`: >= 80
- `Watch`: 60-79
- `High Risk`: < 60

## Acceptance Criteria

- K/L ranking muncul dengan project count per stage.
- Multi-institution project tidak merusak total nasional.
- Tabel dapat sort by pipeline, LA commitment, dan risk.
- Route frontend `/dashboard/kl-portfolio-performance` tersedia.
- Tidak ada filter quarter atau budget year.
