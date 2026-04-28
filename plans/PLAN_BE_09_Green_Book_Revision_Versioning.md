# PLAN BE-09 - Green Book Revision Versioning

> Scope: implementasi logical identity, latest BB resolver, clone revisi, dan history endpoint untuk Green Book.
> Deliverable: GB Project bisa memakai `gb_code` yang sama pada revisi lain, GB selalu link ke latest BB saat dibuat/direvisi, dan history GB Project tersedia.
> Referensi: `plans/PLAN_BE_07_Revision_Versioning_Schema.md`, `plans/PLAN_BE_08_Blue_Book_Revision_Versioning.md`, `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`, `docs/PRISM_Business_Rules.md`, `docs/PRISM_API_Contract.md`.

---

## Prasyarat

- [x] `PLAN_BE_07_Revision_Versioning_Schema.md` selesai.
- [x] `PLAN_BE_08_Blue_Book_Revision_Versioning.md` selesai.
- [x] Latest BB Project resolver tersedia dan sudah teruji.

---

## Step 1 - Model dan Response Contract

Files:

- `prism-backend/internal/model/green_book.go`

Checklist:

- [x] Tambahkan `GBProjectIdentityID` pada GB Project request/response bila diperlukan.
- [x] Tambahkan field response:
  - [x] `gb_project_identity_id`.
  - [x] `is_latest`.
  - [x] `has_newer_revision`.
- [x] Tambahkan model history item untuk `GET /gb-projects/:id/history`.
- [x] Tambahkan field untuk BB relation jika perlu menampilkan concrete BB snapshot yang tersimpan.
- [x] Pastikan JSON field mengikuti contract snake_case.

Acceptance:

- [x] Model dapat merepresentasikan GB snapshot dan logical GB identity.

---

## Step 2 - Create dan Update GB Project

Files:

- `prism-backend/internal/service/green_book_service.go`

Checklist:

- [x] Ganti validasi duplicate dari global code menjadi per `green_book_id`.
- [x] Saat create tanpa `gb_project_identity_id`, buat `gb_project_identity` baru dalam transaksi.
- [x] Saat create dengan `gb_project_identity_id`, validasi identity ada.
- [x] Simpan `gb_project_identity_id` ke `gb_project`.
- [x] Untuk setiap input `bb_project_id`, resolve ke latest BB Project dari identity yang sama.
- [x] Simpan concrete latest `bb_project_id` hasil resolve ke `gb_project_bb_project`.
- [x] Jangan simpan logical id saja pada junction; concrete snapshot tetap wajib.
- [x] Build response mengisi latest flags.
- [x] Pastikan activities, funding source, disbursement plan, dan funding allocation tetap transaksional.

Acceptance:

- [x] GB Project baru selalu terhubung ke latest BB snapshot.
- [x] `gb_code` sama bisa dibuat di revisi Green Book berbeda.
- [x] Duplicate `gb_code` dalam Green Book yang sama tetap conflict.

---

## Step 3 - Create Green Book Revision dan Clone Snapshot

Files:

- `prism-backend/internal/service/green_book_service.go`

Checklist:

- [x] Tentukan source revisi:
  - [x] Jika request membawa `replaces_green_book_id`, gunakan itu.
  - [x] Jika tidak, gunakan Green Book active terakhir pada `publish_year` yang sama sebelum supersede.
- [x] Dalam satu transaksi:
  - [x] Supersede Green Book lama.
  - [x] Create Green Book baru.
  - [x] Clone semua active GB Project dari source.
  - [x] Reuse `gb_project_identity_id` dari source project.
  - [x] Untuk setiap cloned GB Project, resolve BB relation ke latest BB Project saat revisi dibuat.
  - [x] Clone institutions.
  - [x] Clone locations.
  - [x] Clone funding sources.
  - [x] Clone disbursement plan.
  - [x] Clone activities dan simpan mapping old activity id ke new activity id.
  - [x] Clone funding allocation memakai new activity id dari mapping.
- [x] Pastikan clone tidak memakai lookup global `gb_code`.
- [x] Publish SSE setelah commit berhasil.

Acceptance:

- [x] Revisi GB baru berisi snapshot project yang sama dengan GB identity yang sama.
- [x] Relasi BB pada revisi GB menunjuk latest BB snapshot saat revisi dibuat.

---

## Step 4 - History Endpoint dan Route

Files:

- `prism-backend/internal/handler/green_book_handler.go`
- `prism-backend/cmd/api/main.go`

Checklist:

- [x] Implement service method `GetGBProjectHistory`.
- [x] Implement handler untuk `GET /gb-projects/:id/history`.
- [x] Daftarkan route dengan permission `gb_project:read`.
- [x] Response history memuat:
  - [x] snapshot id.
  - [x] `gb_project_identity_id`.
  - [x] Green Book id.
  - [x] publish year + revision.
  - [x] book status.
  - [x] `is_latest`.
  - [x] concrete BB Project snapshots yang dipakai.
  - [x] `used_by_downstream` jika query tersedia.
- [x] Return 404 aman jika snapshot tidak ditemukan.

Acceptance:

- [x] User bisa melihat semua revisi GB yang memuat logical GB project yang sama.

---

## Step 5 - Tests dan Verification

Checklist:

- [ ] Tambah/update backend tests untuk create GB Project duplicate per Green Book.
- [ ] Tambah/update tests untuk GB Project resolve latest BB.
- [ ] Tambah/update tests untuk create GB revision clone.
- [ ] Tambah/update tests untuk funding allocation clone dengan activity mapping.
- [ ] Tambah/update tests untuk GB Project history endpoint.
- [x] Jalankan `go test ./...`.
- [x] Jalankan smoke API:
  - [x] create BB original + revision.
  - [x] create GB using old BB input id.
  - [x] verify stored BB relation memakai latest BB snapshot.
  - [x] create GB revision.
  - [x] call `/gb-projects/:id/history`.

Done Criteria:

- [x] Green Book revision versioning bekerja end-to-end.
- [x] Tidak ada global `gb_code` blocker tersisa di service.
