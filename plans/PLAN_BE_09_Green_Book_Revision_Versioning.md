# PLAN BE-09 - Green Book Revision Versioning

> Scope: implementasi logical identity, latest BB resolver, clone revisi, dan history endpoint untuk Green Book.
> Deliverable: GB Project bisa memakai `gb_code` yang sama pada revisi lain, GB selalu link ke latest BB saat dibuat/direvisi, dan history GB Project tersedia.
> Referensi: `plans/PLAN_BE_07_Revision_Versioning_Schema.md`, `plans/PLAN_BE_08_Blue_Book_Revision_Versioning.md`, `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`, `docs/PRISM_Business_Rules.md`, `docs/PRISM_API_Contract.md`.

---

## Prasyarat

- [ ] `PLAN_BE_07_Revision_Versioning_Schema.md` selesai.
- [ ] `PLAN_BE_08_Blue_Book_Revision_Versioning.md` selesai.
- [ ] Latest BB Project resolver tersedia dan sudah teruji.

---

## Step 1 - Model dan Response Contract

Files:

- `prism-backend/internal/model/green_book.go`

Checklist:

- [ ] Tambahkan `GBProjectIdentityID` pada GB Project request/response bila diperlukan.
- [ ] Tambahkan field response:
  - [ ] `gb_project_identity_id`.
  - [ ] `is_latest`.
  - [ ] `has_newer_revision`.
- [ ] Tambahkan model history item untuk `GET /gb-projects/:id/history`.
- [ ] Tambahkan field untuk BB relation jika perlu menampilkan concrete BB snapshot yang tersimpan.
- [ ] Pastikan JSON field mengikuti contract snake_case.

Acceptance:

- [ ] Model dapat merepresentasikan GB snapshot dan logical GB identity.

---

## Step 2 - Create dan Update GB Project

Files:

- `prism-backend/internal/service/green_book_service.go`

Checklist:

- [ ] Ganti validasi duplicate dari global code menjadi per `green_book_id`.
- [ ] Saat create tanpa `gb_project_identity_id`, buat `gb_project_identity` baru dalam transaksi.
- [ ] Saat create dengan `gb_project_identity_id`, validasi identity ada.
- [ ] Simpan `gb_project_identity_id` ke `gb_project`.
- [ ] Untuk setiap input `bb_project_id`, resolve ke latest BB Project dari identity yang sama.
- [ ] Simpan concrete latest `bb_project_id` hasil resolve ke `gb_project_bb_project`.
- [ ] Jangan simpan logical id saja pada junction; concrete snapshot tetap wajib.
- [ ] Build response mengisi latest flags.
- [ ] Pastikan activities, funding source, disbursement plan, dan funding allocation tetap transaksional.

Acceptance:

- [ ] GB Project baru selalu terhubung ke latest BB snapshot.
- [ ] `gb_code` sama bisa dibuat di revisi Green Book berbeda.
- [ ] Duplicate `gb_code` dalam Green Book yang sama tetap conflict.

---

## Step 3 - Create Green Book Revision dan Clone Snapshot

Files:

- `prism-backend/internal/service/green_book_service.go`

Checklist:

- [ ] Tentukan source revisi:
  - [ ] Jika request membawa `replaces_green_book_id`, gunakan itu.
  - [ ] Jika tidak, gunakan Green Book active terakhir pada `publish_year` yang sama sebelum supersede.
- [ ] Dalam satu transaksi:
  - [ ] Supersede Green Book lama.
  - [ ] Create Green Book baru.
  - [ ] Clone semua active GB Project dari source.
  - [ ] Reuse `gb_project_identity_id` dari source project.
  - [ ] Untuk setiap cloned GB Project, resolve BB relation ke latest BB Project saat revisi dibuat.
  - [ ] Clone institutions.
  - [ ] Clone locations.
  - [ ] Clone funding sources.
  - [ ] Clone disbursement plan.
  - [ ] Clone activities dan simpan mapping old activity id ke new activity id.
  - [ ] Clone funding allocation memakai new activity id dari mapping.
- [ ] Pastikan clone tidak memakai lookup global `gb_code`.
- [ ] Publish SSE setelah commit berhasil.

Acceptance:

- [ ] Revisi GB baru berisi snapshot project yang sama dengan GB identity yang sama.
- [ ] Relasi BB pada revisi GB menunjuk latest BB snapshot saat revisi dibuat.

---

## Step 4 - History Endpoint dan Route

Files:

- `prism-backend/internal/handler/green_book_handler.go`
- `prism-backend/cmd/api/main.go`

Checklist:

- [ ] Implement service method `GetGBProjectHistory`.
- [ ] Implement handler untuk `GET /gb-projects/:id/history`.
- [ ] Daftarkan route dengan permission `gb_project:read`.
- [ ] Response history memuat:
  - [ ] snapshot id.
  - [ ] `gb_project_identity_id`.
  - [ ] Green Book id.
  - [ ] publish year + revision.
  - [ ] book status.
  - [ ] `is_latest`.
  - [ ] concrete BB Project snapshots yang dipakai.
  - [ ] `used_by_downstream` jika query tersedia.
- [ ] Return 404 aman jika snapshot tidak ditemukan.

Acceptance:

- [ ] User bisa melihat semua revisi GB yang memuat logical GB project yang sama.

---

## Step 5 - Tests dan Verification

Checklist:

- [ ] Tambah/update backend tests untuk create GB Project duplicate per Green Book.
- [ ] Tambah/update tests untuk GB Project resolve latest BB.
- [ ] Tambah/update tests untuk create GB revision clone.
- [ ] Tambah/update tests untuk funding allocation clone dengan activity mapping.
- [ ] Tambah/update tests untuk GB Project history endpoint.
- [ ] Jalankan `go test ./...`.
- [ ] Jalankan smoke API:
  - [ ] create BB original + revision.
  - [ ] create GB using old BB input id.
  - [ ] verify stored BB relation memakai latest BB snapshot.
  - [ ] create GB revision.
  - [ ] call `/gb-projects/:id/history`.

Done Criteria:

- [ ] Green Book revision versioning bekerja end-to-end.
- [ ] Tidak ada global `gb_code` blocker tersisa di service.
