# PRISM — Project Loan Integrated Monitoring System
> *Dokumen Ide Awal (Initial Concept Document)*

---

## 1. Latar Belakang

Pengelolaan pinjaman dan hibah luar negeri saat ini menghadapi beberapa tantangan utama:

- **Tidak adanya single source of truth** — data perencanaan pinjaman dan hibah luar negeri tersebar di berbagai unit dan tidak terpusat.
- **Tidak ada dashboard berkualitas** — tidak tersedia visualisasi yang memadai untuk memantau progres dan monitoring pinjaman luar negeri secara real-time.
- **Inefisiensi pengumpulan data** — setiap kali diperlukan analisis, data harus dikumpulkan ulang dari berbagai sumber, memakan waktu dan berpotensi menimbulkan inkonsistensi.

**PRISM** hadir sebagai solusi terpadu untuk menjawab kebutuhan tersebut melalui sistem monitoring dan perencanaan pinjaman luar negeri yang terintegrasi, andal, dan berbasis data.

---

## 2. Pengguna Sistem

Sistem memiliki **2 jenis pengguna** dengan hak akses yang berbeda:

| Role | Hak Akses |
|------|-----------|
| **ADMIN** | Full access ke semua modul (Create, Read, Update, Delete) |
| **STAFF** | Akses CRUD per modul, dikonfigurasi oleh ADMIN — setiap kombinasi modul × operasi (C/R/U/D) dapat diaktifkan atau dinonaktifkan secara independen |

---

## 3. Proses Bisnis

### 3.1 Alur Perencanaan Pinjaman Luar Negeri

Perencanaan pinjaman luar negeri melewati tiga tahapan utama yang berurutan:

```
[Blue Book] ──(LoI dari Lender)──► [Green Book] ──► [Daftar Kegiatan]
```

---

### 3.2 Blue Book (DRPLN-JM)

**Daftar Rencana Pinjaman Luar Negeri Jangka Menengah** — diterbitkan oleh Bappenas setiap **5 tahun sekali** (dapat direvisi).

Memuat kumpulan daftar proyek beserta digest-nya. Pada tahap Blue Book, lender yang tertarik memberikan **indikasi** ketertarikannya — disebut **Lender Indication** — yang bersifat belum pasti. Apabila indikasi berlanjut, lender menerbitkan **Letter of Intent (LoI)** dan proyek dapat diusulkan masuk ke Green Book.

Setiap proyek di Blue Book juga sudah memiliki **unit kerja Bappenas yang bertanggung jawab** (Eselon I dan Eselon II).

#### Atribut Blue Book Project

| Field | Keterangan |
|-------|-----------|
| Blue Book Code | Kode unik proyek |
| Project | Nama proyek |
| Program Title | Dipilih dari entitas master Program Title (shared dengan GB) |
| Executing Agency | Multi-select — satu atau lebih Institution (Kementerian/Eselon I) |
| Implementing Agency | Multi-select — satu atau lebih Institution (Kementerian/Eselon I) |
| Duration | Durasi proyek |
| Location | Multi-select wilayah (Nasional / Provinsi / Kota-Kabupaten); jika Nasional otomatis mencakup seluruh provinsi |
| Objective | Tujuan proyek |
| Scope of Work | Ruang lingkup pekerjaan |
| Outputs | Keluaran proyek |
| Outcomes | Hasil proyek |
| National Priority | Multi-select — satu atau lebih National Priority yang didukung proyek ini |
| Bappenas Partner (Eselon II) | Unit kerja Bappenas penanggung jawab — Eselon I diturunkan otomatis dari hierarki |
| **Lender Indication** | Lender yang memberikan indikasi ketertarikan *(masih bersifat indikasi, belum pasti)* — memuat: Nama Lender + Keterangan |

#### Project Cost

| Sumber Dana | Jenis | Satuan |
|-------------|-------|--------|
| **Foreign Funding** | Loan / Grant | USD |
| **Counterpart Funding** | Central Government / Regional Government / State-Owned Enterprise / Others | USD |

---

### 3.3 Green Book

Diterbitkan oleh **Bappenas** setiap **1 tahun sekali** (dapat direvisi), atas usulan dari Executing Agency. Proyek masuk Green Book setelah adanya LoI dari lender.

> **Catatan Relasi:**
> - Satu proyek Blue Book dapat menghasilkan **lebih dari satu proyek Green Book** (one-to-many, umum terjadi).
> - Satu proyek Green Book dapat mereferensikan **lebih dari satu Blue Book Code** (many-to-many, sangat jarang).
> - Relasi ini bersifat **many-to-many** dan perlu ditangani dengan tabel penghubung di level database.

#### Atribut Green Book Project

| Field | Keterangan |
|-------|-----------|
| Green Book Code | Kode unik Green Book |
| Project | Nama proyek |
| Program Title | Dipilih dari entitas master Program Title (shared dengan BB) |
| Executing Agency | Multi-select — satu atau lebih Institution (Kementerian/Eselon I) |
| Implementing Agency | Multi-select — satu atau lebih Institution (Kementerian/Eselon I) |
| Duration | Durasi proyek |
| Location | Multi-select wilayah (Nasional / Provinsi / Kota-Kabupaten); jika Nasional otomatis mencakup seluruh provinsi |
| Objective | Tujuan |
| Scope of Project | Ruang lingkup proyek |
| Program Reference | Blue Book Code terkait |

#### Tabel Detail dalam Green Book

**Activities**

*(Implementation Location diinput sebagai teks bebas, bukan mengacu ke entitas master Wilayah)*

| Activities | Implementation Location | Project Implementation Units |
|-----------|------------------------|------------------------------|
| ... | ... | ... |

**Funding Source**

*(Satu GB project dapat memiliki lebih dari satu lender — disebut **cofinancing**. Setiap baris mewakili satu lender beserta nominal kontribusinya masing-masing.)*

| Implementing Agency | Loan (USD) | Grant (USD) | Local (USD) | Source (Lender) |
|--------------------|-----------|------------|------------|----------------|
| ... | ... | ... | ... | Lender A |
| ... | ... | ... | ... | Lender B |

**Disbursement Plan (USD)**

*(Total keseluruhan proyek — bukan per lender)*

| Year | Amount |
|------|--------|
| ... | ... |

**Funding Allocation**

*(Kolom Activities mengacu ke baris Activities yang sudah diinput di tabel Activities GB di atas — bukan diinput independen)*

| Activities | Services | Constructions | Goods | Trainings | Other |
|-----------|---------|--------------|-------|-----------|-------|
| ... | ... | ... | ... | ... | ... |

---

### 3.4 Daftar Kegiatan

Surat yang diterbitkan oleh **Bappenas** atas usulan dari Executing Agency, sebagai kelanjutan dari proses Green Book. Satu Daftar Kegiatan (surat) dapat memuat **lebih dari satu proyek Green Book**.

> **Catatan Multi-currency:** Amount dapat berupa USD atau mata uang lender. Apabila dalam mata uang lender, nilai USD dan mata uang yang digunakan dicantumkan bersamaan. **Konversi ke USD dilakukan secara manual oleh Staff** saat input data.

> **Catatan Status:** Daftar Kegiatan bersifat **final** saat diterbitkan — tidak ada mekanisme revisi.

#### Atribut Daftar Kegiatan

| Field | Keterangan |
|-------|-----------|
| Program Title | Judul program |
| Program Reference | Satu atau lebih Green Book Code (dan Blue Book Code terkait) |
| Duration | Durasi |
| Funding Source | Dipilih dari Lender yang sudah terdaftar — bisa dari Lender Indication (BB) maupun Funding Source (GB) |
| Objectives | Tujuan |
| Location | Multi-select wilayah (Nasional / Provinsi / Kota-Kabupaten); jika Nasional otomatis mencakup seluruh provinsi |
| Executing Agency | Dipilih dari entitas Institution |

#### Financing Detail

| Source | Amount (USD) | Grant (USD) | Counterpart Fund (USD) | Remarks |
|--------|-------------|-------------|----------------------|---------|
| ... | ... | ... | ... | ... |

#### Loan Allocation

| Executing Agency | Amount (USD) | Grant (USD) | Counterpart Fund (USD) | Remarks |
|-----------------|-------------|-------------|----------------------|---------|
| ... | ... | ... | ... | ... |

#### Activity Details

Daftar komponen kegiatan — hanya memuat **nama aktivitas**, diinput secara bebas (tidak dipilih dari Activities GB). Secara konteks merupakan realisasi/konfirmasi dari Activities yang direncanakan di Green Book, namun tidak ada relasi teknis langsung.

| No | Nama Aktivitas / Komponen Kegiatan |
|----|-----------------------------------|
| 1 | ... |
| 2 | ... |

---

### 3.5 Loan Agreement (LA)

Perjanjian pinjaman yang diterbitkan setelah Daftar Kegiatan. Setiap proyek dalam Daftar Kegiatan menghasilkan **tepat satu Loan Agreement** (relasi One-to-One dengan proyek di DK).

> **Contoh:** Daftar Kegiatan memuat 3 proyek → akan ada 3 Loan Agreement terpisah, masing-masing merujuk ke satu proyek.

#### Atribut Loan Agreement

| Field | Keterangan |
|-------|-----------|
| Kode Loan | Kode unik Loan Agreement |
| Tanggal Agreement | Tanggal penandatanganan perjanjian |
| Tanggal Efektif | Tanggal perjanjian mulai berlaku |
| Original Closing Date | Tanggal berakhir perjanjian awal *(sebelum perpanjangan)* |
| Closing Date | Tanggal berakhir perjanjian terkini *(diperbarui jika ada perpanjangan)* |
| Lender | Dipilih dari Lender yang sudah terdaftar di GB / DK project terkait |
| Mata Uang | Mata uang lender yang digunakan dalam perjanjian |
| Amount (Mata Uang Original) | Nilai pinjaman dalam mata uang lender |
| Amount (USD) | Ekuivalen USD — dikonversi manual oleh Staff |

> **Catatan:** Apabila Original Closing Date ≠ Closing Date, berarti perjanjian telah mengalami perpanjangan. Sistem dapat mendeteksi ini secara otomatis.

---

### 3.6 Monitoring Disbursement

Monitoring dilakukan setelah LA efektif (berdasarkan Tanggal Efektif LA), dicatat **per triwulan** mengikuti **tahun anggaran** (Apr–Jun, Jul–Sep, Okt–Des, Jan–Mar).

Setiap entri monitoring terdiri dari dua level:
- **Level LA** — total rencana vs realisasi disbursement per triwulan untuk keseluruhan LA *(wajib)*
- **Level Komponen** — breakdown rencana vs realisasi per komponen/aktivitas *(opsional — dapat diisi jika ingin mencatat lebih granular)*

#### Atribut Monitoring Disbursement (Level LA)

| Field | Keterangan |
|-------|-----------|
| Loan Agreement | Referensi ke LA terkait |
| Tahun Anggaran | Tahun anggaran yang dipantau |
| Triwulan | TW1 (Apr–Jun) / TW2 (Jul–Sep) / TW3 (Okt–Des) / TW4 (Jan–Mar) |
| Kurs USD | Kurs USD–IDR yang digunakan — diinput manual oleh Staff per triwulan |
| Kurs Mata Uang LA | Kurs mata uang LA–IDR — diinput manual oleh Staff per triwulan |
| Rencana (Mata Uang LA) | Target disbursement dalam mata uang LA |
| Rencana (USD) | Target disbursement dalam USD |
| Rencana (IDR) | Target disbursement dalam Rupiah |
| Realisasi (Mata Uang LA) | Realisasi disbursement dalam mata uang LA |
| Realisasi (USD) | Realisasi disbursement dalam USD |
| Realisasi (IDR) | Realisasi disbursement dalam Rupiah |

#### Atribut Breakdown per Komponen (Level Komponen)

| Field | Keterangan |
|-------|-----------|
| Nama Komponen / Aktivitas | Diinput bebas per entri monitoring |
| Rencana (Mata Uang LA) | Target per komponen dalam mata uang LA |
| Rencana (USD) | Target per komponen dalam USD |
| Rencana (IDR) | Target per komponen dalam Rupiah |
| Realisasi (Mata Uang LA) | Realisasi per komponen dalam mata uang LA |
| Realisasi (USD) | Realisasi per komponen dalam USD |
| Realisasi (IDR) | Realisasi per komponen dalam Rupiah |

> **Catatan:** Konversi antar mata uang menggunakan kurs yang diinput manual oleh Staff per triwulan. Sistem menyimpan ketiga nilai (mata uang LA, USD, IDR) secara bersamaan.

---

## 4. Fitur Utama Aplikasi

- **Input Data** — Staff menginput data per tahapan (Blue Book, Green Book, Daftar Kegiatan, Loan Agreement, Monitoring Disbursement).
- **Tampilan Tabel** — Data dapat dilihat dalam format tabel yang terstruktur.
- **Visualisasi Timeline / Journey** — Alur proyek dari Blue Book → Green Book → DK → LA → Monitoring ditampilkan secara visual.
- **Dashboard & Insight** — Monitoring progres dan realisasi disbursement pinjaman luar negeri secara agregat.

---

## 5. Entitas Data

### 5.1 Period
Periode perencanaan **5 tahunan** yang menjadi acuan bersama untuk **Blue Book** dan **National Priority**. Green Book tidak menggunakan entitas Period tersendiri — cukup ditambahkan sebagai baris baru setiap tahun.

### 5.2 Blue Book
- List of Blue Book Projects
- Period
- Publish Date
- Versi revisi — format: `[Nama Period] Revisi ke-[N] Tahun [YYYY]`, contoh: *"BB 2025–2029 Revisi ke-1 Tahun 2026"*
- *(Riwayat semua versi disimpan; proyek yang dihapus tetap ada dengan status `deleted`)*

### 5.3 Green Book
- List of Green Book Projects
- Publish Date (tahun terbit)
- Versi revisi — format: `GB [YYYY] Revisi ke-[N]`
- *(Riwayat semua versi disimpan; proyek yang dihapus tetap ada dengan status `deleted`)*

### 5.4 Daftar Kegiatan
Berupa **surat** yang diterbitkan Bappenas dengan struktur dua level:

**Header Surat (level Daftar Kegiatan)**
- Nomor surat *(opsional)*
- Perihal surat *(wajib)*
- Tanggal *(wajib)*

**Detail Proyek (level per proyek di dalam surat)**
- List Green Book Projects yang direferensikan

**Catatan relasi:**
- Satu Daftar Kegiatan dapat memuat **1 atau lebih** Green Book project
- Satu Green Book project dapat menghasilkan **lebih dari satu** Daftar Kegiatan
- Relasi Daftar Kegiatan ↔ Green Book bersifat **many-to-many**

### 5.4b Loan Agreement (LA)
- Kode Loan *(unik)*
- Tanggal Agreement
- Tanggal Efektif
- Original Closing Date
- Closing Date *(jika berbeda dari Original Closing Date → perjanjian telah diperpanjang)*
- Lender *(dipilih dari Lender yang terdaftar di GB/DK project terkait)*
- Mata Uang
- Amount dalam mata uang original
- Amount ekuivalen USD *(dikonversi manual oleh Staff)*
- Relasi ke proyek DK: **One-to-One** — satu LA merujuk tepat ke satu proyek di Daftar Kegiatan

### 5.4c Monitoring Disbursement
Aktif setelah Tanggal Efektif LA. Struktur dua level:

**Level LA (header per triwulan)**
- Referensi LA
- Tahun Anggaran
- Triwulan: `TW1` / `TW2` / `TW3` / `TW4`
- Kurs USD–IDR dan Kurs Mata Uang LA–IDR *(diinput manual per triwulan)*
- Rencana & Realisasi masing-masing dalam: Mata Uang LA / USD / IDR

**Level Komponen (breakdown per aktivitas — opsional)**
- Nama Komponen *(teks bebas)*
- Rencana & Realisasi masing-masing dalam: Mata Uang LA / USD / IDR

### 5.5 Institution *(Hierarki: Parent → Child)*
| Level | Contoh |
|-------|--------|
| **Parent** | Kementerian / Lembaga |
| **Child** | Eselon I |

- Executing Agency dan Implementing Agency di BB maupun GB keduanya mengacu ke entitas Institution yang sama.
- Executing Agency dapat merujuk ke level Parent (Kementerian) maupun Child (Eselon I).

### 5.6 Region *(Hierarki 3 Level)*
| Level | Keterangan |
|-------|-----------|
| `COUNTRY` | Seluruh Indonesia |
| `PROVINCE` | 38 provinsi |
| `CITY` | Child dari Provinsi |

- Location di BB, GB, dan DK bersifat **multi-select** (satu proyek bisa mencakup banyak region).
- Apabila dipilih **Nasional**, maka secara otomatis mencakup seluruh provinsi — tidak perlu memilih satu per satu.
- Pemilihan bisa di level Provinsi saja, atau hingga Kota/Kabupaten.

### 5.7 National Priority
- Title
- Period (terkait dengan periode berlaku)

### 5.8 Lender
- Nama lembaga
- Type: `Bilateral` / `Multilateral` / `KSA`
- Negara asal *(wajib untuk Bilateral dan KSA, tidak relevan untuk Multilateral — dipilih dari entitas master Negara)*

### 5.8b Country
- Entitas master daftar negara
- Digunakan oleh entitas Lender (untuk Bilateral dan KSA)

### 5.9 Letter of Intent (LoI)
- Perihal surat *(wajib)*
- Tanggal *(wajib)*
- Nomor surat *(opsional — tidak selalu ada)*
- Lender terkait *(satu LoI dari satu lender)*
- Blue Book Project terkait
- *(Catatan: satu BB project dapat menerima LoI dari lebih dari satu lender, masing-masing sebagai entitas LoI terpisah)*

> **Catatan alur Lender:** BB Project → Lender Indication *(nama lender + keterangan, belum pasti)* → LoI *(hampir pasti)* → Funding Source GB *(sudah pasti, cofinancing dimungkinkan)* → Funding Source DK *(dipilih dari lender yang sudah ada di GB atau Lender Indication)*

### 5.10 Bappenas Partner *(Hierarki: Parent → Child)*
| Level | Contoh |
|-------|--------|
| **Parent** | Eselon I |
| **Child** | Eselon II |

- Pada BB Project, cukup menyimpan **Eselon II** — Eselon I diturunkan otomatis dari relasi hierarki parent-child.
- Eselon II bersifat **wajib** per proyek BB.

### 5.11 Program Title *(Hierarki: Parent → Child)*
- Parent Program Title
- Child Program Title
- *(Digunakan bersama oleh Blue Book dan Green Book — entitas master yang sama)*

---

## 6. Diagram Alur Proses (Ringkasan)

```
Bappenas Publish Blue Book (per 5 tahun)
         │
         ▼
  Proyek terdaftar di Blue Book
         │
         │ Lender memberikan Indikasi (Lender Indication) ── masih belum pasti
         ▼
  Lender terbitkan LoI → Executing Agency usulkan ke Bappenas
         │
         ▼
  Proyek masuk Green Book (per tahun) ── Lender sudah hampir pasti
         │
         ▼
  Executing Agency usulkan Daftar Kegiatan ke Bappenas
         │
         ▼
  Bappenas terbitkan Daftar Kegiatan
         │
         │ Tiap proyek di DK → 1 Loan Agreement (One-to-One)
         ▼
  Loan Agreement ditandatangani
         │
         │ Setelah Tanggal Efektif LA
         ▼
  Monitoring Disbursement per Triwulan (Tahun Anggaran)
  ├── Rencana vs Realisasi di level LA
  └── Breakdown per Komponen/Aktivitas
```

---

## 7. Catatan Pengembangan Awal

- Sistem bersifat **input-centric**: data diinput oleh Staff, ditampilkan sebagai tabel dan visualisasi.
- **Manajemen versi/revisi** untuk BB dan GB: riwayat semua versi disimpan, proyek yang dihapus tetap ada dengan status `deleted`. Format versi: `[Nama Period] Revisi ke-[N] Tahun [YYYY]`.
- **Multi-currency** pada Daftar Kegiatan: konversi ke USD dilakukan manual oleh Staff, sistem menyimpan nilai original (mata uang lender) dan nilai USD secara bersamaan.
- Relasi **BB ↔ GB bersifat many-to-many** dan **DK ↔ GB bersifat many-to-many** — keduanya perlu tabel penghubung di skema database.
- **LoI** memiliki perihal dan tanggal (wajib) serta nomor surat (opsional); satu BB project bisa menerima banyak LoI dari lender yang berbeda.
- **Daftar Kegiatan** bersifat final (tidak direvisi), terdiri dari header surat dan list proyek GB. Activity Details diinput bebas (tidak dipilih dari Activities GB) — secara konteks merupakan realisasi rencana GB, namun tidak ada relasi teknis langsung.
- **Period** hanya berlaku untuk Blue Book dan National Priority (5 tahunan); Green Book tidak menggunakan entitas Period.
- **Program Title** adalah entitas master bersama (shared) antara BB dan GB — dipilih dari daftar, tidak diketik bebas.
- **Bappenas Partner** pada BB Project cukup menyimpan Eselon II per proyek — Eselon I diturunkan otomatis dari hierarki parent-child.
- **Executing Agency & Implementing Agency** keduanya bersifat multi-select di BB maupun GB, mengacu ke entitas Institution yang sama (Kementerian/Eselon I).
- **Lender Indication** dicatat di level BB Project — lender masih bersifat indikasi (belum pasti). Berbeda dengan Funding Source di GB yang sudah hampir pasti.
- **Funding Source di DK** dipilih dari lender yang sudah terdaftar (dari Lender Indication BB atau Funding Source GB), bukan diinput bebas.
- Tingkat kepastian lender: Lender Indication (BB) → LoI → Funding Source GB → Funding Source DK.
- **Region** di BB, GB, dan DK bersifat multi-select dengan tipe: `COUNTRY` → `PROVINCE` → `CITY`. Pemilihan Nasional otomatis mencakup seluruh provinsi.
- **Country** adalah entitas master tersendiri, digunakan oleh Lender tipe Bilateral dan KSA.
- **Status proyek** BB dan GB tidak perlu field eksplisit — status diturunkan dari relasi yang ada (ada/tidaknya GB, LoI, atau DK terkait).
- **Loan Agreement (LA)** merupakan tahapan setelah Daftar Kegiatan — relasi One-to-One dengan proyek di DK. Memuat Kode Loan, tanggal Agreement, Efektif, Original Closing Date, Closing Date, Lender, Mata Uang, dan Amount (original + USD). Perpanjangan LA terdeteksi otomatis apabila Closing Date ≠ Original Closing Date.
- **Monitoring Disbursement** aktif setelah Tanggal Efektif LA, dicatat per triwulan mengikuti tahun anggaran (TW1: Apr–Jun, TW2: Jul–Sep, TW3: Okt–Des, TW4: Jan–Mar). Mencatat rencana vs realisasi dalam tiga mata uang (Mata Uang LA, USD, IDR) dengan kurs diinput manual per triwulan. Breakdown per komponen bersifat **opsional** — dapat diisi jika ingin pencatatan lebih granular.
- **Disbursement Plan** di GB adalah total keseluruhan proyek per tahun — bukan per lender.
- **Funding Allocation** di GB kolom Activities mengacu ke baris Activities yang sudah diinput di tabel Activities GB — ada relasi teknis antar keduanya.
- **National Priority** di BB Project bersifat multi-select.
- **Implementation Location** di tabel Activities GB adalah teks bebas — tidak mengacu ke entitas master Wilayah.
- Sistem **role & permission** granular hingga level operasi CRUD per modul per user.

---

*Dokumen ini merupakan ide awal (initial concept) untuk pengembangan sistem PRISM.*
*Versi: 0.14 — Draft (updated: breakdown komponen monitoring bersifat opsional)*
