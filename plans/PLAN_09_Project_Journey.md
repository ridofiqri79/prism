# PLAN 09 - Project Journey

> **Scope:** Visualisasi perjalanan proyek sampai Loan Agreement dan Monitoring berdasarkan snapshot downstream yang konkret.
> **Deliverable:** Satu halaman `Project Journey` yang operasional dengan summary, flow, dan timeline.
> **Referensi:** `docs/PRISM_API_Contract.md` bagian Journey.

---

## Prinsip Revisi

- Semua API call lewat `JourneyService`, state lewat `journey.store.ts`, dan tipe lewat `journey.types.ts`.
- Entry point route menerima `bbProjectId` dan tetap valid untuk snapshot lama.
- Node downstream tidak otomatis pindah ke revisi baru; tampilkan indikator jika ada revisi BB/GB yang lebih baru.
- Summary, flow, dan timeline harus membaca path konkret yang benar-benar disimpan downstream.

---

## Project Journey

- Route menerima `bbProjectId`.
- Timeline menampilkan jalur konkret snapshot yang dipakai downstream:
  - Blue Book project
  - Letter of Intent
  - Green Book project
  - Daftar Kegiatan project
  - Loan Agreement
- Node downstream tidak otomatis pindah ke revisi baru; tampilkan badge jika ada revisi BB/GB lebih baru.

---

## Checklist

- [x] `journey.types.ts` sesuai kontrak response journey.
- [x] `journey.service.ts` memuat fetch detail journey.
- [x] `journey.store.ts` mengelola loading/error pencarian dan detail.
- [x] Project Journey memiliki view summary, flow, dan timeline.
- [x] Project Journey menampilkan snapshot path BB -> GB -> DK -> LA.
