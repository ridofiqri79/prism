# AGENTS.md — PRISM Coding Agent Context

> File ini dibaca otomatis oleh Codex di setiap sesi.
> Baca seluruh file ini sebelum mengerjakan task apapun.

---

## 1. Tentang Proyek

**PRISM** (Project Loan Integrated Monitoring System) adalah sistem monitoring pinjaman luar negeri milik Bappenas. Sistem ini mencatat alur perencanaan pinjaman dari Blue Book → Green Book → Daftar Kegiatan → Loan Agreement → Monitoring Disbursement.

**Pengguna:** ADMIN (akses penuh) dan STAFF (akses per modul, dikonfigurasi ADMIN).

---

## 2. Stack & Versi

| Layer | Stack |
|-------|-------|
| Backend | Go · Echo · sqlc · pgx/v5 · golang-migrate |
| Frontend | Vue 3 (Composition API) · Vite · Pinia · PrimeVue 4 · Tailwind CSS v4 |
| Database | PostgreSQL 16 |
| Realtime | Server-Sent Events (SSE) |

---

## 3. Dokumen Referensi

Baca dokumen yang relevan sebelum mengerjakan task. Semua ada di folder `docs/`:

| Dokumen | Kapan dibaca |
|---------|-------------|
| `PRISM_Business_Rules.md` | **SELALU** — sebelum implementasi apapun |
| `prism_ddl.sql` | Sebelum menulis query atau model apapun |
| `PRISM_API_Contract.md` | Sebelum membuat endpoint atau service |
| `PRISM_Backend_Structure.md` | Task backend Go |
| `PRISM_Frontend_Structure.md` | Task frontend Vue |
| `PRISM_Error_Handling.md` | Saat implementasi error handling |
| `PRISM_Dev_Workflow.md` | Cara tambah migration, sqlc generate, dll. |

---

## 4. Struktur Folder

```
prism/
├── docs/                   # Semua dokumen referensi
├── plans/                  # Development plans per fase (PLAN_00 dst.)
├── prism-backend/          # Go project
│   ├── cmd/api/main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── database/queries/   # JANGAN EDIT — hasil generate sqlc
│   │   ├── handler/            # HTTP handler — parsing + response saja
│   │   ├── middleware/         # auth, permission, audit, logger
│   │   ├── model/              # Request/Response DTO
│   │   ├── service/            # Business logic
│   │   ├── sse/
│   │   └── errors/
│   ├── migrations/             # SQL migration files
│   └── sql/
│       ├── queries/            # Source SQL untuk sqlc — EDIT DI SINI
│       └── schema/prism_ddl.sql
└── prism-frontend/         # Vue 3 project
    └── src/
        ├── assets/styles/      # main.css (Tailwind v4) + theme.ts
        ├── components/         # Reusable components
        ├── composables/        # useXxx + composables/forms/
        ├── pages/              # Halaman per modul
        ├── router/             # index.ts + routes/
        ├── schemas/            # Zod validation schemas
        ├── services/           # Axios API calls
        ├── stores/             # Pinia stores
        └── types/              # TypeScript type definitions
```

---

## 5. Aturan Wajib — Backend (Go)

### 5.1 SQL dan sqlc

```
BENAR  → Tulis query di sql/queries/<modul>.sql dengan annotation sqlc
SALAH  → Tulis raw SQL string di file .go
SALAH  → Edit file di internal/database/queries/ (hasil generate, akan tertimpa)
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
Router (permission check) → Handler (parse + response) → Service (business logic) → sqlc Queries
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
// BENAR — wrap error, log internal, return pesan aman ke client
if err != nil {
    log.Error().Err(err).Str("bb_code", req.BBCode).Msg("failed to create")
    return echo.NewHTTPError(http.StatusInternalServerError, "gagal menyimpan data")
}

// SALAH — expose internal error ke client
return c.JSON(500, err.Error())
```

PostgreSQL error code penting:
- `23505` unique_violation → HTTP 409 CONFLICT
- `23503` foreign_key_violation → HTTP 400 VALIDATION_ERROR
- `23514` check_violation → HTTP 400 VALIDATION_ERROR

### 5.5 Audit Trail

Setiap request yang mengubah data, set user aktif di awal transaksi:
```go
conn.Exec(ctx, "SET LOCAL app.current_user_id = $1", user.ID)
```

Middleware `audit.go` sudah handle ini — jangan duplicate di service.

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
| Drop-recreate tabel di migration | Selalu `ALTER TABLE` — migration incremental |

---

## 6. Aturan Wajib — Frontend (Vue 3)

### 6.1 Tailwind v4 — BERBEDA dari v3

| Aspek | v3 (JANGAN) | v4 (BENAR) |
|-------|-------------|------------|
| Config | `tailwind.config.ts` | Tidak ada — semua di `main.css` via `@theme` |
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

<!-- SALAH — Options API tidak digunakan di proyek ini -->
<script lang="ts">
export default { data() { return {} } }
</script>
```

### 6.3 Pembagian Tanggung Jawab

```
types/       → TypeScript interfaces, tidak ada logic
schemas/     → Zod validation schemas
services/    → Axios HTTP calls, tidak ada state
stores/      → Pinia state + actions yang call service
composables/ → Reusable logic (termasuk form state)
pages/       → Gunakan store + composable, tidak ada HTTP langsung
components/  → Presentasi saja, tidak ada store/service langsung
```

### 6.4 Aturan Kritis Frontend

```
BENAR  → Semua API call melalui service (GreenBookService.getProject(...))
SALAH  → axios.get(...) langsung di komponen atau store

BENAR  → Types dari src/types/<modul>.types.ts
SALAH  → Interface didefinisikan di dalam file .vue

BENAR  → Validasi via Zod schema di src/schemas/
SALAH  → Validasi manual (if (!form.field) errors.field = '...')

BENAR  → Permission via can('module', 'action') dari usePermission()
SALAH  → auth.user.role === 'ADMIN' langsung di template

BENAR  → State global di Pinia store
SALAH  → provide/inject untuk state global

BENAR  → Form kompleks via composable di src/composables/forms/
SALAH  → State tabel nested (activities, funding source) di halaman langsung
```

### 6.5 PrimeVue dan Tailwind — Pembagian Peran

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
- `Bilateral` & `KSA` → `country_id` wajib; `Multilateral` → `country_id` NULL
- Lender di DK hanya boleh dari `lender_indication` BB terkait ATAU `gb_funding_source` GB terkait
- Lender di LA harus dari `dk_financing_detail` DK Project terkait

### Blue Book
- Satu `active` per Period — revisi baru → lama jadi `superseded`
- `bb_code` unik global, tidak bisa dipakai ulang meski `deleted`
- Bappenas Partner: simpan Eselon II saja, Eselon I diturunkan dari `parent_id`
- National Priority: filter berdasarkan `period_id` BB yang sama

### Green Book
- Satu `active` per `publish_year`
- GB Project wajib referensikan minimal 1 BB Project
- `gb_funding_allocation` CASCADE dengan `gb_activity` — selalu sinkron
- Disbursement Plan: total proyek per tahun (bukan per lender), `(gb_project_id, year)` unik

### Daftar Kegiatan
- **Final setelah diterbitkan** — backend cegah UPDATE kecuali ADMIN
- Activity Details: input bebas, tidak ada relasi teknis ke GB Activities

### Loan Agreement
- **One-to-One** dengan DK Project — tidak boleh ada LA kedua untuk DK yang sama
- `closing_date >= original_closing_date` (enforced DDL)
- `is_extended` dan `extension_days` adalah **computed**, tidak disimpan di DB
- Konversi mata uang: **manual oleh Staff** — sistem tidak auto-convert

### Monitoring Disbursement
- Hanya bisa dibuat jika `effective_date <= NOW()` — **backend wajib validasi**
- Triwulan: TW1 Apr-Jun, TW2 Jul-Sep, TW3 Okt-Des, TW4 Jan-Mar
- `(loan_agreement_id, budget_year, quarter)` unik
- Kurs diinput manual — sistem simpan 3 nilai (LA, USD, IDR) bersamaan tanpa auto-convert
- `absorption_pct = realized_usd / planned_usd * 100` — computed, bukan disimpan. Jika `planned = 0` → hasil 0

### Region (Wilayah)
- Pilih COUNTRY → otomatis mencakup seluruh PROVINCE — simpan hanya `region_id` COUNTRY di DB
- Frontend: nonaktifkan pilihan PROVINCE/CITY jika COUNTRY sudah dipilih

### Institution
- Satu institution **tidak boleh** jadi EA sekaligus IA pada proyek yang sama

### Permission
- ADMIN: akses penuh, tidak ada entri di `user_permission`
- STAFF: **default deny** — tidak ada entri = tidak ada akses
- Permission dicek di **middleware** (backend) / **router guard + usePermission()** (frontend)

---

## 8. Skema Database — Tabel Utama

> Detail lengkap: `docs/prism_ddl.sql`

| Tabel | Keterangan |
|-------|-----------|
| `country` | Master negara (bukan `negara`) |
| `lender` | Master lender dengan `country_id`, `short_name` |
| `institution` | Hierarki Kementerian → Eselon I, level: Kementerian/Badan/Lembaga, Eselon I, BUMN, Pemerintah Daerah, BUMD, Lainnya |
| `region` | Hierarki wilayah via `type` (COUNTRY/PROVINCE/CITY) + `parent_code` (bukan `parent_id`) |
| `blue_book` | Header BB, relasi ke `period` |
| `bb_project` | Proyek BB, `status: active/deleted`, `bb_code` UNIQUE |
| `lender_indication` | Indikasi lender per BB project, field `remarks` (bukan `keterangan`) |
| `loi` | LoI per BB project, field `subject/date/letter_number` (bukan `perihal/tanggal/nomor_surat`) |
| `green_book` | Header GB, tidak pakai `period` |
| `gb_project` | Proyek GB, many-to-many dengan BB via `gb_project_bb_project` |
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

**Urutan pengerjaan:** Frontend (FE-00 → FE-09) dikerjakan lebih dulu. Backend (BE-00 → BE-06) menyusul setelahnya.

---

### 10.1 Frontend Plans

| Plan | File | Deliverable |
|------|------|-------------|
| **FE-00** | `plans/PLAN_00_Setup.md` | `vite.config.ts` (Tailwind v4 plugin), `main.css` (@import tailwindcss + tailwindcss-primeui), `theme.ts` (definePreset Aura), `main.ts` (cssLayer: 'theme, base, primevue'), router + semua route files (placeholder), `AppLayout`, `AppSidebar`, `AppTopbar`, `auth.store.ts`, `http.ts` (Axios interceptors) |
| **FE-01** | `plans/PLAN_01_Foundation.md` | Composables: `usePermission`, `usePagination`, `useToast`, `useConfirm`. Components: `PageHeader`, `DataTable`, `EmptyState`, `StatusBadge`, `CurrencyDisplay`, `ConfirmDialog`. Form components: `LocationMultiSelect`, `InstitutionSelect`, `LenderSelect`, `ProgramTitleSelect`, `NationalPriorityMultiSelect`, `CurrencyInput`. `master.store.ts` + `master.service.ts` + `master.types.ts` |
| **FE-02** | `plans/PLAN_02_Auth.md` | `LoginPage.vue` fungsional, router guard (requiresAuth + adminOnly), session restore dari localStorage, `UserListPage`, `UserFormPage`, `UserPermissionPage` (permission matrix checkbox), `ForbiddenPage`, `NotFoundPage` |
| **FE-03** | `plans/PLAN_03_Master_Data.md` | 8 halaman CRUD master: `CountryPage`, `LenderPage` (country_id conditional), `InstitutionPage` (TreeTable 6 level), `RegionPage` (TreeTable COUNTRY/PROVINCE/CITY), `ProgramTitlePage`, `BappenasPartnerPage`, `PeriodPage`, `NationalPriorityPage` (filter by period). `master.schema.ts` dengan Zod refine |
| **FE-04** | `plans/PLAN_04_Blue_Book.md` | `BlueBookListPage`, `BlueBookDetailPage`, `BBProjectFormPage` (5 section: info umum, pihak terlibat, lokasi+prioritas, project cost tabel, lender indication tabel), `BBProjectDetailPage`, LoI dialog, `useBBProjectForm.ts`, komponen: `ProjectCostTable`, `LenderIndicationTable`, `LoITable` |
| **FE-05** | `plans/PLAN_05_Green_Book.md` | `GBProjectFormPage` (5 tab: info umum, activities, funding source, disbursement plan, funding allocation), `useGBProjectForm.ts` (activities↔allocationValues sync via `watch`), komponen: `ActivitiesTable` (drag reorder), `FundingSourceTable`, `DisbursementPlanTable`, `FundingAllocationTable` (computed dari activities) |
| **FE-06** | `plans/PLAN_06_Daftar_Kegiatan.md` | `DKListPage`, `DKDetailPage` (accordion per proyek), `DKProjectFormPage` (4 section: header + financing multi-currency + loan allocation + activity details), `useDKProjectForm.ts` (`allowedLenderIds` computed dari GB funding source + BB lender indication) |
| **FE-07** | `plans/PLAN_07_Loan_Agreement.md` | `LAListPage` (filter is_extended, closing_date_before), `LAFormPage` (indikator perpanjangan real-time: `isExtended` + `extensionDays` computed), `LADetailPage`, `loan-agreement.schema.ts` (refine closing_date >= original_closing_date) |
| **FE-08** | `plans/PLAN_08_Monitoring.md` | `MonitoringListPage` (guard: disable tombol jika LA belum efektif), `MonitoringFormPage` (3 section: periode + rencana/realisasi tabel 3×2 + komponen opsional), `useMonitoringForm.ts` (`absorptionPct` computed, div-by-zero safe), `AbsorptionBar` (color coding), `MonitoringCard`, `KomponenTable`, `MonitoringChart` (ECharts grouped bar) |
| **FE-09** | `plans/PLAN_09_Dashboard_Journey.md` | `DashboardPage` (summary cards + filter budget_year/quarter/lender + AbsorptionBar + MonitoringChart + tabel by-lender), `ProjectJourneyPage` (search BB + timeline), `ProjectTimeline.vue` (hierarki vertikal expand/collapse, node status: completed/pending/extended), `SummaryCard.vue` |

---

### 10.2 Backend Plans

| Plan | File | Deliverable |
|------|------|-------------|
| **BE-00** | `plans/PLAN_BE_00_Foundation.md` | `sqlc.yaml`, `Makefile`, `.air.toml`, `internal/config/config.go` (viper), `internal/database/db.go` (pgxpool 20 koneksi), `internal/errors/errors.go` (AppError + FromPgError pg code mapping), middleware: `logger.go`, `auth.go` (JWT), `permission.go` (stub), `audit.go` (SET LOCAL), `error_handler.go`. `internal/model/common.go` (generic ListResponse/DataResponse). `internal/sse/broker.go` (channel-based). `cmd/api/main.go` wiring + `/health` endpoint |
| **BE-01** | `plans/PLAN_BE_01_Auth.md` | `sql/queries/user.sql` (GetUserByUsername, GetUserByID, ListUsers, CreateUser, UpdateUser, GetUserPermissions, DeleteUserPermissions, CreateUserPermission, GetUserPermissionByModule), `make generate`, `internal/model/auth.go`, `internal/service/auth_service.go` (bcrypt + JWT sign), `internal/service/user_service.go` (UpdatePermissions replace-all transaksional), handler auth + user, `permission.go` middleware implementasi full (cek DB), routes terdaftar, seed ADMIN migration |
| **BE-02** | `plans/PLAN_BE_02_Master_Data.md` | `sql/queries/master.sql` (CRUD semua 8 tabel master, ListLenders dengan JOIN country), `make generate`, `internal/model/master.go`, `internal/service/master_service.go` (validasi lender: country_id wajib Bilateral/KSA, NULL Multilateral), `internal/handler/master_handler.go`, semua routes master terdaftar |
| **BE-03** | `plans/PLAN_BE_03_Blue_Book.md` | `sql/queries/bb_project.sql` (CRUD BB + BB Project + junction tables institution/location/priority + costs + lender_indication + LoI + SupersedeBlueBooksByPeriod), `make generate`, `internal/model/blue_book.go`, `internal/service/blue_book_service.go` (validasi: bb_code unik global termasuk deleted, EA≠IA, transaksi multi-tabel, SSE publish), handler + routes |
| **BE-04** | `plans/PLAN_BE_04_Green_Book.md` | `sql/queries/gb_project.sql` (CRUD GB + junction + activities ordered by sort_order + funding_source + UpsertGBDisbursementPlan + funding_allocation), `make generate`, `internal/service/green_book_service.go` (validasi: min 1 BB, EA≠IA, tahun disbursement tidak duplikat, `activity_index` mapping ke `activityIDs[]` dalam transaksi), handler + routes |
| **BE-05** | `plans/PLAN_BE_05_DK_LA.md` | `sql/queries/dk_project.sql` (CRUD DK + junction + financing multi-currency + loan_allocation + activity_detail + `GetAllowedLenderIDsForDK` UNION query), `sql/queries/loan_agreement.sql` (CRUD LA + `GetAllowedLenderIDsForLA` + `GetLoanAgreementByDKProject`), `make generate`, service DK (validasi lender dari allowed set setelah GB relations tersimpan), service LA (cek one-to-one + validasi lender + `is_extended` computed + SSE `loan_agreement.extended`), handler + routes |
| **BE-06** | `plans/PLAN_BE_06_Monitoring.md` | `sql/queries/monitoring.sql` (CRUD monitoring + komponen + `GetMonitoringByLAAndPeriod` + `GetDashboardSummary` aggregate query + `GetMonitoringSummary` GROUP BY lender), `make generate`, `internal/service/monitoring_service.go` (guard: `effective_date <= NOW()`, cek duplikat quarter, `absorption_pct` computed div-by-zero safe), `internal/service/dashboard_service.go`, `internal/service/journey_service.go` (multi-level response assembly), handler monitoring + dashboard + journey, semua routes |

---

### 10.3 Urutan Pengerjaan & Dependensi

```
FE-00 → FE-01 → FE-02 → FE-03 → FE-04 → FE-05 → FE-06 → FE-07 → FE-08 → FE-09
                                                                              ↓
BE-00 → BE-01 → BE-02 → BE-03 → BE-04 → BE-05 ─────────────────────────── BE-06
```

**Dependensi kritis:**
- FE-01 harus selesai sebelum FE-02+ (semua modul pakai shared components)
- FE-02 harus selesai sebelum FE-03+ (semua modul pakai auth store + permission)
- BE-00 harus selesai sebelum BE-01+ (middleware + error handler dipakai semua)
- BE-01 harus selesai sebelum BE-02+ (permission middleware butuh tabel user_permission)
- BE-05 butuh BE-03 dan BE-04 selesai (GetAllowedLenderIDsForDK query gabungkan data BB + GB)

### 10.4 Aturan Per Sesi

**Saat memulai plan baru:**
1. Baca file plan dari `plans/` folder
2. Pastikan checklist plan sebelumnya sudah semua ter-centang
3. Baca dokumen referensi yang disebutkan di bagian "Instruksi untuk Codex" plan tersebut

**Backend — urutan wajib dalam setiap plan:**
```
sql/queries/<modul>.sql → make generate → model → service → handler → register route
```

**Frontend — urutan wajib dalam setiap plan:**
```
API Contract → types → schema → service → store → composable → component → page → route
```

**Business rules** dari `docs/PRISM_Business_Rules.md` selalu jadi referensi utama. Jangan implementasi logic yang bertentangan meski tidak disebutkan eksplisit dalam plan.

---

## 11. Docker — Environment

```bash
# Development
docker compose -f docker-compose.dev.yml up --build

# Service yang berjalan:
# PostgreSQL → localhost:5432  (DDL diapply otomatis dari docs/prism_ddl.sql)
# Backend    → http://localhost:8080  (Air hot reload)
# Frontend   → http://localhost:5173  (Vite HMR)

# Perintah harian
docker compose -f docker-compose.dev.yml up -d          # start background
docker compose -f docker-compose.dev.yml down            # stop
docker compose -f docker-compose.dev.yml down -v         # reset DB
docker compose -f docker-compose.dev.yml logs -f frontend

# sqlc generate — SELALU dari lokal, bukan dari dalam Docker
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

*AGENTS.md — PRISM v1.2 | 17 plans (10 frontend + 7 backend) | Dibaca otomatis oleh Codex setiap sesi*
