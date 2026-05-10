# PLAN 07 — Loan Agreement Module

> **Scope:** CRUD Loan Agreement multi DK Project dengan alokasi, deteksi perpanjangan otomatis, dan validasi lender.
> **Deliverable:** Staff bisa input LA terhubung ke satu atau beberapa DK Project, mengalokasikan nilai komitmen per project, dan melihat indikator perpanjangan real-time.
> **Referensi:** docs/PRISM_API_Contract.md (Loan Agreement), docs/PRISM_Business_Rules.md (bagian 6)

---

## Task 1 — Types & Schema

**`src/types/loan-agreement.types.ts`:**
```typescript
export interface LoanAgreement {
  id: string
  dk_projects: LoanAgreementDKProject[]
  lender: Lender
  loan_code: string
  agreement_date: string
  effective_date: string
  original_closing_date: string | null
  closing_date: string
  is_extended: boolean
  extension_days: number
  currency: string
  amount_original: number
  amount_usd: number
  cumulative_disbursement: number
}
```

**`src/schemas/loan-agreement.schema.ts`:**
```typescript
export const loanAgreementSchema = z.object({
  dk_project_allocations: z.array(z.object({
    dk_project_id: z.string().uuid('DK Project wajib dipilih'),
    allocation_original: z.number().positive('Alokasi harus lebih dari 0'),
  })).min(1, 'Minimal satu DK Project wajib dipilih'),
  lender_id: z.string().uuid('Lender wajib dipilih'),
  loan_code: z.string().min(1, 'Kode Loan wajib diisi'),
  agreement_date: z.string().min(1),
  effective_date: z.string().min(1),
  original_closing_date: z.string().optional().default(''),
  closing_date: z.string().min(1),
  currency: z.string().min(3, 'Kode mata uang minimal 3 karakter (ISO 4217)'),
  amount_original: z.number().positive('Amount harus lebih dari 0'),
  amount_usd: z.number().positive('Amount USD harus lebih dari 0'),
  cumulative_disbursement: z.number().nonnegative('Cumulative disbursement tidak boleh negatif'),
}).refine(d => d.dk_project_allocations.reduce((sum, item) => sum + item.allocation_original, 0) === d.amount_original, {
  message: 'Total alokasi harus sama dengan Amount Original',
  path: ['dk_project_allocations'],
}).refine(d => !d.original_closing_date || new Date(d.closing_date) >= new Date(d.original_closing_date), {
  message: 'Closing Date tidak boleh lebih awal dari Original Closing Date',
  path: ['closing_date'],
})
```

---

## Task 2 — Service & Store

**`src/services/loan-agreement.service.ts`** — semua endpoint API Contract.
**`src/stores/loan-agreement.store.ts`** — state dan actions standar.

---

## Task 3 — LAListPage.vue

- `<PageHeader title="Loan Agreement">` + tombol "Buat LA"
- Tabel: loan_code, daftar DK Project, lender name, effective_date, closing_date, currency, amount_usd (`<CurrencyDisplay>`), cumulative_disbursement dalam currency Loan Agreement, `<StatusBadge status="extended">` jika is_extended, actions
- Filter: lender (dropdown), is_extended (toggle), closing_date_before (date picker)

---

## Task 4 — LAFormPage.vue

Gunakan `useForm` dengan `loanAgreementSchema`.

**Field form:**
- `dk_project_allocations`: Autocomplete multiple DK Project by project name, objectives, DK letter, atau GB code
- Tabel alokasi: `allocation_original` per project, total harus sama dengan `amount_original`
- `lender_id`: `<LenderSelect>` - setelah DK Project dipilih, filter `allowedIds` dari irisan financing_details semua DK Project tersebut
- `loan_code`: text
- `agreement_date`, `effective_date`, `original_closing_date` (opsional), `closing_date`: DatePicker PrimeVue
- `currency`: text (ISO 4217, contoh: JPY, USD, EUR, CNY)
- `amount_original`: `<CurrencyInput>` dengan label "[currency] (mata uang lender)"
- `amount_usd`: `<CurrencyInput>` USD — dengan note "Diisi manual oleh Staff"
- `cumulative_disbursement`: `<CurrencyInput>` berdasarkan `currency` yang dipilih. Nilai ini manual, tidak dikonversi otomatis.

**Indikator perpanjangan real-time:**
```vue
<div v-if="isExtended" class="p-3 bg-orange-50 border border-orange-200 rounded-lg">
  <span class="text-orange-700">Perpanjangan terdeteksi: +{{ extensionDays }} hari</span>
</div>
```
Computed `isExtended = original_closing_date ? closing_date !== original_closing_date : false`, `extensionDays = diff in days`.

---

## Task 5 — LADetailPage.vue

- Info lengkap LA
- Badge `<StatusBadge status="extended">` dan "extension_days hari" jika is_extended
- Daftar DK Project terkait, dikelompokkan per header DK, dengan nilai alokasi per project
- Tombol "Lihat Monitoring" → navigate ke `/loan-agreements/:id/monitoring`
- Tombol Edit, Hapus

---

## Checklist

- [x] `loan-agreement.types.ts`
- [x] `loan-agreement.schema.ts` - allocations minimal satu project, total alokasi sama dengan amount original, original_closing_date opsional
- [x] `loan-agreement.service.ts`
- [x] `loan-agreement.store.ts`
- [x] `LAListPage.vue` — filter is_extended
- [x] `LAFormPage.vue` - multi DK Project, tabel alokasi, indikator perpanjangan real-time + lender intersection dari DK
- [x] `LADetailPage.vue`
