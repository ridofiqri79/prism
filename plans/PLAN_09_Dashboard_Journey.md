# PLAN 09 — Dashboard & Project Journey

> **Scope:** Dashboard agregat monitoring dan visualisasi alur proyek BB → Monitoring.
> **Deliverable:** Halaman dashboard informatif dan journey timeline per proyek.
> **Referensi:** docs/PRISM_API_Contract.md (Dashboard & Aggregasi)
> **Revision update:** Ikuti `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`. Journey harus menampilkan concrete snapshot path yang dipakai downstream dan memberi indikator jika ada BB/GB revisi lebih baru.

---

## Task 1 — Types

**`src/types/dashboard.types.ts`:**
```typescript
export interface DashboardSummary {
  total_bb_projects: number
  total_gb_projects: number
  total_loan_agreements: number
  total_amount_usd: number
  total_realisasi_usd: number
  overall_absorption_pct: number
  active_monitoring: number
}

export interface MonitoringSummary {
  tahun_anggaran: number
  triwulan: string
  total_rencana_usd: number
  total_realisasi_usd: number
  absorption_pct: number
  by_lender: LenderSummary[]
}

export interface LenderSummary {
  lender: Lender
  rencana_usd: number
  realisasi_usd: number
  absorption_pct: number
}

export interface JourneyResponse {
  bb_project: BBProjectSummary
  loi: LoI[]
  gb_projects: GBProjectJourney[]
}

export interface GBProjectJourney {
  id: string; gb_code: string; project_name: string
  dk_projects: DKProjectJourney[]
}

export interface DKProjectJourney {
  id: string; objectives: string
  daftar_kegiatan: { subject: string; date: string }
  loan_agreement: LAJourney | null
}

export interface LAJourney {
  id: string; loan_code: string; effective_date: string
  closing_date: string; is_extended: boolean
  monitoring: MonitoringSummaryItem[]
}
```

---

## Task 2 — Service

**`src/services/dashboard.service.ts`:**
```typescript
export const DashboardService = {
  getSummary: () => http.get<ApiResponse<DashboardSummary>>('/dashboard/summary').then(r => r.data.data),
  getMonitoringSummary: (params: { budget_year?: number; quarter?: string; lender_id?: string }) =>
    http.get<ApiResponse<MonitoringSummary>>('/dashboard/monitoring-summary', { params }).then(r => r.data.data),
  getJourney: (bbProjectId: string) =>
    http.get<ApiResponse<JourneyResponse>>(`/projects/${bbProjectId}/journey`).then(r => r.data.data),
}
```

---

## Task 3 — src/components/common/SummaryCard.vue

```vue
<template>
  <div class="bg-surface-0 rounded-xl p-5 border border-surface-200 shadow-sm">
    <p class="text-sm text-muted-color">{{ label }}</p>
    <p class="text-2xl font-semibold text-surface-900 mt-1">{{ formattedValue }}</p>
    <p v-if="unit" class="text-xs text-muted-color mt-0.5">{{ unit }}</p>
  </div>
</template>
```

Props: `label: string`, `value: number | string`, `unit?: string`, `format?: 'number' | 'currency' | 'percent'`

---

## Task 4 — DashboardPage.vue

**Baris 1 — Summary Cards (7 card, grid 4 kolom):**
- Total BB Projects, Total GB Projects, Total Loan Agreements
- Total Amount (USD), Total Realisasi (USD), Overall Absorption (%), Active Monitoring

**Baris 2 — Filter Bar:**
- budget_year (Select tahun), quarter (Select TW1–TW4), lender_id (`<LenderSelect>`)
- Tombol "Terapkan Filter"

**Baris 3 — Monitoring Overview:**
- Kiri: `<AbsorptionBar>` besar + angka rencana vs realisasi (USD)
- Kanan: `<MonitoringChart>` — bar chart dari monitoring data yang terfilter

**Baris 4 — Tabel Breakdown by Lender:**
Kolom: lender name, type badge, rencana_usd (`<CurrencyDisplay>`), realisasi_usd, absorption_pct (`<AbsorptionBar compact>`)

---

## Task 5 — ProjectJourneyPage.vue

Input: BB Project ID dari route param `/journey/:bbProjectId`.

Layout: search bar di atas + timeline vertikal.

**Search BB Project:**
- Autocomplete search by bb_code atau project_name
- Setelah pilih → load journey dari API

**`<ProjectTimeline :journey="journeyData">`:**

```
┌─ [Blue Book]
│   └─ BB-2025-001 — Trans Sumatra Toll Road
│       ├─ Lender Indication: JICA, ADB
│       └─ LoI: JICA (10 Mar 2025) ✓
│
└─ [Green Book]
    └─ GB-2025-001 — Trans Sumatra Section 1
        ├─ Funding: JICA USD 300M, ADB USD 100M
        └─ [Daftar Kegiatan: B-001/2025]
            └─ [Loan Agreement: IP-603]
                ├─ Effective: 1 Jun 2025
                ├─ Closing: 31 Des 2030
                └─ [Monitoring]
                    ├─ TW1 2025: 84% ✓
                    └─ TW2 2025: 76% ✓
```

---

## Task 6 — src/components/journey/ProjectTimeline.vue

- Props: `journey: JourneyResponse`
- Render hierarki vertikal dengan connector line (border-left style)
- Setiap node: icon + label + summary chip + expand/collapse button
- Node status warna:
  - Completed (ada data) → `text-green-600`
  - Pending (belum ada data) → `text-surface-400` (abu-abu, italic "Belum ada")
  - Extended LA → `text-orange-500`
- Node yang bisa diklik: link ke halaman detail masing-masing
- Expand/collapse sub-tree

---

## Checklist

- [x] `dashboard.types.ts` — DashboardSummary, MonitoringSummary, JourneyResponse
- [x] `dashboard.service.ts` — getSummary, getMonitoringSummary, getJourney
- [x] `SummaryCard.vue`
- [x] `DashboardPage.vue` — cards + filter + chart + tabel
- [x] `ProjectJourneyPage.vue` — search + timeline
- [x] `ProjectTimeline.vue` — hierarki vertikal dengan expand/collapse
