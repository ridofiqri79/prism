# PLAN 05 — Green Book Module

> **Scope:** CRUD GB Project dengan 4 tabel nested (Activities, Funding Source, Disbursement Plan, Funding Allocation).
> **Deliverable:** Staff bisa input GB Project lengkap. Funding Allocation sinkron otomatis dengan Activities.
> **Referensi:** docs/PRISM_API_Contract.md (Green Book), docs/PRISM_Business_Rules.md (bagian 4)
> **Revision update:** Ikuti `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`. GB Project perlu `gb_project_identity_id`, picker BB Project harus default ke latest snapshot, dan detail perlu histori revisi.

---

## Task 1 — Types & Schema

**`src/types/green-book.types.ts`:**
```typescript
export interface GreenBook { id: string; publish_year: number; revision_number: number; status: 'active' | 'superseded' }
export interface GBProject { id: string; gb_code: string; project_name: string; program_title?: ProgramTitle; bb_projects: BBProjectSummary[]; executing_agencies: Institution[]; implementing_agencies: Institution[]; locations: Region[]; activities: GBActivity[]; funding_sources: GBFundingSource[]; disbursement_plan: GBDisbursementPlan[]; funding_allocations: GBFundingAllocation[]; status: 'active' | 'deleted' }
export interface GBActivity { id: string; activity_name: string; implementation_location?: string; piu?: string; sort_order: number }
export interface GBFundingSource { id: string; lender: Lender; institution?: Institution; loan_usd: number; grant_usd: number; local_usd: number }
export interface GBDisbursementPlan { id: string; year: number; amount_usd: number }
export interface GBFundingAllocation { id: string; gb_activity_id: string; services: number; constructions: number; goods: number; trainings: number; other: number }
```

**`src/schemas/green-book.schema.ts`:**
```typescript
export const gbProjectSchema = z.object({
  program_title_id: z.string().uuid(),
  gb_code: z.string().min(1),
  project_name: z.string().min(1),
  duration: z.string().optional(),
  objective: z.string().optional(),
  scope_of_project: z.string().optional(),
  bb_project_ids: z.array(z.string().uuid()).min(1, 'Minimal 1 BB Project'),
  executing_agency_ids: z.array(z.string().uuid()).min(1),
  implementing_agency_ids: z.array(z.string().uuid()).min(1),
  location_ids: z.array(z.string().uuid()).min(1),
})
```

---

## Task 2 — Service & Store

**`src/services/green-book.service.ts`** — semua endpoint dari API Contract.
**`src/stores/green-book.store.ts`** — state dan actions standar.

---

## Task 3 — composables/forms/useGBProjectForm.ts

```typescript
export function useGBProjectForm(initialData?) {
  const { handleSubmit, errors, values } = useForm({ validationSchema: toTypedSchema(gbProjectSchema), initialValues: initialData ?? {} })

  // Activities — rows dengan sort_order
  const activities = ref<ActivityRow[]>(initialData?.activities ?? [])
  const addActivity = () => activities.value.push({ activity_name: '', implementation_location: '', piu: '', sort_order: activities.value.length })
  const removeActivity = (i: number) => { activities.value.splice(i, 1); activities.value.forEach((a, idx) => a.sort_order = idx) }
  const reorderActivities = (from: number, to: number) => { /* swap + recalc sort_order */ }

  // Funding Source — per lender
  const fundingSources = ref<FundingSourceRow[]>(initialData?.funding_sources ?? [])
  const addFundingSource = () => fundingSources.value.push({ lender_id: '', institution_id: null, loan_usd: 0, grant_usd: 0, local_usd: 0 })
  const removeFundingSource = (i: number) => fundingSources.value.splice(i, 1)

  // Disbursement Plan — per tahun, auto-sort
  const disbursementPlan = ref<DisbursementRow[]>(initialData?.disbursement_plan ?? [])
  const addDisbursementYear = (year: number) => {
    if (disbursementPlan.value.find(d => d.year === year)) return  // no duplicate year
    disbursementPlan.value.push({ year, amount_usd: 0 })
    disbursementPlan.value.sort((a, b) => a.year - b.year)
  }

  // Funding Allocation — COMPUTED dari activities, one-to-one
  // PENTING: ini computed, bukan state — agent jangan override ini
  const fundingAllocations = computed(() =>
    activities.value.map((_, idx) => ({
      activity_index: idx,
      services: 0, constructions: 0, goods: 0, trainings: 0, other: 0,
    }))
  )
  // Tapi kita perlu bisa edit nilai allocation, jadi pakai ref yang di-sync:
  const allocationValues = ref<AllocationValues[]>([])
  watch(activities, (newActivities) => {
    // tambah row baru jika activity bertambah, hapus jika berkurang
    while (allocationValues.value.length < newActivities.length)
      allocationValues.value.push({ services: 0, constructions: 0, goods: 0, trainings: 0, other: 0 })
    allocationValues.value.length = newActivities.length
  }, { deep: true })

  const submit = handleSubmit(values => ({
    ...values,
    activities: activities.value,
    funding_sources: fundingSources.value,
    disbursement_plan: disbursementPlan.value,
    funding_allocations: activities.value.map((_, i) => ({ activity_index: i, ...allocationValues.value[i] })),
  }))

  return { values, errors, activities, addActivity, removeActivity, reorderActivities, fundingSources, addFundingSource, removeFundingSource, disbursementPlan, addDisbursementYear, allocationValues, submit }
}
```

---

## Task 4 — GBProjectFormPage.vue (5 Tab)

Gunakan PrimeVue `<TabView>`:

**Tab 1 — Informasi Umum:** program_title_id, gb_code, project_name, duration, objective, scope_of_project, bb_project_ids (MultiSelect), executing_agency_ids, implementing_agency_ids, location_ids

**Tab 2 — Activities:** `<ActivitiesTable>` — editable, drag reorder via `sortOrder` + PrimeVue OrderList atau manual up/down button

**Tab 3 — Funding Source:** `<FundingSourceTable>` — per lender, tampilkan total di footer

**Tab 4 — Disbursement Plan:** `<DisbursementPlanTable>` — input year + amount, auto-sort, duplicate year ditolak

**Tab 5 — Funding Allocation:** `<FundingAllocationTable>` — baris dari Tab 2, kolom = 5 kategori, input per sel. Readonly activity_name (dari activities), editable values.

Tombol "Simpan" di footer (bukan di dalam tab) — submit semua tab sekaligus.

---

## Task 5 — Komponen Tabel GB

**`ActivitiesTable.vue`** — tabel editable: activity_name (input), implementation_location (input), piu (input), tombol hapus, tombol up/down untuk reorder

**`FundingSourceTable.vue`** — tabel editable: `<LenderSelect>`, `<InstitutionSelect>`, `<CurrencyInput>` x3, tombol hapus. Footer: total loan, grant, local

**`DisbursementPlanTable.vue`** — tabel: year (input number), amount_usd (`<CurrencyInput>`), tombol hapus. Footer: grand total

**`FundingAllocationTable.vue`** — tabel: kolom pertama = activity_name (read-only dari activities), 5 kolom `<CurrencyInput>`, row footer total per kolom

---

## Task 6 — GBProjectDetailPage.vue

- Header info + status badge
- Referensi BB Projects (link ke BBProjectDetailPage)
- `<TabView>`: Activities, Funding Source, Disbursement Plan, Funding Allocation (semua read-only)
- Tombol Edit, Hapus

---

## Checklist

- [x] `green-book.types.ts`
- [x] `green-book.schema.ts`
- [x] `green-book.service.ts`
- [x] `green-book.store.ts`
- [x] `useGBProjectForm.ts` — activities/funding/disbursement/allocation sync via watch
- [x] `GBProjectFormPage.vue` — 5 tab
- [x] `GBProjectDetailPage.vue`
- [x] `ActivitiesTable.vue`, `FundingSourceTable.vue`, `DisbursementPlanTable.vue`, `FundingAllocationTable.vue`
