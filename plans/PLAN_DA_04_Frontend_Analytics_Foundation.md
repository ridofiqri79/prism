# PLAN DA-04 - Frontend Analytics Foundation

> **Scope:** Menyiapkan fondasi frontend Dashboard Analytics: types, service, composable/store, layout, filter bar, dan shared components.
> **Deliverable:** Frontend bisa consume endpoint analytics tanpa placeholder dan punya shell UI stabil untuk phase berikutnya.
> **Dependencies:** `PLAN_DA_01_Backend_Analytics_Contract_Foundation.md`, minimal route backend tersedia.

---

## Instruksi Untuk Agent

Baca sebelum mulai:

- `AGENTS.md`
- `docs/PRISM_Business_Rules.md`
- `docs/PRISM_API_Contract.md`
- `docs/PRISM_Frontend_Structure.md`
- `plans/PLAN_DA_00_Dashboard_Analytics_Roadmap.md`
- `plans/PLAN_DA_01_Backend_Analytics_Contract_Foundation.md`

Aturan frontend:

- Semua API call lewat service.
- Types di `src/types/`, bukan di file `.vue`.
- State dan fetch logic di store/composable, bukan langsung di komponen presentasi.
- Gunakan Composition API.
- Gunakan PrimeVue untuk komponen UI.
- Gunakan Tailwind v4 sesuai setup repo.
- Jangan pakai placeholder data.
- Label UI pakai "Kementerian/Lembaga", bukan "KL".

---

## Task 1 - Types

Update:

```text
prism-frontend/src/types/dashboard.types.ts
```

Tambahkan types:

- `DashboardAnalyticsFilterParams`
- `DashboardAnalyticsOverview`
- `DashboardInstitutionAnalytics`
- `DashboardLenderAnalytics`
- `DashboardAbsorptionAnalytics`
- `DashboardYearlyAnalytics`
- `DashboardLenderProportionAnalytics`
- `DashboardRiskAnalytics`
- `DashboardDrilldownQuery`
- reusable `AnalyticsMoneyMetric`, `AnalyticsRankedItem`, `AnalyticsStageBreakdown`.

Rules:

- Reuse `Lender`, `Institution`, `Region` dari `master.types.ts` jika tersedia.
- Jangan definisikan `interface Lender` lokal di file dashboard.
- Hindari `any`; untuk drilldown query pakai `Record<string, string[]>` agar selaras dengan query parameter backend.

---

## Task 2 - Service

Update:

```text
prism-frontend/src/services/dashboard.service.ts
```

Tambahkan method:

```ts
getAnalyticsOverview(params?: DashboardAnalyticsFilterParams)
getAnalyticsInstitutions(params?: DashboardAnalyticsFilterParams)
getAnalyticsLenders(params?: DashboardAnalyticsFilterParams)
getAnalyticsAbsorption(params?: DashboardAnalyticsFilterParams)
getAnalyticsYearly(params?: DashboardAnalyticsFilterParams)
getAnalyticsLenderProportion(params?: DashboardAnalyticsFilterParams)
getAnalyticsRisks(params?: DashboardAnalyticsFilterParams)
```

Rules:

- Semua response dinormalisasi jika backend masih punya alias field lama.
- Jangan silent fallback ke angka palsu. Jika data kosong, return array kosong dan summary 0.

---

## Task 3 - Composable/Store

Pilih salah satu pola yang paling sesuai dengan repo:

1. `src/stores/dashboard.store.ts`, atau
2. `src/composables/useDashboardAnalytics.ts`.

Tanggung jawab:

- menyimpan filter draft/applied,
- fetch data per tab,
- refresh semua section ketika filter berubah,
- expose loading/error per section,
- cache master data yang dibutuhkan,
- expose `applyFilters`, `resetFilters`, `openDrilldown`.

Jangan fetch endpoint analytics langsung dari `DashboardPage.vue`.

---

## Task 4 - Filter Bar

Buat component:

```text
prism-frontend/src/components/dashboard/DashboardAnalyticsFilterBar.vue
```

Controls:

- Tahun Anggaran,
- Triwulan,
- Lender multi-select,
- Tipe Lender multi-select: Bilateral, Multilateral, KSA,
- Kementerian/Lembaga multi-select,
- Status Pipeline multi-select,
- Status Project multi-select,
- Wilayah multi-select,
- Program Title multi-select,
- range nilai pinjaman USD,
- toggle "Tampilkan histori revisi".

Rules:

- Filter advanced boleh collapsible.
- Jangan menampilkan ID mentah.
- Opsi Kementerian/Lembaga harus label manusiawi.
- Toggle histori harus jelas sebagai mode audit/history.

---

## Task 5 - Shared Components

Buat component presentasi minimal:

```text
src/components/dashboard/AnalyticsMetricGrid.vue
src/components/dashboard/AnalyticsMetricCard.vue
src/components/dashboard/AnalyticsBreakdownTable.vue
src/components/dashboard/AnalyticsEmptyState.vue
src/components/dashboard/AnalyticsDrilldownButton.vue
src/components/dashboard/AnalyticsStatusBadge.vue
```

Jika component common yang sudah ada cukup, reuse:

- `SummaryCard`
- `CurrencyDisplay`
- `StatusBadge`
- `EmptyState`
- `AbsorptionBar`
- `MonitoringChart`

Rules:

- Jangan membuat card di dalam card.
- Table harus readable untuk data panjang.
- Angka USD jangan disingkat jika konteksnya membutuhkan nilai penuh.
- Text tidak boleh overflow di mobile.

---

## Task 6 - Dashboard Layout Shell

Update:

```text
prism-frontend/src/pages/dashboard/DashboardPage.vue
```

Tambahkan layout analytics dengan tab/section:

- Portfolio,
- Kementerian/Lembaga,
- Lender,
- Penyerapan,
- Tahunan,
- Risiko & Data Quality.

Phase ini cukup menyiapkan struktur dan wiring data kosong/real dari endpoint foundation. Jangan membangun semua chart final dulu.

---

## Task 7 - Drilldown Navigation Helper

Implement helper:

```text
src/composables/useAnalyticsDrilldown.ts
```

Target mapping:

| Target backend | Route frontend |
|----------------|----------------|
| `projects` | Project Master |
| `monitoring` | Monitoring List/Overview |
| `loan_agreements` | Loan Agreement List |
| `spatial_distribution` | Sebaran Wilayah |

Rules:

- Preserve query dari backend.
- Query yang belum didukung target harus diabaikan dengan aman atau dicatat sebagai follow-up plan berikutnya.
- Jangan hardcode ID contoh.

---

## Acceptance Criteria

- Types analytics lengkap dan dipakai service.
- Service analytics tersedia dan tidak dipanggil langsung dari component page.
- Filter bar reusable dan bisa apply/reset.
- Dashboard page punya tab/section analytics foundation.
- Empty/loading/error states tampil benar.
- Tidak ada placeholder sample data.
- `npm.cmd run type-check` berhasil.
- `npm.cmd run build` berhasil.

---

## Verification

```powershell
cd prism-frontend
npm.cmd run type-check
npm.cmd run build
```

Jika backend live tersedia, smoke halaman `/dashboard` di browser.

---

## Checklist

- [ ] `dashboard.types.ts` update analytics types
- [ ] `dashboard.service.ts` update analytics methods
- [ ] Store/composable dashboard analytics dibuat
- [ ] Filter bar dibuat
- [ ] Shared analytics components dibuat/reuse
- [ ] Dashboard layout tab/section dibuat
- [ ] Drilldown helper dibuat
- [ ] Loading/empty/error states tersedia
- [ ] `npm.cmd run type-check` berhasil
- [ ] `npm.cmd run build` berhasil
