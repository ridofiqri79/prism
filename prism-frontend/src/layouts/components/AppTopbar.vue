<script setup lang="ts">
import Button from 'primevue/button'
import { useAuthStore } from '@/stores/auth.store'

const auth = useAuthStore()

const emit = defineEmits<{
  'toggle-sidebar': []
}>()
</script>

<template>
  <header class="prism-topbar flex items-center justify-between border-b px-4 sm:px-6">
    <!-- Hamburger — visible only on mobile/tablet (below lg) -->
    <div class="flex items-center gap-3">
      <Button
        icon="pi pi-bars"
        text
        rounded
        aria-label="Toggle navigasi"
        class="lg:hidden"
        @click="emit('toggle-sidebar')"
      />
      <div class="hidden sm:block lg:block">
        <p class="prism-topbar-kicker text-xs font-semibold uppercase">PRISM</p>
        <h2 class="prism-topbar-title text-lg font-semibold">Monitoring Pinjaman Luar Negeri</h2>
      </div>
      <!-- Compact title on very small screens -->
      <p class="prism-topbar-title text-base font-semibold sm:hidden">PRISM</p>
    </div>

    <div class="flex items-center gap-3">
      <Button label="Keluar" icon="pi pi-sign-out" outlined @click="auth.logout()" />
    </div>
  </header>
</template>
