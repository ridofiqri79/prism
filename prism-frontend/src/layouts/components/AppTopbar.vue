<script setup lang="ts">
import { computed } from 'vue'
import Button from 'primevue/button'
import Tag from 'primevue/tag'
import { useAuthStore } from '@/stores/auth.store'

const auth = useAuthStore()

const roleSeverity = computed(() => (auth.user?.role === 'ADMIN' ? 'success' : 'info'))
</script>

<template>
  <header class="flex items-center justify-between border-b border-surface-200 bg-white px-6 py-4">
    <div>
      <p class="text-xs font-semibold uppercase tracking-[0.2em] text-primary">PRISM</p>
      <h2 class="text-lg font-semibold text-surface-950">Monitoring Pinjaman Luar Negeri</h2>
    </div>

    <div class="flex items-center gap-3">
      <div class="text-right">
        <p class="text-sm font-medium text-surface-900">
          {{ auth.user?.username ?? 'Tamu' }}
        </p>
        <p class="text-xs text-surface-500">
          {{ auth.user?.email ?? 'Belum ada sesi aktif' }}
        </p>
      </div>
      <Tag :severity="roleSeverity" :value="auth.user?.role ?? 'STAFF'" rounded />
      <Button label="Keluar" icon="pi pi-sign-out" outlined @click="auth.logout()" />
    </div>
  </header>
</template>
