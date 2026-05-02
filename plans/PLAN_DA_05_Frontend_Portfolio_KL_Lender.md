# PLAN DA-05 - Frontend Portfolio, Kementerian/Lembaga, Lender, and Absorption UI

> **Scope:** Membangun UI utama Dashboard Analytics untuk portfolio, Kementerian/Lembaga, lender, penyerapan, tahunan, dan proporsi lender.
> **Deliverable:** Dashboard analytics utama usable, API-backed, dan clickable ke workspace terkait.
> **Dependencies:** `PLAN_DA_02_Backend_KL_Lender_Absorption.md`, `PLAN_DA_04_Frontend_Analytics_Foundation.md`.

---

## Instruksi Untuk Agent

Baca sebelum mulai:

- `AGENTS.md`
- `docs/PRISM_API_Contract.md`
- `docs/PRISM_Frontend_Structure.md`
- `plans/PLAN_DA_00_Dashboard_Analytics_Roadmap.md`
- `plans/PLAN_DA_02_Backend_KL_Lender_Absorption.md`
- `plans/PLAN_DA_04_Frontend_Analytics_Foundation.md`

Aturan UI:

- Label visible gunakan "Kementerian/Lembaga".
- Jangan gunakan data dummy.
- Card/tabel yang menunjukkan insight harus bisa drilldown jika backend menyediakan query.
- UI operational, bukan landing page.
- Gunakan komponen shared yang sudah dibuat.
- Jangan menambah dependency chart baru sebelum cek ECharts yang sudah ada.

---

## Task 1 - Portfolio Section

Section "Portfolio" menampilkan:

- total project,
- assignment Kementerian/Lembaga,
- total nilai pipeline USD,
- total nilai Loan Agreement USD,
- rencana USD,
- realisasi USD,
- penyerapan.

Tambahkan funnel pipeline:

- Blue Book,
- Green Book,
- Daftar Kegiatan,
- Loan Agreement,
- Monitoring.

UI:

- Metric cards ringkas.
- Funnel bisa chart bar horizontal atau compact table.
- Tiap stage clickable ke Project Master dengan `pipeline_statuses`.

Acceptance:

- Funnel tidak double count revisi.
- Empty state jelas jika data kosong.

---

## Task 2 - Kementerian/Lembaga Section

Section "Kementerian/Lembaga" menampilkan:

- sebaran Kementerian/Lembaga yang mendapatkan project,
- performa Kementerian/Lembaga,
- project count,
- assignment count,
- nilai Loan Agreement,
- planned USD,
- realized USD,
- absorption,
- pipeline breakdown.

UI components:

- leaderboard table,
- bar chart top 10 by project count,
- bar chart/top table by absorption,
- badges untuk low/normal/high absorption.

Columns table:

| Column | Keterangan |
|--------|------------|
| Kementerian/Lembaga | root institution label |
| Project | distinct project count |
| Assignment | overlap-aware count |
| Loan Agreement | count |
| Nilai Pinjaman USD | agreement amount |
| Rencana USD | monitoring planned |
| Realisasi USD | monitoring realized |
| Penyerapan | absorption bar |
| Aksi | drilldown |

Rules:

- Jangan tampilkan UUID.
- Jika name panjang, truncate dengan title tooltip.
- Drilldown ke Project Master memakai filter institution.

---

## Task 3 - Lender Section

Section "Lender" menampilkan:

- performa lender,
- lender per Kementerian/Lembaga matrix,
- total Loan Agreement,
- institution coverage,
- project coverage,
- amount USD,
- planned/realisasi/penyerapan.

UI:

- summary cards,
- lender performance table,
- matrix table dengan row Kementerian/Lembaga dan column/top lender jika data banyak,
- filter lokal "Top N" untuk matrix agar tidak terlalu lebar.

Rules:

- Lender type badge: Bilateral, Multilateral, KSA.
- KSA harus punya badge sendiri.
- Jangan gabungkan KSA ke Bilateral.
- Drilldown lender ke Project Master atau Loan Agreement sesuai target backend.

---

## Task 4 - Absorption Section

Section "Penyerapan" menampilkan performa penyerapan di tiga level:

- Kementerian/Lembaga,
- Project,
- Lender.

UI:

- segmented control level,
- ranked table,
- status badge low/normal/high,
- variance USD,
- absorption bar,
- top low absorption list.

Rules:

- planned 0 harus ditampilkan sebagai 0%, bukan NaN/Infinity.
- Low absorption harus mudah dilihat tetapi tidak memakai warna berlebihan.
- Tiap row punya drilldown jika tersedia.

---

## Task 5 - Yearly Performance Section

Section "Tahunan" menampilkan:

- trend planned vs realized per tahun/triwulan,
- absorption trend,
- jumlah Loan Agreement/project aktif per periode.

UI:

- grouped bar chart planned vs realized,
- line atau bar untuk absorption,
- table detail per year/quarter.

Rules:

- Order quarter: TW1, TW2, TW3, TW4.
- Jika filter tahun aktif, tampilkan semua triwulan pada tahun tersebut.
- Jika filter triwulan aktif, tampilkan konteks jelas.

---

## Task 6 - Lender Proportion Section

Section "Proporsi Lender" menampilkan:

- proporsi Bilateral, Multilateral, KSA.
- tampil per stage:
  - Lender Indication,
  - Green Book Funding Source,
  - Loan Agreement,
  - Monitoring Realization.

UI:

- stacked bar atau donut per stage,
- table dengan count, amount USD, share percent.

Rules:

- Label stage harus eksplisit agar user tidak mengira indication sama dengan legal binding.
- Jika amount tidak tersedia di stage tertentu, pakai count dan tampilkan tooltip/empty note singkat.

---

## Task 7 - Responsive Behavior

Pastikan:

- desktop: sections dense dan scan-friendly,
- tablet/mobile: table horizontal scroll hanya jika diperlukan,
- filter advanced collapsible,
- chart tetap terbaca,
- text tidak overlap.

---

## Acceptance Criteria

- Semua section mengambil data dari backend analytics.
- Tidak ada placeholder sample data.
- Kementerian/Lembaga dan lender analytics tampil lengkap.
- Penyerapan by Kementerian/Lembaga, project, lender tampil dan div-by-zero safe.
- Proporsi Bilateral/Multilateral/KSA tampil dengan KSA sebagai kategori terpisah.
- Drilldown bekerja untuk cards/table row yang memiliki query.
- `npm.cmd run type-check` dan `npm.cmd run build` berhasil.

---

## Verification

```powershell
cd prism-frontend
npm.cmd run type-check
npm.cmd run build
```

Browser smoke:

- buka `/dashboard`,
- apply filter tahun,
- apply filter lender type KSA,
- apply filter Kementerian/Lembaga,
- klik drilldown stage pipeline,
- klik drilldown lender,
- cek mobile width.

---

## Checklist

- [x] Portfolio section selesai
- [x] Funnel pipeline selesai
- [x] Kementerian/Lembaga section selesai
- [x] Lender section selesai
- [x] Lender per Kementerian/Lembaga matrix selesai
- [x] Penyerapan section selesai
- [x] Performa tahunan section selesai
- [x] Proporsi lender section selesai
- [x] Drilldown utama bekerja
- [x] Responsive behavior dicek
- [x] `npm.cmd run type-check` berhasil
- [x] `npm.cmd run build` berhasil
