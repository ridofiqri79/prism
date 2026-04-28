# PLAN BE-10 - DK/LA Latest Resolver & Frozen Snapshot

> Scope: DK memakai latest GB saat link dibuat, tetapi DK/LA yang sudah tersimpan tetap frozen ke concrete snapshot.
> Deliverable: create/update eksplisit DK resolve ke latest GB Project, downstream tidak auto-pindah, dan lender validation memakai concrete stored version.
> Referensi: `plans/PLAN_BE_09_Green_Book_Revision_Versioning.md`, `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`, `docs/PRISM_Business_Rules.md`, `docs/PRISM_API_Contract.md`.

---

## Prasyarat

- [x] `PLAN_BE_07_Revision_Versioning_Schema.md` selesai.
- [x] `PLAN_BE_08_Blue_Book_Revision_Versioning.md` selesai.
- [x] `PLAN_BE_09_Green_Book_Revision_Versioning.md` selesai.
- [x] Latest GB Project resolver tersedia dan sudah teruji.

---

## Step 1 - Query Resolver dan Concrete Read Path

Files:

- `prism-backend/sql/queries/dk_project.sql`

Checklist:

- [x] Tambahkan query resolve latest GB Project dari input `gb_project_id`.
- [x] Pastikan resolver memakai `gb_project_identity_id`.
- [x] Pastikan resolver hanya memilih snapshot aktif dan dokumen/revisi yang valid untuk dipilih.
- [x] Pastikan query detail DK tetap membaca concrete `gb_project_id` dari `dk_project_gb_project`.
- [x] Pastikan query allowed lender membaca concrete GB/BB path yang tersimpan.
- [x] Jalankan `make generate` setelah query berubah.

Acceptance:

- [x] DK service punya query untuk resolve latest saat menulis.
- [x] DK read/detail tidak berubah menjadi dynamic latest.

---

## Step 2 - DK Create/Update Behavior

Files:

- `prism-backend/internal/service/daftar_kegiatan_service.go`

Checklist:

- [x] Saat create DK Project, resolve setiap input `gb_project_id` ke latest GB snapshot sebelum insert junction.
- [x] Saat update DK Project dan user eksplisit mengganti GB selection, resolve input baru ke latest GB snapshot.
- [x] Jangan pernah update `dk_project_gb_project` otomatis hanya karena ada revisi BB/GB baru.
- [x] Simpan concrete `gb_project_id` hasil resolve di `dk_project_gb_project`.
- [x] Jika DK sudah final dan update dibatasi, pertahankan aturan final yang sudah ada.
- [x] Response DK detail menampilkan snapshot yang tersimpan, bukan latest dynamic.
- [x] Jika snapshot tersimpan punya newer revision, response boleh membawa `has_newer_revision` untuk UI.

Acceptance:

- [x] DK baru selalu mengambil latest GB saat dipilih.
- [x] DK lama tidak berubah setelah revisi GB dibuat.

---

## Step 3 - Lender Validation

Files:

- `prism-backend/internal/service/daftar_kegiatan_service.go`
- `prism-backend/internal/service/loan_agreement_service.go`
- `prism-backend/sql/queries/dk_project.sql`
- `prism-backend/sql/queries/loan_agreement.sql`

Checklist:

- [x] DK financing lender validation memakai concrete GB Project yang tersimpan.
- [x] DK lender validation tetap mengambil lender dari:
  - [x] `gb_funding_source` pada concrete GB Project.
  - [x] `lender_indication` pada concrete BB Project yang terhubung ke concrete GB Project.
- [x] LA lender validation tetap memakai `dk_financing_detail` dari DK Project terkait.
- [x] Jangan resolve latest saat membuat LA atau monitoring.
- [x] Pastikan validation error tetap memakai format aman dari error handling repo.

Acceptance:

- [x] Revisi BB/GB setelah DK dibuat tidak mengubah allowed lender untuk DK/LA tersebut.

---

## Step 4 - Tests dan Verification

Checklist:

- [ ] Tambah/update test DK create dengan input GB lama tetapi identity punya snapshot baru.
- [x] Verify junction tersimpan menunjuk latest GB snapshot saat create.
- [ ] Tambah/update test DK tetap menunjuk snapshot lama setelah GB revisi baru dibuat.
- [ ] Tambah/update test lender validation memakai concrete snapshot.
- [ ] Tambah/update test LA tetap mengikuti DK Project yang tersimpan.
- [x] Jalankan `go test ./...`.
- [x] Jalankan smoke API:
  - [x] create BB original + revision.
  - [x] create GB original + revision.
  - [x] create DK dengan input GB identity/snapshot lama.
  - [x] verify DK stored relation adalah latest saat create.
  - [x] create GB revision baru.
  - [x] verify DK relation tidak berubah.

Done Criteria:

- [x] DK/LA downstream freeze rule terjaga.
- [x] Semua write baru memakai latest resolver hanya pada saat link dibuat.
