# AGENTS.md â€” PRISM Coding Agent Context

> File ini dibaca otomatis oleh Codex di setiap sesi.
> Baca seluruh file ini sebelum mengerjakan task apapun.

---

## 1. Tentang Proyek

**PRISM** (Project Loan Integrated Monitoring System) adalah sistem monitoring pinjaman luar negeri milik Bappenas. Sistem ini mencatat alur perencanaan pinjaman dari Blue Book â†’ Green Book â†’ Daftar Kegiatan â†’ Loan Agreement â†’ Monitoring Disbursement.

**Pengguna:** ADMIN (akses penuh) dan STAFF (akses per modul, dikonfigurasi ADMIN).

---

## 2. Stack & Versi

| Layer | Stack |
|-------|-------|
| Backend | Go Â· Echo Â· sqlc Â· pgx/v5 Â· golang-migrate |
| Frontend | Vue 3 (Composition API) Â· Vite Â· Pinia Â· PrimeVue 4 Â· Tailwind CSS v4 |
| Database | PostgreSQL 16 |
| Realtime | Server-Sent Events (SSE) |

---

## 3. Dokumen Referensi

Baca dokumen yang relevan sebelum mengerjakan task. Semua ada di folder `docs/`:

| Dokumen | Kapan dibaca |
|---------|-------------|
| `PRISM_Business_Rules.md` | **SELALU** â€” sebelum implementasi apapun |
| `prism_ddl.sql` | Sebelum menulis query atau model apapun |
| `PRISM_API_Contract.md` | Sebelum membuat endpoint atau service |
| `PRISM_BB_GB_Revision_Versioning_Plan.md` | Sebelum mengerjakan revisi Blue Book/Green Book, logical project, latest resolver, DK/LA frozen snapshot, atau journey history |
| `PRISM_Backend_Structure.md` | Task backend Go |
| `PRISM_Frontend_Structure.md` | Task frontend Vue |
| `PRISM_Error_Handling.md` | Saat implementasi error handling |
| `PRISM_Dev_Workflow.md` | Cara tambah migration, sqlc generate, dll. |

---

## 4. Struktur Folder

```
prism/
â”œâ”€â”€ docs/                   # Semua dokumen referensi
â”œâ”€â”€ plans/                  # Development plans per fase (PLAN_00 dst.)
â”œâ”€â”€ prism-backend/          # Go project
â”‚   â”œâ”€â”€ cmd/api/main.go
â”‚   â”œâ”€â”€ internal/
â”‚   â”‚   â”œâ”€â”€ config/
â”‚   â”‚   â”œâ”€â”€ database/queries/   # JANGAN EDIT â€” hasil generate sqlc
â”‚   â”‚   â”œâ”€â”€ handler/            # HTTP handler â€” parsing + response saja
â”‚   â”‚   â”œâ”€â”€ middleware/         # auth, permission, audit, logger
â”‚   â”‚   â”œâ”€â”€ model/              # Request/Response DTO
â”‚   â”‚   â”œâ”€â”€ service/            # Business logic
â”‚   â”‚   â”œâ”€â”€ sse/
â”‚   â”‚   â””â”€â”€ errors/
â”‚   â”œâ”€â”€ migrations/             # SQL migration files
â”‚   â””â”€â”€ sql/
â”‚       â”œâ”€â”€ queries/            # Source SQL untuk sqlc â€” EDIT DI SINI
â”‚       â””â”€â”€ schema/prism_ddl.sql
â””â”€â”€ prism-frontend/         # Vue 3 project
    â””â”€â”€ src/
        â”œâ”€â”€ assets/styles/      # main.css (Tailwind v4) + theme.ts
        â”œâ”€â”€ components/         # Reusable components
        â”œâ”€â”€ composables/        # useXxx + composables/forms/
        â”œâ”€â”€ pages/              # Halaman per modul
        â”œâ”€â”€ router/             # index.ts + routes/
        â”œâ”€â”€ schemas/            # Zod validation schemas
        â”œâ”€â”€ services/           # Axios API calls
        â”œâ”€â”€ stores/             # Pinia stores
        â””â”€â”€ types/              # TypeScript type definitions
```

---

## 5. Aturan Wajib â€” Backend (Go)

### 5.1 SQL dan sqlc

```
BENAR  â†’ Tulis query di sql/queries/<modul>.sql dengan annotation sqlc
SALAH  â†’ Tulis raw SQL string di file .go
SALAH  â†’ Edit file di internal/database/queries/ (hasil generate, akan tertimpa)
```

Format query sqlc:
```sql
-- name: GetBBProject :one
SELECT * FROM bb_project WHERE id = $1 AND status = 'active';

-- name: ListBBProjectByBlueBook :many
SELECT * FROM bb_project
WHERE blue_book_id = $1 AND status = 'active'
ORDER BY bb_code ASC;
```

Setelah ubah query: jalankan `make generate` (atau `sqlc generate`).

### 5.2 Layer Architecture

```
Router (permission check) â†’ Handler (parse + response) â†’ Service (business logic) â†’ sqlc Queries
```

- **Handler:** hanya `c.Bind(&req)`, call service, return JSON. Tidak ada logic bisnis.
- **Service:** semua validasi bisnis, transaksi DB, trigger SSE setelah commit.
- **Permission check:** di layer router (middleware), bukan di service atau handler.

### 5.3 Transaksi DB

Semua operasi yang mengubah lebih dari satu tabel WAJIB dalam transaksi:

```go
tx, err := s.db.Begin(ctx)
if err != nil { return nil, err }
defer tx.Rollback(ctx)

qtx := s.queries.WithTx(tx)
// ... operasi DB ...
tx.Commit(ctx)
```

### 5.4 Error Handling

```go
// BENAR â€” wrap error, log internal, return pesan aman ke client
if err != nil {
    log.Error().Err(err).Str("bb_code", req.BBCode).Msg("failed to create")
    return echo.NewHTTPError(http.StatusInternalServerError, "gagal menyimpan data")
}

// SALAH â€” expose internal error ke client
return c.JSON(500, err.Error())
```

PostgreSQL error code penting:
- `23505` unique_violation â†’ HTTP 409 CONFLICT
- `23503` foreign_key_violation â†’ HTTP 400 VALIDATION_ERROR
- `23514` check_violation â†’ HTTP 400 VALIDATION_ERROR

### 5.5 Audit Trail

Setiap request yang mengubah data, set user aktif di awal transaksi:
```go
conn.Exec(ctx, "SET LOCAL app.current_user_id = $1", user.ID)
```

Middleware `audit.go` sudah handle ini â€” jangan duplicate di service.

### 5.6 Urutan Menambah Fitur Backend

```
1. Cek skema di sql/schema/prism_ddl.sql
2. Tulis query di sql/queries/<modul>.sql
3. make generate
4. Buat/update struct di internal/model/<modul>.go
5. Implementasi service di internal/service/<modul>_service.go
6. Implementasi handler di internal/handler/<modul>_handler.go
7. Daftarkan route di cmd/api/main.go dengan permission middleware
```

### 5.7 Larangan Backend

| Larangan | Alasan |
|----------|--------|
| Edit `internal/database/queries/*.go` | Hasil generate, tertimpa saat `make generate` |
| Raw SQL di file `.go` | Semua SQL via sqlc |
| `interface{}` / `any` untuk data DB | Gunakan struct strongly-typed dari sqlc |
| Tambah dependency baru tanpa konfirmasi | Jaga dependency minimal |
| Return `err.Error()` ke client | Selalu wrap dengan pesan aman |
| Cek permission di service | Permission check hanya di router layer |
| Drop-recreate tabel di migration | Selalu `ALTER TABLE` â€” migration incremental |

---

## 6. Aturan Wajib â€” Frontend (Vue 3)

### 6.1 Tailwind v4 â€” BERBEDA dari v3

| Aspek | v3 (JANGAN) | v4 (BENAR) |
|-------|-------------|------------|
| Config | `tailwind.config.ts` | Tidak ada â€” semua di `main.css` via `@theme` |
| Import CSS | `@tailwind base/components/utilities` | `@import "tailwindcss"` |
| Vite | PostCSS plugin | `@tailwindcss/vite` Vite plugin |
| Custom token | `theme.extend` di JS | `@theme` directive di CSS |
| PrimeVue bridge | JS plugin di config | `@import "tailwindcss-primeui"` di CSS |
| cssLayer order | `tailwind-base, primevue, tailwind-utilities` | `theme, base, primevue` |

`main.css` yang benar:
```css
@import "tailwindcss";
@import "tailwindcss-primeui";

@theme {
  --font-sans: 'Inter Variable', ui-sans-serif, system-ui;
}

@layer base {
  body { @apply bg-surface-50 text-surface-900; }
}
```

`main.ts` yang benar:
```typescript
app.use(PrimeVue, {
  theme: {
    preset: prismPreset,
    options: {
      darkModeSelector: '.dark',
      cssLayer: { name: 'primevue', order: 'theme, base, primevue' },
    },
  },
})
```

### 6.2 Vue 3 Composition API

```vue
<!-- BENAR -->
<script setup lang="ts">
import { ref, computed } from 'vue'
</script>

<!-- SALAH â€” Options API tidak digunakan di proyek ini -->
<script lang="ts">
export default { data() { return {} } }
</script>
```

### 6.3 Pembagian Tanggung Jawab

```
types/       â†’ TypeScript interfaces, tidak ada logic
schemas/     â†’ Zod validation schemas
services/    â†’ Axios HTTP calls, tidak ada state
stores/      â†’ Pinia state + actions yang call service
composables/ â†’ Reusable logic (termasuk form state)
pages/       â†’ Gunakan store + composable, tidak ada HTTP langsung
components/  â†’ Presentasi saja, tidak ada store/service langsung
```

### 6.4 Aturan Kritis Frontend

```
BENAR  â†’ Semua API call melalui service (GreenBookService.getProject(...))
SALAH  â†’ axios.get(...) langsung di komponen atau store

BENAR  â†’ Types dari src/types/<modul>.types.ts
SALAH  â†’ Interface didefinisikan di dalam file .vue

BENAR  â†’ Validasi via Zod schema di src/schemas/
SALAH  â†’ Validasi manual (if (!form.field) errors.field = '...')

BENAR  â†’ Permission via can('module', 'action') dari usePermission()
SALAH  â†’ auth.user.role === 'ADMIN' langsung di template

BENAR  â†’ State global di Pinia store
SALAH  â†’ provide/inject untuk state global

BENAR  â†’ Form kompleks via composable di src/composables/forms/
SALAH  â†’ State tabel nested (activities, funding source) di halaman langsung
```

### 6.5 PrimeVue dan Tailwind â€” Pembagian Peran

| Keperluan | Gunakan |
|-----------|---------|
| Komponen UI (Button, Table, Dialog) | PrimeVue |
| Layout & spacing | Tailwind (`flex gap-4`, `p-6`, `grid`) |
| Warna semantik tema | `tailwindcss-primeui` (`bg-primary`, `text-surface-500`) |
| Custom design token | `@theme` di `main.css` |
| Style internal komponen PrimeVue | Pass-Through (PT) API atau `definePreset` |

**Jangan** override style PrimeVue dengan Tailwind class atau `!important`.

### 6.6 Urutan Menambah Fitur Frontend

```
1. Cek API Contract di docs/PRISM_API_Contract.md
2. Update types di src/types/<modul>.types.ts
3. Update Zod schema di src/schemas/<modul>.schema.ts
4. Update service di src/services/<modul>.service.ts
5. Update Pinia store di src/stores/<modul>.store.ts
6. Buat composable di src/composables/forms/ jika ada form kompleks
7. Buat komponen di src/components/<modul>/
8. Buat page di src/pages/<modul>/
9. Daftarkan route di src/router/routes/<modul>.routes.ts
```

### 6.7 Larangan Frontend

| Larangan | Alasan |
|----------|--------|
| Buat atau edit `tailwind.config.ts` | Tidak digunakan di Tailwind v4 |
| Buat `postcss.config.ts` | Tailwind v4 pakai Vite plugin |
| `@tailwind base/components/utilities` | Syntax v3 |
| `theme.extend` untuk custom token | Gunakan `@theme` di CSS |
| `cssLayer order: 'tailwind-base, ...'` | Itu v3, gunakan `'theme, base, primevue'` |
| Override PrimeVue via Tailwind class | Gunakan PT API atau `definePreset` |
| Options API | Proyek sepenuhnya Composition API |
| axios langsung di komponen | Melalui service |
| Interface di file `.vue` | Di `src/types/` |
| `any` sebagai tipe | Selalu strongly-typed |

---

## 7. Business Rules Ringkas

> Detail lengkap: `docs/PRISM_Business_Rules.md`

### Lender
- `Bilateral` & `KSA` â†’ `country_id` wajib; `Multilateral` â†’ `country_id` NULL
- Lender di DK hanya boleh dari `lender_indication` BB terkait ATAU `gb_funding_source` GB terkait
- Lender di LA harus dari `dk_financing_detail` DK Project terkait

### Blue Book
- Satu `active` per Period â€” revisi baru â†’ lama jadi `superseded`
- BB Project adalah snapshot per Blue Book/revisi dan dihubungkan lintas revisi oleh logical identity
- `bb_code` unik hanya dalam Blue Book yang sama; kode yang sama boleh muncul di revisi lain untuk logical project yang sama
- Mitra Kerja Bappenas: opsional, boleh lebih dari satu; simpan Eselon II saja, Eselon I diturunkan dari `parent_id`
- National Priority pada proyek Blue Book boleh menggunakan master National Priority dari period mana pun

### Green Book
- Satu `active` per `publish_year`
- GB Project adalah snapshot per Green Book/revisi dan dihubungkan lintas revisi oleh logical identity
- `gb_code` unik hanya dalam Green Book yang sama; kode yang sama boleh muncul di revisi lain untuk logical GB Project yang sama
- Saat GB dibuat/direvisi, relasi ke BB Project harus memakai versi BB Project terbaru
- GB Project wajib referensikan minimal 1 BB Project
- GB Project boleh mereferensikan lebih dari satu BB Project hanya jika seluruh BB Project berasal dari header Blue Book yang sama
- Satu BB Project boleh dipakai oleh lebih dari satu GB Project
- `gb_funding_allocation` CASCADE dengan `gb_activity` â€” selalu sinkron
- Disbursement Plan: total proyek per tahun (bukan per lender), `(gb_project_id, year)` unik

### Daftar Kegiatan
- **Final setelah diterbitkan** â€” backend cegah UPDATE kecuali ADMIN
- Saat DK dibuat, relasi ke GB Project harus memakai versi GB Project terbaru
- Setelah DK/LA dibuat, downstream tetap menunjuk snapshot yang dicantumkan saat DK dibuat dan tidak auto-pindah saat ada revisi BB/GB baru
- Activity Details: input bebas, tidak ada relasi teknis ke GB Activities

### Loan Agreement
- **One-to-One** dengan DK Project â€” tidak boleh ada LA kedua untuk DK yang sama
- `closing_date >= original_closing_date` (enforced DDL)
- `is_extended` dan `extension_days` adalah **computed**, tidak disimpan di DB
- Konversi mata uang: **manual oleh Staff** â€” sistem tidak auto-convert

### Monitoring Disbursement
- Hanya bisa dibuat jika `effective_date <= NOW()` â€” **backend wajib validasi**
- Triwulan: TW1 Apr-Jun, TW2 Jul-Sep, TW3 Okt-Des, TW4 Jan-Mar
- `(loan_agreement_id, budget_year, quarter)` unik
- Kurs diinput manual â€” sistem simpan 3 nilai (LA, USD, IDR) bersamaan tanpa auto-convert
- `absorption_pct = realized_usd / planned_usd * 100` â€” computed, bukan disimpan. Jika `planned = 0` â†’ hasil 0

### Region (Wilayah)
- Pilih COUNTRY â†’ otomatis mencakup seluruh PROVINCE â€” simpan hanya `region_id` COUNTRY di DB
- Frontend: nonaktifkan pilihan PROVINCE/CITY jika COUNTRY sudah dipilih

### Institution
- Satu institution **boleh** jadi EA sekaligus IA pada proyek yang sama bila sesuai data proyek

### Permission
- ADMIN: akses penuh, tidak ada entri di `user_permission`
- STAFF: **default deny** â€” tidak ada entri = tidak ada akses
- Permission dicek di **middleware** (backend) / **router guard + usePermission()** (frontend)

---

## 8. Skema Database â€” Tabel Utama

> Detail lengkap: `docs/prism_ddl.sql`

| Tabel | Keterangan |
|-------|-----------|
| `country` | Master negara (bukan `negara`) |
| `lender` | Master lender dengan `country_id`, `short_name` |
| `institution` | Hierarki Kementerian â†’ Eselon I, level: Kementerian/Badan/Lembaga, Eselon I, Eselon II, BUMN, Pemerintah Daerah Tk. I, Pemerintah Daerah Tk. II, BUMD, Lainya |
| `region` | Hierarki wilayah via `type` (COUNTRY/PROVINCE/CITY) + `parent_code` (bukan `parent_id`) |
| `blue_book` | Header BB, relasi ke `period` |
| `project_identity` | Identitas logical BB Project lintas revisi |
| `bb_project` | Snapshot proyek BB, `status: active/deleted`, `bb_code` unik per `blue_book_id` |
| `lender_indication` | Indikasi lender per BB project, field `remarks` (bukan `keterangan`) |
| `loi` | LoI per BB project, field `subject/date/letter_number` (bukan `perihal/tanggal/nomor_surat`) |
| `green_book` | Header GB, tidak pakai `period` |
| `gb_project_identity` | Identitas logical GB Project lintas revisi |
| `gb_project` | Snapshot proyek GB, many-to-many dengan BB via `gb_project_bb_project`, `gb_code` unik per `green_book_id` |
| `gb_activity` | Activities GB, field `sort_order` |
| `gb_funding_allocation` | Relasi ke `gb_activity` (bukan independen) |
| `daftar_kegiatan` | Header surat, field `subject/date/letter_number` |
| `dk_project` | Proyek dalam surat, many-to-many dengan GB |
| `dk_financing_detail` | Multi-currency: `currency`, `amount_original`, `amount_usd` |
| `dk_loan_allocation` | Multi-currency: sama dengan financing_detail |
| `dk_activity_detail` | Activity bebas, field `activity_number/activity_name` |
| `loan_agreement` | One-to-One dengan `dk_project`, field `loan_code/agreement_date/effective_date/currency` |
| `monitoring_disbursement` | Field `budget_year/quarter/exchange_rate_usd_idr/planned_*/realized_*` |
| `monitoring_komponen` | Field `component_name`, breakdown opsional |

**Kolom yang sering salah:**

| Salah (lama) | Benar (baru) |
|-------------|-------------|
| `negara` (tabel) | `country` |
| `wilayah` (tabel) | `region` |
| `negara_id` | `country_id` |
| `wilayah_id` | `region_id` |
| `keterangan` (lender_indication) | `remarks` |
| `perihal` (loi/dk) | `subject` |
| `tanggal` (loi/dk) | `date` |
| `nomor_surat` | `letter_number` |
| `kode_loan` | `loan_code` |
| `tanggal_agreement` | `agreement_date` |
| `tanggal_efektif` | `effective_date` |
| `mata_uang` | `currency` |
| `tahun_anggaran` | `budget_year` |
| `triwulan` | `quarter` |
| `kurs_usd_idr` | `exchange_rate_usd_idr` |
| `kurs_la_idr` | `exchange_rate_la_idr` |
| `rencana_la/usd/idr` | `planned_la/usd/idr` |
| `realisasi_la/usd/idr` | `realized_la/usd/idr` |
| `nama_komponen` | `component_name` |
| `nomor` (dk_activity_detail) | `activity_number` |
| `nama_aktivitas` | `activity_name` |

---

## 9. API Contract Ringkas

> Detail lengkap: `docs/PRISM_API_Contract.md`

Base URL: `/api/v1`
Auth: `Authorization: Bearer <token>`

| Modul | Endpoint prefix |
|-------|----------------|
| Auth | `/auth/login`, `/auth/logout`, `/auth/me` |
| Master | `/master/countries`, `/master/lenders`, `/master/institutions`, `/master/regions`, `/master/program-titles`, `/master/bappenas-partners`, `/master/periods`, `/master/national-priorities` |
| Blue Book | `/blue-books`, `/blue-books/:id/projects`, `/bb-projects/:id/loi` |
| Green Book | `/green-books`, `/green-books/:id/projects` |
| DK | `/daftar-kegiatan`, `/daftar-kegiatan/:id/projects` |
| LA | `/loan-agreements` |
| Monitoring | `/loan-agreements/:laId/monitoring` |
| Dashboard | `/dashboard/summary`, `/dashboard/monitoring-summary` |
| Journey | `/projects/:bbProjectId/journey` |
| User | `/users`, `/users/:id/permissions` |
| SSE | `/events` |

Error response format:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Pesan yang aman untuk user",
    "details": [{ "field": "bb_code", "message": "sudah digunakan" }]
  }
}
```

---

## 10. Development Plans

Kerjakan **satu plan per sesi**. Selesaikan semua task dan checklist sebelum pindah ke plan berikutnya.

**Urutan pengerjaan:** Frontend baseline (FE-00 -> FE-09) dikerjakan lebih dulu. Backend baseline (BE-00 -> BE-06) menyusul setelahnya. Fase revision versioning dikerjakan setelah baseline: BE-07 -> BE-11, lalu FE-10.

---

### 10.1 Frontend Plans

| Plan | File | Deliverable |
|------|------|-------------|
| **FE-00** | `plans/PLAN_00_Setup.md` | `vite.config.ts` (Tailwind v4 plugin), `main.css` (@import tailwindcss + tailwindcss-primeui), `theme.ts` (definePreset Aura), `main.ts` (cssLayer: 'theme, base, primevue'), router + semua route files (placeholder), `AppLayout`, `AppSidebar`, `AppTopbar`, `auth.store.ts`, `http.ts` (Axios interceptors) |
| **FE-01** | `plans/PLAN_01_Foundation.md` | Composables: `usePermission`, `usePagination`, `useToast`, `useConfirm`. Components: `PageHeader`, `DataTable`, `EmptyState`, `StatusBadge`, `CurrencyDisplay`, `ConfirmDialog`. Form components: `LocationMultiSelect`, `InstitutionSelect`, `LenderSelect`, `ProgramTitleSelect`, `NationalPriorityMultiSelect`, `CurrencyInput`. `master.store.ts` + `master.service.ts` + `master.types.ts` |
| **FE-02** | `plans/PLAN_02_Auth.md` | `LoginPage.vue` fungsional, router guard (requiresAuth + adminOnly), session restore dari localStorage, `UserListPage`, `UserFormPage`, `UserPermissionPage` (permission matrix checkbox), `ForbiddenPage`, `NotFoundPage` |
| **FE-03** | `plans/PLAN_03_Master_Data.md` | 8 halaman CRUD master: `CountryPage`, `LenderPage` (country_id conditional), `InstitutionPage` (TreeTable 8 level), `RegionPage` (TreeTable COUNTRY/PROVINCE/CITY), `ProgramTitlePage`, `BappenasPartnerPage`, `PeriodPage`, `NationalPriorityPage` (filter by period). `master.schema.ts` dengan Zod refine |
| **FE-04** | `plans/PLAN_04_Blue_Book.md` | `BlueBookListPage`, `BlueBookDetailPage`, `BBProjectFormPage` (5 section: info umum, pihak terlibat, lokasi+prioritas, project cost tabel, lender indication tabel), `BBProjectDetailPage`, LoI dialog, `useBBProjectForm.ts`, komponen: `ProjectCostTable`, `LenderIndicationTable`, `LoITable` |
| **FE-05** | `plans/PLAN_05_Green_Book.md` | `GBProjectFormPage` (5 tab: info umum, activities, funding source, disbursement plan, funding allocation), `useGBProjectForm.ts` (activitiesâ†”allocationValues sync via `watch`), komponen: `ActivitiesTable` (drag reorder), `FundingSourceTable`, `DisbursementPlanTable`, `FundingAllocationTable` (computed dari activities) |
| **FE-06** | `plans/PLAN_06_Daftar_Kegiatan.md` | `DKListPage`, `DKDetailPage` (accordion per proyek), `DKProjectFormPage` (4 section: header + financing multi-currency + loan allocation + activity details), `useDKProjectForm.ts` (`allowedLenderIds` computed dari GB funding source + BB lender indication) |
| **FE-07** | `plans/PLAN_07_Loan_Agreement.md` | `LAListPage` (filter is_extended, closing_date_before), `LAFormPage` (indikator perpanjangan real-time: `isExtended` + `extensionDays` computed), `LADetailPage`, `loan-agreement.schema.ts` (refine closing_date >= original_closing_date) |
| **FE-08** | `plans/PLAN_08_Monitoring.md` | `MonitoringListPage` (guard: disable tombol jika LA belum efektif), `MonitoringFormPage` (3 section: periode + rencana/realisasi tabel 3Ã—2 + komponen opsional), `useMonitoringForm.ts` (`absorptionPct` computed, div-by-zero safe), `AbsorptionBar` (color coding), `MonitoringCard`, `KomponenTable`, `MonitoringChart` (ECharts grouped bar) |
| **FE-09** | `plans/PLAN_09_Dashboard_Journey.md` | `DashboardPage` (summary cards + filter budget_year/quarter/lender + AbsorptionBar + MonitoringChart + tabel by-lender), `ProjectJourneyPage` (search BB + timeline), `ProjectTimeline.vue` (hierarki vertikal expand/collapse, node status: completed/pending/extended), `SummaryCard.vue` |

---

### 10.1a Revision Versioning Frontend Plan

| Plan | File | Deliverable |
|------|------|-------------|
| **FE-10** | `plans/PLAN_10_BB_GB_Revision_UI.md` | UI revision versioning: history section BB/GB, latest BB/GB pickers, DK concrete snapshot display, Journey newer-revision badge |

---

### 10.2 Backend Plans

| Plan | File | Deliverable |
|------|------|-------------|
| **BE-00** | `plans/PLAN_BE_00_Foundation.md` | `sqlc.yaml`, `Makefile`, `.air.toml`, `internal/config/config.go` (viper), `internal/database/db.go` (pgxpool 20 koneksi), `internal/errors/errors.go` (AppError + FromPgError pg code mapping), middleware: `logger.go`, `auth.go` (JWT), `permission.go` (stub), `audit.go` (SET LOCAL), `error_handler.go`. `internal/model/common.go` (generic ListResponse/DataResponse). `internal/sse/broker.go` (channel-based). `cmd/api/main.go` wiring + `/health` endpoint |
| **BE-01** | `plans/PLAN_BE_01_Auth.md` | `sql/queries/user.sql` (GetUserByUsername, GetUserByID, ListUsers, CreateUser, UpdateUser, GetUserPermissions, DeleteUserPermissions, CreateUserPermission, GetUserPermissionByModule), `make generate`, `internal/model/auth.go`, `internal/service/auth_service.go` (bcrypt + JWT sign), `internal/service/user_service.go` (UpdatePermissions replace-all transaksional), handler auth + user, `permission.go` middleware implementasi full (cek DB), routes terdaftar, seed ADMIN migration |
| **BE-02** | `plans/PLAN_BE_02_Master_Data.md` | `sql/queries/master.sql` (CRUD semua 8 tabel master, ListLenders dengan JOIN country), `make generate`, `internal/model/master.go`, `internal/service/master_service.go` (validasi lender: country_id wajib Bilateral/KSA, NULL Multilateral), `internal/handler/master_handler.go`, semua routes master terdaftar |
| **BE-03** | `plans/PLAN_BE_03_Blue_Book.md` | `sql/queries/bb_project.sql` (CRUD BB + BB Project snapshot + logical identity + junction tables institution/location/priority + costs + lender_indication + LoI + SupersedeBlueBooksByPeriod), `make generate`, `internal/model/blue_book.go`, `internal/service/blue_book_service.go` (validasi: bb_code unik per Blue Book, clone revisi, transaksi multi-tabel, SSE publish), handler + routes |
| **BE-04** | `plans/PLAN_BE_04_Green_Book.md` | `sql/queries/gb_project.sql` (CRUD GB + GB Project snapshot + logical identity + latest BB resolver + junction + activities ordered by sort_order + funding_source + UpsertGBDisbursementPlan + funding_allocation), `make generate`, `internal/service/green_book_service.go` (validasi: min 1 BB, gb_code unik per Green Book, tahun disbursement tidak duplikat, `activity_index` mapping ke `activityIDs[]` dalam transaksi), handler + routes |
| **BE-05** | `plans/PLAN_BE_05_DK_LA.md` | `sql/queries/dk_project.sql` (CRUD DK + latest GB resolver + frozen concrete snapshot junction + financing multi-currency + loan_allocation + activity_detail + `GetAllowedLenderIDsForDK` UNION query), `sql/queries/loan_agreement.sql` (CRUD LA + `GetAllowedLenderIDsForLA` + `GetLoanAgreementByDKProject`), `make generate`, service DK (validasi lender dari allowed set setelah GB relations tersimpan), service LA (cek one-to-one + validasi lender + `is_extended` computed + SSE `loan_agreement.extended`), handler + routes |
| **BE-06** | `plans/PLAN_BE_06_Monitoring.md` | `sql/queries/monitoring.sql` (CRUD monitoring + komponen + `GetMonitoringByLAAndPeriod` + `GetDashboardSummary` aggregate query + `GetMonitoringSummary` GROUP BY lender), `make generate`, `internal/service/monitoring_service.go` (guard: `effective_date <= NOW()`, cek duplikat quarter, `absorption_pct` computed div-by-zero safe), `internal/service/dashboard_service.go`, `internal/service/journey_service.go` (multi-level response assembly), handler monitoring + dashboard + journey, semua routes |

---

### 10.2a Revision Versioning Backend Plans

| Plan | File | Deliverable |
|------|------|-------------|
| **BE-07** | `plans/PLAN_BE_07_Revision_Versioning_Schema.md` | Schema, DDL/migration, sqlc queries, latest/history resolver, API contract awal untuk BB/GB revision versioning |
| **BE-08** | `plans/PLAN_BE_08_Blue_Book_Revision_Versioning.md` | Blue Book logical identity, duplicate per dokumen, clone revisi, BB history endpoint, import BB |
| **BE-09** | `plans/PLAN_BE_09_Green_Book_Revision_Versioning.md` | Green Book logical identity, latest BB resolver, clone revisi, GB history endpoint |
| **BE-10** | `plans/PLAN_BE_10_DK_LA_Frozen_Snapshot.md` | DK latest GB resolver, downstream frozen snapshot, lender validation berdasarkan concrete version |
| **BE-11** | `plans/PLAN_BE_11_Journey_Import_Project_List_Versioning.md` | Project list latest default, Journey concrete path, dashboard count safety, import final, backend smoke |

---

### 10.3 Urutan Pengerjaan & Dependensi

```
FE-00 â†’ FE-01 â†’ FE-02 â†’ FE-03 â†’ FE-04 â†’ FE-05 â†’ FE-06 â†’ FE-07 â†’ FE-08 â†’ FE-09
                                                                              â†“
BE-00 â†’ BE-01 â†’ BE-02 â†’ BE-03 â†’ BE-04 â†’ BE-05 â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€â”€ BE-06
```

**Dependensi kritis:**
- FE-01 harus selesai sebelum FE-02+ (semua modul pakai shared components)
- FE-02 harus selesai sebelum FE-03+ (semua modul pakai auth store + permission)
- BE-00 harus selesai sebelum BE-01+ (middleware + error handler dipakai semua)
- BE-01 harus selesai sebelum BE-02+ (permission middleware butuh tabel user_permission)
- BE-05 butuh BE-03 dan BE-04 selesai (GetAllowedLenderIDsForDK query gabungkan data BB + GB)
- Revision versioning dikerjakan setelah baseline BE-06: BE-07 -> BE-08 -> BE-09 -> BE-10 -> BE-11 -> FE-10.
- BE-08 butuh BE-07 selesai karena service BB bergantung pada schema/query identity.
- BE-09 butuh BE-08 selesai karena Green Book latest resolver bergantung pada BB identity/latest resolver.
- BE-10 butuh BE-09 selesai karena DK resolver bergantung pada GB identity/latest resolver.
- BE-11 butuh BE-08 sampai BE-10 selesai.
- FE-10 butuh BE-11 selesai dan `docs/PRISM_API_Contract.md` sudah sinkron.

### 10.4 Aturan Per Sesi

**Saat memulai plan baru:**
1. Baca file plan dari `plans/` folder
2. Pastikan checklist plan sebelumnya sudah semua ter-centang
3. Baca dokumen referensi yang disebutkan di bagian "Instruksi untuk Codex" plan tersebut

**Backend â€” urutan wajib dalam setiap plan:**
```
sql/queries/<modul>.sql â†’ make generate â†’ model â†’ service â†’ handler â†’ register route
```

**Frontend â€” urutan wajib dalam setiap plan:**
```
API Contract â†’ types â†’ schema â†’ service â†’ store â†’ composable â†’ component â†’ page â†’ route
```

**Business rules** dari `docs/PRISM_Business_Rules.md` selalu jadi referensi utama. Jangan implementasi logic yang bertentangan meski tidak disebutkan eksplisit dalam plan.

---

## 11. Docker â€” Environment

```bash
# Development
docker compose -f docker-compose.dev.yml up --build

# Service yang berjalan:
# PostgreSQL â†’ localhost:5432  (DDL diapply otomatis dari docs/prism_ddl.sql)
# Backend    â†’ http://localhost:8080  (Air hot reload)
# Frontend   â†’ http://localhost:5173  (Vite HMR)

# Perintah harian
docker compose -f docker-compose.dev.yml up -d          # start background
docker compose -f docker-compose.dev.yml down            # stop
docker compose -f docker-compose.dev.yml down -v         # reset DB
docker compose -f docker-compose.dev.yml logs -f frontend

# sqlc generate â€” SELALU dari lokal, bukan dari dalam Docker
cd prism-backend && sqlc generate
```

---

## 12. Git Convention

```bash
# Branch naming
feature/bb-project-form
fix/lender-validation
migration/add-short-name-to-lender
chore/update-dependencies

# Commit message (Conventional Commits)
feat: tambah endpoint monitoring disbursement
fix: validasi lender di DK tidak periksa lender indication BB
migration: tambah kolom short_name di tabel lender
docs: update API contract bagian journey
```

---

*AGENTS.md â€” PRISM v1.3 | 23 plans (11 frontend + 12 backend) | Dibaca otomatis oleh Codex setiap sesi*
