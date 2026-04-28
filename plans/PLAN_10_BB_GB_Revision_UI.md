# PLAN 10 - BB/GB Revision History UI

> Scope: frontend untuk history revision, latest picker, concrete downstream snapshot display, dan journey newer indicator.
> Deliverable: UI bisa menampilkan histori BB/GB project, picker memakai latest snapshot, dan DK/Journey tetap menunjukkan versi yang benar.
> Referensi: `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`, `docs/PRISM_API_Contract.md`, `docs/PRISM_Frontend_Structure.md`, `plans/PLAN_BE_11_Journey_Import_Project_List_Versioning.md`.

---

## Prasyarat

- [ ] Backend BE-07 sampai BE-11 selesai.
- [ ] `docs/PRISM_API_Contract.md` sudah memuat field dan endpoint versioning.
- [ ] Frontend base plans FE-00 sampai FE-09 sudah selesai atau modul terkait sudah ada.

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

- [ ] Tambahkan field BB:
  - [ ] `project_identity_id`.
  - [ ] `is_latest`.
  - [ ] `has_newer_revision`.
- [ ] Tambahkan field GB:
  - [ ] `gb_project_identity_id`.
  - [ ] `is_latest`.
  - [ ] `has_newer_revision`.
- [ ] Tambahkan type history item BB/GB.
- [ ] Tambahkan service:
  - [ ] `getBBProjectHistory(id)`.
  - [ ] `getGBProjectHistory(id)`.
- [ ] Tambahkan type journey identity/newer indicator.
- [ ] Jangan definisikan interface di file `.vue`.

Acceptance:

- [ ] TypeScript contract frontend selaras dengan API versioning.

---

## Step 2 - Stores dan Data Flow

Files:

- `prism-frontend/src/stores/blue-book.store.ts`
- `prism-frontend/src/stores/green-book.store.ts`
- `prism-frontend/src/stores/daftar-kegiatan.store.ts`
- store dashboard/journey jika ada.

Checklist:

- [ ] Tambahkan state history BB Project.
- [ ] Tambahkan state history GB Project.
- [ ] Tambahkan actions load history.
- [ ] Tambahkan support query latest/default untuk picker jika API memakai parameter.
- [ ] Pastikan page memanggil store/service, bukan axios langsung.
- [ ] Pastikan loading/error state konsisten dengan pola existing.

Acceptance:

- [ ] History dan picker data bisa di-load tanpa bypass service layer.

---

## Step 3 - Blue Book UI

Files:

- `prism-frontend/src/pages/blue-book/BBProjectDetailPage.vue`
- Komponen Blue Book terkait jika ada.

Checklist:

- [ ] Tambahkan section atau tab "Histori Revisi".
- [ ] Tampilkan daftar snapshot:
  - [ ] Blue Book period/revision.
  - [ ] `bb_code`.
  - [ ] status book.
  - [ ] latest badge.
  - [ ] used by downstream indicator jika tersedia.
- [ ] Jika current snapshot bukan latest, tampilkan badge "Ada revisi lebih baru".
- [ ] Action view detail menuju concrete snapshot id.
- [ ] Jangan membuat tampilan history sebagai popup jika halaman detail sudah punya ruang.

Acceptance:

- [ ] User bisa memahami project ini muncul di revisi BB mana saja.

---

## Step 4 - Green Book UI

Files:

- `prism-frontend/src/pages/green-book/GBProjectFormPage.vue`
- `prism-frontend/src/pages/green-book/GBProjectDetailPage.vue`
- Komponen picker BB Project jika ada.

Checklist:

- [ ] Picker BB Project hanya menampilkan latest BB Project per identity secara default.
- [ ] Jika API menyediakan historical mode, jangan aktifkan di picker create default.
- [ ] Detail GB Project menampilkan concrete BB snapshots yang tersimpan.
- [ ] Tambahkan section atau tab "Histori Revisi" untuk GB Project.
- [ ] Jika current GB snapshot bukan latest, tampilkan badge "Ada revisi lebih baru".

Acceptance:

- [ ] GB create/edit tidak memilih snapshot BB lama secara tidak sengaja.
- [ ] Detail GB tetap transparan soal concrete BB version yang dipakai.

---

## Step 5 - Daftar Kegiatan UI

Files:

- `prism-frontend/src/pages/daftar-kegiatan/DKProjectFormPage.vue`
- `prism-frontend/src/pages/daftar-kegiatan/DKDetailPage.vue`
- `prism-frontend/src/composables/forms/useDKProjectForm.ts`

Checklist:

- [ ] Picker GB Project hanya menampilkan latest GB Project per identity secara default.
- [ ] Saat DK tersimpan, detail menampilkan concrete GB snapshots yang dipakai.
- [ ] Jika concrete snapshot punya newer revision, tampilkan badge informatif.
- [ ] Jangan auto-rewrite pilihan DK di frontend setelah revisi muncul.
- [ ] Allowed lender UI tetap berdasarkan data dari backend, bukan dihitung manual dari latest dynamic.

Acceptance:

- [ ] DK create memakai latest, tetapi DK detail menunjukkan frozen snapshot.

---

## Step 6 - Journey dan Project List UI

Files:

- `prism-frontend/src/pages/dashboard/ProjectJourneyPage.vue`
- `prism-frontend/src/components/dashboard/ProjectTimeline.vue`
- Project list page jika ada.

Checklist:

- [ ] Journey menampilkan concrete path BB -> GB -> DK -> LA -> Monitoring.
- [ ] Tampilkan badge "Ada revisi lebih baru" pada BB/GB node jika API mengirim indicator.
- [ ] Jangan mengganti path journey ke latest jika downstream menyimpan versi lama.
- [ ] Project list default memakai latest snapshots.
- [ ] Jika ada toggle historical view, label harus jelas dan tidak aktif default.

Acceptance:

- [ ] Journey membedakan versi historis dan revisi terbaru secara eksplisit.

---

## Step 7 - Verification

Checklist:

- [ ] Jalankan typecheck/build frontend sesuai script repo.
- [ ] Jalankan lint jika tersedia.
- [ ] Smoke UI:
  - [ ] BB detail history tampil.
  - [ ] GB picker hanya latest.
  - [ ] GB detail history tampil.
  - [ ] DK picker hanya latest.
  - [ ] DK detail tetap concrete snapshot.
  - [ ] Journey badge newer revision tampil.
- [ ] Jika melakukan perubahan visual besar, verifikasi di browser lokal.

Done Criteria:

- [ ] Frontend selaras dengan contract versioning.
- [ ] Tidak ada direct axios di page/component.
- [ ] Tidak ada interface baru di `.vue`.
- [ ] Build frontend lulus.
