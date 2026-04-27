# PLAN 02 — Authentication & User Management

> **Scope:** Halaman login fungsional, session management, user management (ADMIN only).
> **Deliverable:** Login bekerja end-to-end, permission dicek di setiap route dan komponen.
> **Referensi:** docs/PRISM_API_Contract.md (Auth & User Management), docs/PRISM_Business_Rules.md (bagian 10)

---

## Task 1 — src/schemas/auth.schema.ts

```typescript
export const loginSchema = z.object({
  username: z.string().min(3, 'Username minimal 3 karakter'),
  password: z.string().min(6, 'Password minimal 6 karakter'),
})
```

---

## Task 2 — src/services/auth.service.ts

```typescript
export const AuthService = {
  login: (data: { username: string; password: string }) =>
    http.post<ApiResponse<LoginResponse>>('/auth/login', data).then(r => r.data.data),
  logout: () => http.post('/auth/logout'),
  getMe: () => http.get<ApiResponse<MeResponse>>('/auth/me').then(r => r.data.data),
}
```

---

## Task 3 — src/pages/auth/LoginPage.vue

- Gunakan `AuthLayout`
- Card centered: logo PRISM + form login
- Field: username, password (type=password dengan toggle show/hide)
- Validasi via `useForm({ validationSchema: toTypedSchema(loginSchema) })`
- Submit: `auth.login(values)` → success: push `/dashboard` → error: tampilkan error inline
- Loading state di tombol submit
- Pesan error 401: "Username atau password salah"

---

## Task 4 — Update AppSidebar.vue

- Sembunyikan menu item yang user tidak punya permission `read`:
  ```vue
  <li v-if="can('bb_project', 'read')">
    <RouterLink to="/blue-books">Blue Book</RouterLink>
  </li>
  ```
- Sembunyikan "Users" jika bukan ADMIN
- Sembunyikan "Master Data" section jika tidak ada satupun master module yang bisa di-read

---

## Task 5 — src/schemas/user.schema.ts

```typescript
export const createUserSchema = z.object({
  username: z.string().min(3),
  email: z.string().email('Format email tidak valid'),
  password: z.string().min(8, 'Password minimal 8 karakter'),
  role: z.enum(['ADMIN', 'STAFF']),
})

export const updateUserSchema = createUserSchema.omit({ password: true }).extend({
  is_active: z.boolean(),
})
```

---

## Task 6 — src/services/user.service.ts

- `getUsers(params?)`, `getUser(id)`
- `createUser(data)`, `updateUser(id, data)`, `deleteUser(id)`
- `getUserPermissions(id)`, `updatePermissions(id, permissions)`

---

## Task 7 — src/stores/user.store.ts

- State: `users: AppUser[]`, `currentUser: AppUser | null`, `loading: boolean`, `total: number`
- Actions: `fetchUsers`, `fetchUser`, `createUser`, `updateUser`, `deleteUser`, `updatePermissions`

---

## Task 8 — src/pages/user/UserListPage.vue

- `<PageHeader title="Manajemen User">` dengan slot actions: tombol "Tambah User"
- `<DataTable>` dengan kolom: username, email, role badge, status badge (active/inactive), actions (Edit, Set Permission, Nonaktifkan)
- Tombol "Nonaktifkan" dengan `<confirmDelete>` sebelum eksekusi
- Route actions ke halaman form dan permission

---

## Task 9 — src/pages/user/UserFormPage.vue

- Mode create: semua field termasuk password
- Mode edit: tanpa field password, ada toggle `is_active`
- Setelah save: kembali ke `UserListPage` dengan toast success

---

## Task 10 — src/pages/user/UserPermissionPage.vue

Halaman set permission per user (ADMIN only):

- Header: nama user yang sedang di-set permission
- Tabel matrix: baris = modul, kolom = Create / Read / Update / Delete
- `<Checkbox>` per sel
- Modul yang ditampilkan:
  `bb_project`, `gb_project`, `daftar_kegiatan`, `loan_agreement`, `monitoring_disbursement`,
  `institution`, `lender`, `region`, `national_priority`, `program_title`, `bappenas_partner`, `period`, `country`
- Tombol "Simpan Semua" → `PUT /users/:id/permissions` (replace-all, transaksional)
- Tampilkan toast sukses setelah save

---

## Task 11 — Halaman Forbidden & NotFound

**`src/pages/common/ForbiddenPage.vue`:**
- Icon + text "403 — Anda tidak memiliki izin mengakses halaman ini"
- Tombol "Kembali ke Dashboard"

**`src/pages/common/NotFoundPage.vue`:**
- Icon + text "404 — Halaman tidak ditemukan"
- Tombol "Kembali ke Dashboard"

---

## Checklist

- [ ] `auth.schema.ts`
- [ ] `auth.service.ts`
- [ ] `LoginPage.vue` — form fungsional + error handling
- [ ] `AppSidebar.vue` — menu conditional berdasarkan permission
- [ ] `user.schema.ts` — create + update schema
- [ ] `user.service.ts`
- [ ] `user.store.ts`
- [ ] `UserListPage.vue`
- [ ] `UserFormPage.vue` — create & edit mode
- [ ] `UserPermissionPage.vue` — permission matrix
- [ ] `ForbiddenPage.vue` + `NotFoundPage.vue`
