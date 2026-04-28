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

- [x] Baca `docs/PRISM_BB_GB_Revision_Versioning_Plan.md` sampai selesai.
- [x] Baca ulang aturan BB/GB/DK di `docs/PRISM_Business_Rules.md`.
- [x] Cek `docs/prism_ddl.sql` untuk tabel `blue_book`, `bb_project`, `green_book`, `gb_project`, `gb_project_bb_project`, dan `dk_project_gb_project`.
- [x] Catat constraint lama yang harus diganti:
  - [x] `bb_project.bb_code UNIQUE`.
  - [x] `gb_project.gb_code UNIQUE`.
- [x] Cek query lama yang masih memakai lookup global:
  - [x] `GetBBProjectByCode`.
  - [x] query duplicate/check existing GB Code.
  - [x] query import yang skip code global.

Acceptance:

- [x] Daftar constraint/query lama yang terdampak sudah jelas sebelum edit SQL.

---

## Step 2 - Update DDL dan Migration

Checklist:

- [x] Tambahkan tabel `project_identity` di `docs/prism_ddl.sql`.
- [x] Tambahkan tabel `gb_project_identity` di `docs/prism_ddl.sql`.
- [x] Tambahkan `project_identity_id UUID NOT NULL REFERENCES project_identity(id)` ke `bb_project`.
- [x] Tambahkan `gb_project_identity_id UUID NOT NULL REFERENCES gb_project_identity(id)` ke `gb_project`.
- [x] Ganti uniqueness BB dari global menjadi `UNIQUE (blue_book_id, bb_code)`.
- [x] Ganti uniqueness GB dari global menjadi `UNIQUE (green_book_id, gb_code)`.
- [x] Tambahkan index:
  - [x] `idx_bb_project_identity`.
  - [x] `idx_bb_project_book_code`.
  - [x] `idx_gb_project_identity`.
  - [x] `idx_gb_project_book_code`.
- [x] Tambahkan optional lineage header jika dipilih:
  - [x] `blue_book.replaces_blue_book_id`.
  - [x] `green_book.replaces_green_book_id`.
- [x] Jika stack migration digunakan untuk DB lokal yang sudah hidup, buat migration incremental di `prism-backend/migrations/`.
- [x] Pastikan migration tidak melakukan drop-recreate tabel.

Acceptance:

- [x] DDL target sudah mencerminkan snapshot + identity.
- [x] Constraint global code sudah tidak menjadi sumber kebenaran.

---

## Step 3 - Update Query Blue Book

File utama: `prism-backend/sql/queries/bb_project.sql`.

Checklist:

- [x] Tambahkan query create/get identity:
  - [x] `CreateProjectIdentity`.
  - [x] `GetProjectIdentity`.
- [x] Tambahkan lookup duplicate per dokumen:
  - [x] `GetBBProjectByBlueBookAndCode`.
- [x] Sesuaikan `CreateBBProject` agar menerima `project_identity_id`.
- [x] Sesuaikan list/get response query agar membawa `project_identity_id`.
- [x] Tambahkan latest resolver:
  - [x] `GetLatestBBProjectByIdentity`.
  - [x] `GetLatestBBProjectByProject`.
- [x] Tambahkan history query:
  - [x] `ListBBProjectHistoryByIdentity`.
  - [x] `ListBBProjectHistoryByProject`.
- [x] Tambahkan query clone helper jika perlu:
  - [x] list source projects by Blue Book.
  - [x] list source child rows by BB Project.

Acceptance:

- [x] Query BB bisa membedakan duplicate dalam dokumen vs kode sama di revisi lain.
- [x] Query BB bisa mengambil latest snapshot dan history snapshot.

---

## Step 4 - Update Query Green Book

File utama: `prism-backend/sql/queries/gb_project.sql`.

Checklist:

- [x] Tambahkan query create/get identity:
  - [x] `CreateGBProjectIdentity`.
  - [x] `GetGBProjectIdentity`.
- [x] Tambahkan lookup duplicate per dokumen:
  - [x] `GetGBProjectByGreenBookAndCode`.
- [x] Sesuaikan `CreateGBProject` agar menerima `gb_project_identity_id`.
- [x] Sesuaikan list/get response query agar membawa `gb_project_identity_id`.
- [x] Tambahkan latest resolver:
  - [x] `GetLatestGBProjectByIdentity`.
  - [x] `GetLatestGBProjectByProject`.
- [x] Tambahkan history query:
  - [x] `ListGBProjectHistoryByIdentity`.
  - [x] `ListGBProjectHistoryByProject`.
- [x] Tambahkan query clone helper untuk activities, funding source, disbursement plan, funding allocation, institutions, locations, dan BB relations.

Acceptance:

- [x] Query GB bisa membedakan duplicate dalam dokumen vs kode sama di revisi lain.
- [x] Query GB bisa mengambil latest snapshot dan history snapshot.

---

## Step 5 - Update Query DK, Journey, dan List Master

Files:

- `prism-backend/sql/queries/dk_project.sql`
- `prism-backend/sql/queries/project.sql`
- `prism-backend/sql/queries/monitoring.sql`

Checklist:

- [x] Tambahkan resolver latest GB untuk input DK:
  - [x] from `gb_project_id` ke latest by `gb_project_identity_id`.
- [x] Pastikan query DK detail tetap membaca concrete `gb_project_id` yang tersimpan.
- [x] Tambahkan query history/newer indicator untuk journey:
  - [x] ada newer BB snapshot untuk identity yang sama.
  - [x] ada newer GB snapshot untuk identity yang sama.
- [x] Update query project list agar bisa default ke latest snapshot per identity.
- [x] Jika perlu historical mode, tambahkan parameter eksplisit seperti `include_history`.

Acceptance:

- [x] Query DK bisa resolve latest saat create/update eksplisit.
- [x] Query journey/detail tetap bisa membaca concrete historical path.

---

## Step 6 - Generate dan Contract

Checklist:

- [x] Jalankan `make generate` dari `prism-backend`.
- [x] Jika `make generate` gagal karena environment Windows, coba `sqlc generate` sesuai workflow repo.
- [x] Jangan edit `internal/database/queries/*.go` manual.
- [x] Update `docs/PRISM_API_Contract.md` untuk field baru:
  - [x] `project_identity_id`.
  - [x] `gb_project_identity_id`.
  - [x] `is_latest`.
  - [x] `has_newer_revision`.
- [x] Tambahkan endpoint contract:
  - [x] `GET /bb-projects/:id/history`.
  - [x] `GET /gb-projects/:id/history`.
- [x] Tambahkan catatan contract untuk default latest picker/list vs historical detail.

Acceptance:

- [x] `make generate` berhasil.
- [x] API contract menyatakan field dan endpoint versioning.
- [x] Tidak ada raw SQL baru di file Go.

---

## Step 7 - Verification

Checklist:

- [x] Jalankan `go test ./...` dari `prism-backend`.
- [x] Jika test belum ada untuk fase ini, minimal pastikan build package backend lulus.
- [x] Cek `git diff` untuk memastikan tidak ada edit manual di generated query files selain hasil sqlc.
- [x] Cek `docs/PRISM_BB_GB_Revision_Versioning_Plan.md` masih konsisten dengan DDL/API contract.

Done Criteria:

- [x] Schema target siap.
- [x] sqlc code generated.
- [x] Query latest/history tersedia untuk fase service berikutnya.
- [x] API contract siap dipakai frontend dan backend phases berikutnya.
