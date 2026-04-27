<script setup lang="ts">
import { reactive } from 'vue'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'

const credentials = reactive({
  username: '',
  password: '',
})
</script>

<template>
  <div class="w-full max-w-md rounded-[2rem] bg-white p-8 shadow-xl ring-1 ring-surface-200">
    <p class="text-sm font-semibold uppercase tracking-[0.24em] text-primary">PRISM</p>
    <h1 class="mt-3 text-3xl font-semibold text-surface-950">Login</h1>
    <p class="mt-2 text-sm text-surface-500">
      Scaffold autentikasi FE-00. Flow final akan dilengkapi pada Plan 02.
    </p>

    <form class="mt-8 space-y-5" @submit.prevent>
      <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Username</span>
        <InputText v-model="credentials.username" class="w-full" placeholder="Masukkan username" />
      </label>

      <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Password</span>
        <InputText
          v-model="credentials.password"
          class="w-full"
          placeholder="Masukkan password"
          type="password"
        />
      </label>

      <div class="flex items-center justify-between rounded-2xl bg-primary px-4 py-3 text-primary-contrast">
        <div>
          <p class="text-sm font-semibold">PrimeVue + Tailwind v4 aktif</p>
          <p class="text-xs text-primary-contrast/80">Verifikasi awal FE-00</p>
        </div>
        <Button label="Masuk" disabled />
      </div>
    </form>
  </div>
</template>
