# PLAN BE-07 - BB/GB Revision Versioning Schema & Queries

> Scope: foundation schema, query sqlc, dan API contract untuk BB/GB project snapshot versioning.
> Deliverable: identity tables, per-document code uniqueness, latest/history queries, generated sqlc code, dan kontrak API awal.
> Referensi: `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`, `docs/PRISM_Business_Rules.md`, `docs/prism_ddl.sql`, `docs/PRISM_API_Contract.md`, `docs/PRISM_Dev_Workflow.md`.

---

## Prinsip Fase

- Jangan implement service Go sebelum semua SQL selesai dan `make generate` berhasil.
- Data fresh boleh di-reset, tetapi schema target tetap harus terdokumentasi di `docs/prism_ddl.sql`.
- Jika membuat migration, tetap incremental. Jangan drop-recreate tabel.
- `bb_project` dan `gb_project` adalah snapshot per dokumen/revisi.
- Identity lintas revisi wajib eksplisit melalui `project_identity` dan `gb_project_identity`.

---

## Step 1 - Audit Schema dan Constraint Saat Ini

Checklist:

- [ ] Baca `docs/PRISM_BB_GB_Revision_Versioning_Plan.md` sampai selesai.
- [ ] Baca ulang aturan BB/GB/DK di `docs/PRISM_Business_Rules.md`.
- [ ] Cek `docs/prism_ddl.sql` untuk tabel `blue_book`, `bb_project`, `green_book`, `gb_project`, `gb_project_bb_project`, dan `dk_project_gb_project`.
- [ ] Catat constraint lama yang harus diganti:
  - [ ] `bb_project.bb_code UNIQUE`.
  - [ ] `gb_project.gb_code UNIQUE`.
- [ ] Cek query lama yang masih memakai lookup global:
  - [ ] `GetBBProjectByCode`.
  - [ ] query duplicate/check existing GB Code.
  - [ ] query import yang skip code global.

Acceptance:

- [ ] Daftar constraint/query lama yang terdampak sudah jelas sebelum edit SQL.

---

## Step 2 - Update DDL dan Migration

Checklist:

- [ ] Tambahkan tabel `project_identity` di `docs/prism_ddl.sql`.
- [ ] Tambahkan tabel `gb_project_identity` di `docs/prism_ddl.sql`.
- [ ] Tambahkan `project_identity_id UUID NOT NULL REFERENCES project_identity(id)` ke `bb_project`.
- [ ] Tambahkan `gb_project_identity_id UUID NOT NULL REFERENCES gb_project_identity(id)` ke `gb_project`.
- [ ] Ganti uniqueness BB dari global menjadi `UNIQUE (blue_book_id, bb_code)`.
- [ ] Ganti uniqueness GB dari global menjadi `UNIQUE (green_book_id, gb_code)`.
- [ ] Tambahkan index:
  - [ ] `idx_bb_project_identity`.
  - [ ] `idx_bb_project_book_code`.
  - [ ] `idx_gb_project_identity`.
  - [ ] `idx_gb_project_book_code`.
- [ ] Tambahkan optional lineage header jika dipilih:
  - [ ] `blue_book.replaces_blue_book_id`.
  - [ ] `green_book.replaces_green_book_id`.
- [ ] Jika stack migration digunakan untuk DB lokal yang sudah hidup, buat migration incremental di `prism-backend/migrations/`.
- [ ] Pastikan migration tidak melakukan drop-recreate tabel.

Acceptance:

- [ ] DDL target sudah mencerminkan snapshot + identity.
- [ ] Constraint global code sudah tidak menjadi sumber kebenaran.

---

## Step 3 - Update Query Blue Book

File utama: `prism-backend/sql/queries/bb_project.sql`.

Checklist:

- [ ] Tambahkan query create/get identity:
  - [ ] `CreateProjectIdentity`.
  - [ ] `GetProjectIdentity`.
- [ ] Tambahkan lookup duplicate per dokumen:
  - [ ] `GetBBProjectByBlueBookAndCode`.
- [ ] Sesuaikan `CreateBBProject` agar menerima `project_identity_id`.
- [ ] Sesuaikan list/get response query agar membawa `project_identity_id`.
- [ ] Tambahkan latest resolver:
  - [ ] `GetLatestBBProjectByIdentity`.
  - [ ] `GetLatestBBProjectByProject`.
- [ ] Tambahkan history query:
  - [ ] `ListBBProjectHistoryByIdentity`.
  - [ ] `ListBBProjectHistoryByProject`.
- [ ] Tambahkan query clone helper jika perlu:
  - [ ] list source projects by Blue Book.
  - [ ] list source child rows by BB Project.

Acceptance:

- [ ] Query BB bisa membedakan duplicate dalam dokumen vs kode sama di revisi lain.
- [ ] Query BB bisa mengambil latest snapshot dan history snapshot.

---

## Step 4 - Update Query Green Book

File utama: `prism-backend/sql/queries/gb_project.sql`.

Checklist:

- [ ] Tambahkan query create/get identity:
  - [ ] `CreateGBProjectIdentity`.
  - [ ] `GetGBProjectIdentity`.
- [ ] Tambahkan lookup duplicate per dokumen:
  - [ ] `GetGBProjectByGreenBookAndCode`.
- [ ] Sesuaikan `CreateGBProject` agar menerima `gb_project_identity_id`.
- [ ] Sesuaikan list/get response query agar membawa `gb_project_identity_id`.
- [ ] Tambahkan latest resolver:
  - [ ] `GetLatestGBProjectByIdentity`.
  - [ ] `GetLatestGBProjectByProject`.
- [ ] Tambahkan history query:
  - [ ] `ListGBProjectHistoryByIdentity`.
  - [ ] `ListGBProjectHistoryByProject`.
- [ ] Tambahkan query clone helper untuk activities, funding source, disbursement plan, funding allocation, institutions, locations, dan BB relations.

Acceptance:

- [ ] Query GB bisa membedakan duplicate dalam dokumen vs kode sama di revisi lain.
- [ ] Query GB bisa mengambil latest snapshot dan history snapshot.

---

## Step 5 - Update Query DK, Journey, dan List Master

Files:

- `prism-backend/sql/queries/dk_project.sql`
- `prism-backend/sql/queries/project.sql`
- `prism-backend/sql/queries/monitoring.sql`

Checklist:

- [ ] Tambahkan resolver latest GB untuk input DK:
  - [ ] from `gb_project_id` ke latest by `gb_project_identity_id`.
- [ ] Pastikan query DK detail tetap membaca concrete `gb_project_id` yang tersimpan.
- [ ] Tambahkan query history/newer indicator untuk journey:
  - [ ] ada newer BB snapshot untuk identity yang sama.
  - [ ] ada newer GB snapshot untuk identity yang sama.
- [ ] Update query project list agar bisa default ke latest snapshot per identity.
- [ ] Jika perlu historical mode, tambahkan parameter eksplisit seperti `include_history`.

Acceptance:

- [ ] Query DK bisa resolve latest saat create/update eksplisit.
- [ ] Query journey/detail tetap bisa membaca concrete historical path.

---

## Step 6 - Generate dan Contract

Checklist:

- [ ] Jalankan `make generate` dari `prism-backend`.
- [ ] Jika `make generate` gagal karena environment Windows, coba `sqlc generate` sesuai workflow repo.
- [ ] Jangan edit `internal/database/queries/*.go` manual.
- [ ] Update `docs/PRISM_API_Contract.md` untuk field baru:
  - [ ] `project_identity_id`.
  - [ ] `gb_project_identity_id`.
  - [ ] `is_latest`.
  - [ ] `has_newer_revision`.
- [ ] Tambahkan endpoint contract:
  - [ ] `GET /bb-projects/:id/history`.
  - [ ] `GET /gb-projects/:id/history`.
- [ ] Tambahkan catatan contract untuk default latest picker/list vs historical detail.

Acceptance:

- [ ] `make generate` berhasil.
- [ ] API contract menyatakan field dan endpoint versioning.
- [ ] Tidak ada raw SQL baru di file Go.

---

## Step 7 - Verification

Checklist:

- [ ] Jalankan `go test ./...` dari `prism-backend`.
- [ ] Jika test belum ada untuk fase ini, minimal pastikan build package backend lulus.
- [ ] Cek `git diff` untuk memastikan tidak ada edit manual di generated query files selain hasil sqlc.
- [ ] Cek `docs/PRISM_BB_GB_Revision_Versioning_Plan.md` masih konsisten dengan DDL/API contract.

Done Criteria:

- [ ] Schema target siap.
- [ ] sqlc code generated.
- [ ] Query latest/history tersedia untuk fase service berikutnya.
- [ ] API contract siap dipakai frontend dan backend phases berikutnya.
