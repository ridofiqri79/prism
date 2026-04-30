# PRISM — Frontend Project Structure (Vue 3)

> Stack: **Vue 3** · **Vite** · **Pinia** · **Vue Router 4** · **PrimeVue 4** · **Tailwind CSS v4** · **Axios** · **VeeValidate + Zod**

---

## Stack Overview

| Komponen | Library | Alasan |
|----------|---------|--------|
| Framework | `Vue 3` + Composition API | Modern, reactivity system yang efisien |
| Build Tool | `Vite` | Cold start < 1 detik, HMR instan |
| Router | `Vue Router 4` | Official, nested routes untuk layout kompleks |
| State Management | `Pinia` | Official Vuex replacement, lebih simpel dan type-safe |
| UI Component | `PrimeVue 4` | DataTable, Tree, MultiSelect, Form — semua tersedia |
| HTTP Client | `axios` | Interceptor untuk JWT inject & token refresh |
| Form Validation | `VeeValidate` + `Zod` | Schema-based, sinergi dengan API contract |
| Table Kompleks | `TanStack Table v8` | Untuk tabel dengan nested editing (Funding Source, Activities) |
| Charts | `vue-echarts` + `echarts` | Terbaik untuk dashboard monitoring disbursement |
| SSE Client | Native `EventSource` | Built-in browser, tidak butuh library tambahan |
| Date | `date-fns` | Ringan, tree-shakeable |
| CSS | `Tailwind CSS v4` | Utility-first via Vite plugin — tidak perlu PostCSS |
| PrimeVue–Tailwind Bridge | `tailwindcss-primeui` | Plugin resmi PrimeTek — expose design token PrimeVue sebagai Tailwind utility |

---

## Project Structure

```
prism-frontend/
├── public/
│   └── favicon.ico
│
├── src/
│   ├── main.ts                        # Entry point — init Vue, plugins, router, pinia
│   ├── App.vue                        # Root component
│   │
│   ├── assets/
│   │   ├── styles/
│   │   │   ├── main.css               # @import tailwindcss + tailwindcss-primeui + @theme
│   │   │   └── theme.ts               # PrimeVue definePreset (design token kustom)
│   │   └── images/
│   │
│   ├── router/
│   │   ├── index.ts                   # Router instance + navigation guards
│   │   └── routes/
│   │       ├── auth.routes.ts
│   │       ├── master.routes.ts
│   │       ├── blue-book.routes.ts
│   │       ├── green-book.routes.ts
│   │       ├── daftar-kegiatan.routes.ts
│   │       ├── loan-agreement.routes.ts
│   │       ├── monitoring.routes.ts
│   │       └── user.routes.ts
│   │
│   ├── stores/                        # Pinia stores
│   │   ├── auth.store.ts              # User session, JWT, permissions
│   │   ├── blue-book.store.ts
│   │   ├── green-book.store.ts
│   │   ├── daftar-kegiatan.store.ts
│   │   ├── loan-agreement.store.ts
│   │   ├── monitoring.store.ts
│   │   ├── master.store.ts            # Lender, institution, region, dll (cached)
│   │   └── notification.store.ts     # SSE events
│   │
│   ├── services/                      # Axios API calls — satu file per modul
│   │   ├── http.ts                    # Axios instance + interceptors
│   │   ├── auth.service.ts
│   │   ├── blue-book.service.ts
│   │   ├── green-book.service.ts
│   │   ├── daftar-kegiatan.service.ts
│   │   ├── loan-agreement.service.ts
│   │   ├── monitoring.service.ts
│   │   ├── master.service.ts
│   │   └── user.service.ts
│   │
│   ├── composables/                   # Reusable logic (Composition API)
│   │   ├── useAuth.ts                 # Login, logout, cek permission
│   │   ├── usePermission.ts           # can('bb_project', 'create') helper
│   │   ├── usePagination.ts           # Pagination state & handler
│   │   ├── useSSE.ts                  # Subscribe/unsubscribe SSE events
│   │   ├── useToast.ts                # Wrapper PrimeVue toast
│   │   ├── useConfirm.ts              # Wrapper dialog konfirmasi delete
│   │   │
│   │   └── forms/                     # Form composables per modul
│   │       ├── useBBProjectForm.ts    # Form BB Project (termasuk multi-select, costs, lender indications)
│   │       ├── useGBProjectForm.ts    # Form GB Project (activities, funding source, disbursement, allocation)
│   │       ├── useDKProjectForm.ts    # Form DK Project
│   │       ├── useLAForm.ts           # Form Loan Agreement
│   │       └── useMonitoringForm.ts   # Form Monitoring (level LA + komponen opsional)
│   │
│   ├── types/                         # TypeScript type definitions
│   │   ├── api.types.ts               # Generic API response types (Paginated<T>, ApiError, dll)
│   │   ├── auth.types.ts
│   │   ├── blue-book.types.ts
│   │   ├── green-book.types.ts
│   │   ├── daftar-kegiatan.types.ts
│   │   ├── loan-agreement.types.ts
│   │   ├── monitoring.types.ts
│   │   └── master.types.ts  # Country, Currency, Lender, Institution, Region, dll
│   │
│   ├── schemas/                       # Zod validation schemas — satu file per modul
│   │   ├── blue-book.schema.ts
│   │   ├── green-book.schema.ts
│   │   ├── daftar-kegiatan.schema.ts
│   │   ├── loan-agreement.schema.ts
│   │   ├── monitoring.schema.ts
│   │   └── master.schema.ts
│   │
│   ├── layouts/
│   │   ├── AppLayout.vue              # Layout utama — sidebar + topbar + content slot
│   │   ├── AuthLayout.vue             # Layout halaman login
│   │   └── components/
│   │       ├── AppSidebar.vue
│   │       ├── AppTopbar.vue
│   │       └── AppBreadcrumb.vue
│   │
│   ├── pages/                         # Halaman — satu folder per modul
│   │   ├── auth/
│   │   │   └── LoginPage.vue
│   │   │
│   │   ├── dashboard/
│   │   │   └── DashboardPage.vue
│   │   │
│   │   ├── blue-book/
│   │   │   ├── BlueBookListPage.vue
│   │   │   ├── BlueBookDetailPage.vue
│   │   │   ├── BBProjectListPage.vue
│   │   │   ├── BBProjectDetailPage.vue
│   │   │   └── BBProjectFormPage.vue  # Create & Edit
│   │   │
│   │   ├── green-book/
│   │   │   ├── GreenBookListPage.vue
│   │   │   ├── GreenBookDetailPage.vue
│   │   │   ├── GBProjectListPage.vue
│   │   │   ├── GBProjectDetailPage.vue
│   │   │   └── GBProjectFormPage.vue
│   │   │
│   │   ├── daftar-kegiatan/
│   │   │   ├── DKListPage.vue
│   │   │   ├── DKDetailPage.vue
│   │   │   └── DKProjectFormPage.vue
│   │   │
│   │   ├── loan-agreement/
│   │   │   ├── LAListPage.vue
│   │   │   ├── LADetailPage.vue
│   │   │   └── LAFormPage.vue
│   │   │
│   │   ├── monitoring/
│   │   │   ├── MonitoringListPage.vue
│   │   │   └── MonitoringFormPage.vue
│   │   │
│   │   ├── journey/
│   │   │   └── ProjectJourneyPage.vue # Visualisasi alur BB → GB → DK → LA → Monitoring
│   │   │
│   │   ├── master/
│   │   │   ├── LenderPage.vue
│   │   │   ├── CurrencyPage.vue
│   │   │   ├── InstitutionPage.vue
│   │   │   ├── RegionPage.vue
│   │   │   ├── ProgramTitlePage.vue
│   │   │   ├── BappenasPartnerPage.vue
│   │   │   ├── PeriodPage.vue
│   │   │   └── NationalPriorityPage.vue
│   │   │
│   │   └── user/
│   │       ├── UserListPage.vue
│   │       ├── UserFormPage.vue
│   │       └── UserPermissionPage.vue
│   │
│   └── components/                    # Reusable components
│       ├── common/
│       │   ├── DataTable.vue          # Wrapper PrimeVue DataTable dengan pagination
│       │   ├── SearchFilterBar.vue    # Search + filter drawer standar untuk halaman paginated
│       │   ├── TableReloadShell.vue   # Shell animasi reload tabel yang mempertahankan data lama
│       │   ├── PageHeader.vue         # Title + breadcrumb + action button
│       │   ├── StatusBadge.vue        # Badge status active/deleted/extended
│       │   ├── CurrencyDisplay.vue    # Format angka USD / IDR / mata uang lender
│       │   ├── EmptyState.vue         # Empty state reusable
│       │   └── ConfirmDialog.vue      # Dialog konfirmasi delete
│       │
│       ├── forms/
│       │   ├── LocationMultiSelect.vue    # Multi-select region dengan hierarki
│       │   ├── InstitutionSelect.vue      # Select institution dengan filter level
│       │   ├── LenderSelect.vue           # Select lender dengan filter type
│       │   ├── CurrencySelect.vue         # Select currency dari Master Currency aktif
│       │   ├── ProgramTitleSelect.vue     # Select program title (parent-child)
│       │   ├── NationalPriorityMultiSelect.vue
│       │   └── CurrencyInput.vue          # Input angka dengan format currency
│       │
│       ├── blue-book/
│       │   ├── BBProjectCard.vue          # Card ringkas untuk list
│       │   ├── LenderIndicationTable.vue  # Tabel lender indication (editable)
│       │   ├── ProjectCostTable.vue       # Tabel project cost (editable)
│       │   └── LoITable.vue               # Tabel LoI per BB project
│       │
│       ├── green-book/
│       │   ├── GBProjectCard.vue
│       │   ├── ActivitiesTable.vue        # Tabel activities (editable, dengan sort_order)
│       │   ├── FundingSourceTable.vue     # Tabel funding source cofinancing + currency/original/USD
│       │   ├── DisbursementPlanTable.vue  # Tabel disbursement plan per tahun (editable)
│       │   └── FundingAllocationTable.vue # Tabel alokasi per activity (editable)
│       │
│       ├── monitoring/
│       │   ├── MonitoringCard.vue         # Card per quarter
│       │   ├── AbsorptionBar.vue          # Progress bar rencana vs realisasi
│       │   ├── KomponenTable.vue          # Tabel breakdown komponen (opsional, editable)
│       │   └── MonitoringChart.vue        # Chart realisasi per quarter (ECharts)
│       │
│       └── journey/
│           └── ProjectTimeline.vue        # Visualisasi timeline BB → Monitoring
│
├── .env
├── .env.example
├── index.html
├── vite.config.ts                     # Gunakan @tailwindcss/vite — BUKAN postcss.config.ts
├── tsconfig.json
└── package.json
```

---

## Alur Data

```
Page/Form Component
  │
  ▼
Composable (useXxxForm / usePagination)
  │
  ├── Validasi via Zod schema
  │
  ▼
Service (xxx.service.ts)
  │
  ├── Axios HTTP call
  ├── Token di-inject otomatis via interceptor
  │
  ▼
Pinia Store (update state)
  │
  ▼
Component re-render via reactivity
```

## Pola Global Reload Tabel

- Gunakan `src/components/common/DataTable.vue` untuk tabel list standar; komponen ini sudah mempertahankan data lama saat `loading` dan menampilkan animasi reload yang sama di semua halaman.
- DataTable list standar harus meneruskan `page`, `limit`, `total`, `sort`, dan `order` dari API/store agar pagination server-side, sorting header, dan resize column tetap konsisten.
- Untuk tabel custom, bungkus markup tabel dengan `src/components/common/TableReloadShell.vue` dan kirim `refreshing` saat data sedang di-fetch ulang.
- Skeleton hanya untuk load awal ketika data belum ada. Saat search/filter/pagination memicu fetch ulang, tabel lama tetap tampil dengan opacity transition dan indikator reload global.
- Jika tabel custom perlu animasi baris, gunakan `<TransitionGroup name="prism-table-row-fade">` agar timing dan geraknya konsisten dengan tabel lain.

## Pola Search, Filter, dan Pagination

- Halaman yang memiliki tabel paginated dan filter wajib memakai `src/components/common/SearchFilterBar.vue` di atas `DataTable.vue` atau table custom paginated.
- Search bar tampil full-width sebagai kontrol utama, tanpa label visual terpisah. Gunakan placeholder yang eksplisit dan ikon `pi pi-search` terintegrasi, tetapi input box tetap mengikuti styling standar `InputText` PrimeVue seperti field lain.
- Search boleh auto-apply dengan debounce pendek. Saat search berubah, reset page ke `1` dan kirim parameter `search` ke endpoint list agar pagination tetap server-side.
- Filter lanjutan dibuka lewat tombol `Filter` di dalam card yang sama. Drawer harus in-place, bukan modal/pop-up terpisah, agar konteks tabel tidak terputus.
- Isi drawer memakai grid responsif `xl:grid-cols-6`. Setiap field boleh memakai `xl:col-span-*` sesuai kompleksitas, tetapi layout harus tetap rapi dan tidak membuat card bersarang.
- Nilai filter dropdown disiapkan sebagai draft. Terapkan filter hanya ketika user menekan `Terapkan`; tombol `Reset` membersihkan search, draft filter, dan filter yang sudah diterapkan.
- Filter dropdown yang sudah diterapkan wajib muncul sebagai pill aktif langsung di bawah search bar. Search cukup tetap terlihat di input, tidak perlu dibuat pill terpisah. Pill menampilkan nama filter dan ringkasan nilai, serta bisa diklik untuk menghapus filter tersebut.
- Query parameter multi-value memakai array/repeated values, misalnya `executing_agency_ids` atau `location_ids`, sesuai kontrak API. Jangan mengambil semua data lalu filter lokal jika endpoint sudah paginated.

---

## Perbedaan Kritis Tailwind v4 vs v3

> **WAJIB DIBACA AGENT** — Tailwind v4 berbeda secara fundamental dari v3. Jangan menerapkan pola v3.

| Aspek | Tailwind v3 (lama) | Tailwind v4 (sekarang) |
|-------|-------------------|----------------------|
| Konfigurasi | `tailwind.config.ts` (JS/TS) | **CSS-first** — via `@theme` di dalam file CSS |
| Import di CSS | `@tailwind base/components/utilities` | `@import "tailwindcss"` |
| Vite integration | PostCSS plugin | **`@tailwindcss/vite`** Vite plugin — tidak perlu PostCSS |
| Content scanning | Array `content` di config | **Otomatis** — Vite plugin handle sendiri |
| CSS layer order (PrimeVue) | `tailwind-base, primevue, tailwind-utilities` | **`theme, base, primevue`** |
| Custom tokens | `theme.extend` di JS config | `@theme` directive di CSS |
| PrimeVue bridge | JS plugin di `tailwind.config.ts` | `@import "tailwindcss-primeui"` di CSS |

---

## Aturan Penggunaan PrimeVue dan Tailwind

> **PENTING:** PrimeVue dan Tailwind memiliki peran yang berbeda dan **tidak boleh dicampur untuk hal yang sama**.

| Keperluan | Gunakan | Contoh |
|-----------|---------|--------|
| Komponen UI (Button, Table, Dialog, dll.) | **PrimeVue** | `<Button>`, `<DataTable>`, `<Dialog>` |
| Layout & spacing | **Tailwind** | `flex gap-4`, `p-6`, `grid grid-cols-3` |
| Warna semantik dari tema PrimeVue | **`tailwindcss-primeui` utilities** | `bg-primary`, `text-surface-500`, `text-muted-color` |
| Custom design token baru | **`@theme` di CSS** | `--color-brand-accent: ...` |
| Style internal komponen PrimeVue | **Pass-Through (PT) API** | `pt:` props atau `definePreset` |
| Animasi komponen | **PrimeVue** | Bawaan PrimeVue (Dialog transition, dll.) |

**Jangan override style PrimeVue menggunakan Tailwind class langsung** — gunakan **Pass-Through (PT) API** atau `definePreset`.

---

## Setup Tailwind v4 — `vite.config.ts`

Tailwind v4 menggunakan **Vite plugin**, bukan PostCSS. Hapus `postcss.config.ts` jika ada.

```typescript
// vite.config.ts
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'   // ← Tailwind v4 Vite plugin
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [
    tailwindcss(),  // ← HARUS sebelum vue()
    vue(),
  ],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
})
```

> **Tidak ada `postcss.config.ts`** — Tailwind v4 tidak membutuhkan PostCSS.
> **Tidak ada `tailwind.config.ts`** — Semua konfigurasi pindah ke `main.css`.

---

## Setup CSS — `main.css`

Ini adalah file konfigurasi utama Tailwind v4. Semua yang dulu ada di `tailwind.config.ts` sekarang ada di sini.

```css
/* src/assets/styles/main.css */

/* 1. Import Tailwind v4 */
@import "tailwindcss";

/* 2. Import plugin bridge PrimeVue ↔ Tailwind */
/*    Ini yang membuat bg-primary, text-surface-500, dst. tersedia sebagai Tailwind utility */
@import "tailwindcss-primeui";

/* 3. Custom design tokens tambahan (opsional) */
/*    Gunakan @theme — BUKAN theme.extend di tailwind.config.ts (file itu tidak ada) */
@theme {
  /* Font custom PRISM */
  --font-sans: 'Inter Variable', ui-sans-serif, system-ui;

  /* Warna brand tambahan di luar PrimeVue token */
  --color-brand-muted: oklch(0.92 0.01 240);

  /* Breakpoint tambahan jika dibutuhkan */
  --breakpoint-3xl: 1920px;
}

/* 4. Global base styles */
@layer base {
  body {
    @apply bg-surface-50 text-surface-900;
  }
}
```

> **Tidak ada** `@tailwind base`, `@tailwind components`, atau `@tailwind utilities` — itu syntax v3.
> **Tidak ada** konfigurasi `cssLayer` di CSS — layer order diatur di `main.ts` via PrimeVue options.

---

## Setup PrimeVue di `main.ts`

```typescript
// src/main.ts
import { createApp } from 'vue'
import { createPinia } from 'pinia'
import PrimeVue from 'primevue/config'
import ToastService from 'primevue/toastservice'
import ConfirmationService from 'primevue/confirmationservice'
import Tooltip from 'primevue/tooltip'
import { prismPreset } from '@/assets/styles/theme'
import router from './router'

import '@/assets/styles/main.css'  // ← Import CSS sebelum mount
import App from './App.vue'

const app = createApp(App)

app.use(createPinia())
app.use(router)
app.use(PrimeVue, {
  theme: {
    preset: prismPreset,
    options: {
      darkModeSelector: '.dark',
      // Tailwind v4: layer order BERBEDA dari v3
      // v4 pakai: 'theme, base, primevue'
      // v3 pakai: 'tailwind-base, primevue, tailwind-utilities' — JANGAN dipakai untuk v4
      cssLayer: {
        name: 'primevue',
        order: 'theme, base, primevue',
      },
    },
  },
})
app.use(ToastService)
app.use(ConfirmationService)
app.directive('tooltip', Tooltip)

app.mount('#app')
```

---

## Kustomisasi Theme PrimeVue — `theme.ts`

Kustomisasi warna dan style PrimeVue dilakukan **hanya** di `src/assets/styles/theme.ts` menggunakan `definePreset`. Jangan override via `@theme` di CSS atau Tailwind class langsung.

```typescript
// src/assets/styles/theme.ts
import { definePreset } from '@primevue/themes'
import Aura from '@primevue/themes/aura'

export const prismPreset = definePreset(Aura, {
  semantic: {
    primary: {
      50:  '{blue.50}',
      100: '{blue.100}',
      200: '{blue.200}',
      300: '{blue.300}',
      400: '{blue.400}',
      500: '{blue.500}',
      600: '{blue.600}',
      700: '{blue.700}',
      800: '{blue.800}',
      900: '{blue.900}',
      950: '{blue.950}',
    },
  },
})
```

> Design token PrimeVue (warna `primary`, `surface`, dll.) **otomatis tersedia sebagai Tailwind utility**
> berkat `tailwindcss-primeui` — misalnya: `bg-primary`, `text-surface-500`, `border-primary-200`.

---


```typescript
// services/http.ts
import axios from 'axios'
import { useAuthStore } from '@/stores/auth.store'
import router from '@/router'

const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000,
})

// Inject JWT ke setiap request
http.interceptors.request.use(config => {
  const auth = useAuthStore()
  if (auth.token) {
    config.headers.Authorization = `Bearer ${auth.token}`
  }
  return config
})

// Handle 401 — redirect ke login
http.interceptors.response.use(
  response => response,
  error => {
    if (error.response?.status === 401) {
      const auth = useAuthStore()
      auth.logout()
      router.push('/login')
    }
    return Promise.reject(error)
  }
)

export default http
```

---

## Permission Guard

Cek permission dilakukan di dua tempat: route guard dan dalam komponen.

```typescript
// router/index.ts — route guard global
router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return next('/login')
  }

  if (to.meta.permission) {
    const { module, action } = to.meta.permission as { module: string; action: string }
    if (!auth.can(module, action)) {
      return next('/forbidden')
    }
  }

  next()
})
```

```typescript
// composables/usePermission.ts
export function usePermission() {
  const auth = useAuthStore()

  const can = (module: string, action: 'create' | 'read' | 'update' | 'delete') => {
    if (auth.user?.role === 'ADMIN') return true
    const perm = auth.permissions.find(p => p.module === module)
    if (!perm) return false
    return perm[`can_${action}`]
  }

  return { can }
}
```

```vue
<!-- Penggunaan di komponen -->
<script setup>
const { can } = usePermission()
</script>

<template>
  <Button
    v-if="can('bb_project', 'create')"
    label="Tambah Proyek"
    @click="openForm"
  />
</template>
```

---

## SSE Integration

```typescript
// composables/useSSE.ts
import { onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth.store'
import { useNotificationStore } from '@/stores/notification.store'

export function useSSE() {
  const auth = useAuthStore()
  const notif = useNotificationStore()
  let source: EventSource | null = null

  const connect = () => {
    const url = `${import.meta.env.VITE_API_BASE_URL}/events?token=${auth.token}`
    source = new EventSource(url)

    source.addEventListener('bb_project.created', (e) => {
      notif.add({ type: 'bb_project.created', data: JSON.parse(e.data) })
    })

    source.addEventListener('monitoring.updated', (e) => {
      notif.add({ type: 'monitoring.updated', data: JSON.parse(e.data) })
    })

    source.onerror = () => {
      source?.close()
      // Reconnect setelah 5 detik
      setTimeout(connect, 5000)
    }
  }

  const disconnect = () => source?.close()

  onUnmounted(disconnect)

  return { connect, disconnect }
}
```

---

## Form Composable — Contoh GB Project

Form GB Project adalah form paling kompleks di PRISM karena memiliki 4 tabel nested yang saling berelasi.

```typescript
// composables/forms/useGBProjectForm.ts
import { ref, computed } from 'vue'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { gbProjectSchema } from '@/schemas/green-book.schema'
import type { GBProjectRequest } from '@/types/green-book.types'

export function useGBProjectForm(initialData?: Partial<GBProjectRequest>) {
  // Form utama
  const { handleSubmit, resetForm, values, errors } = useForm({
    validationSchema: toTypedSchema(gbProjectSchema),
    initialValues: initialData ?? {},
  })

  // Activities — tabel dengan sort_order
  const activities = ref<ActivityRow[]>(initialData?.activities ?? [])

  const addActivity = () => {
    activities.value.push({
      activity_name: '',
      implementation_location: '',
      piu: '',
      sort_order: activities.value.length,
    })
  }

  const removeActivity = (index: number) => {
    activities.value.splice(index, 1)
    // Recalculate sort_order
    activities.value.forEach((a, i) => (a.sort_order = i))
  }

  // Funding Source — cofinancing, satu baris per lender
  const fundingSources = ref<FundingSourceRow[]>(initialData?.funding_sources ?? [])

  const addFundingSource = () => {
    fundingSources.value.push({
      lender_id: '',
      institution_id: null,
      loan_usd: 0,
      grant_usd: 0,
      local_usd: 0,
    })
  }

  // Funding Allocation — activity_index merujuk ke index activities
  const fundingAllocations = computed(() =>
    activities.value.map((_, index) => ({
      activity_index: index,
      services: 0,
      constructions: 0,
      goods: 0,
      trainings: 0,
      other: 0,
    }))
  )

  // Disbursement Plan — per tahun
  const disbursementPlan = ref<DisbursementPlanRow[]>(initialData?.disbursement_plan ?? [])

  const addDisbursementYear = (year: number) => {
    if (disbursementPlan.value.find(d => d.year === year)) return
    disbursementPlan.value.push({ year, amount_usd: 0 })
    disbursementPlan.value.sort((a, b) => a.year - b.year)
  }

  const submit = handleSubmit(async (formValues) => {
    return {
      ...formValues,
      activities: activities.value,
      funding_sources: fundingSources.value,
      disbursement_plan: disbursementPlan.value,
      funding_allocations: fundingAllocations.value,
    }
  })

  return {
    values, errors,
    activities, addActivity, removeActivity,
    fundingSources, addFundingSource,
    fundingAllocations,
    disbursementPlan, addDisbursementYear,
    submit, resetForm,
  }
}
```

---

## Zod Schema — Contoh

```typescript
// schemas/green-book.schema.ts
import { z } from 'zod'

export const gbProjectSchema = z.object({
  program_title_id:    z.string().uuid('Program Title wajib dipilih'),
  gb_code:             z.string().min(1, 'GB Code wajib diisi'),
  project_name:        z.string().min(1, 'Nama proyek wajib diisi'),
  duration:            z.number().int().positive().optional().nullable(),
  objective:           z.string().optional(),
  scope_of_project:    z.string().optional(),
  bb_project_ids:      z.array(z.string().uuid()).min(1, 'Minimal 1 BB Project'),
  executing_agency_ids:   z.array(z.string().uuid()).min(1, 'Executing Agency wajib diisi'),
  implementing_agency_ids: z.array(z.string().uuid()).min(1, 'Implementing Agency wajib diisi'),
  location_ids:        z.array(z.string().uuid()).min(1, 'Lokasi wajib dipilih'),
})

export type GBProjectFormValues = z.infer<typeof gbProjectSchema>
```

`LocationMultiSelect` wajib memuat seluruh level region (`COUNTRY`, `PROVINCE`, `CITY`) melalui `fetchAllRegionLevels()`, bukan hanya daftar default/paginated. Durasi proyek di form BB Project, GB Project, dan DK Project adalah angka bulan (`number | null`) dan dikirim sebagai integer. Currency mulai dicatat pada Funding Source Green Book dengan `CurrencySelect` dari Master Currency aktif; jika currency `USD`, form tidak meminta input USD terpisah dan payload menyamakan nilai USD dengan nilai original. Pada DK Project, picker `GB Project` ditempatkan paling atas dan perubahan pilihan memanggil autofill di `useDKProjectForm()` untuk mengisi field turunan dari GB Project terpilih; hasilnya tetap editable sebelum submit.

---

## Pola Penamaan

| Konteks | Konvensi | Contoh |
|---------|----------|--------|
| File page | `<Noun>Page.vue` atau `<Noun><Action>Page.vue` | `BBProjectFormPage.vue` |
| File komponen | `PascalCase.vue` | `FundingSourceTable.vue` |
| File composable | `use<Noun>.ts` | `useGBProjectForm.ts` |
| File service | `<modul>.service.ts` | `green-book.service.ts` |
| File store | `<modul>.store.ts` | `green-book.store.ts` |
| File schema | `<modul>.schema.ts` | `green-book.schema.ts` |
| File types | `<modul>.types.ts` | `green-book.types.ts` |
| Route path | `kebab-case` | `/green-books/:id/projects` |
| Pinia store id | `kebab-case` | `defineStore('green-book', ...)` |
| Env variable | `VITE_<NAMA>` | `VITE_API_BASE_URL` |

---

## Panduan untuk Coding Agent

Bagian ini ditujukan khusus untuk coding agent (Claude Code, Copilot, Cursor, dll.) agar output yang dihasilkan konsisten dengan arsitektur PRISM frontend.

---

### Konteks Proyek

Ini adalah frontend untuk **PRISM** — sistem internal Bappenas. Ditulis dalam **Vue 3** dengan Composition API dan TypeScript. Semua komponen menggunakan `<script setup>` syntax.

Referensi utama sebelum menulis kode:
- API contract: `PRISM_API_Contract.md` — untuk mengetahui struktur request/response
- Backend DDL: `prism_ddl.sql` — untuk memahami relasi data
- Types: `src/types/` — selalu gunakan type yang sudah ada, jangan definisikan ulang

---

### Aturan Wajib

**1. Selalu gunakan `<script setup>` — tidak ada Options API**

```vue
<!-- BENAR ✓ -->
<script setup lang="ts">
import { ref, computed } from 'vue'
const count = ref(0)
</script>

<!-- SALAH ✗ -->
<script lang="ts">
export default {
  data() { return { count: 0 } }
}
</script>
```

**2. Semua API call melalui service — tidak ada axios di komponen atau store**

```typescript
// BENAR ✓ — di komponen atau store
import { GreenBookService } from '@/services/green-book.service'
const project = await GreenBookService.getProject(id)

// SALAH ✗ — axios langsung di komponen
import axios from 'axios'
const project = await axios.get(`/api/v1/green-books/${id}/projects/${projectId}`)
```

**3. Semua state global di Pinia store — tidak ada provide/inject untuk state global**

```typescript
// BENAR ✓
const gbStore = useGreenBookStore()
await gbStore.fetchProject(id)

// SALAH ✗ — state di komponen parent, di-provide ke child
provide('currentProject', project)
```

**4. Validasi form selalu menggunakan Zod schema yang sudah ada di `src/schemas/`**

```typescript
// BENAR ✓
import { gbProjectSchema } from '@/schemas/green-book.schema'
const { handleSubmit } = useForm({
  validationSchema: toTypedSchema(gbProjectSchema),
})

// SALAH ✗ — validasi manual di dalam komponen
if (!form.gb_code) {
  errors.gb_code = 'GB Code wajib diisi'
}
```

**5. Cek permission sebelum render tombol aksi — selalu gunakan `usePermission()`**

```vue
<!-- BENAR ✓ -->
<script setup>
const { can } = usePermission()
</script>
<template>
  <Button v-if="can('gb_project', 'create')" label="Tambah" />
  <Button v-if="can('gb_project', 'delete')" label="Hapus" severity="danger" />
</template>

<!-- SALAH ✗ — cek role langsung -->
<template>
  <Button v-if="auth.user.role === 'ADMIN'" label="Tambah" />
</template>
```

**6. Jangan pernah hardcode URL API — selalu gunakan service**

```typescript
// BENAR ✓
const response = await GreenBookService.listProjects(greenBookId, params)

// SALAH ✗
const response = await http.get(`/api/v1/green-books/${greenBookId}/projects`)
```

**7. Form dengan tabel nested (Activities, Funding Source, dll.) wajib menggunakan composable dari `src/composables/forms/`**

```vue
<!-- BENAR ✓ -->
<script setup>
const { activities, addActivity, removeActivity, fundingSources } = useGBProjectForm()
</script>

<!-- SALAH ✗ — manage state tabel langsung di komponen page -->
<script setup>
const activities = ref([])
const addActivity = () => activities.value.push({ ... })
</script>
```

**8. Tipe data selalu diimport dari `src/types/` — tidak mendefinisikan interface baru di file komponen**

```typescript
// BENAR ✓
import type { GBProject, GBProjectRequest } from '@/types/green-book.types'

// SALAH ✗ — interface didefinisikan di komponen
interface GBProject {
  id: string
  gb_code: string
}
```

**9. Jangan gunakan syntax Tailwind v3 di CSS**

```css
/* BENAR ✓ — Tailwind v4 */
@import "tailwindcss";
@import "tailwindcss-primeui";

@theme {
  --font-sans: 'Inter Variable', ui-sans-serif;
}

/* SALAH ✗ — ini syntax Tailwind v3, tidak akan bekerja di v4 */
@tailwind base;
@tailwind components;
@tailwind utilities;
```

**10. Jangan buat atau edit `tailwind.config.ts` — file ini tidak digunakan di v4**

```typescript
// SALAH ✗ — file ini tidak ada dan tidak boleh dibuat
// tailwind.config.ts
export default {
  content: ['./src/**/*.vue'],  // tidak perlu — Vite plugin handle otomatis
  theme: { extend: {} },        // gunakan @theme di CSS
}
```

**11. Gunakan `cssLayer order` yang benar untuk Tailwind v4**

```typescript
// BENAR ✓ — untuk Tailwind v4
cssLayer: {
  name: 'primevue',
  order: 'theme, base, primevue',
}

// SALAH ✗ — ini untuk Tailwind v3
cssLayer: {
  name: 'primevue',
  order: 'tailwind-base, primevue, tailwind-utilities',
}
```

---

### Urutan Langkah Saat Menambah Fitur Baru

```
1. Cek API contract di PRISM_API_Contract.md
         ↓
2. Tambah/update types di src/types/<modul>.types.ts
         ↓
3. Tambah/update Zod schema di src/schemas/<modul>.schema.ts
         ↓
4. Implementasi service di src/services/<modul>.service.ts
         ↓
5. Buat/update Pinia store di src/stores/<modul>.store.ts
         ↓
6. Buat composable jika ada form kompleks di src/composables/forms/
         ↓
7. Buat komponen reusable di src/components/<modul>/
         ↓
8. Buat page di src/pages/<modul>/
         ↓
9. Daftarkan route di src/router/routes/<modul>.routes.ts
```

---

### Hal yang Tidak Boleh Dilakukan Agent

| Larangan | Alasan |
|----------|--------|
| Gunakan Options API | Proyek ini sepenuhnya Composition API dengan `<script setup>` |
| Panggil axios langsung di komponen | Semua HTTP call harus melalui layer service |
| Definisikan interface di file `.vue` | Types harus di `src/types/` agar bisa di-reuse |
| Manage state form tabel nested di page langsung | Gunakan composable di `src/composables/forms/` |
| Cek `user.role === 'ADMIN'` untuk render tombol | Selalu gunakan `usePermission()` |
| Hardcode string URL API | Semua endpoint harus ada di file service |
| Buat store baru tanpa struktur `state/getters/actions` yang konsisten | Ikuti pola store yang sudah ada |
| Import komponen PrimeVue tanpa melalui plugin | PrimeVue sudah di-register global di `main.ts` |
| Gunakan `any` sebagai tipe | Selalu gunakan type yang strongly-typed dari `src/types/` |
| Buat atau edit `tailwind.config.ts` | File ini tidak digunakan di Tailwind v4 |
| Buat `postcss.config.ts` | Tailwind v4 menggunakan Vite plugin, bukan PostCSS |
| Gunakan `@tailwind base/components/utilities` di CSS | Syntax Tailwind v3 — gunakan `@import "tailwindcss"` |
| Gunakan `theme.extend` untuk custom token | Gunakan `@theme` directive di `main.css` |
| Gunakan `cssLayer order: 'tailwind-base, primevue, tailwind-utilities'` | Itu syntax v3 — gunakan `'theme, base, primevue'` untuk v4 |
| Override style PrimeVue via Tailwind class atau `!important` | Gunakan PT API atau `definePreset` |

---

### Struktur Pinia Store yang Konsisten

Semua store mengikuti pola ini:

```typescript
// stores/green-book.store.ts
import { defineStore } from 'pinia'
import { ref } from 'vue'
import { GreenBookService } from '@/services/green-book.service'
import type { GBProject, GBProjectListParams } from '@/types/green-book.types'

export const useGreenBookStore = defineStore('green-book', () => {
  // State
  const projects = ref<GBProject[]>([])
  const currentProject = ref<GBProject | null>(null)
  const loading = ref(false)
  const total = ref(0)

  // Actions
  async function fetchProjects(greenBookId: string, params: GBProjectListParams) {
    loading.value = true
    try {
      const res = await GreenBookService.listProjects(greenBookId, params)
      projects.value = res.data
      total.value = res.meta.total
    } finally {
      loading.value = false
    }
  }

  async function fetchProject(greenBookId: string, id: string) {
    loading.value = true
    try {
      currentProject.value = await GreenBookService.getProject(greenBookId, id)
    } finally {
      loading.value = false
    }
  }

  function $reset() {
    projects.value = []
    currentProject.value = null
    loading.value = false
    total.value = 0
  }

  return { projects, currentProject, loading, total, fetchProjects, fetchProject, $reset }
})
```

---

### Catatan Khusus untuk Form GB Project

Form GB Project adalah form paling kompleks — selalu gunakan `useGBProjectForm()` dan pecah tampilan menjadi beberapa komponen tab:

```
GBProjectFormPage.vue
  ├── Tab 1: Informasi Umum    → field utama GB Project
  ├── Tab 2: Activities        → <ActivitiesTable /> (editable, drag reorder)
  ├── Tab 3: Funding Source    → <FundingSourceTable /> (satu baris per lender)
  ├── Tab 4: Disbursement Plan → <DisbursementPlanTable /> (satu baris per tahun)
  └── Tab 5: Funding Allocation → <FundingAllocationTable /> (baris = activities dari Tab 2)
```

Kolom Activities di Funding Allocation **harus selalu sinkron** dengan baris di tabel Activities. Ini sudah ditangani di `useGBProjectForm()` via `computed` — agent tidak perlu mengimplementasikan ulang logika ini.
