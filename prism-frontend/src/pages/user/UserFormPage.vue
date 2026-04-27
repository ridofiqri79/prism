<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Select from 'primevue/select'
import ToggleSwitch from 'primevue/toggleswitch'
import PageHeader from '@/components/common/PageHeader.vue'
import { useToast } from '@/composables/useToast'
import { createUserSchema, updateUserSchema } from '@/schemas/user.schema'
import { useUserStore } from '@/stores/user.store'
import type { UserRole } from '@/types/auth.types'

interface UserFormValues {
  username: string
  email: string
  password?: string
  role: UserRole
  is_active: boolean
}

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()
const toast = useToast()

const isEditMode = computed(() => route.name === 'user-edit')
const userId = computed(() => String(route.params.id ?? ''))
const roleOptions: UserRole[] = ['ADMIN', 'STAFF']

const { defineField, errors, handleSubmit, setValues } = useForm<UserFormValues>({
  validationSchema: toTypedSchema(isEditMode.value ? updateUserSchema : createUserSchema),
  initialValues: {
    username: '',
    email: '',
    password: '',
    role: 'STAFF',
    is_active: true,
  },
})

const [username] = defineField('username')
const [email] = defineField('email')
const [password] = defineField('password')
const [role] = defineField('role')
const [isActive] = defineField('is_active')

const pageTitle = computed(() => (isEditMode.value ? 'Edit User' : 'Tambah User'))

const onSubmit = handleSubmit(async (values) => {
  if (isEditMode.value) {
    await userStore.updateUser(userId.value, {
      username: values.username,
      email: values.email,
      role: values.role,
      is_active: values.is_active,
    })
    toast.success('Berhasil', 'User berhasil diperbarui')
  } else {
    await userStore.createUser({
      username: values.username,
      email: values.email,
      password: values.password ?? '',
      role: values.role,
    })
    toast.success('Berhasil', 'User berhasil dibuat')
  }

  await router.push({ name: 'users' })
})

onMounted(async () => {
  if (!isEditMode.value) {
    return
  }

  const user = await userStore.fetchUser(userId.value)

  setValues({
    username: user.username,
    email: user.email,
    role: user.role,
    is_active: user.is_active,
  })
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader :title="pageTitle" subtitle="Akun ADMIN memiliki akses penuh, STAFF mengikuti permission matrix">
      <template #actions>
        <Button label="Kembali" icon="pi pi-arrow-left" outlined @click="router.push({ name: 'users' })" />
      </template>
    </PageHeader>

    <form class="max-w-2xl space-y-5 rounded-lg border border-surface-200 bg-white p-6" @submit.prevent="onSubmit">
      <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Username</span>
        <InputText v-model="username" class="w-full" :invalid="Boolean(errors.username)" />
        <small v-if="errors.username" class="text-red-600">{{ errors.username }}</small>
      </label>

      <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Email</span>
        <InputText v-model="email" class="w-full" :invalid="Boolean(errors.email)" />
        <small v-if="errors.email" class="text-red-600">{{ errors.email }}</small>
      </label>

      <label v-if="!isEditMode" class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Password</span>
        <InputText v-model="password" class="w-full" type="password" :invalid="Boolean(errors.password)" />
        <small v-if="errors.password" class="text-red-600">{{ errors.password }}</small>
      </label>

      <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Role</span>
        <Select v-model="role" :options="roleOptions" class="w-full" />
        <small v-if="errors.role" class="text-red-600">{{ errors.role }}</small>
      </label>

      <label v-if="isEditMode" class="flex items-center justify-between rounded-lg border border-surface-200 p-4">
        <span class="text-sm font-medium text-surface-700">User aktif</span>
        <ToggleSwitch v-model="isActive" />
      </label>

      <div class="flex justify-end gap-2 border-t border-surface-200 pt-5">
        <Button label="Batal" severity="secondary" outlined @click="router.push({ name: 'users' })" />
        <Button type="submit" label="Simpan" icon="pi pi-save" :loading="userStore.loading" />
      </div>
    </form>
  </section>
</template>
