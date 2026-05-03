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

# Phase 9 — Testing, Performance, and Release Hardening

## Objective

Memastikan 7 dashboard stabil, cepat, aman, tidak double count, dan siap dipakai sebagai control tower PRISM.

## Backend Test Plan

### Unit/Service Tests

1. Executive summary tidak double count BB/GB revisi.
2. Pipeline stage derivation benar untuk:
   - BB only
   - indication no LoI
   - LoI no GB
   - GB no DK
   - DK no LA
   - LA not effective
   - effective no monitoring
   - monitoring active
3. Green Book readiness score benar.
4. Cofinancing dihitung jika lender > 1 dalam satu GB project.
5. Lender conversion benar dari indication sampai LA.
6. K/L multi-institution tidak merusak distinct count.
7. LA under-disbursement risk aman untuk pembagi 0.
8. Data quality issues muncul sesuai rules.
9. STAFF tidak menerima audit detail.
10. ADMIN menerima audit summary.

### SQL Performance Tests

Gunakan seed minimal besar:

- 500 BB projects
- 300 GB projects
- 200 DK projects
- 150 LA
- 600 monitoring rows
- multiple lenders/institutions/locations

Target awal:

| Endpoint | Target Response Time Lokal |
|---|---:|
| `/dashboard/executive-portfolio` | < 800 ms |
| `/dashboard/pipeline-bottleneck` | < 1000 ms |
| `/dashboard/green-book-readiness` | < 1000 ms |
| `/dashboard/lender-financing-mix` | < 1000 ms |
| `/dashboard/kl-portfolio-performance` | < 1200 ms |
| `/dashboard/la-disbursement` | < 1000 ms |
| `/dashboard/data-quality-governance` | < 1200 ms |

Jika lambat:

1. Jalankan `EXPLAIN ANALYZE`.
2. Tambah index hanya jika query plan membutuhkan.
3. Pertimbangkan view/materialized view hanya setelah bottleneck jelas.
4. Jangan premature optimize.

## Frontend Test Plan

1. Semua route dashboard render tanpa error.
2. Filter mengubah query API, bukan filter lokal dataset besar.
3. Empty state muncul jika data kosong.
4. Error state muncul jika API error.
5. Loading state konsisten.
6. Chart resize responsive.
7. Table pagination server-side berjalan.
8. Role STAFF tidak melihat audit tab.
9. Route click menuju detail/journey valid jika route tersedia.
10. Frontend build lulus.

## Security Checks

- Endpoint dashboard wajib authenticated.
- Data quality audit detail hanya ADMIN.
- Tidak expose `password_hash`.
- Tidak expose raw `old_data`/`new_data` untuk STAFF.
- Tidak ada raw SQL string baru di Go service.
- Tidak ada axios langsung di component.

## Release Checklist

- [ ] `make generate` sudah dijalankan.
- [ ] `go test ./...` lulus.
- [ ] Frontend typecheck/build lulus.
- [ ] Manual smoke test semua dashboard.
- [ ] Tidak ada double count pada sample data revisi BB/GB.
- [ ] Risk rules tervalidasi manual pada sample LA.
- [ ] Permission ADMIN/STAFF dicek.
- [ ] Dokumentasi endpoint dashboard diperbarui di `PRISM_API_Contract.md`.
- [ ] Screenshot dashboard utama disimpan untuk referensi UI.

## Suggested Smoke Flow

1. Login sebagai ADMIN.
2. Buka `/dashboard`.
3. Buka setiap dashboard dari card home.
4. Terapkan filter period/GB year/budget year.
5. Cek angka funnel terhadap data sample.
6. Buka risk item ke detail LA atau journey.
7. Login sebagai STAFF.
8. Pastikan audit tab tidak muncul.
9. Cek network tab: semua call ke `/api/v1/dashboard/*`.
10. Pastikan tidak ada console error.

## Prompt Codex

```text
Implement PRISM dashboard Phase 9: Testing, Performance, and Release Hardening.
Add/complete backend tests for all dashboard endpoints, including no double counting across BB/GB revisions, pipeline stage derivation, readiness score, lender conversion, institution distinct counting, LA risk, and data quality issues. Add frontend checks where available and fix build/type errors.
Run make generate, go test ./..., and frontend typecheck/build. Update PRISM_API_Contract.md with final dashboard endpoint contracts. Do not add unrelated features.
```
