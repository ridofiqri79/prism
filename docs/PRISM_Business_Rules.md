# PRISM — Business Rules

> Agent wajib membaca dokumen ini sebelum mengimplementasikan logic apapun.

---

## 1. Aturan Umum

- Semua `id` menggunakan UUID v4.
- `created_at` dan `updated_at` di-set oleh database, bukan aplikasi.
- Penghapusan BB Project dan GB Project memakai hard delete. Backend wajib menolak penghapusan jika record masih dipakai downstream, dan response harus menampilkan relasi turunan yang perlu dibersihkan terlebih dahulu.

---

## 2. Aturan Lender

- `Bilateral` dan `KSA` → `country_id` wajib. `Multilateral` → `country_id` NULL.
- Funding Source di DK hanya boleh dipilih dari lender yang ada di `lender_indication` BB terkait ATAU `gb_funding_source` GB terkait.
- `lender_id` di LA harus berasal dari `dk_financing_detail` DK Project yang direferensikan.
- Alur kepastian: Lender Indication (belum pasti) → LoI → Funding Source GB (pasti) → DK → LA (legal binding).

---

## 2a. Aturan Currency

- `currency` memakai kode ISO 4217 dari Master Currency.
- Currency yang dapat dipilih pada Green Book, DK, dan LA hanya currency dengan `is_active = true`.
- Pencatatan currency dimulai dari Green Book Funding Source. DK melakukan autofill currency dan nilai original/USD dari Funding Source GB, lalu tetap dapat diedit sebelum disimpan.
- Jika currency adalah `USD`, user tidak perlu mengisi nilai USD terpisah. Backend menyimpan nilai USD sama dengan nilai original.
- Seed awal Master Currency berisi mata uang negara donor/lender dan mata uang yang lazim dipakai lembaga multilateral. `XDR` disediakan nonaktif sebagai referensi Special Drawing Rights.
- Konversi ke USD dan IDR tetap manual oleh Staff. Sistem tidak melakukan konversi otomatis.

---

## 3. Aturan Blue Book

> Detail versioning BB/GB: `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`.

- Status Blue Book yang tampil ke user adalah `Berlaku` dan `Tidak Berlaku`.
- Status Blue Book dipilih user saat create/edit, tidak lagi diubah otomatis saat Blue Book baru dibuat.
- Format revisi: `BB 2025-2029 Revisi ke-1 Tahun 2026`.
- Kombinasi Period + revision number + revision year harus unik; Blue Book dengan period dan versi yang sama tidak boleh dibuat dua kali.
- BB Project adalah snapshot di dalam satu Blue Book/revisi, bukan identitas logical tunggal.
- Project yang sama lintas revisi harus dihubungkan dengan logical identity.
- `bb_code` unik hanya dalam Blue Book yang sama. Kode yang sama boleh muncul kembali pada revisi Blue Book lain untuk logical project yang sama.
- Pada Blue Book detail untuk revisi, user dapat memakai tombol `Impor Proyek dari Blue Book Lain` untuk memilih Project Blue Book dari Blue Book sumber periode yang sama dan membawanya ke Blue Book tujuan.
- Project Blue Book yang dibawa dari revisi sebelumnya di-clone sebagai snapshot baru dengan `project_identity_id` yang sama.
- Mitra Kerja Bappenas bersifat opsional dan boleh lebih dari satu. Simpan Eselon II saja; Eselon I diturunkan dari `parent_id`.
- National Priority pada proyek Blue Book boleh menggunakan master National Priority dari period mana pun.

---

## 4. Aturan Green Book

- Hanya satu GB berstatus `active` per `publish_year`.
- Format revisi: `GB 2025 Revisi ke-1`.
- Kombinasi `publish_year` + `revision_number` harus unik; Green Book dengan Publish Year dan Revision number yang sama tidak boleh dibuat dua kali.
- GB Project adalah snapshot di dalam satu Green Book/revisi, bukan identitas logical tunggal.
- Project yang sama lintas revisi Green Book harus dihubungkan dengan logical identity.
- `gb_code` unik hanya dalam Green Book yang sama. Kode yang sama boleh muncul kembali pada revisi Green Book lain untuk logical GB Project yang sama.
- Revisi Green Book boleh menyalin GB Project yang sama persis dari revisi sebelumnya.
- GB Project wajib mereferensikan minimal 1 BB Project.
- GB Project boleh menggabungkan lebih dari satu BB Project hanya jika seluruh BB Project berasal dari header Blue Book yang sama (Period, revision number, dan revision year sama). Backend menolak relasi dari header Blue Book berbeda.
- Satu BB Project boleh dipakai oleh lebih dari satu GB Project dalam header Green Book yang sama untuk mendukung pemecahan satu proyek BB menjadi beberapa proyek GB.
- Saat GB Project dibuat atau direvisi, relasi ke BB Project harus memakai versi BB Project terbaru untuk logical project terkait.
- Mitra Kerja Bappenas pada GB Project bersifat opsional dan boleh lebih dari satu.
- Funding Source Green Book menyimpan `currency`, nilai original, dan nilai USD. Untuk `USD`, nilai original dan USD harus sama.
- `gb_funding_allocation` mereferensikan `gb_activity` — selalu sinkron, jika activity dihapus, allocation ikut terhapus (CASCADE).
- Disbursement Plan: total proyek per tahun — bukan per lender. Kombinasi `(gb_project_id, year)` unik.

---

## 5. Aturan Daftar Kegiatan

- Final setelah diterbitkan — tidak bisa direvisi. Backend cegah UPDATE kecuali ADMIN.
- Saat DK Project dibuat, relasi ke GB Project harus memakai versi GB Project terbaru untuk logical project terkait.
- Setelah DK/LA dibuat, downstream tetap menunjuk ke snapshot GB/BB yang dicantumkan saat DK dibuat; tidak auto-pindah ketika ada revisi BB/GB baru.
- Nama proyek pada DK Project dicatat ulang sebagai snapshot Daftar Kegiatan dan boleh berbeda dari nama proyek Green Book; default frontend diisi dari GB Project terpilih lalu tetap dapat diedit sebelum disimpan.
- Mitra Kerja Bappenas pada DK Project bersifat opsional, boleh lebih dari satu, dapat diisi otomatis dari GB Project terpilih, dan tetap dapat diedit sebelum disimpan.
- Activity Details diinput bebas — tidak ada relasi teknis ke Activities GB.

---

## 6. Aturan Loan Agreement

- One-to-One dengan DK Project — tidak boleh ada LA kedua untuk DK yang sama.
- `closing_date >= original_closing_date` (enforced di DDL).
- `is_extended` dan `extension_days` adalah computed, tidak disimpan di DB.
- Saat `closing_date` diupdate → kirim SSE `loan_agreement.extended`.
- `currency`: kode ISO 4217. Konversi ke USD dilakukan manual oleh Staff — sistem tidak konversi otomatis.

---

## 7. Aturan Monitoring Disbursement

- Hanya boleh dibuat jika `effective_date <= NOW()`. Backend wajib validasi ini.
- Triwulan tahun anggaran: TW1 Apr-Jun, TW2 Jul-Sep, TW3 Okt-Des, TW4 Jan-Mar.
- Kombinasi `(loan_agreement_id, budget_year, quarter)` unik.
- Kurs diinput manual — sistem tidak auto-fetch kurs. Ketiga nilai (LA, USD, IDR) disimpan bersamaan tanpa kalkulasi otomatis.
- Breakdown komponen: opsional. Total komponen tidak harus sama dengan level LA.
- `penyerapan_pct = (realisasi / rencana) * 100` — computed di server, jika `rencana = 0` maka hasil 0.

---

## 8. Aturan Wilayah

- Pilih Nasional → otomatis mencakup seluruh provinsi. Simpan hanya `region_id` Nasional di DB.
- COUNTRY/Nasional juga mencakup seluruh CITY di bawah provinsi turunannya.
- Frontend nonaktifkan pilihan PROVINCE/CITY jika COUNTRY/Nasional sudah dipilih.
- Location BB, GB, DK: multi-select dan boleh memilih seluruh level region (`COUNTRY`, `PROVINCE`, `CITY`).
- Untuk analitik choropleth kabupaten/kota, data yang disimpan di level `COUNTRY`/`PROVINCE` tidak boleh digandakan ke setiap `CITY`. Drilldown kab/kota hanya menghitung lokasi `CITY` eksplisit agar angka peta tidak membuat alokasi palsu.

---

## 8a. Aturan Durasi Proyek

- Durasi proyek pada BB Project, GB Project, dan DK Project disimpan sebagai integer jumlah bulan.
- Durasi kosong boleh `NULL`; jika diisi harus lebih dari 0 bulan.

---

## 9. Aturan Institution

- Nama institution top-level (`parent_id` NULL) harus unik.
- Nama institution child harus unik dalam parent yang sama. Nama yang sama boleh digunakan di parent berbeda.
- Satu institution boleh menjadi EA sekaligus IA pada proyek yang sama bila memang sesuai data proyek.

---

## 10. Aturan Permission

- ADMIN: akses penuh, tidak ada entri di `user_permission`, cukup cek `role = 'ADMIN'`.
- STAFF: default deny — tidak ada entri = tidak ada akses. Dicek di middleware, bukan di service/handler.

| Module | Cakupan |
|--------|---------|
| `blue_book` | Blue Book header |
| `bb_project` | BB Project + LoI + Lender Indication + Project Cost |
| `green_book` | Green Book header |
| `gb_project` | GB Project + Activities + Funding Source + Disbursement Plan + Funding Allocation |
| `daftar_kegiatan` | DK header + DK Project + semua sub-tabel |
| `loan_agreement` | Loan Agreement |
| `monitoring_disbursement` | Monitoring + Komponen |
| `institution` | Master Institution |
| `lender` | Master Lender |
| `currency` | Master Currency |
| `region` | Master Wilayah |
| `national_priority` | Master National Priority |
| `program_title` | Master Program Title |
| `bappenas_partner` | Master Bappenas Partner |
| `period` | Master Period |
| `country` | Master Negara |
| `user` | User Management (ADMIN only) |

---

## 11. Aturan Audit Trail

- Setiap request ubah data: `SET LOCAL app.current_user_id = '<uuid>'` di awal transaksi.
- `audit_log` hanya bisa diakses ADMIN — tidak boleh diekspos ke STAFF.
- Tabel junction proyek tanpa kolom `id` tetap diaudit memakai owner id (`bb_project_id`, `gb_project_id`, atau `dk_project_id`) supaya relasi turunan yang ikut terhapus tetap tercatat.

---

## 12. Aturan SSE

- Event dikirim setelah DB commit berhasil — jika rollback, event tidak dikirim.
- Client wajib auto-reconnect dengan delay minimal 5 detik.
