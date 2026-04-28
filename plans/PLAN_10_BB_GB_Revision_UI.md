# PLAN 10 - BB/GB Revision History UI

> Scope: frontend untuk history revision, latest picker, concrete downstream snapshot display, dan journey newer indicator.
> Deliverable: UI bisa menampilkan histori BB/GB project, picker memakai latest snapshot, dan DK/Journey tetap menunjukkan versi yang benar.
> Referensi: `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`, `docs/PRISM_API_Contract.md`, `docs/PRISM_Frontend_Structure.md`, `plans/PLAN_BE_11_Journey_Import_Project_List_Versioning.md`.

---

## Prasyarat

- [x] Backend BE-07 sampai BE-11 selesai.
- [x] `docs/PRISM_API_Contract.md` sudah memuat field dan endpoint versioning.
- [x] Frontend base plans FE-00 sampai FE-09 sudah selesai atau modul terkait sudah ada.

---

## Step 1 - Types dan Services

Files:

- `prism-frontend/src/types/blue-book.types.ts`
- `prism-frontend/src/types/green-book.types.ts`
- `prism-frontend/src/types/daftar-kegiatan.types.ts`
- `prism-frontend/src/types/dashboard.types.ts`
- `prism-frontend/src/services/blue-book.service.ts`
- `prism-frontend/src/services/green-book.service.ts`
- `prism-frontend/src/services/dashboard.service.ts`

Checklist:

- [x] Tambahkan field BB:
  - [x] `project_identity_id`.
  - [x] `is_latest`.
  - [x] `has_newer_revision`.
- [x] Tambahkan field GB:
  - [x] `gb_project_identity_id`.
  - [x] `is_latest`.
  - [x] `has_newer_revision`.
- [x] Tambahkan type history item BB/GB.
- [x] Tambahkan service:
  - [x] `getBBProjectHistory(id)`.
  - [x] `getGBProjectHistory(id)`.
- [x] Tambahkan type journey identity/newer indicator.
- [x] Jangan definisikan interface di file `.vue`.

Acceptance:

- [x] TypeScript contract frontend selaras dengan API versioning.

---

## Step 2 - Stores dan Data Flow

Files:

- `prism-frontend/src/stores/blue-book.store.ts`
- `prism-frontend/src/stores/green-book.store.ts`
- `prism-frontend/src/stores/daftar-kegiatan.store.ts`
- store dashboard/journey jika ada.

Checklist:

- [x] Tambahkan state history BB Project.
- [x] Tambahkan state history GB Project.
- [x] Tambahkan actions load history.
- [x] Tambahkan support query latest/default untuk picker jika API memakai parameter.
- [x] Pastikan page memanggil store/service, bukan axios langsung.
- [x] Pastikan loading/error state konsisten dengan pola existing.

Acceptance:

- [x] History dan picker data bisa di-load tanpa bypass service layer.

---

## Step 3 - Blue Book UI

Files:

- `prism-frontend/src/pages/blue-book/BBProjectDetailPage.vue`
- Komponen Blue Book terkait jika ada.

Checklist:

- [x] Tambahkan section atau tab "Histori Revisi".
- [x] Tampilkan daftar snapshot:
  - [x] Blue Book period/revision.
  - [x] `bb_code`.
  - [x] status book.
  - [x] latest badge.
  - [x] used by downstream indicator jika tersedia.
- [x] Jika current snapshot bukan latest, tampilkan badge "Ada revisi lebih baru".
- [x] Action view detail menuju concrete snapshot id.
- [x] Jangan membuat tampilan history sebagai popup jika halaman detail sudah punya ruang.

Acceptance:

- [x] User bisa memahami project ini muncul di revisi BB mana saja.

---

## Step 4 - Green Book UI

Files:

- `prism-frontend/src/pages/green-book/GBProjectFormPage.vue`
- `prism-frontend/src/pages/green-book/GBProjectDetailPage.vue`
- Komponen picker BB Project jika ada.

Checklist:

- [x] Picker BB Project hanya menampilkan latest BB Project per identity secara default.
- [x] Jika API menyediakan historical mode, jangan aktifkan di picker create default.
- [x] Detail GB Project menampilkan concrete BB snapshots yang tersimpan.
- [x] Tambahkan section atau tab "Histori Revisi" untuk GB Project.
- [x] Jika current GB snapshot bukan latest, tampilkan badge "Ada revisi lebih baru".

Acceptance:

- [x] GB create/edit tidak memilih snapshot BB lama secara tidak sengaja.
- [x] Detail GB tetap transparan soal concrete BB version yang dipakai.

---

## Step 5 - Daftar Kegiatan UI

Files:

- `prism-frontend/src/pages/daftar-kegiatan/DKProjectFormPage.vue`
- `prism-frontend/src/pages/daftar-kegiatan/DKDetailPage.vue`
- `prism-frontend/src/composables/forms/useDKProjectForm.ts`

Checklist:

- [x] Picker GB Project hanya menampilkan latest GB Project per identity secara default.
- [x] Saat DK tersimpan, detail menampilkan concrete GB snapshots yang dipakai.
- [x] Jika concrete snapshot punya newer revision, tampilkan badge informatif.
- [x] Jangan auto-rewrite pilihan DK di frontend setelah revisi muncul.
- [x] Allowed lender UI tetap berdasarkan data dari backend, bukan dihitung manual dari latest dynamic.

Acceptance:

- [x] DK create memakai latest, tetapi DK detail menunjukkan frozen snapshot.

---

## Step 6 - Journey dan Project List UI

Files:

- `prism-frontend/src/pages/dashboard/ProjectJourneyPage.vue`
- `prism-frontend/src/components/dashboard/ProjectTimeline.vue`
- Project list page jika ada.

Checklist:

- [x] Journey menampilkan concrete path BB -> GB -> DK -> LA -> Monitoring.
- [x] Tampilkan badge "Ada revisi lebih baru" pada BB/GB node jika API mengirim indicator.
- [x] Jangan mengganti path journey ke latest jika downstream menyimpan versi lama.
- [x] Project list default memakai latest snapshots.
- [x] Jika ada toggle historical view, label harus jelas dan tidak aktif default.

Acceptance:

- [x] Journey membedakan versi historis dan revisi terbaru secara eksplisit.

---

## Step 7 - Verification

Checklist:

- [x] Jalankan typecheck/build frontend sesuai script repo.
- [x] Jalankan lint jika tersedia.
- [ ] Smoke UI:
  - [ ] BB detail history tampil.
  - [ ] GB picker hanya latest.
  - [ ] GB detail history tampil.
  - [ ] DK picker hanya latest.
  - [ ] DK detail tetap concrete snapshot.
  - [ ] Journey badge newer revision tampil.
- [ ] Jika melakukan perubahan visual besar, verifikasi di browser lokal.

Done Criteria:

- [x] Frontend selaras dengan contract versioning.
- [x] Tidak ada direct axios di page/component.
- [x] Tidak ada interface baru di `.vue`.
- [x] Build frontend lulus.
