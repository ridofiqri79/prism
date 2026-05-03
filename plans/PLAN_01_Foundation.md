# PLAN 01 — Shared Components & Utilities

> **Scope:** Semua komponen, composables, stores, dan services yang dipakai ulang di semua modul.
> **Deliverable:** Library komponen internal siap dipakai di Plan 02+.
> **Referensi:** docs/PRISM_Frontend_Structure.md, docs/PRISM_Error_Handling.md

---

## Task 1 — Composables

**`src/composables/usePermission.ts`**
```typescript
export function usePermission() {
  const auth = useAuthStore()
  const can = (module: string, action: 'create' | 'read' | 'update' | 'delete') => {
    if (auth.user?.role === 'ADMIN') return true
    const perm = auth.permissions.find(p => p.module === module)
    return perm ? perm[`can_${action}` as keyof UserPermission] as boolean : false
  }
  return { can }
}
```

**`src/composables/usePagination.ts`**
- State: `page` (default 1), `limit` (default 20), `sort` (default 'created_at'), `order` (default 'desc')
- Computed: `queryParams` → `{ page, limit, sort, order }`
- Actions: `setPage(n)`, `nextPage()`, `prevPage()`, `resetPage()`

**`src/composables/useToast.ts`**
- Wrapper `useToast()` dari PrimeVue
- `success(summary, detail?, life?)` → life default 3000
- `error(summary, detail?, life?)` → life default 5000
- `warn(summary, detail?, life?)` → life default 5000
- `info(summary, detail?)` → life default 3000

**`src/composables/useConfirm.ts`**
- Wrapper `useConfirm()` dari PrimeVue
- `confirmDelete(label: string, onAccept: () => void)` → dialog konfirmasi dengan teks: "Hapus [label]? Data yang dihapus tidak dapat dikembalikan."

---

## Task 2 — Common Components

**`src/components/common/PageHeader.vue`**
- Props: `title: string`, `subtitle?: string`
- Slot default: `actions` (right side)
- Style: `flex items-center justify-between pb-4 border-b border-surface-200`

**`src/components/common/DataTable.vue`**
- Wrapper PrimeVue `<DataTable>` dengan built-in pagination via `<Paginator>`
- Props: `data: T[]`, `columns: ColumnDef[]`, `loading: boolean`, `total: number`, `page: number`, `limit: number`
- Emit: `@update:page`, `@update:limit`, `@sort`
- Slot: `body-row` untuk custom row rendering
- Show `<EmptyState>` jika `data.length === 0` dan tidak loading
- Show skeleton rows jika loading

**`src/components/common/EmptyState.vue`**
- Props: `title?: string` (default "Belum ada data"), `description?: string`
- Icon: `pi pi-inbox` ukuran besar, text muted

**`src/components/common/StatusBadge.vue`**
- Props: `status: string`
- Map status ke severity PrimeVue Tag:
  - `active` → severity="success"
  - `deleted` → severity="danger"
  - `superseded` → severity="secondary"
  - `extended` → severity="warn"
  - `TW1|TW2|TW3|TW4` → severity="info"

**`src/components/common/CurrencyDisplay.vue`**
- Props: `amount: number`, `currency?: string` (default 'USD'), `compact?: boolean`
- Format: `Intl.NumberFormat('en-US', { minimumFractionDigits: 2 })`
- Output: `USD 1,234,567.00` atau jika compact: `USD 1.23M`

**`src/components/common/ConfirmDialog.vue`**
- Letakkan sekali di `App.vue`: `<ConfirmDialog />`
- Hanya wrapper `<ConfirmDialog />` dari PrimeVue

---

## Task 3 — Form Components

**`src/components/forms/LocationMultiSelect.vue`**
- `v-model`: array region_id
- Tampilkan TreeSelect atau MultiSelect PrimeVue dengan hierarki COUNTRY → PROVINCE → CITY
- Jika COUNTRY dipilih: nonaktifkan semua PROVINCE/CITY, tampilkan "(Nasional)"
- Load seluruh level via `masterStore.fetchAllRegionLevels()` agar pilihan COUNTRY, PROVINCE, dan CITY tersedia di form BB, GB, dan DK.

**`src/components/forms/InstitutionSelect.vue`**
- Props: `modelValue`, `multiple?: boolean`, `levelFilter?: string[]`
- Tampilkan `name` + `(short_name)` jika short_name ada
- Load dari `masterStore.institutions`

**`src/components/forms/LenderSelect.vue`**
- Props: `modelValue`, `multiple?: boolean`, `allowedIds?: string[]`
- Jika `allowedIds` ada, filter hanya lender dengan id tersebut
- Tampilkan nama + badge type (Bilateral/Multilateral/KSA)
- Load dari `masterStore.lenders`

**`src/components/forms/ProgramTitleSelect.vue`**
- Grouped select: parent sebagai optgroup, child sebagai option
- Load dari `masterStore.programTitles`

**`src/components/forms/NationalPriorityMultiSelect.vue`**
- Props: `modelValue`, `periodId?: string`
- Filter berdasarkan periodId jika ada
- Load dari `masterStore.nationalPriorities`

**`src/components/forms/CurrencyInput.vue`**
- Props: `modelValue: number`, `currency?: string`, `placeholder?: string`, `disabled?: boolean`
- Input tipe number dengan format ribuan saat blur
- Emit `update:modelValue` sebagai number (bukan string)
- Tampilkan prefix currency di kiri input

---

## Task 4 — src/types/master.types.ts

```typescript
export interface Country { id: string; name: string; code: string }
export interface Lender { id: string; name: string; short_name?: string; type: 'Bilateral' | 'Multilateral' | 'KSA'; country?: Country }
export interface Institution { id: string; name: string; short_name?: string; level: string; parent_id?: string; parent?: Institution }
export interface Region { id: string; code: string; name: string; type: 'COUNTRY' | 'PROVINCE' | 'CITY'; parent_code?: string }
export interface ProgramTitle { id: string; title: string; parent_id?: string; parent?: ProgramTitle }
export interface BappenasPartner { id: string; name: string; level: 'Eselon I' | 'Eselon II'; parent_id?: string; parent?: BappenasPartner }
export interface Period { id: string; name: string; year_start: number; year_end: number }
export interface NationalPriority { id: string; title: string; period_id: string; period?: Period }
```

---

## Task 5 — src/services/master.service.ts

Satu function per endpoint dari `PRISM_API_Contract.md` bagian Master Data:
- `getCountries(params?)`, `createCountry(data)`, `updateCountry(id, data)`, `deleteCountry(id)`
- `getLenders(params?)`, `createLender(data)`, `updateLender(id, data)`, `deleteLender(id)`
- `getInstitutions(params?)`, `createInstitution(data)`, `updateInstitution(id, data)`, `deleteInstitution(id)`
- `getRegions(params?)`, `createRegion(data)`, `updateRegion(id, data)`, `deleteRegion(id)`
- `getProgramTitles()`, `createProgramTitle(data)`, `updateProgramTitle(id, data)`, `deleteProgramTitle(id)`
- `getBappenasPartners()`, `createBappenasPartner(data)`, `updateBappenasPartner(id, data)`, `deleteBappenasPartner(id)`
- `getPeriods()`, `createPeriod(data)`, `updatePeriod(id, data)`, `deletePeriod(id)`
- `getNationalPriorities(params?)`, `createNationalPriority(data)`, `updateNationalPriority(id, data)`, `deleteNationalPriority(id)`

---

## Task 6 — src/stores/master.store.ts

```typescript
export const useMasterStore = defineStore('master', () => {
  const countries = ref<Country[]>([])
  const lenders = ref<Lender[]>([])
  const institutions = ref<Institution[]>([])
  const regions = ref<Region[]>([])
  const programTitles = ref<ProgramTitle[]>([])
  const bappenasPartners = ref<BappenasPartner[]>([])
  const periods = ref<Period[]>([])
  const nationalPriorities = ref<NationalPriority[]>([])

  // loaded flags — fetch hanya sekali, kecuali force = true
  const loaded = ref<Record<string, boolean>>({})

  async function fetchCountries(force = false) { ... }
  async function fetchLenders(force = false) { ... }
  // dst untuk semua entitas

  function $reset() { /* clear semua state dan loaded flags */ }
  return { countries, lenders, institutions, regions, programTitles, bappenasPartners, periods, nationalPriorities, loaded, fetchCountries, fetchLenders, /* ... */ $reset }
})
```

Setiap fetch action: jika `loaded[key]` dan tidak force, skip. Setelah mutasi (create/update/delete), set `loaded[key] = false` agar fetch berikutnya fresh.

---

## Checklist

- [x] `usePermission.ts`
- [x] `usePagination.ts`
- [x] `useToast.ts`
- [x] `useConfirm.ts`
- [x] `PageHeader.vue`
- [x] `DataTable.vue` — wrapper dengan pagination + empty state + skeleton
- [x] `EmptyState.vue`
- [x] `StatusBadge.vue`
- [x] `CurrencyDisplay.vue`
- [x] `ConfirmDialog.vue` — di App.vue
- [x] `LocationMultiSelect.vue` — dengan logika COUNTRY
- [x] `InstitutionSelect.vue`
- [x] `LenderSelect.vue` — dengan allowedIds filter
- [x] `ProgramTitleSelect.vue`
- [x] `NationalPriorityMultiSelect.vue`
- [x] `CurrencyInput.vue`
- [x] `master.types.ts`
- [x] `master.service.ts` — semua endpoint
- [x] `master.store.ts` — cache + force refresh
