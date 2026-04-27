# PLAN 08 — Monitoring Disbursement Module

> **Scope:** Input dan view monitoring per triwulan (rencana vs realisasi) dengan breakdown komponen opsional.
> **Deliverable:** Staff bisa input monitoring setelah LA efektif.
> **Referensi:** docs/PRISM_API_Contract.md (Monitoring), docs/PRISM_Business_Rules.md (bagian 7)

---

## Task 1 — Types & Schema

**`src/types/monitoring.types.ts`:**
```typescript
export type Quarter = 'TW1' | 'TW2' | 'TW3' | 'TW4'

export interface MonitoringDisbursement {
  id: string
  loan_agreement_id: string
  budget_year: number
  quarter: Quarter
  exchange_rate_usd_idr: number
  exchange_rate_la_idr: number
  planned_la: number; planned_usd: number; planned_idr: number
  realized_la: number; realized_usd: number; realized_idr: number
  absorption_pct: number
  komponen: MonitoringKomponen[]
}

export interface MonitoringKomponen {
  id?: string
  component_name: string
  planned_la: number; planned_usd: number; planned_idr: number
  realized_la: number; realized_usd: number; realized_idr: number
}
```

**`src/schemas/monitoring.schema.ts`:**
```typescript
export const monitoringSchema = z.object({
  budget_year: z.number().int().min(2000),
  quarter: z.enum(['TW1', 'TW2', 'TW3', 'TW4']),
  exchange_rate_usd_idr: z.number().positive('Kurs USD/IDR harus lebih dari 0'),
  exchange_rate_la_idr: z.number().positive('Kurs LA/IDR harus lebih dari 0'),
  planned_la: z.number().min(0), planned_usd: z.number().min(0), planned_idr: z.number().min(0),
  realized_la: z.number().min(0), realized_usd: z.number().min(0), realized_idr: z.number().min(0),
})
```

---

## Task 2 — Service & Store

**`src/services/monitoring.service.ts`** — semua endpoint API Contract bagian Monitoring.
**`src/stores/monitoring.store.ts`** — state: `monitorings`, `currentLA`, `loading`, `total`.

---

## Task 3 — composables/forms/useMonitoringForm.ts

```typescript
export function useMonitoringForm(initialData?) {
  const { handleSubmit, errors, values } = useForm({ validationSchema: toTypedSchema(monitoringSchema), initialValues: initialData ?? {} })

  const showKomponen = ref(false)
  const komponen = ref<KomponenRow[]>(initialData?.komponen ?? [])

  const addKomponen = () => komponen.value.push({ component_name: '', planned_la: 0, planned_usd: 0, planned_idr: 0, realized_la: 0, realized_usd: 0, realized_idr: 0 })
  const removeKomponen = (i: number) => komponen.value.splice(i, 1)

  // Real-time absorption pct — hindari division by zero
  const absorptionPct = computed(() => {
    const planned = values.planned_usd ?? 0
    const realized = values.realized_usd ?? 0
    if (planned === 0) return 0
    return Math.round((realized / planned) * 1000) / 10  // 1 desimal
  })

  const submit = handleSubmit(values => ({
    ...values,
    komponen: showKomponen.value ? komponen.value : [],
  }))

  return { values, errors, showKomponen, komponen, addKomponen, removeKomponen, absorptionPct, submit }
}
```

---

## Task 4 — MonitoringListPage.vue

Diakses dari `LADetailPage` — list monitoring untuk LA tertentu:

- Info header LA (loan_code, lender, currency, effective_date)
- **Guard:** Jika `effective_date > today`, tampilkan banner "LA belum efektif — monitoring belum bisa diinput" dan disable tombol "Tambah"
- Tabel: budget_year, quarter badge, exchange_rate_usd_idr, exchange_rate_la_idr, planned_usd, realized_usd, `<AbsorptionBar :pct="absorption_pct">`, actions
- Tombol "Tambah Monitoring" → navigate ke form

---

## Task 5 — MonitoringFormPage.vue

Gunakan `useMonitoringForm()`.

**Section 1 — Periode:**
budget_year (number input), quarter (Select dengan label: "TW1 (Apr–Jun)", "TW2 (Jul–Sep)", "TW3 (Okt–Des)", "TW4 (Jan–Mar)"), exchange_rate_usd_idr, exchange_rate_la_idr

**Section 2 — Rencana vs Realisasi (Level LA):**

Tabel 3×2 (3 mata uang × 2 kolom Rencana/Realisasi):
|  | Rencana | Realisasi |
|--|---------|-----------|
| Mata Uang LA | planned_la | realized_la |
| USD | planned_usd | realized_usd |
| IDR | planned_idr | realized_idr |

Tampilkan di bawah tabel: `<AbsorptionBar :pct="absorptionPct">` real-time

**Section 3 — Breakdown Komponen (Opsional):**
Toggle switch "Tambah Breakdown per Komponen". Jika ON: tampilkan `<KomponenTable>` editable. Tombol "Tambah Komponen".

---

## Task 6 — Komponen Monitoring

**`src/components/monitoring/AbsorptionBar.vue`:**
- Props: `pct: number`
- PrimeVue `<ProgressBar :value="pct">`
- Warna via class: pct < 50 = danger, 50–79 = warn, >= 80 = success
- Tampilkan label persentase

**`src/components/monitoring/MonitoringCard.vue`:**
- Props: `monitoring: MonitoringDisbursement`
- Card dengan: quarter badge, budget_year, planned_usd, realized_usd, `<AbsorptionBar>`

**`src/components/monitoring/KomponenTable.vue`:**
- Props: `komponen: MonitoringKomponen[]`, `editable?: boolean`
- Tabel: component_name, planned_la, planned_usd, planned_idr, realized_la, realized_usd, realized_idr
- Jika editable: input per sel + tombol hapus baris
- Footer total per kolom

**`src/components/monitoring/MonitoringChart.vue`:**
- Props: `data: MonitoringDisbursement[]`
- ECharts bar chart grouped: X = triwulan (TW1–TW4), Y = USD
- 2 seri: Planned (biru) dan Realized (hijau)
- Tooltip menampilkan kedua nilai + absorption pct

---

## Checklist

- [x] `monitoring.types.ts`
- [x] `monitoring.schema.ts`
- [x] `monitoring.service.ts`
- [x] `monitoring.store.ts`
- [x] `useMonitoringForm.ts` — absorptionPct computed + komponen optional
- [x] `MonitoringListPage.vue` — guard effective_date + tabel dengan AbsorptionBar
- [x] `MonitoringFormPage.vue` — 3 section
- [x] `AbsorptionBar.vue` — color coding 3 level
- [x] `MonitoringCard.vue`
- [x] `KomponenTable.vue` — editable + read-only mode
- [x] `MonitoringChart.vue` — ECharts grouped bar
