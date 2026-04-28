# PLAN BE-08 - Blue Book Revision Versioning

> Scope: implementasi logical identity, clone revisi, history endpoint, dan import behavior untuk Blue Book.
> Deliverable: BB Project bisa memakai `bb_code` yang sama pada revisi lain, revisi BB bisa clone snapshot, dan history BB Project tersedia.
> Referensi: `plans/PLAN_BE_07_Revision_Versioning_Schema.md`, `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`, `docs/PRISM_Business_Rules.md`, `docs/PRISM_API_Contract.md`.

---

## Prasyarat

- [x] `PLAN_BE_07_Revision_Versioning_Schema.md` selesai.
- [x] `make generate` sudah berhasil setelah perubahan SQL.
- [x] Query identity/latest/history BB tersedia di generated sqlc package.

---

## Step 1 - Model dan Response Contract

Files:

- `prism-backend/internal/model/blue_book.go`

Checklist:

- [x] Tambahkan `ProjectIdentityID` pada BB Project request/response bila diperlukan.
- [x] Tambahkan field response:
  - [x] `project_identity_id`.
  - [x] `is_latest`.
  - [x] `has_newer_revision`.
- [x] Tambahkan model history item untuk `GET /bb-projects/:id/history`.
- [x] Pastikan semua tipe strongly typed, tidak memakai `any`/`interface{}` untuk payload DB.
- [x] Pastikan JSON field mengikuti contract snake_case.

Acceptance:

- [x] Model dapat merepresentasikan snapshot dan logical identity.

---

## Step 2 - Create dan Update BB Project

Files:

- `prism-backend/internal/service/blue_book_service.go`

Checklist:

- [x] Ganti validasi duplicate dari global code menjadi per `blue_book_id`.
- [x] Saat create tanpa `project_identity_id`, buat `project_identity` baru dalam transaksi.
- [x] Saat create dengan `project_identity_id`, validasi identity ada.
- [x] Simpan `project_identity_id` ke `bb_project`.
- [x] Build response mengisi latest flags:
  - [x] `is_latest`.
  - [x] `has_newer_revision`.
- [x] Update BB Project tidak mengubah `bb_code` kecuali memang contract memperbolehkan; jika diperbolehkan, validasi tetap per Blue Book.
- [x] Pastikan child rows tetap transaksional: institutions, locations, priorities, costs, lender indications.

Acceptance:

- [x] `bb_code` sama bisa dibuat di revisi Blue Book berbeda.
- [x] Duplicate `bb_code` dalam Blue Book yang sama tetap conflict.

---

## Step 3 - Create Blue Book Revision dan Clone Snapshot

Files:

- `prism-backend/internal/service/blue_book_service.go`

Checklist:

- [x] Tentukan source revisi:
  - [x] Jika request membawa `replaces_blue_book_id`, gunakan itu.
  - [x] Jika tidak, gunakan Blue Book active terakhir pada period yang sama sebelum supersede.
- [x] Dalam satu transaksi:
  - [x] Supersede Blue Book lama.
  - [x] Create Blue Book baru.
  - [x] Clone semua active BB Project dari source.
  - [x] Reuse `project_identity_id` dari source project.
  - [x] Clone institutions.
  - [x] Clone locations.
  - [x] Clone national priorities.
  - [x] Clone project costs.
  - [x] Clone lender indications.
  - [x] Clone LoI sesuai keputusan plan default.
- [x] Pastikan clone tidak memakai lookup global `bb_code`.
- [x] Publish SSE setelah commit berhasil.

Acceptance:

- [x] Revisi BB baru berisi snapshot project yang sama dengan identity yang sama.
- [x] BB lama menjadi `superseded`.

---

## Step 4 - History Endpoint dan Route

Files:

- `prism-backend/internal/handler/blue_book_handler.go`
- `prism-backend/cmd/api/main.go`

Checklist:

- [x] Implement service method `GetBBProjectHistory`.
- [x] Implement handler untuk `GET /bb-projects/:id/history`.
- [x] Daftarkan route dengan permission `bb_project:read`.
- [x] Response history memuat:
  - [x] snapshot id.
  - [x] `project_identity_id`.
  - [x] Blue Book id.
  - [x] book label atau period + revision.
  - [x] `revision_number`.
  - [x] `revision_year`.
  - [x] book status.
  - [x] `is_latest`.
  - [x] `used_by_downstream` jika query tersedia.
- [x] Return 404 aman jika snapshot tidak ditemukan.

Acceptance:

- [x] User bisa melihat semua revisi BB yang memuat logical project yang sama.

---

## Step 5 - Blue Book Import Behavior

Files:

- `prism-backend/internal/service/blue_book_import_service.go`

Checklist:

- [x] Duplicate `BB Code` dalam workbook tetap error.
- [x] `BB Code` yang sudah ada di revisi lama tidak di-skip otomatis.
- [x] Jika import ke Blue Book revisi dan kode ditemukan pada revisi/source sebelumnya, reuse `project_identity_id`.
- [x] Jika kode tidak ditemukan pada source identity manapun, create identity baru.
- [x] Relasi workbook tetap memakai `BB Code` hanya dalam konteks workbook target.
- [x] Import summary membedakan:
  - [x] created new logical project.
  - [x] created revision snapshot.
  - [x] skipped due to duplicate within workbook/target document.

Acceptance:

- [x] Import ke revisi bisa membuat snapshot dengan kode yang sama tanpa kehilangan history.

---

## Step 6 - Tests dan Verification

Checklist:

- [ ] Tambah/update backend tests untuk create BB Project duplicate per Blue Book.
- [ ] Tambah/update tests untuk create BB revision clone.
- [ ] Tambah/update tests untuk BB Project history endpoint.
- [ ] Tambah/update tests untuk import code reuse behavior.
- [x] Jalankan `go test ./...`.
- [x] Jalankan smoke API:
  - [x] create BB original.
  - [x] create BB Project `BB-001`.
  - [x] create BB revision.
  - [x] verify revisi memuat `BB-001` dengan `project_identity_id` sama.
  - [x] call `/bb-projects/:id/history`.

Done Criteria:

- [x] Blue Book revision versioning bekerja end-to-end.
- [x] Tidak ada global `bb_code` blocker tersisa di service/import.
