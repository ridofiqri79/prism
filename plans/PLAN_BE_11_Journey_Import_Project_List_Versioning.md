# PLAN BE-11 - Journey, Import, Project List & Versioning Smoke

> Scope: integrasi akhir backend untuk project list, journey, import, dashboard counts, dan smoke test end-to-end.
> Deliverable: list default tidak menduplikasi snapshot lama, journey menampilkan concrete path + newer indicator, import mendukung code reuse lintas revisi, dan smoke flow BB->GB->DK->LA stabil.
> Referensi: `plans/PLAN_BE_08_Blue_Book_Revision_Versioning.md`, `plans/PLAN_BE_09_Green_Book_Revision_Versioning.md`, `plans/PLAN_BE_10_DK_LA_Frozen_Snapshot.md`, `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`, `docs/PRISM_API_Contract.md`.

---

## Prasyarat

- [ ] `PLAN_BE_08_Blue_Book_Revision_Versioning.md` selesai.
- [ ] `PLAN_BE_09_Green_Book_Revision_Versioning.md` selesai.
- [ ] `PLAN_BE_10_DK_LA_Frozen_Snapshot.md` selesai.
- [ ] History endpoints BB/GB tersedia.

---

## Step 1 - Project List Latest Default

Files:

- `prism-backend/sql/queries/project.sql`
- `prism-backend/internal/model/project.go`
- `prism-backend/internal/service/project_service.go`
- `prism-backend/internal/handler/project_handler.go`

Checklist:

- [ ] Default project list hanya menampilkan latest BB Project snapshot per `project_identity_id`.
- [ ] Tambahkan parameter eksplisit untuk melihat semua snapshot jika dibutuhkan, misalnya `include_history=true`.
- [ ] Tambahkan field response:
  - [ ] `project_identity_id`.
  - [ ] `is_latest`.
  - [ ] `has_newer_revision`.
  - [ ] `blue_book_revision_label`.
- [ ] Sorting/filter existing tetap bekerja.
- [ ] Dashboard/list count tidak double count snapshot lama kecuali historical mode aktif.

Acceptance:

- [ ] Project list default tidak membingungkan user dengan duplikasi revisi.

---

## Step 2 - Journey Concrete Path dan Newer Indicator

Files:

- `prism-backend/sql/queries/monitoring.sql` atau file journey query terkait
- `prism-backend/internal/model/journey.go`
- `prism-backend/internal/service/journey_service.go`

Checklist:

- [ ] Journey tetap menerima concrete `bb_project_id` atau resolve entry point sesuai contract.
- [ ] Journey menampilkan path historis berdasarkan concrete relation yang tersimpan.
- [ ] Tambahkan identity metadata:
  - [ ] `project_identity_id` untuk BB node.
  - [ ] `gb_project_identity_id` untuk GB node.
- [ ] Tambahkan indicator:
  - [ ] BB node punya newer snapshot atau tidak.
  - [ ] GB node punya newer snapshot atau tidak.
- [ ] Jangan auto-ganti DK/LA path ke latest.
- [ ] Jika entry point adalah old BB snapshot, response tetap valid dan bisa menunjukkan newer BB revision.

Acceptance:

- [ ] Journey bisa membedakan "versi yang dipakai" dan "ada revisi lebih baru".

---

## Step 3 - Import Blue Book dan Green Book

Files:

- `prism-backend/internal/service/blue_book_import_service.go`
- Green Book import service jika tersedia atau saat dibuat.

Checklist:

- [ ] Blue Book import:
  - [ ] Duplicate `BB Code` dalam workbook tetap error.
  - [ ] Code yang ada pada revisi lama tidak di-skip global.
  - [ ] Code yang cocok dengan source revisi reuse `project_identity_id`.
  - [ ] Code baru create `project_identity`.
- [ ] Green Book import:
  - [ ] Duplicate `GB Code` dalam workbook tetap error.
  - [ ] Code yang ada pada revisi lama tidak di-skip global.
  - [ ] Code yang cocok dengan source revisi reuse `gb_project_identity_id`.
  - [ ] BB relations di-resolve ke latest BB snapshot.
- [ ] Import preview/summary memberi pesan yang membedakan snapshot revisi dan project logical baru.

Acceptance:

- [ ] Import tidak lagi bertentangan dengan behavior revisi.

---

## Step 4 - Dashboard dan Aggregate Safety

Files:

- `prism-backend/sql/queries/monitoring.sql`
- `prism-backend/internal/service/dashboard_service.go`

Checklist:

- [ ] `total_bb_projects` menghitung latest logical BB Project, bukan semua snapshot, kecuali contract meminta historical count.
- [ ] `total_gb_projects` menghitung latest logical GB Project.
- [ ] Monitoring/LA aggregates tetap berdasarkan concrete downstream records.
- [ ] Tidak ada double count karena snapshot revisi lama.
- [ ] Contract dashboard diperbarui jika ada field baru untuk historical counts.

Acceptance:

- [ ] Dashboard summary tidak naik palsu hanya karena revisi dokumen.

---

## Step 5 - End-to-End Smoke

Checklist:

- [ ] Reset DB dev jika dibutuhkan karena data fresh diperbolehkan.
- [ ] Jalankan backend dan DB.
- [ ] Create BB original.
- [ ] Create BB Project `BB-001`.
- [ ] Create BB revision dan verify `BB-001` diclone dengan identity sama.
- [ ] Create GB original dengan input BB old/latest.
- [ ] Verify stored BB relation memakai latest BB snapshot saat GB dibuat.
- [ ] Create GB revision dan verify `GB-001` identity sama.
- [ ] Create DK dengan input GB old/latest.
- [ ] Verify DK stored relation memakai latest GB saat DK dibuat.
- [ ] Create revisi BB/GB baru setelah DK.
- [ ] Verify DK/LA path tidak berubah.
- [ ] Call BB history endpoint.
- [ ] Call GB history endpoint.
- [ ] Call journey endpoint dan verify newer indicator.
- [ ] Run `go test ./...`.

Done Criteria:

- [ ] Backend versioning flow lulus dari schema sampai journey.
- [ ] Tidak ada double count pada project list/dashboard default.
- [ ] Import dan smoke flow selaras dengan business rules baru.
