# PLAN 09 - Dashboard Analytics & Project Journey

> **Scope:** Dashboard analytics pipeline BB -> GB -> DK -> LA dan visualisasi perjalanan proyek sampai Loan Agreement.
> **Deliverable:** Satu halaman `Dashboard` dengan tab analytics yang operasional, plus Project Journey yang mengikuti snapshot downstream.
> **Referensi:** `docs/PRISM_API_Contract.md` bagian Dashboard & Aggregasi.

---

## Prinsip Revisi

- Dashboard analytics hanya membaca data Blue Book, Green Book, Daftar Kegiatan, dan Loan Agreement.
- Filter analytics utama: `period_id`, `publish_year`, `lender_id`, `institution_id`, `include_history`.
- Tidak ada filter quarter atau budget year di analytics dashboard.
- Green Book disbursement plan tetap boleh tampil pada Green Book Readiness karena itu rencana tahunan Green Book.
- Semua API call lewat `DashboardService`, state lewat `dashboard.store.ts`, dan tipe lewat `dashboard.types.ts`.

---

## Tab Dashboard

1. Ringkasan Eksekutif
   - KPI pipeline dan commitment.
   - Funnel BB -> GB -> DK -> LA.
   - Top K/L, top lender, risk item pipeline, dan insight otomatis.

2. Pipeline & Bottleneck
   - Worklist bottleneck: `BB_NO_LENDER`, `INDICATION_NO_LOI`, `LOI_NO_GB`, `GB_NO_DK`, `DK_NO_LA`, `LA_NOT_EFFECTIVE`.
   - Filter: search, periode Blue Book, tahun Green Book, K/L, lender, min age.

3. Green Book Readiness
   - Kelengkapan funding source, activities, rencana pencairan Green Book, cofinancing, dan funding allocation.

4. Lender & Financing Mix
   - Certainty ladder lender indication -> Letter of Intent -> Green Book funding source -> Daftar Kegiatan financing -> Loan Agreement.
   - Filter: lender type, lender, currency, periode Blue Book, tahun Green Book.

5. K/L Portfolio Performance
   - Ranking K/L berdasarkan pipeline, LA commitment, risk count, dan performance score.
   - Sort valid: `pipeline_usd`, `la_commitment_usd`, `risk_count`.

6. Data Quality & Governance
   - Issue kelengkapan data untuk BB/GB/DK/LA dan audit summary untuk ADMIN.

---

## Project Journey

- Route menerima `bbProjectId`.
- Timeline menampilkan jalur konkret snapshot yang dipakai downstream:
  - Blue Book project
  - Letter of Intent
  - Green Book project
  - Daftar Kegiatan project
  - Loan Agreement
- Node downstream tidak otomatis pindah ke revisi baru; tampilkan badge jika ada revisi BB/GB lebih baru.

---

## Checklist

- [x] `dashboard.types.ts` sesuai kontrak analytics tanpa quarter/budget year.
- [x] `dashboard.service.ts` memuat summary, stage funnel, filter options, dan tab analytics.
- [x] `dashboard.store.ts` mengelola loading/error per tab.
- [x] Dashboard tab ringkasan, pipeline, readiness, lender mix, K/L performance, dan data quality.
- [x] Project Journey menampilkan snapshot path BB -> GB -> DK -> LA.
