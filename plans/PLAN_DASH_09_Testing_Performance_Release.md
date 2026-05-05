# PLAN DASH 09 - Testing, Performance, and Release Hardening

## Objective

Memastikan dashboard analytics stabil, cepat, aman, tidak double count, dan siap dipakai sebagai control tower PRISM.

## Backend Test Plan

1. Executive summary tidak double count revisi Blue Book/Green Book.
2. Pipeline stage derivation benar untuk:
   - `BB_NO_LENDER`
   - `INDICATION_NO_LOI`
   - `LOI_NO_GB`
   - `GB_NO_DK`
   - `DK_NO_LA`
   - `LA_NOT_EFFECTIVE`
3. Green Book readiness score benar.
4. Cofinancing dihitung jika lender lebih dari satu dalam satu proyek.
5. Lender conversion benar dari indication sampai Loan Agreement.
6. K/L multi-institution tidak merusak distinct count.
7. Data quality issues muncul sesuai rules.
8. STAFF tidak menerima audit detail.
9. ADMIN menerima audit summary.

## SQL Performance Targets

| Endpoint | Target Response Time Lokal |
|---|---:|
| `/dashboard/executive-portfolio` | < 800 ms |
| `/dashboard/pipeline-bottleneck` | < 1000 ms |
| `/dashboard/green-book-readiness` | < 1000 ms |
| `/dashboard/lender-financing-mix` | < 1000 ms |
| `/dashboard/kl-portfolio-performance` | < 1200 ms |
| `/dashboard/data-quality-governance` | < 1200 ms |

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

## Release Checklist

- [ ] `sqlc generate` sudah dijalankan.
- [ ] `go test ./...` lulus.
- [ ] Frontend type-check/build lulus.
- [ ] Manual smoke test semua dashboard.
- [ ] Tidak ada double count pada sample data revisi Blue Book/Green Book.
- [ ] Permission ADMIN/STAFF dicek.
- [ ] Dokumentasi endpoint dashboard diperbarui di `PRISM_API_Contract.md`.
