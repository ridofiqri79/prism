# PLAN 00 — Project Setup & Foundation Files

> **Scope:** Inisialisasi frontend, konfigurasi semua tools, buat semua file foundation.
> **Deliverable:** Frontend berjalan di Docker dev, routing bekerja, PrimeVue + Tailwind v4 aktif.
> **Referensi:** docs/PRISM_Frontend_Structure.md

---

## Instruksi untuk Codex

Baca dulu:
- `docs/PRISM_Frontend_Structure.md` — struktur folder, aturan Tailwind v4, setup PrimeVue
- `docs/PRISM_Business_Rules.md` — aturan bisnis dasar

Semua file dibuat di dalam `prism-frontend/`.

---

## Task 1 — vite.config.ts

- Plugin `@tailwindcss/vite` harus SEBELUM `@vitejs/plugin-vue`
- Alias `@` ke `./src`
- `server.host: '0.0.0.0'`, `server.port: 5173` (wajib untuk Docker)
- Tidak ada `postcss.config.ts`, tidak ada `tailwind.config.ts`

---

## Task 2 — src/assets/styles/main.css

```css
@import "tailwindcss";
@import "tailwindcss-primeui";

@theme {
  --font-sans: 'Inter Variable', ui-sans-serif, system-ui;
}

@layer base {
  body { @apply bg-surface-50 text-surface-900; }
}
```

Tidak ada `@tailwind base/components/utilities` — itu syntax v3.

---

## Task 3 — src/assets/styles/theme.ts

```typescript
import { definePreset } from '@primevue/themes'
import Aura from '@primevue/themes/aura'

export const prismPreset = definePreset(Aura, {
  semantic: {
    primary: {
      50: '{blue.50}', 100: '{blue.100}', 200: '{blue.200}',
      300: '{blue.300}', 400: '{blue.400}', 500: '{blue.500}',
      600: '{blue.600}', 700: '{blue.700}', 800: '{blue.800}',
      900: '{blue.900}', 950: '{blue.950}',
    },
  },
})
```

---

## Task 4 — src/main.ts

- Import `main.css` sebelum mount
- `app.use(PrimeVue, { theme: { preset: prismPreset, options: { darkModeSelector: '.dark', cssLayer: { name: 'primevue', order: 'theme, base, primevue' } } } })`
- Register: `ToastService`, `ConfirmationService`, directive `Tooltip`
- Urutan: createPinia → router → PrimeVue → services → mount

---

## Task 5 — src/types/api.types.ts

```typescript
export interface ApiResponse<T> { data: T }
export interface PaginatedResponse<T> { data: T[]; meta: PaginationMeta }
export interface PaginationMeta { page: number; limit: number; total: number; total_pages: number }
export interface ApiError { code: string; message: string; details?: FieldError[] }
export interface FieldError { field: string; message: string }
```

---

## Task 6 — src/types/auth.types.ts

```typescript
export interface AuthUser { id: string; username: string; email: string; role: 'ADMIN' | 'STAFF' }
export interface UserPermission { module: string; can_create: boolean; can_read: boolean; can_update: boolean; can_delete: boolean }
export interface LoginResponse { access_token: string; expires_in: number; user: AuthUser }
```

---

## Task 7 — src/services/http.ts

Axios instance:
- `baseURL: import.meta.env.VITE_API_BASE_URL`, `timeout: 30000`
- Request interceptor: inject `Authorization: Bearer <token>` dari localStorage jika ada
- Response interceptor:
  - 401 → clear localStorage, redirect `/login`
  - 403 → tampilkan toast "Akses Ditolak"
  - 500 → tampilkan toast "Terjadi Kesalahan Server"

---

## Task 8 — src/stores/auth.store.ts

State: `user: AuthUser | null`, `token: string | null`, `permissions: UserPermission[]`, `loading: boolean`

Computed: `isAuthenticated` → `token !== null`

Actions:
- `login(username, password)` → POST `/auth/login`, simpan token ke store + localStorage
- `logout()` → clear store + localStorage, push `/login`
- `fetchMe()` → GET `/auth/me`, update user dan permissions
- `can(module, action)` → ADMIN selalu true; STAFF cek permissions array
- `restoreSession()` → baca token dari localStorage, call `fetchMe()` jika ada

---

## Task 9 — src/router/index.ts + routes/

Router dengan `createWebHistory`. Lazy loading semua route.

Navigation guard:
1. Panggil `auth.restoreSession()` sekali saat init (beforeEach hanya jika belum done)
2. Route dengan `meta.requiresAuth` → redirect `/login` jika belum auth
3. Route dengan `meta.adminOnly` → redirect `/forbidden` jika bukan ADMIN

Buat file route terpisah di `src/router/routes/`:
- `auth.routes.ts` → `/login` (AuthLayout, tidak requiresAuth)
- `home.routes.ts` → `/` redirect ke halaman utama pertama yang boleh diakses user (requiresAuth)
- `master.routes.ts` → `/master/countries`, `/master/lenders`, `/master/institutions`, `/master/regions`, `/master/program-titles`, `/master/bappenas-partners`, `/master/periods`, `/master/national-priorities`
- `blue-book.routes.ts` → `/blue-books`, `/blue-books/:id`, `/blue-books/:bbId/projects/new`, `/blue-books/:bbId/projects/:id`, `/blue-books/:bbId/projects/:id/edit`
- `green-book.routes.ts` → pola sama dengan blue book
- `daftar-kegiatan.routes.ts` → `/daftar-kegiatan`, `/daftar-kegiatan/:id`, `/daftar-kegiatan/:dkId/projects/new`, dst.
- `loan-agreement.routes.ts` → `/loan-agreements`, `/loan-agreements/new`, `/loan-agreements/:id`, `/loan-agreements/:id/edit`
- `monitoring.routes.ts` → `/loan-agreements/:laId/monitoring`, `/loan-agreements/:laId/monitoring/new`
- `journey.routes.ts` → `/journey/:bbProjectId`
- `user.routes.ts` → `/users`, `/users/new`, `/users/:id/edit`, `/users/:id/permissions` (adminOnly)

Semua halaman pakai komponen placeholder `defineComponent({ template: '<div class="p-6">TODO: {{ $route.name }}</div>' })` dulu — akan diganti di plan berikutnya.

---

## Task 10 — src/layouts/

**AppLayout.vue:**
```vue
<template>
  <div class="flex h-screen overflow-hidden">
    <AppSidebar />
    <div class="flex flex-col flex-1 overflow-hidden">
      <AppTopbar />
      <main class="flex-1 overflow-y-auto p-6">
        <RouterView />
      </main>
    </div>
  </div>
</template>
```

**AuthLayout.vue:**
```vue
<template>
  <div class="min-h-screen flex items-center justify-center bg-surface-100">
    <RouterView />
  </div>
</template>
```

**AppSidebar.vue:**
- Menu item list statis dengan `<RouterLink>` ke semua route utama
- Icon PrimeIcons per item
- Sembunyikan "Users" jika bukan ADMIN (`auth.user?.role !== 'ADMIN'`)

**AppTopbar.vue:**
- Tampilkan `auth.user?.username` dan badge role
- Tombol logout → call `auth.logout()`

---

## Task 11 — Halaman Sederhana

Buat halaman berikut (bisa minimal, hanya scaffold):
- `src/pages/auth/LoginPage.vue` — form login placeholder (akan dilengkapi Plan 02)
- `src/pages/common/HomeRedirectPage.vue` — text "Mengalihkan..."
- `src/pages/common/ForbiddenPage.vue` — text "403 Forbidden"
- `src/pages/common/NotFoundPage.vue` — text "404 Not Found"

---

## Verifikasi

```bash
docker compose -f docker-compose.dev.yml up --build
```

- `http://localhost:5173` terbuka, redirect ke `/login`
- Navigasi ke `/` redirect ke `/login`
- PrimeVue `<Button label="Test" />` bisa dirender
- `bg-primary` dan `text-surface-500` bekerja di elemen apapun

---

## Checklist

- [x] `vite.config.ts` — tailwindcss plugin + host 0.0.0.0
- [x] `main.css` — @import tailwindcss + tailwindcss-primeui + @theme
- [x] `theme.ts` — definePreset Aura + primary blue
- [x] `main.ts` — cssLayer order: 'theme, base, primevue'
- [x] `api.types.ts` + `auth.types.ts`
- [x] `http.ts` — Axios instance + interceptors
- [x] `auth.store.ts` — login/logout/fetchMe/can/restoreSession
- [x] `router/index.ts` + semua route files
- [x] `AppLayout.vue`, `AuthLayout.vue`, `AppSidebar.vue`, `AppTopbar.vue`
- [x] Halaman placeholder untuk semua route
- [ ] Docker dev berjalan tanpa error

Catatan verifikasi 2026-04-27:
- `docker compose -f docker-compose.dev.yml up --build` berhasil menjalankan `frontend` dan `postgres`.
- `http://localhost:5173` terbuka dan route guard me-redirect ke `/login` tanpa error browser.
- `backend` masih gagal start karena repo `prism-backend` belum memiliki entrypoint Go (`cmd/api` masih kosong), jadi poin terakhir belum bisa dicentang dari scope FE-00 saja.
