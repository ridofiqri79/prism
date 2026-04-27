<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useForm } from 'vee-validate'
import { toTypedSchema } from '@vee-validate/zod'
import { isAxiosError } from 'axios'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import { loginSchema, type LoginFormValues } from '@/schemas/auth.schema'
import { useAuthStore } from '@/stores/auth.store'

const router = useRouter()
const auth = useAuthStore()
const showPassword = ref(false)
const loginError = ref<string | null>(null)

const { defineField, errors, handleSubmit } = useForm<LoginFormValues>({
  validationSchema: toTypedSchema(loginSchema),
  initialValues: {
    username: '',
    password: '',
  },
})

const [username] = defineField('username')
const [password] = defineField('password')

const passwordType = computed(() => (showPassword.value ? 'text' : 'password'))

const onSubmit = handleSubmit(async (values) => {
  loginError.value = null

  try {
    await auth.login(values)
    await router.push({ name: 'dashboard' })
  } catch (err) {
    if (isAxiosError(err) && err.response?.status === 401) {
      loginError.value = 'Username atau password salah'
      return
    }

    loginError.value = 'Login gagal. Silakan coba lagi.'
  }
})
</script>

<template>
  <div class="w-full max-w-sm rounded-lg bg-white p-8 shadow-lg ring-1 ring-surface-200">
    <div class="mb-8 text-center">
      <p class="text-xs font-semibold uppercase tracking-[0.24em] text-primary">PRISM</p>
      <h1 class="mt-3 text-2xl font-semibold text-surface-950">Masuk ke PRISM</h1>
      <p class="mt-2 text-sm text-surface-500">Gunakan akun yang sudah diberikan admin.</p>
    </div>

    <form class="space-y-5" @submit.prevent="onSubmit">
      <Message v-if="loginError" severity="error" size="small" :closable="false">
        {{ loginError }}
      </Message>

      <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Username</span>
        <InputText
          v-model="username"
          class="w-full"
          autocomplete="username"
          :invalid="Boolean(errors.username)"
        />
        <small v-if="errors.username" class="text-red-600">{{ errors.username }}</small>
      </label>

      <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Password</span>
        <div class="flex">
          <InputText
            v-model="password"
            class="w-full rounded-r-none"
            :type="passwordType"
            autocomplete="current-password"
            :invalid="Boolean(errors.password)"
          />
          <Button
            type="button"
            :icon="showPassword ? 'pi pi-eye-slash' : 'pi pi-eye'"
            severity="secondary"
            outlined
            class="rounded-l-none"
            aria-label="Tampilkan password"
            @click="showPassword = !showPassword"
          />
        </div>
        <small v-if="errors.password" class="text-red-600">{{ errors.password }}</small>
      </label>

      <Button
        type="submit"
        label="Masuk"
        icon="pi pi-sign-in"
        class="w-full"
        :loading="auth.loading"
      />
    </form>
  </div>
</template>
