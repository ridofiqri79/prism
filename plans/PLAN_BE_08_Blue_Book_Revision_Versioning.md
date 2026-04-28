# PLAN BE-08 - Blue Book Revision Versioning

> Scope: implementasi logical identity, clone revisi, history endpoint, dan import behavior untuk Blue Book.
> Deliverable: BB Project bisa memakai `bb_code` yang sama pada revisi lain, revisi BB bisa clone snapshot, dan history BB Project tersedia.
> Referensi: `plans/PLAN_BE_07_Revision_Versioning_Schema.md`, `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`, `docs/PRISM_Business_Rules.md`, `docs/PRISM_API_Contract.md`.

---

## Prasyarat

- [ ] `PLAN_BE_07_Revision_Versioning_Schema.md` selesai.
- [ ] `make generate` sudah berhasil setelah perubahan SQL.
- [ ] Query identity/latest/history BB tersedia di generated sqlc package.

---

## Step 1 - Model dan Response Contract

Files:

- `prism-backend/internal/model/blue_book.go`

Checklist:

- [ ] Tambahkan `ProjectIdentityID` pada BB Project request/response bila diperlukan.
- [ ] Tambahkan field response:
  - [ ] `project_identity_id`.
  - [ ] `is_latest`.
  - [ ] `has_newer_revision`.
- [ ] Tambahkan model history item untuk `GET /bb-projects/:id/history`.
- [ ] Pastikan semua tipe strongly typed, tidak memakai `any`/`interface{}` untuk payload DB.
- [ ] Pastikan JSON field mengikuti contract snake_case.

Acceptance:

- [ ] Model dapat merepresentasikan snapshot dan logical identity.

---

## Step 2 - Create dan Update BB Project

Files:

- `prism-backend/internal/service/blue_book_service.go`

Checklist:

- [ ] Ganti validasi duplicate dari global code menjadi per `blue_book_id`.
- [ ] Saat create tanpa `project_identity_id`, buat `project_identity` baru dalam transaksi.
- [ ] Saat create dengan `project_identity_id`, validasi identity ada.
- [ ] Simpan `project_identity_id` ke `bb_project`.
- [ ] Build response mengisi latest flags:
  - [ ] `is_latest`.
  - [ ] `has_newer_revision`.
- [ ] Update BB Project tidak mengubah `bb_code` kecuali memang contract memperbolehkan; jika diperbolehkan, validasi tetap per Blue Book.
- [ ] Pastikan child rows tetap transaksional: institutions, locations, priorities, costs, lender indications.

Acceptance:

- [ ] `bb_code` sama bisa dibuat di revisi Blue Book berbeda.
- [ ] Duplicate `bb_code` dalam Blue Book yang sama tetap conflict.

---

## Step 3 - Create Blue Book Revision dan Clone Snapshot

Files:

- `prism-backend/internal/service/blue_book_service.go`

Checklist:

- [ ] Tentukan source revisi:
  - [ ] Jika request membawa `replaces_blue_book_id`, gunakan itu.
  - [ ] Jika tidak, gunakan Blue Book active terakhir pada period yang sama sebelum supersede.
- [ ] Dalam satu transaksi:
  - [ ] Supersede Blue Book lama.
  - [ ] Create Blue Book baru.
  - [ ] Clone semua active BB Project dari source.
  - [ ] Reuse `project_identity_id` dari source project.
  - [ ] Clone institutions.
  - [ ] Clone locations.
  - [ ] Clone national priorities.
  - [ ] Clone project costs.
  - [ ] Clone lender indications.
  - [ ] Clone LoI sesuai keputusan plan default.
- [ ] Pastikan clone tidak memakai lookup global `bb_code`.
- [ ] Publish SSE setelah commit berhasil.

Acceptance:

- [ ] Revisi BB baru berisi snapshot project yang sama dengan identity yang sama.
- [ ] BB lama menjadi `superseded`.

---

## Step 4 - History Endpoint dan Route

Files:

- `prism-backend/internal/handler/blue_book_handler.go`
- `prism-backend/cmd/api/main.go`

Checklist:

- [ ] Implement service method `GetBBProjectHistory`.
- [ ] Implement handler untuk `GET /bb-projects/:id/history`.
- [ ] Daftarkan route dengan permission `bb_project:read`.
- [ ] Response history memuat:
  - [ ] snapshot id.
  - [ ] `project_identity_id`.
  - [ ] Blue Book id.
  - [ ] book label atau period + revision.
  - [ ] `revision_number`.
  - [ ] `revision_year`.
  - [ ] book status.
  - [ ] `is_latest`.
  - [ ] `used_by_downstream` jika query tersedia.
- [ ] Return 404 aman jika snapshot tidak ditemukan.

Acceptance:

- [ ] User bisa melihat semua revisi BB yang memuat logical project yang sama.

---

## Step 5 - Blue Book Import Behavior

Files:

- `prism-backend/internal/service/blue_book_import_service.go`

Checklist:

- [ ] Duplicate `BB Code` dalam workbook tetap error.
- [ ] `BB Code` yang sudah ada di revisi lama tidak di-skip otomatis.
- [ ] Jika import ke Blue Book revisi dan kode ditemukan pada revisi/source sebelumnya, reuse `project_identity_id`.
- [ ] Jika kode tidak ditemukan pada source identity manapun, create identity baru.
- [ ] Relasi workbook tetap memakai `BB Code` hanya dalam konteks workbook target.
- [ ] Import summary membedakan:
  - [ ] created new logical project.
  - [ ] created revision snapshot.
  - [ ] skipped due to duplicate within workbook/target document.

Acceptance:

- [ ] Import ke revisi bisa membuat snapshot dengan kode yang sama tanpa kehilangan history.

---

## Step 6 - Tests dan Verification

Checklist:

- [ ] Tambah/update backend tests untuk create BB Project duplicate per Blue Book.
- [ ] Tambah/update tests untuk create BB revision clone.
- [ ] Tambah/update tests untuk BB Project history endpoint.
- [ ] Tambah/update tests untuk import code reuse behavior.
- [ ] Jalankan `go test ./...`.
- [ ] Jalankan smoke API:
  - [ ] create BB original.
  - [ ] create BB Project `BB-001`.
  - [ ] create BB revision.
  - [ ] verify revisi memuat `BB-001` dengan `project_identity_id` sama.
  - [ ] call `/bb-projects/:id/history`.

Done Criteria:

- [ ] Blue Book revision versioning bekerja end-to-end.
- [ ] Tidak ada global `bb_code` blocker tersisa di service/import.
