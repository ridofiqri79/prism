# PLAN 07 — Loan Agreement Module

> **Scope:** CRUD Loan Agreement dengan deteksi perpanjangan otomatis dan validasi lender.
> **Deliverable:** Staff bisa input LA terhubung ke DK Project dengan indikator perpanjangan real-time.
> **Referensi:** docs/PRISM_API_Contract.md (Loan Agreement), docs/PRISM_Business_Rules.md (bagian 6)

---

## Task 1 — Types & Schema

**`src/types/loan-agreement.types.ts`:**
```typescript
export interface LoanAgreement {
  id: string
  dk_project: DKProjectSummary
  lender: Lender
  loan_code: string
  agreement_date: string
  effective_date: string
  original_closing_date: string
  closing_date: string
  is_extended: boolean
  extension_days: number
  currency: string
  amount_original: number
  amount_usd: number
}
```

**`src/schemas/loan-agreement.schema.ts`:**
```typescript
export const loanAgreementSchema = z.object({
  dk_project_id: z.string().uuid('DK Project wajib dipilih'),
  lender_id: z.string().uuid('Lender wajib dipilih'),
  loan_code: z.string().min(1, 'Kode Loan wajib diisi'),
  agreement_date: z.string().min(1),
  effective_date: z.string().min(1),
  original_closing_date: z.string().min(1),
  closing_date: z.string().min(1),
  currency: z.string().min(3, 'Kode mata uang minimal 3 karakter (ISO 4217)'),
  amount_original: z.number().positive('Amount harus lebih dari 0'),
  amount_usd: z.number().positive('Amount USD harus lebih dari 0'),
}).refine(d => new Date(d.closing_date) >= new Date(d.original_closing_date), {
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
- Tabel: loan_code, lender name, effective_date, closing_date, currency, amount_usd (`<CurrencyDisplay>`), `<StatusBadge status="extended">` jika is_extended, actions
- Filter: lender (dropdown), is_extended (toggle), closing_date_before (date picker)

---

## Task 4 — LAFormPage.vue

Gunakan `useForm` dengan `loanAgreementSchema`.

**Field form:**
- `dk_project_id`: Autocomplete search DK Project by objectives atau GB code
- `lender_id`: `<LenderSelect>` — setelah dk_project_id dipilih, filter `allowedIds` dari financing_details DK Project tersebut
- `loan_code`: text
- `agreement_date`, `effective_date`, `original_closing_date`, `closing_date`: DatePicker PrimeVue
- `currency`: text (ISO 4217, contoh: JPY, USD, EUR, CNY)
- `amount_original`: `<CurrencyInput>` dengan label "[currency] (mata uang lender)"
- `amount_usd`: `<CurrencyInput>` USD — dengan note "Diisi manual oleh Staff"

**Indikator perpanjangan real-time:**
```vue
<div v-if="isExtended" class="p-3 bg-orange-50 border border-orange-200 rounded-lg">
  <span class="text-orange-700">Perpanjangan terdeteksi: +{{ extensionDays }} hari</span>
</div>
```
Computed `isExtended = closing_date !== original_closing_date`, `extensionDays = diff in days`.

---

## Task 5 — LADetailPage.vue

- Info lengkap LA
- Badge `<StatusBadge status="extended">` dan "extension_days hari" jika is_extended
- Link ke DK Project terkait
- Tombol "Lihat Monitoring" → navigate ke `/loan-agreements/:id/monitoring`
- Tombol Edit, Hapus

---

## Checklist

- [x] `loan-agreement.types.ts`
- [x] `loan-agreement.schema.ts` — dengan refine closing_date >= original_closing_date
- [x] `loan-agreement.service.ts`
- [x] `loan-agreement.store.ts`
- [x] `LAListPage.vue` — filter is_extended
- [x] `LAFormPage.vue` — indikator perpanjangan real-time + lender filter dari DK
- [x] `LADetailPage.vue`
