<script setup lang="ts">
import { onMounted } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import DataTable, { type ColumnDef } from '@/components/common/DataTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { useConfirm } from '@/composables/useConfirm'
import { usePagination } from '@/composables/usePagination'
import { useToast } from '@/composables/useToast'
import { useUserStore } from '@/stores/user.store'
import type { AppUser } from '@/types/user.types'

const router = useRouter()
const userStore = useUserStore()
const pagination = usePagination()
const confirm = useConfirm()
const toast = useToast()

const columns: ColumnDef[] = [
  { field: 'username', header: 'Username', sortable: true },
  { field: 'email', header: 'Email', sortable: true },
  { field: 'role', header: 'Peran' },
  { field: 'is_active', header: 'Status' },
  { field: 'actions', header: 'Aksi' },
]

async function loadUsers() {
  await userStore.fetchUsers(pagination.queryParams.value)
}

function goToCreate() {
  void router.push({ name: 'user-create' })
}

function goToEdit(user: AppUser) {
  void router.push({ name: 'user-edit', params: { id: user.id } })
}

function goToPermissions(user: AppUser) {
  void router.push({ name: 'user-permissions', params: { id: user.id } })
}

function confirmDeactivate(user: AppUser) {
  confirm.confirmDelete(`pengguna ${user.username}`, async () => {
    await userStore.updateUser(user.id, {
      username: user.username,
      email: user.email,
      role: user.role,
      is_active: false,
    })
    await loadUsers()
    toast.success('Berhasil', 'Pengguna berhasil dinonaktifkan')
  })
}

function handleSort(payload: { sort: string; order: 'asc' | 'desc' }) {
  pagination.sort.value = payload.sort
  pagination.order.value = payload.order
  pagination.resetPage()
  void loadUsers()
}

function handlePage(value: number) {
  pagination.setPage(value)
  void loadUsers()
}

function handleLimit(value: number) {
  pagination.limit.value = value
  pagination.resetPage()
  void loadUsers()
}

onMounted(() => {
  void loadUsers()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Manajemen Pengguna" subtitle="Kelola akun dan hak akses staff PRISM">
      <template #actions>
        <Button label="Tambah Pengguna" icon="pi pi-plus" @click="goToCreate" />
      </template>
    </PageHeader>

    <DataTable
      :data="userStore.users"
      :columns="columns"
      :loading="userStore.loading"
      :total="userStore.total"
      :page="pagination.page.value"
      :limit="pagination.limit.value"
      @update:page="handlePage"
      @update:limit="handleLimit"
      @sort="handleSort"
    >
      <template #body-row="{ row, column }">
        <Tag
          v-if="column.field === 'role'"
          :value="row.role"
          :severity="row.role === 'ADMIN' ? 'success' : 'info'"
          rounded
        />
        <StatusBadge
          v-else-if="column.field === 'is_active'"
          :status="row.is_active ? 'active' : 'deleted'"
        />
        <div v-else-if="column.field === 'actions'" class="flex flex-wrap gap-2">
          <Button icon="pi pi-pencil" label="Edit" size="small" outlined @click="goToEdit(row as AppUser)" />
          <Button
            icon="pi pi-shield"
            label="Atur Hak Akses"
            size="small"
            outlined
            @click="goToPermissions(row as AppUser)"
          />
          <Button
            icon="pi pi-user-minus"
            label="Nonaktifkan"
            size="small"
            severity="danger"
            outlined
            :disabled="!(row as AppUser).is_active"
            @click="confirmDeactivate(row as AppUser)"
          />
        </div>
        <span v-else>{{ row[column.field] }}</span>
      </template>
    </DataTable>
  </section>
</template>
