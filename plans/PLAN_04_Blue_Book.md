# PLAN 04 — Blue Book Module

> **Scope:** CRUD Blue Book (header + BB Project + LoI + Lender Indication + Project Cost).
> **Deliverable:** Staff bisa input dan melihat Blue Book lengkap.
> **Referensi:** docs/PRISM_API_Contract.md (Blue Book), docs/PRISM_Business_Rules.md (bagian 3)
> **Revision update:** Ikuti `docs/PRISM_BB_GB_Revision_Versioning_Plan.md`. BB Project perlu `project_identity_id`, indikator latest/newer revision, dan section histori revisi.

---

## Task 1 — Types

**`src/types/blue-book.types.ts`:**
```typescript
export interface BlueBook { id: string; period: Period; publish_date: string; revision_number: number; revision_year?: number; status: 'active' | 'superseded' }
export interface BBProject { id: string; bb_code: string; project_name: string; program_title?: ProgramTitle; bappenas_partner?: BappenasPartner; executing_agencies: Institution[]; implementing_agencies: Institution[]; locations: Region[]; national_priorities: NationalPriority[]; project_costs: BBProjectCost[]; lender_indications: LenderIndication[]; duration?: string; objective?: string; scope_of_work?: string; outputs?: string; outcomes?: string; status: 'active' | 'deleted' }
export interface LenderIndication { id: string; lender: Lender; remarks?: string }
export interface LoI { id: string; lender: Lender; subject: string; date: string; letter_number?: string }
export interface BBProjectCost { id: string; funding_type: 'Foreign' | 'Counterpart'; funding_category: string; amount_usd: number }
```

---

## Task 2 — Schema

**`src/schemas/blue-book.schema.ts`:**
```typescript
export const blueBookSchema = z.object({
  period_id: z.string().uuid('Period wajib dipilih'),
  publish_date: z.string().min(1, 'Tanggal terbit wajib diisi'),
  revision_number: z.number().int().min(0),
  revision_year: z.number().int().optional(),
})

export const bbProjectSchema = z.object({
  program_title_id: z.string().uuid('Program Title wajib dipilih'),
  bappenas_partner_id: z.string().uuid('Bappenas Partner wajib dipilih'),
  bb_code: z.string().min(1, 'BB Code wajib diisi'),
  project_name: z.string().min(1, 'Nama proyek wajib diisi'),
  duration: z.string().optional(),
  objective: z.string().optional(),
  scope_of_work: z.string().optional(),
  outputs: z.string().optional(),
  outcomes: z.string().optional(),
  executing_agency_ids: z.array(z.string().uuid()).min(1, 'Minimal 1 Executing Agency'),
  implementing_agency_ids: z.array(z.string().uuid()).min(1, 'Minimal 1 Implementing Agency'),
  location_ids: z.array(z.string().uuid()).min(1, 'Lokasi wajib dipilih'),
  national_priority_ids: z.array(z.string().uuid()),
})

export const loiSchema = z.object({
  lender_id: z.string().uuid('Lender wajib dipilih'),
  subject: z.string().min(1, 'Perihal wajib diisi'),
  date: z.string().min(1, 'Tanggal wajib diisi'),
  letter_number: z.string().optional(),
})
```

---

## Task 3 — Service & Store

**`src/services/blue-book.service.ts`** — semua endpoint dari API Contract bagian Blue Book.

**`src/stores/blue-book.store.ts`** — state: `blueBooks`, `currentBlueBook`, `projects`, `currentProject`, `loading`, `total`. Actions per entitas.

---

## Task 4 — BlueBookListPage.vue

- `<PageHeader title="Blue Book">` + tombol "Buat Blue Book" (can create)
- `<DataTable>`: period name, publish_date, revision, status badge, total projects, actions
- Click baris → `BlueBookDetailPage`
- Filter: period dropdown, status

---

## Task 5 — BlueBookDetailPage.vue

- Info header Blue Book (period, publish_date, revision_number, status badge)
- Tombol "Edit BB" dan "Tambah Proyek"
- `<DataTable>` BB Projects: bb_code, project_name, executing agency (first), status badge, actions (View, Edit, Delete)

---

## Task 6 — composables/forms/useBBProjectForm.ts

```typescript
export function useBBProjectForm(initialData?: Partial<BBProjectFormValues>) {
  const { handleSubmit, errors, values } = useForm({
    validationSchema: toTypedSchema(bbProjectSchema),
    initialValues: initialData ?? {},
  })

  // Project Costs rows
  const projectCosts = ref<ProjectCostRow[]>(initialData?.project_costs ?? [])
  const addCost = () => projectCosts.value.push({ funding_type: 'Foreign', funding_category: 'Loan', amount_usd: 0 })
  const removeCost = (i: number) => projectCosts.value.splice(i, 1)

  // Lender Indication rows
  const lenderIndications = ref<LenderIndicationRow[]>(initialData?.lender_indications ?? [])
  const addIndication = () => lenderIndications.value.push({ lender_id: '', remarks: '' })
  const removeIndication = (i: number) => lenderIndications.value.splice(i, 1)

  const submit = handleSubmit(values => ({
    ...values,
    project_costs: projectCosts.value,
    lender_indications: lenderIndications.value,
  }))

  return { values, errors, projectCosts, addCost, removeCost, lenderIndications, addIndication, removeIndication, submit }
}
```

---

## Task 7 — BBProjectFormPage.vue

Gunakan `useBBProjectForm()`. Layout section vertikal:

**Section 1 — Informasi Umum:**
program_title_id (`<ProgramTitleSelect>`), bappenas_partner_id (Select Eselon II + tampilkan parent Eselon I read-only), bb_code, project_name, duration, objective, scope_of_work, outputs, outcomes (textarea)

**Section 2 — Pihak Terlibat:**
executing_agency_ids (`<InstitutionSelect multiple>`), implementing_agency_ids (`<InstitutionSelect multiple>`)

**Section 3 — Lokasi & Prioritas:**
location_ids (`<LocationMultiSelect>`), national_priority_ids (`<NationalPriorityMultiSelect :periodId>`)

**Section 4 — Project Cost (tabel editable):**
Tabel dengan baris dynamic: funding_type (Select: Foreign/Counterpart), funding_category (Select tergantung funding_type), amount_usd (`<CurrencyInput>`). Tombol "Tambah Baris".

Foreign categories: Loan, Grant
Counterpart categories: Central Government, Regional Government, State-Owned Enterprise, Others

**Section 5 — Lender Indication (tabel editable):**
Tabel: lender_id (`<LenderSelect>`), remarks (text). Tombol "Tambah Indikasi".

Footer: tombol "Simpan" dan "Batal".

---

## Task 8 — BBProjectDetailPage.vue

- Header: bb_code, project_name, status badge
- Grid info: Executing Agency, Implementing Agency, Bappenas Partner (Eselon II + I parent), Lokasi, National Priority
- Section Project Cost: tabel read-only
- Section Lender Indication: tabel read-only
- Section LoI: tabel dengan kolom lender, subject, date, letter_number + tombol "Tambah LoI"
- Tombol: Edit, Hapus, "Lihat Journey"

---

## Task 9 — LoI Dialog

Dialog untuk create LoI (dari BBProjectDetailPage):
- lender_id: `<LenderSelect>` (filter hanya dari lender indication proyek ini)
- subject, date (DatePicker), letter_number (opsional)
- Validasi via loiSchema

---

## Komponen Spesifik

**`src/components/blue-book/ProjectCostTable.vue`** — tabel editable project cost
**`src/components/blue-book/LenderIndicationTable.vue`** — tabel editable lender indication
**`src/components/blue-book/LoITable.vue`** — tabel read-only + tombol tambah LoI

---

## Checklist

- [x] `blue-book.types.ts`
- [x] `blue-book.schema.ts` — bbProject + loI schemas
- [x] `blue-book.service.ts`
- [x] `blue-book.store.ts`
- [x] `BlueBookListPage.vue`
- [x] `BlueBookDetailPage.vue`
- [x] `useBBProjectForm.ts`
- [x] `BBProjectFormPage.vue` — 5 section
- [x] `BBProjectDetailPage.vue`
- [x] `ProjectCostTable.vue`, `LenderIndicationTable.vue`, `LoITable.vue`
- [x] LoI Dialog
