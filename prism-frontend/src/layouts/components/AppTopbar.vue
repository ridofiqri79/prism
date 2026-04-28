<script setup lang="ts">
import Button from 'primevue/button'
import { useAuthStore } from '@/stores/auth.store'

const auth = useAuthStore()
</script>

<template>
  <header class="prism-topbar flex items-center justify-between border-b px-6">
    <div>
      <p class="prism-topbar-kicker text-xs font-semibold uppercase">PRISM</p>
      <h2 class="prism-topbar-title text-lg font-semibold">Monitoring Pinjaman Luar Negeri</h2>
    </div>

    <div class="flex items-center gap-3">
      <Button label="Keluar" icon="pi pi-sign-out" outlined @click="auth.logout()" />
    </div>
  </header>
</template>
