# PLAN DA-06 - Frontend Risk, Data Quality, QA, and Handover

> **Scope:** Menyelesaikan Dashboard Analytics dengan UI risk/data quality, drilldown penuh, testing, verifikasi live, dan catatan handover.
> **Deliverable:** Dashboard Analytics siap dipakai dan bisa diverifikasi end-to-end.
> **Dependencies:** `PLAN_DA_03_Backend_Risk_Data_Quality_Drilldown.md`, `PLAN_DA_05_Frontend_Portfolio_KL_Lender.md`.

---

## Instruksi Untuk Agent

Baca sebelum mulai:

- `AGENTS.md`
- `docs/PRISM_API_Contract.md`
- `docs/PRISM_Frontend_Structure.md`
- `plans/PLAN_DA_00_Dashboard_Analytics_Roadmap.md`
- `plans/PLAN_DA_03_Backend_Risk_Data_Quality_Drilldown.md`
- `plans/PLAN_DA_05_Frontend_Portfolio_KL_Lender.md`

Aturan:

- Jangan menyelesaikan hanya dengan build pass. Lakukan browser smoke jika memungkinkan.
- Risk/data quality cards harus actionable.
- Drilldown query harus tetap bisa diubah user setelah masuk target workspace.
- Jangan membuat route/permission baru tanpa menyesuaikan nav dan guard.

---

## Task 1 - Risk & Data Quality Section

Tambahkan section/tab:

```text
Risiko & Data Quality
```

Tampilkan:

- low absorption projects,
- Loan Agreement efektif tanpa monitoring,
- closing risk,
- extended loans,
- pipeline bottlenecks,
- data quality cards.

UI:

- cards summary untuk risk count,
- watchlist table per kategori,
- severity badge,
- drilldown button,
- empty state jika tidak ada issue.

Labels:

- "Penyerapan rendah"
- "Loan Agreement efektif tanpa monitoring"
- "Mendekati closing date"
- "Loan Agreement diperpanjang"
- "Belum berlanjut ke tahap berikutnya"
- "Kelengkapan data"

---

## Task 2 - Risk Watchlist Tables

Buat table per watchlist atau satu table dengan kategori.

Columns minimal:

| Column | Keterangan |
|--------|------------|
| Kategori | risk/data quality type |
| Project | nama project |
| Kementerian/Lembaga | root institution |
| Lender | lender name/type |
| Nilai USD | amount/planned/realisasi sesuai context |
| Penyerapan | absorption bar jika relevan |
| Status | severity |
| Aksi | drilldown |

Rules:

- Jika field tidak relevan, tampilkan `-`, bukan 0 palsu.
- Jangan tampilkan raw JSON drilldown.
- Jangan tampilkan UUID.

---

## Task 3 - Data Quality Drilldown

Pastikan cards:

- `NO_EXECUTING_AGENCY`
- `NO_LENDER`
- `NO_REGION`
- `NO_FUNDING_AMOUNT`
- `EFFECTIVE_NO_MONITORING`
- `PLANNED_ZERO_REALIZED_POSITIVE`

bisa klik ke target workspace.

Jika Project Master atau Monitoring belum support filter dari backend:

- tambah filter support di target dengan scope sempit,
- update types/service/store target,
- jangan hardcode filter di URL tanpa parser target.

Acceptance:

- User masuk ke workspace target dan masih bisa adjust filter.
- Active filter pill muncul jika target page mendukung.

---

## Task 4 - Cross-Section Loading and Error UX

Pastikan setiap section punya state:

- loading,
- ready,
- empty,
- error.

Rules:

- Error satu section tidak membuat seluruh dashboard blank.
- Tombol retry tersedia pada section yang gagal.
- Filter apply tidak menyebabkan layout jump berat.

---

## Task 5 - Browser Smoke Checklist

Dengan backend dan frontend running:

1. Login sebagai admin/staff yang punya akses.
2. Buka `/dashboard`.
3. Cek semua tab/section bisa dimuat.
4. Apply filter tahun anggaran.
5. Apply filter triwulan.
6. Apply filter tipe lender `KSA`.
7. Apply filter Kementerian/Lembaga.
8. Reset filter.
9. Klik drilldown:
   - pipeline stage,
   - Kementerian/Lembaga,
   - lender,
   - low absorption,
   - data quality.
10. Cek mobile viewport.

Catat hasil smoke di final handover.

---

## Task 6 - Automated Checks

Backend:

```powershell
cd prism-backend
make generate
go test ./...
```

Frontend:

```powershell
cd prism-frontend
npm.cmd run type-check
npm.cmd run build
```

Jika ada test yang gagal karena existing issue di luar scope, catat jelas:

- test name,
- error ringkas,
- alasan out-of-scope,
- apakah dashboard analytics tetap terverifikasi.

---

## Task 7 - Update Documentation and Handover Notes

Update docs yang relevan:

- `docs/PRISM_API_Contract.md` jika contract berubah saat implementasi.
- Plan checklist aktif.
- Jika ada filter baru di Project Master/Monitoring, update bagian endpoint terkait.

Tulis handover notes di final response:

- endpoint yang ditambah,
- file backend utama,
- file frontend utama,
- verifikasi yang dijalankan,
- known limitations.

Jangan membuat file dokumentasi baru jika informasi cukup di plan dan API contract, kecuali user meminta.

---

## Acceptance Criteria

- Risk & Data Quality section tampil dan API-backed.
- Semua card/table actionable dengan drilldown.
- Drilldown masuk ke workspace target dengan filter aktif.
- Loading/error/empty state stabil.
- Backend tests pass atau blocker dicatat.
- Frontend type-check/build pass.
- Browser smoke selesai atau blocker dicatat.

---

## Checklist

- [ ] Risk & Data Quality section selesai
- [ ] Watchlist tables selesai
- [ ] Data quality cards clickable
- [ ] Drilldown target Project Master bekerja
- [ ] Drilldown target Monitoring bekerja
- [ ] Drilldown target Loan Agreement bekerja jika digunakan
- [ ] Section-level loading/error/empty state selesai
- [ ] Backend verification selesai
- [ ] Frontend verification selesai
- [ ] Browser smoke selesai atau blocker dicatat
- [ ] API contract final sinkron
- [ ] Handover notes siap
