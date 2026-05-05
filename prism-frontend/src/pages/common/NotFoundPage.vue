<script setup lang="ts">
import { useRouter } from 'vue-router'
import Button from 'primevue/button'

import { useAuthStore } from '@/stores/auth.store'
import { resolveDefaultAuthenticatedRoute } from '@/utils/default-route'

const router = useRouter()
const auth = useAuthStore()
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-surface-100 px-4">
    <section class="w-full max-w-xl rounded-lg bg-white p-8 text-center shadow-lg ring-1 ring-surface-200">
      <i class="pi pi-search text-5xl text-primary" />
      <h1 class="mt-6 text-3xl font-semibold text-surface-950">404 - Halaman tidak ditemukan</h1>
      <Button
        class="mt-6"
        label="Kembali ke Halaman Utama"
        icon="pi pi-arrow-left"
        @click="router.push(resolveDefaultAuthenticatedRoute({ user: auth.user, permissions: auth.permissions }))"
      />
    </section>
  </div>
</template>
