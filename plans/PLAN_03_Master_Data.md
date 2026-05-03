# PLAN 03 — Master Data Pages

> **Scope:** CRUD UI untuk semua entitas master.
> **Deliverable:** Semua master data bisa dikelola via UI, data tersimpan di store untuk dipakai modul lain.
> **Referensi:** docs/PRISM_API_Contract.md (Master Data), docs/PRISM_Business_Rules.md (bagian 2, 8, 9)

---

## Pola Umum Semua Halaman Master

Setiap halaman master mengikuti pola:
1. `<PageHeader>` dengan tombol "Tambah" di slot actions (proteksi `can(module, 'create')`)
2. `<DataTable>` atau `<TreeTable>` (untuk hierarki)
3. Dialog form untuk create/edit — satu halaman, tidak navigate ke halaman baru
4. `<ConfirmDialog>` sebelum delete
5. Setelah create/update/delete: refresh store dengan `force = true`

---

## Task 1 — src/schemas/master.schema.ts

Zod schemas untuk semua form master:

```typescript
export const countrySchema = z.object({
  name: z.string().min(1),
  code: z.string().length(3).transform(s => s.toUpperCase()),
})

export const lenderSchema = z.object({
  name: z.string().min(1),
  short_name: z.string().optional(),
  type: z.enum(['Bilateral', 'Multilateral', 'KSA']),
  country_id: z.string().uuid().optional(),
}).refine(data => {
  if (data.type !== 'Multilateral') return !!data.country_id
  return true
}, { message: 'Negara wajib diisi untuk Bilateral dan KSA', path: ['country_id'] })

export const institutionSchema = z.object({
  name: z.string().min(1),
  short_name: z.string().optional(),
  level: z.enum(['Kementerian/Badan/Lembaga', 'Eselon I', 'Eselon II', 'BUMN', 'Pemerintah Daerah Tk. I', 'Pemerintah Daerah Tk. II', 'BUMD', 'Lainya']),
  parent_id: z.string().uuid().optional(),
})

export const regionSchema = z.object({
  code: z.string().min(1).max(10),
  name: z.string().min(1),
  type: z.enum(['COUNTRY', 'PROVINCE', 'CITY']),
  parent_code: z.string().optional(),
})

export const periodSchema = z.object({
  name: z.string().min(1),
  year_start: z.number().int().min(2000),
  year_end: z.number().int().min(2000),
}).refine(d => d.year_end > d.year_start, { message: 'Tahun akhir harus lebih besar dari tahun awal', path: ['year_end'] })

export const nationalPrioritySchema = z.object({
  period_id: z.string().uuid('Period wajib dipilih'),
  title: z.string().min(1),
})

export const programTitleSchema = z.object({
  title: z.string().min(1),
  parent_id: z.string().uuid().optional(),
})

export const bappenasPartnerSchema = z.object({
  name: z.string().min(1),
  level: z.enum(['Eselon I', 'Eselon II']),
  parent_id: z.string().uuid().optional(),
}).refine(d => {
  if (d.level === 'Eselon II') return !!d.parent_id
  return true
}, { message: 'Eselon I parent wajib diisi untuk Eselon II', path: ['parent_id'] })
```

---

## Task 2 — CountryPage.vue

- Tabel: code, name, actions
- Dialog form: name, code (3 karakter uppercase)
- Module: `country`

---

## Task 3 — LenderPage.vue

- Tabel: name, short_name, type badge, country name, actions
- Dialog form:
  - name, short_name
  - type (Select)
  - `country_id` (`<CountrySelect>`) — render hanya jika type ≠ Multilateral
- **Business rule:** validasi Zod refine seperti di schema

---

## Task 4 — InstitutionPage.vue

- PrimeVue `<TreeTable>` — tampilkan hierarki institution berdasarkan `parent_id`
- Kolom: name, short_name, level badge, actions
- Dialog form: name, short_name, level (Select 8 pilihan), parent_id (muncul jika level bukan Kementerian/Badan/Lembaga)

---

## Task 5 — RegionPage.vue

- `<TreeTable>`: COUNTRY → PROVINCE → CITY
- Kolom: code, name, type badge, actions
- Dialog form: code, name, type (Select), parent_code (muncul jika PROVINCE atau CITY)

---

## Task 6 — ProgramTitlePage.vue

- `<TreeTable>`: Parent → Child
- Dialog form: title, parent_id (opsional Select program title level parent)

---

## Task 7 — BappenasPartnerPage.vue

- `<TreeTable>`: Eselon I → Eselon II
- Dialog form: name, level (Select), parent_id (wajib jika level = Eselon II)

---

## Task 8 — PeriodPage.vue

- Tabel: name, year_start, year_end, actions
- Dialog form: name, year_start, year_end
- Refine: year_end > year_start

---

## Task 9 — NationalPriorityPage.vue

- Filter dropdown Period di atas tabel
- Tabel: title, period name, actions
- Dialog form: period_id (Select Period), title
- Load ulang tabel saat filter period berubah

---

## Checklist

- [x] `master.schema.ts` — semua Zod schemas dengan refine
- [x] `CountryPage.vue`
- [x] `LenderPage.vue` — conditional country_id field
- [x] `InstitutionPage.vue` — TreeTable 8 level
- [x] `RegionPage.vue` — TreeTable 3 level
- [x] `ProgramTitlePage.vue` — TreeTable
- [x] `BappenasPartnerPage.vue` — TreeTable
- [x] `PeriodPage.vue`
- [x] `NationalPriorityPage.vue` — filter by period
