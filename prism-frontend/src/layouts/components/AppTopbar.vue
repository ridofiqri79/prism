<script setup lang="ts">
import { computed } from 'vue'
import Button from 'primevue/button'
import { useAuthStore } from '@/stores/auth.store'

const auth = useAuthStore()

const roleLabel = computed(() => auth.user?.role ?? 'STAFF')
const roleClass = computed(() => (auth.user?.role === 'ADMIN' ? 'is-admin' : 'is-staff'))
</script>

<template>
  <header class="prism-topbar flex items-center justify-between border-b px-6 py-4">
    <div>
      <p class="prism-topbar-kicker text-xs font-semibold uppercase">PRISM</p>
      <h2 class="prism-topbar-title text-lg font-semibold">Monitoring Pinjaman Luar Negeri</h2>
    </div>

    <div class="flex items-center gap-3">
      <div class="text-right">
        <p class="prism-topbar-user text-sm font-medium">
          {{ auth.user?.username ?? 'Tamu' }}
        </p>
        <p class="prism-topbar-muted text-xs">
          {{ auth.user?.email ?? 'Belum ada sesi aktif' }}
        </p>
      </div>
      <span
        class="prism-role-pill rounded-full px-2.5 py-1 text-xs font-semibold"
        :class="roleClass"
      >
        {{ roleLabel }}
      </span>
      <Button label="Keluar" icon="pi pi-sign-out" outlined @click="auth.logout()" />
    </div>
  </header>
</template>
