# PRISM BB/GB Revision Versioning Plan

> Scope: rencana implementasi versioning project untuk Blue Book dan Green Book.
> Status: planning document. Belum mengubah DDL/API aktif atau kode aplikasi.

---

## 1. Tujuan

Blue Book dan Green Book dapat direvisi. Dalam revisi, project bisa muncul kembali dengan kode yang sama dan isi yang sama persis dengan versi sebelumnya. Sistem harus bisa:

- Menyimpan setiap project sebagai snapshot di dokumen/revisi tertentu.
- Mengetahui logical project yang sama lintas revisi.
- Menampilkan histori project muncul di Blue Book dan Green Book revisi mana saja.
- Memastikan link baru memakai versi terbaru.
- Memastikan downstream yang sudah masuk DK/LA tetap menunjuk ke versi lama yang dipilih saat itu.

---

## 2. Prinsip Desain

1. `bb_project` dan `gb_project` adalah snapshot, bukan identitas logical tunggal.
2. Logical identity disimpan terpisah agar project yang sama bisa muncul di banyak revisi.
3. `bb_code` boleh sama lintas revisi Blue Book, tetapi tidak boleh duplikat dalam Blue Book yang sama.
4. `gb_code` boleh sama lintas revisi Green Book, tetapi tidak boleh duplikat dalam Green Book yang sama.
5. Green Book baru/revisi harus resolve ke versi terbaru BB Project.
6. DK baru harus resolve ke versi terbaru GB Project, dan secara tidak langsung memakai BB Project versi yang tersimpan pada GB tersebut.
7. Setelah DK/LA dibuat, relasi downstream tidak auto-pindah ketika BB/GB direvisi.
8. Karena data fresh dan bisa digenerate ulang, tidak perlu backfill data lama. Tetap gunakan DDL/migration yang jelas agar schema target terdokumentasi.

---

## 2.1 Execution Phase Files

Gunakan plan di folder `plans/` berikut untuk implementasi step by step:

| Urutan | File | Fokus |
|--------|------|-------|
| 1 | `plans/PLAN_BE_07_Revision_Versioning_Schema.md` | Schema, DDL, migration, sqlc queries, latest/history resolver, API contract awal |
| 2 | `plans/PLAN_BE_08_Blue_Book_Revision_Versioning.md` | Blue Book logical identity, duplicate per dokumen, clone revisi, BB history endpoint, import BB |
| 3 | `plans/PLAN_BE_09_Green_Book_Revision_Versioning.md` | Green Book logical identity, latest BB resolver, clone revisi, GB history endpoint |
| 4 | `plans/PLAN_BE_10_DK_LA_Frozen_Snapshot.md` | DK latest GB resolver, downstream frozen snapshot, lender validation berdasarkan concrete version |
| 5 | `plans/PLAN_BE_11_Journey_Import_Project_List_Versioning.md` | Project list latest default, journey concrete path, aggregate safety, import final, smoke backend |
| 6 | `plans/PLAN_10_BB_GB_Revision_UI.md` | Frontend history UI, latest pickers, DK/Journey concrete snapshot display |

Setiap file plan memiliki checklist per step. Kerjakan berurutan; jangan lompat ke fase frontend sebelum backend contract dan smoke BE-11 selesai.

---

## 3. Target Schema

### 3.1 Logical Blue Book Project

Tambahkan tabel:

```sql
CREATE TABLE project_identity (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Tambahkan ke `bb_project`:

```sql
project_identity_id UUID NOT NULL REFERENCES project_identity(id)
```

Ubah constraint:

```sql
-- lama
UNIQUE (bb_code)

-- baru
UNIQUE (blue_book_id, bb_code)
```

Index yang dibutuhkan:

```sql
CREATE INDEX idx_bb_project_identity ON bb_project(project_identity_id);
CREATE INDEX idx_bb_project_book_code ON bb_project(blue_book_id, bb_code);
```

### 3.2 Logical Green Book Project

Tambahkan tabel:

```sql
CREATE TABLE gb_project_identity (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Tambahkan ke `gb_project`:

```sql
gb_project_identity_id UUID NOT NULL REFERENCES gb_project_identity(id)
```

Ubah constraint:

```sql
-- lama
UNIQUE (gb_code)

-- baru
UNIQUE (green_book_id, gb_code)
```

Index yang dibutuhkan:

```sql
CREATE INDEX idx_gb_project_identity ON gb_project(gb_project_identity_id);
CREATE INDEX idx_gb_project_book_code ON gb_project(green_book_id, gb_code);
```

### 3.3 Header Revision Lineage

Header `blue_book` dan `green_book` sudah punya `revision_number`, `revision_year` atau `publish_year`, dan `status`. Untuk cloning yang eksplisit, tambahkan field opsional:

```sql
replaces_blue_book_id UUID REFERENCES blue_book(id)
replaces_green_book_id UUID REFERENCES green_book(id)
```

Jika field ini tidak ditambahkan, source revisi tetap bisa ditentukan dari header `Berlaku` terakhir pada period/publish_year yang sama.

---

## 4. Backend Plan

### Phase BE-1: Schema dan sqlc

Files:

- `docs/prism_ddl.sql`
- `prism-backend/sql/queries/bb_project.sql`
- `prism-backend/sql/queries/gb_project.sql`
- `prism-backend/sql/queries/dk_project.sql`
- `prism-backend/sql/queries/monitoring.sql` atau file journey query terkait

Tasks:

1. Tambah tabel identity dan foreign key identity ke `bb_project` dan `gb_project`.
2. Ubah uniqueness code menjadi per dokumen:
   - `UNIQUE (blue_book_id, bb_code)`
   - `UNIQUE (green_book_id, gb_code)`
3. Tambah query create/get identity.
4. Tambah query latest resolver:
   - latest BB Project by `project_identity_id`
   - latest BB Project by current `bb_project_id`
   - latest GB Project by `gb_project_identity_id`
   - latest GB Project by current `gb_project_id`
5. Tambah query history:
   - list BB Project snapshots by `project_identity_id`
   - list GB Project snapshots by `gb_project_identity_id`
6. Jalankan `make generate` setelah SQL selesai.

### Phase BE-2: Blue Book Service

Files:

- `prism-backend/internal/model/blue_book.go`
- `prism-backend/internal/service/blue_book_service.go`
- `prism-backend/internal/handler/blue_book_handler.go`

Tasks:

1. Create BB Project:
   - Jika request membawa `project_identity_id`, pakai identity itu.
   - Jika tidak, create `project_identity` baru.
   - Validasi duplicate hanya dalam `blue_book_id` yang sama.
2. Create Blue Book revision:
   - Create header baru dengan status eksplisit pilihan user (`Berlaku`/`Tidak Berlaku`).
   - Tidak mengubah status header lama secara otomatis.
   - Clone project pilihan user dari source Blue Book periode yang sama.
   - Clone children: institutions, locations, national priorities, project costs, lender indications, dan LoI.
   - Setiap clone memakai `project_identity_id` yang sama dengan snapshot asal.
3. Tambah endpoint/history service:
   - `GET /bb-projects/:id/history`
   - Response berisi semua snapshot dari identity yang sama, urut revision.

### Phase BE-3: Green Book Service

Files:

- `prism-backend/internal/model/green_book.go`
- `prism-backend/internal/service/green_book_service.go`
- `prism-backend/internal/handler/green_book_handler.go`

Tasks:

1. Create GB Project:
   - Jika request membawa `gb_project_identity_id`, pakai identity itu.
   - Jika tidak, create `gb_project_identity` baru.
   - Validasi duplicate hanya dalam `green_book_id` yang sama.
2. Saat menyimpan relasi `gb_project_bb_project`:
   - Resolve setiap input BB Project ke latest BB Project dari identity yang sama.
   - Simpan concrete `bb_project_id` hasil resolve.
3. Create Green Book revision:
   - Supersede header lama pada `publish_year` yang sama.
   - Create header baru.
   - Clone GB Project dan semua children.
   - Clone activities, funding source, disbursement plan, dan funding allocation.
   - Untuk relasi BB, simpan latest BB Project saat revisi GB dibuat.
   - Setiap clone memakai `gb_project_identity_id` yang sama dengan snapshot asal.
4. Tambah endpoint/history service:
   - `GET /gb-projects/:id/history`
   - Response berisi semua snapshot dari identity yang sama, urut revision.

### Phase BE-4: DK/LA Resolver dan Freeze Rule

Files:

- `prism-backend/sql/queries/dk_project.sql`
- `prism-backend/internal/service/daftar_kegiatan_service.go`
- `prism-backend/internal/service/loan_agreement_service.go`

Tasks:

1. Saat DK dibuat atau project GB dipilih:
   - Resolve input GB Project ke latest GB Project dari identity yang sama.
   - Simpan concrete `gb_project_id` hasil resolve di `dk_project_gb_project`.
2. Jangan update `dk_project_gb_project` otomatis saat GB/BB direvisi.
3. LA tetap mengikuti DK Project yang sudah tersimpan.
4. Lender validation DK/LA harus memakai concrete version yang tersimpan, bukan latest version saat request LA/monitoring dibuat.

### Phase BE-5: Journey, Aggregate Safety, dan Import

Files:

- `prism-backend/sql/queries/project.sql`
- `prism-backend/sql/queries/monitoring.sql`
- `prism-backend/internal/service/journey_service.go`
- `prism-backend/internal/service/blue_book_import_service.go`

Tasks:

1. Journey menampilkan concrete path yang dipakai downstream.
2. Journey/detail memberi indikator jika ada BB/GB snapshot yang lebih baru untuk identity yang sama.
3. Project list bisa filter default ke latest snapshot agar daftar tidak berisi duplikasi revisi, dengan opsi melihat histori.
4. Import Blue Book:
   - Duplicate `BB Code` dalam workbook tetap error.
   - `BB Code` yang sudah ada di revisi lama tidak di-skip.
   - Jika import ke revisi baru dan kode ditemukan pada revisi sebelumnya, reuse `project_identity_id`.
5. Import Green Book mengikuti aturan yang sama untuk `GB Code`.

---

## 5. API Contract Plan

Tambahkan field response BB Project:

```json
{
  "id": "uuid-bb-project-snapshot",
  "project_identity_id": "uuid-logical-project",
  "blue_book_id": "uuid",
  "bb_code": "BB-2025-001",
  "is_latest": true,
  "has_newer_revision": false
}
```

Tambahkan field response GB Project:

```json
{
  "id": "uuid-gb-project-snapshot",
  "gb_project_identity_id": "uuid-logical-gb-project",
  "green_book_id": "uuid",
  "gb_code": "GB-2025-001",
  "is_latest": true,
  "has_newer_revision": false
}
```

Endpoint baru:

| Method | Endpoint | Keterangan |
|--------|----------|------------|
| `GET` | `/bb-projects/:id/history` | Semua snapshot BB untuk logical project yang sama |
| `GET` | `/gb-projects/:id/history` | Semua snapshot GB untuk logical GB project yang sama |

History response minimal:

```json
{
  "data": [
    {
      "id": "uuid",
      "book_id": "uuid",
      "code": "BB-2025-001",
      "book_label": "BB 2025-2029",
      "revision_number": 0,
      "revision_year": null,
      "book_status": "superseded",
      "is_latest": false,
      "used_by_downstream": true
    }
  ]
}
```

---

## 6. Frontend Plan

Affected plans:

- `plans/PLAN_04_Blue_Book.md`
- `plans/PLAN_05_Green_Book.md`
- `plans/PLAN_06_Daftar_Kegiatan.md`
- `plans/PLAN_09_Project_Journey.md`

Tasks:

1. Update types:
   - `project_identity_id`
   - `gb_project_identity_id`
   - `is_latest`
   - `has_newer_revision`
   - history item types
2. Blue Book:
   - Detail BB Project tampilkan tab/section "Histori Revisi".
   - List default menampilkan snapshot pada Blue Book yang sedang dibuka.
3. Green Book:
   - Picker BB Project hanya menampilkan latest BB Project per identity.
   - Detail GB Project tampilkan history lintas revisi.
4. Daftar Kegiatan:
   - Picker GB Project hanya menampilkan latest GB Project per identity.
   - Setelah tersimpan, detail DK menampilkan versi snapshot yang dipakai.
5. Journey:
   - Tampilkan concrete snapshot path.
   - Jika ada versi baru, tampilkan badge "Ada revisi lebih baru" tanpa mengubah path historis.

---

## 7. Test Plan

Backend tests:

1. Bisa create BB revisi dengan `bb_code` sama.
2. Duplicate `bb_code` dalam Blue Book yang sama tetap conflict.
3. BB revision clone mempertahankan `project_identity_id`.
4. GB create/revision resolve BB Project ke latest version.
5. Bisa create GB revisi dengan `gb_code` sama.
6. Duplicate `gb_code` dalam Green Book yang sama tetap conflict.
7. DK create resolve GB Project ke latest version.
8. DK/LA tetap menunjuk concrete GB/BB version lama setelah revisi baru dibuat.
9. Lender validation DK/LA memakai concrete version yang tersimpan.
10. Journey menampilkan historical path dan indikator newer revision.

Frontend tests:

1. BB/GB picker tidak menampilkan duplikasi snapshot lama secara default.
2. History section tampil dengan urutan revisi yang benar.
3. DK detail tetap menampilkan snapshot yang dipakai saat create.
4. Journey menampilkan badge newer revision tanpa mengubah downstream path.

Smoke tests:

1. `make generate`
2. `go test ./...`
3. Frontend typecheck/build.
4. Live API smoke:
   - create BB original
   - create BB revision
   - create GB using latest BB
   - create GB revision
   - create DK using latest GB
   - create another revision
   - verify DK/LA path stays on stored version

---

## 8. Acceptance Criteria

- `bb_code` dan `gb_code` bisa sama lintas revisi, tetapi tidak duplikat dalam dokumen yang sama.
- BB/GB revisions clone project snapshots dengan identity yang sama.
- Green Book baru/revisi selalu memakai latest BB Project saat link dibuat.
- DK baru selalu memakai latest GB Project saat link dibuat.
- DK/LA yang sudah dibuat tidak berubah ketika revisi baru muncul.
- Project detail dan journey bisa menunjukkan histori revisi dan concrete version yang dipakai downstream.
- Import workbook tidak lagi skip kode yang hanya ada di revisi lama.

---

## 9. Risiko dan Keputusan yang Perlu Dijaga

- Jangan hanya melepas unique global tanpa identity table; itu membuat history ambigu.
- Jangan auto-update downstream relation setelah DK/LA dibuat; audit path harus stabil.
- Jangan menyamakan "latest visible" dengan "historical stored relation"; list/picker boleh latest, tetapi detail/journey harus memakai concrete IDs.
- Jika LoI tidak ingin dicopy saat revisi BB, ubah Phase BE-2 sebelum implementasi. Default plan ini meng-copy LoI agar snapshot revisi self-contained.
