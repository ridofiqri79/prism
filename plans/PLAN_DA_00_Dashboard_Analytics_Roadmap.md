# PLAN DA-00 - Dashboard Analytics Roadmap

> **Scope:** Roadmap implementasi Dashboard Analytics PRISM setelah baseline `PLAN_09` dan `PLAN_BE_06`.
> **Deliverable:** Pembagian fase yang bisa di-handover ke Claude Code atau agent lain secara bertahap.
> **Status:** Planning only. Jangan implementasi kode dari file ini secara langsung.

---

## 1. Tujuan

Dashboard saat ini baru menampilkan ringkasan dasar:

- total proyek Blue Book,
- total proyek Green Book,
- total Loan Agreement,
- total nilai pinjaman,
- total realisasi,
- overall absorption,
- monitoring aktif,
- filter tahun/triwulan/lender,
- rincian monitoring per lender.

Dashboard Analytics baru harus memberi gambaran portfolio pinjaman luar negeri yang lebih operasional:

- pembagian lender per Kementerian/Lembaga,
- sebaran Kementerian/Lembaga yang mendapatkan proyek,
- performa Kementerian/Lembaga,
- performa lender,
- performa tahunan,
- performa penyerapan di tingkat Kementerian/Lembaga, project, dan lender,
- proporsi lender Bilateral, Multilateral, dan KSA,
- insight tambahan: pipeline funnel, bottleneck stage, risk watchlist, closing risk, monitoring compliance, data quality, dan drilldown ke workspace terkait.

---

## 2. Prinsip Data

### 2.1 Jangan campur stage lender

Lender punya tingkat kepastian berbeda:

| Stage | Sumber data | Makna |
|-------|-------------|-------|
| Blue Book | `lender_indication` | Indikasi, belum pasti |
| Green Book | `gb_funding_source` | Funding source, lebih pasti |
| Daftar Kegiatan | `dk_financing_detail` | Funding detail pada surat DK |
| Loan Agreement | `loan_agreement.lender_id` | Legal binding |
| Monitoring | `monitoring_disbursement` lewat `loan_agreement` | Performa realisasi |

Dashboard eksekutif default memakai Loan Agreement dan Monitoring untuk angka performa. Data Blue Book dan Green Book boleh tampil sebagai pipeline, tetapi labelnya harus eksplisit.

### 2.2 Kementerian/Lembaga

Definisi dashboard:

- Kementerian/Lembaga adalah root institution hasil roll-up dari `institution.parent_id`.
- Label utama memakai nama lengkap "Kementerian/Lembaga", bukan hanya "K/L" di UI.
- Jika source institution adalah Eselon I/Eselon II, naikkan ke root parent.
- Jika source root bukan level `Kementerian/Badan/Lembaga`, tampilkan sebagai "Instansi Non Kementerian/Lembaga" atau label root aslinya, tergantung hasil validasi data.

Sumber stage:

| Stage | Sumber institution |
|-------|--------------------|
| Blue Book | `bb_project_institution` role `Executing Agency` |
| Green Book | `gb_project_institution` role `Executing Agency`; `gb_funding_source.institution_id` untuk implementing agency per funding row |
| Daftar Kegiatan | `dk_project.institution_id` sebagai Executing Agency |
| Monitoring | `loan_agreement -> dk_project.institution_id` |

### 2.3 Anti double count revisi

Dashboard portfolio default harus mengikuti latest snapshot:

- Blue Book dan Green Book adalah snapshot revisi, bukan logical identity tunggal.
- Untuk hitungan portfolio, default ke latest `project_identity_id`.
- `include_history=true` hanya untuk mode audit/history, bukan default dashboard.
- Detail/journey tetap memakai concrete IDs yang tersimpan downstream.

### 2.4 Dua jenis hitungan project

Untuk analitik Kementerian/Lembaga, tampilkan dua angka jika relevan:

| Metric | Makna |
--------|-------|
| `project_count` | Jumlah project logical unik, deduplicated |
| `assignment_count` | Jumlah assignment Kementerian/Lembaga, overlap-aware |

Ini penting karena satu project bisa punya lebih dari satu Executing Agency atau Implementing Agency.

---

## 3. Phase Plan

Kerjakan satu plan per sesi. Jangan lompat ke plan berikutnya sebelum checklist plan aktif selesai.

| Urutan | File | Fokus |
|--------|------|-------|
| 1 | `plans/PLAN_DA_01_Backend_Analytics_Contract_Foundation.md` | API contract, DTO, route, filter umum, query foundation |
| 2 | `plans/PLAN_DA_02_Backend_KL_Lender_Absorption.md` | Aggregasi Kementerian/Lembaga, lender, tahunan, penyerapan, proporsi lender |
| 3 | `plans/PLAN_DA_03_Backend_Risk_Data_Quality_Drilldown.md` | Insight tambahan, risk/data quality, drilldown query |
| 4 | `plans/PLAN_DA_04_Frontend_Analytics_Foundation.md` | Types, service, store/composable, layout dashboard analytics |
| 5 | `plans/PLAN_DA_05_Frontend_Portfolio_KL_Lender.md` | UI portfolio, Kementerian/Lembaga, lender, tahunan, penyerapan |
| 6 | `plans/PLAN_DA_06_Frontend_Risk_QA_Handover.md` | UI risk/data quality, clickable drilldown, QA, handover |

---

## 4. Endpoint Target

Endpoint final yang disarankan:

| Endpoint | Tujuan |
|----------|--------|
| `GET /dashboard/analytics/overview` | Summary portfolio, funnel pipeline, top insights |
| `GET /dashboard/analytics/institutions` | Kementerian/Lembaga distribution and performance |
| `GET /dashboard/analytics/lenders` | Lender performance and lender per Kementerian/Lembaga matrix |
| `GET /dashboard/analytics/absorption` | Penyerapan by Kementerian/Lembaga, project, lender |
| `GET /dashboard/analytics/yearly` | Performa tahunan/triwulan |
| `GET /dashboard/analytics/lender-proportion` | Proporsi Bilateral, Multilateral, KSA by stage |
| `GET /dashboard/analytics/risks` | Risk watchlist, monitoring compliance, data quality |

Permission awal mengikuti dashboard saat ini: authenticated. Jangan tambah permission module baru kecuali user menyetujui eksplisit.

---

## 5. Filter Umum

Semua endpoint analytics memakai filter yang konsisten:

| Param | Keterangan |
|-------|------------|
| `budget_year` | Tahun monitoring |
| `quarter` | `TW1`, `TW2`, `TW3`, `TW4` |
| `lender_ids` | Multi-value UUID lender |
| `lender_types` | Multi-value: `Bilateral`, `Multilateral`, `KSA` |
| `institution_ids` | Multi-value UUID institution, resolve ke root Kementerian/Lembaga bila perlu |
| `pipeline_statuses` | Multi-value: `BB`, `GB`, `DK`, `LA`, `Monitoring` |
| `project_statuses` | Multi-value: `Pipeline`, `Ongoing` |
| `region_ids` | Multi-value UUID region |
| `program_title_ids` | Multi-value UUID program title |
| `foreign_loan_min`, `foreign_loan_max` | Range nilai USD |
| `include_history` | Default `false`; `true` hanya untuk audit/history |

Multi-value mengikuti pola `GET /projects`: repeated query param, comma-separated, atau array suffix.

---

## 6. Definisi Metric

| Metric | Formula |
|--------|---------|
| `planned_usd` | `SUM(monitoring_disbursement.planned_usd)` |
| `realized_usd` | `SUM(monitoring_disbursement.realized_usd)` |
| `absorption_pct` | `realized_usd / planned_usd * 100`; jika planned 0 maka 0 |
| `agreement_amount_usd` | `SUM(loan_agreement.amount_usd)` |
| `pipeline_loan_usd` | Ikuti logic Project Master: BB memakai `bb_project_cost`, GB+ memakai `gb_funding_source` |
| `monitoring_count` | `COUNT(monitoring_disbursement.id)` |
| `effective_without_monitoring_count` | LA efektif tetapi belum punya monitoring pada filter periode |
| `extended_loan_count` | `closing_date > original_closing_date` |
| `closing_risk_count` | Closing date dalam threshold dan absorption di bawah threshold |

---

## 7. Handover Prompt Untuk Claude Code

Gunakan prompt berikut saat mengeksekusi salah satu phase:

```text
Baca AGENTS.md, docs/PRISM_Business_Rules.md, docs/prism_ddl.sql,
docs/PRISM_API_Contract.md, docs/PRISM_BB_GB_Revision_Versioning_Plan.md,
dan plan aktif di plans/PLAN_DA_*.md.

Kerjakan hanya plan aktif, jangan lompat ke phase berikutnya.
Ikuti aturan query-first backend PRISM: SQL di sql/queries -> make generate -> model -> service -> handler -> route.
Untuk frontend, ikuti urutan: API contract -> types -> service -> store/composable -> components -> page -> route.
Jangan pakai placeholder data. Jangan campur lender indication, funding source, dan Loan Agreement tanpa label stage.
Default analytics harus latest snapshot dan tidak double-count revisi.
Setelah selesai, update checklist di plan aktif dan jalankan verifikasi yang diminta.
```

---

## Checklist

- [x] Phase order disetujui user
- [x] Definisi Kementerian/Lembaga disetujui user
- [x] Definisi stage lender disetujui user
- [x] Permission strategy disetujui user
- [x] Lanjut ke `PLAN_DA_01_Backend_Analytics_Contract_Foundation.md`
