# PLAN 06 — Daftar Kegiatan Module

> **Scope:** CRUD Daftar Kegiatan (header surat + DK Project + sub-tabel multi-currency).
> **Deliverable:** Staff bisa input DK lengkap. Lender difilter dari BB/GB terkait.
> **Referensi:** docs/PRISM_API_Contract.md (Daftar Kegiatan), docs/PRISM_Business_Rules.md (bagian 5)

---

## Task 1 — Types & Schema

**`src/types/daftar-kegiatan.types.ts`:**
```typescript
export interface DaftarKegiatan { id: string; letter_number?: string; subject: string; date: string }
export interface DKProject { id: string; dk: DaftarKegiatan; program_title?: ProgramTitle; institution?: Institution; duration?: string; objectives?: string; gb_projects: GBProjectSummary[]; locations: Region[]; financing_details: DKFinancingDetail[]; loan_allocations: DKLoanAllocation[]; activity_details: DKActivityDetail[] }
export interface DKFinancingDetail { id: string; lender?: Lender; currency: string; amount_original: number; grant_original: number; counterpart_original: number; amount_usd: number; grant_usd: number; counterpart_usd: number; remarks?: string }
export interface DKLoanAllocation { id: string; institution?: Institution; currency: string; amount_original: number; grant_original: number; counterpart_original: number; amount_usd: number; grant_usd: number; counterpart_usd: number; remarks?: string }
export interface DKActivityDetail { id: string; activity_number: number; activity_name: string }
```

**`src/schemas/daftar-kegiatan.schema.ts`:**
```typescript
export const daftarKegiatanSchema = z.object({
  subject: z.string().min(1, 'Perihal wajib diisi'),
  date: z.string().min(1, 'Tanggal wajib diisi'),
  letter_number: z.string().optional(),
})

export const dkProjectSchema = z.object({
  program_title_id: z.string().uuid().optional(),
  institution_id: z.string().uuid('Executing Agency wajib dipilih'),
  duration: z.string().optional(),
  objectives: z.string().optional(),
  gb_project_ids: z.array(z.string().uuid()).min(1, 'Minimal 1 GB Project'),
  location_ids: z.array(z.string().uuid()).min(1, 'Lokasi wajib dipilih'),
})
```

---

## Task 2 — Service & Store

**`src/services/daftar-kegiatan.service.ts`** — semua endpoint dari API Contract.
**`src/stores/daftar-kegiatan.store.ts`** — state dan actions standar.

---

## Task 3 — composables/forms/useDKProjectForm.ts

```typescript
export function useDKProjectForm(initialData?) {
  const { handleSubmit, errors, values, setFieldValue } = useForm({ ... })

  // Computed allowed lenders dari GB yang dipilih
  const allowedLenderIds = computed(async () => {
    const gbIds = values.gb_project_ids ?? []
    if (!gbIds.length) return []
    // fetch funding_sources dari GB yang dipilih + lender_indications dari BB terkait
    // return array of lender IDs yang diperbolehkan
  })

  // Financing Details
  const financingDetails = ref<FinancingRow[]>([])
  const addFinancing = () => financingDetails.value.push({ lender_id: '', currency: 'USD', amount_original: 0, grant_original: 0, counterpart_original: 0, amount_usd: 0, grant_usd: 0, counterpart_usd: 0 })

  // Loan Allocations
  const loanAllocations = ref<AllocationRow[]>([])
  const addAllocation = () => loanAllocations.value.push({ institution_id: '', currency: 'USD', amount_original: 0, grant_original: 0, counterpart_original: 0, amount_usd: 0, grant_usd: 0, counterpart_usd: 0 })

  // Activity Details — nomor urut otomatis
  const activityDetails = ref<ActivityRow[]>([])
  const addActivity = () => activityDetails.value.push({ activity_number: activityDetails.value.length + 1, activity_name: '' })
  const removeActivity = (i: number) => { activityDetails.value.splice(i, 1); activityDetails.value.forEach((a, idx) => a.activity_number = idx + 1) }

  return { values, errors, allowedLenderIds, financingDetails, addFinancing, loanAllocations, addAllocation, activityDetails, addActivity, removeActivity, handleSubmit }
}
```

---

## Task 4 — DKListPage.vue

- `<PageHeader title="Daftar Kegiatan">` + tombol "Buat Daftar Kegiatan"
- Tabel: subject, date, letter_number, jumlah proyek, actions
- Click baris → DKDetailPage

---

## Task 5 — DKDetailPage.vue

- Header surat: subject, date, letter_number
- Daftar DK Projects dalam surat: accordion per proyek
- Setiap proyek accordion memuat: info proyek, GB references, Financing Detail table, Loan Allocation table, Activity Details list
- Tombol "Tambah Proyek ke Surat" → form DK Project

---

## Task 6 — DKProjectFormPage.vue (4 Section)

Gunakan `useDKProjectForm()`.

**Section 1 — Header Proyek:**
program_title_id (`<ProgramTitleSelect>`), institution_id (`<InstitutionSelect>` — Executing Agency), duration, objectives, gb_project_ids (MultiSelect GB Project), location_ids (`<LocationMultiSelect>`)

**Section 2 — Financing Detail (tabel multi-currency):**
Kolom: lender_id (`<LenderSelect :allowedIds>`), currency (text input ISO), amount_original, grant_original, counterpart_original, amount_usd, grant_usd, counterpart_usd (`<CurrencyInput>`), remarks. Tombol "Tambah Baris". Tampilkan catatan: "Konversi ke USD dilakukan manual".

**Section 3 — Loan Allocation (tabel multi-currency):**
Kolom: institution_id (`<InstitutionSelect>`), currency, amount_original, grant_original, counterpart_original, amount_usd, grant_usd, counterpart_usd, remarks.

**Section 4 — Activity Details:**
List sederhana dengan nomor urut otomatis + nama aktivitas (input bebas). Tombol "Tambah" dan "Hapus" per baris. Nomor urut auto-recalculate.

---

## Checklist

- [ ] `daftar-kegiatan.types.ts`
- [ ] `daftar-kegiatan.schema.ts`
- [ ] `daftar-kegiatan.service.ts`
- [ ] `daftar-kegiatan.store.ts`
- [ ] `useDKProjectForm.ts` — allowedLenderIds computed
- [ ] `DKListPage.vue`
- [ ] `DKDetailPage.vue` — accordion per DK Project
- [ ] `DKProjectFormPage.vue` — 4 section + multi-currency
