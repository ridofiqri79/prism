<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import Button from 'primevue/button'
import Checkbox from 'primevue/checkbox'
import PageHeader from '@/components/common/PageHeader.vue'
import { useToast } from '@/composables/useToast'
import { useUserStore } from '@/stores/user.store'
import type { PermissionAction, UserPermission } from '@/types/auth.types'
import { permissionModules } from '@/types/user.types'

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const toast = useToast()

const userId = computed(() => String(route.params.id ?? ''))
const actions: PermissionAction[] = ['create', 'read', 'update', 'delete']
const actionLabels: Record<PermissionAction, string> = {
  create: 'Tambah',
  read: 'Lihat',
  update: 'Ubah',
  delete: 'Hapus',
}

const matrix = reactive<Record<string, UserPermission>>({})

const title = computed(() =>
  userStore.currentUser ? `Hak Akses ${userStore.currentUser.username}` : 'Hak Akses Pengguna',
)

function ensurePermission(module: string) {
  matrix[module] ??= {
    module,
    can_create: false,
    can_read: false,
    can_update: false,
    can_delete: false,
  }

  return matrix[module]
}

function getPermission(module: string, action: PermissionAction) {
  return ensurePermission(module)[`can_${action}`]
}

function setPermission(module: string, action: PermissionAction, value: boolean) {
  ensurePermission(module)[`can_${action}`] = value
}

function applyPermissions(permissions: UserPermission[]) {
  permissionModules.forEach((item) => {
    const existing = permissions.find((permission) => permission.module === item.module)
    matrix[item.module] = existing
      ? { ...existing }
      : {
          module: item.module,
          can_create: false,
          can_read: false,
          can_update: false,
          can_delete: false,
        }
  })
}

async function savePermissions() {
  const payload = permissionModules.map((item) => ensurePermission(item.module))

  await userStore.updatePermissions(userId.value, payload)
  toast.success('Berhasil', 'Hak akses pengguna berhasil disimpan')
}

onMounted(async () => {
  const [permissionsResult] = await Promise.allSettled([
    userStore.fetchUserPermissions(userId.value),
    userStore.fetchUser(userId.value),
  ])

  const permissions = permissionsResult.status === 'fulfilled' ? permissionsResult.value : []

  applyPermissions(permissions)
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader :title="title" subtitle="Setiap perubahan hak akses disimpan secara transaksional">
      <template #actions>
        <Button label="Kembali" icon="pi pi-arrow-left" outlined @click="router.push({ name: 'users' })" />
        <Button label="Simpan Semua" icon="pi pi-save" :loading="userStore.loading" @click="savePermissions" />
      </template>
    </PageHeader>

    <div class="overflow-hidden rounded-lg border border-surface-200 bg-white">
      <table class="w-full border-collapse text-sm">
        <thead class="bg-surface-50 text-left text-xs font-semibold uppercase tracking-wide text-surface-500">
          <tr>
            <th class="border-b border-surface-200 px-4 py-3 font-semibold">Modul</th>
            <th
              v-for="action in actions"
              :key="action"
              class="border-b border-surface-200 px-4 py-3 text-center font-semibold"
            >
              {{ actionLabels[action] }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in permissionModules" :key="item.module" class="border-b border-surface-100 last:border-b-0">
            <td class="px-4 py-3 font-medium text-surface-800">{{ item.label }}</td>
            <td v-for="action in actions" :key="action" class="px-4 py-3 text-center">
              <Checkbox
                binary
                :model-value="getPermission(item.module, action)"
                @update:model-value="setPermission(item.module, action, Boolean($event))"
              />
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>
