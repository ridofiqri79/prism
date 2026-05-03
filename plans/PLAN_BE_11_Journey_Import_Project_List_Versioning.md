# PLAN BE-11 - Journey, Import, Project List & Versioning Smoke

> Scope: integrasi akhir backend untuk project list, journey, import, dashboard counts, dan smoke test end-to-end.
> Deliverable: list default tidak menduplikasi snapshot lama, journey menampilkan concrete path + newer indicator, import mendukung code reuse lintas revisi, dan smoke flow BB->GB->DK->LA stabil.
> Referensi: `plans/PLAN_BE_08_Blue_Book_Revision_Versioning.md`, `plans/PLAN_BE_09_Green_Book_Revision_Versioning.md`, `plans/PLAN_BE_10_DK_LA_Frozen_Snapshot.md`, `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`, `docs/PRISM_API_Contract.md`.

---

## Prasyarat

- [x] `PLAN_BE_08_Blue_Book_Revision_Versioning.md` selesai.
- [x] `PLAN_BE_09_Green_Book_Revision_Versioning.md` selesai.
- [x] `PLAN_BE_10_DK_LA_Frozen_Snapshot.md` selesai.
- [x] History endpoints BB/GB tersedia.

---

## Step 1 - Project List Latest Default

Files:

- `prism-backend/sql/queries/project.sql`
- `prism-backend/internal/model/project.go`
- `prism-backend/internal/service/project_service.go`
- `prism-backend/internal/handler/project_handler.go`

Checklist:

- [x] Default project list hanya menampilkan latest BB Project snapshot per `project_identity_id`.
- [x] Tambahkan parameter eksplisit untuk melihat semua snapshot jika dibutuhkan, misalnya `include_history=true`.
- [x] Tambahkan field response:
  - [x] `project_identity_id`.
  - [x] `is_latest`.
  - [x] `has_newer_revision`.
  - [x] `blue_book_revision_label`.
- [x] Sorting/filter existing tetap bekerja.
- [x] Dashboard/list count tidak double count snapshot lama kecuali historical mode aktif.

Acceptance:

- [x] Project list default tidak membingungkan user dengan duplikasi revisi.

---

## Step 2 - Journey Concrete Path dan Newer Indicator

Files:

- `prism-backend/sql/queries/monitoring.sql` atau file journey query terkait
- `prism-backend/internal/model/journey.go`
- `prism-backend/internal/service/journey_service.go`

Checklist:

- [x] Journey tetap menerima concrete `bb_project_id` atau resolve entry point sesuai contract.
- [x] Journey menampilkan path historis berdasarkan concrete relation yang tersimpan.
- [x] Tambahkan identity metadata:
  - [x] `project_identity_id` untuk BB node.
  - [x] `gb_project_identity_id` untuk GB node.
- [x] Tambahkan indicator:
  - [x] BB node punya newer snapshot atau tidak.
  - [x] GB node punya newer snapshot atau tidak.
- [x] Jangan auto-ganti DK/LA path ke latest.
- [x] Jika entry point adalah old BB snapshot, response tetap valid dan bisa menunjukkan newer BB revision.

Acceptance:

- [x] Journey bisa membedakan "versi yang dipakai" dan "ada revisi lebih baru".

---

## Step 3 - Import Blue Book dan Green Book

Files:

- `prism-backend/internal/service/blue_book_import_service.go`
- Green Book import service jika tersedia atau saat dibuat.

Checklist:

- [x] Blue Book import:
  - [x] Duplicate `BB Code` dalam workbook tetap error.
  - [x] Code yang ada pada revisi lama tidak di-skip global.
  - [x] Code yang cocok dengan source revisi reuse `project_identity_id`.
  - [x] Code baru create `project_identity`.
- [x] Green Book import:
  - [x] Duplicate `GB Code` dalam workbook tetap error.
  - [x] Code yang ada pada revisi lama tidak di-skip global.
  - [x] Code yang cocok dengan source revisi reuse `gb_project_identity_id`.
  - [x] BB relations di-resolve ke latest BB snapshot.
- [x] Import preview/summary memberi pesan yang membedakan snapshot revisi dan project logical baru.

Acceptance:

- [x] Import tidak lagi bertentangan dengan behavior revisi.

---

## Step 4 - Dashboard dan Aggregate Safety

Files:

- `prism-backend/sql/queries/monitoring.sql`
- `prism-backend/internal/service/dashboard_service.go`

Checklist:

- [x] `total_bb_projects` menghitung latest logical BB Project, bukan semua snapshot, kecuali contract meminta historical count.
- [x] `total_gb_projects` menghitung latest logical GB Project.
- [x] Monitoring/LA aggregates tetap berdasarkan concrete downstream records.
- [x] Tidak ada double count karena snapshot revisi lama.
- [x] Contract dashboard diperbarui jika ada field baru untuk historical counts.

Acceptance:

- [x] Dashboard summary tidak naik palsu hanya karena revisi dokumen.

---

## Step 5 - End-to-End Smoke

Checklist:

- [x] Reset DB dev jika dibutuhkan karena data fresh diperbolehkan.
- [x] Jalankan backend dan DB.
- [x] Create BB original.
- [x] Create BB Project `BB-001`.
- [x] Create BB revision dan verify `BB-001` diclone dengan identity sama.
- [x] Create GB original dengan input BB old/latest.
- [x] Verify stored BB relation memakai latest BB snapshot saat GB dibuat.
- [x] Create GB revision dan verify `GB-001` identity sama.
- [x] Create DK dengan input GB old/latest.
- [x] Verify DK stored relation memakai latest GB saat DK dibuat.
- [x] Create revisi BB/GB baru setelah DK.
- [x] Verify DK/LA path tidak berubah.
- [x] Call BB history endpoint.
- [x] Call GB history endpoint.
- [x] Call journey endpoint dan verify newer indicator.
- [x] Run `go test ./...`.

Done Criteria:

- [x] Backend versioning flow lulus dari schema sampai journey.
- [x] Tidak ada double count pada project list/dashboard default.
- [x] Import dan smoke flow selaras dengan business rules baru.
