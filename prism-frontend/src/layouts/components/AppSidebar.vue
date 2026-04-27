<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth.store'

interface NavigationItem {
  label: string
  to: string
  icon: string
  adminOnly?: boolean
}

const route = useRoute()
const auth = useAuthStore()

const items = computed<NavigationItem[]>(() =>
  [
    { label: 'Dashboard', to: '/dashboard', icon: 'pi pi-home' },
    { label: 'Countries', to: '/master/countries', icon: 'pi pi-globe' },
    { label: 'Lenders', to: '/master/lenders', icon: 'pi pi-building-columns' },
    { label: 'Institutions', to: '/master/institutions', icon: 'pi pi-sitemap' },
    { label: 'Regions', to: '/master/regions', icon: 'pi pi-map' },
    { label: 'Program Titles', to: '/master/program-titles', icon: 'pi pi-book' },
    { label: 'Bappenas Partners', to: '/master/bappenas-partners', icon: 'pi pi-users' },
    { label: 'Periods', to: '/master/periods', icon: 'pi pi-calendar' },
    { label: 'National Priorities', to: '/master/national-priorities', icon: 'pi pi-flag' },
    { label: 'Blue Books', to: '/blue-books', icon: 'pi pi-folder-open' },
    { label: 'Green Books', to: '/green-books', icon: 'pi pi-folder' },
    { label: 'Daftar Kegiatan', to: '/daftar-kegiatan', icon: 'pi pi-list' },
    { label: 'Loan Agreements', to: '/loan-agreements', icon: 'pi pi-file-edit' },
    { label: 'Journey', to: '/journey/demo-project', icon: 'pi pi-share-alt' },
    { label: 'Users', to: '/users', icon: 'pi pi-user', adminOnly: true },
  ].filter((item) => !item.adminOnly || auth.user?.role === 'ADMIN'),
)

function isActive(path: string) {
  return route.path === path || route.path.startsWith(`${path}/`)
}
</script>

<template>
  <aside class="hidden w-72 shrink-0 border-r border-surface-200 bg-white lg:block">
    <div class="border-b border-surface-200 px-6 py-5">
      <p class="text-xs font-semibold uppercase tracking-[0.24em] text-primary">PRISM</p>
      <h1 class="mt-2 text-xl font-semibold text-surface-950">Frontend Setup</h1>
      <p class="mt-1 text-sm text-surface-500">Plan FE-00 foundation scaffold</p>
    </div>

    <nav class="space-y-1 px-3 py-4">
      <RouterLink
        v-for="item in items"
        :key="item.to"
        :to="item.to"
        class="flex items-center gap-3 rounded-2xl px-3 py-2.5 text-sm font-medium transition-colors"
        :class="
          isActive(item.to)
            ? 'bg-primary text-primary-contrast'
            : 'text-surface-700 hover:bg-surface-100 hover:text-surface-950'
        "
      >
        <i :class="[item.icon, 'text-sm']" />
        <span>{{ item.label }}</span>
      </RouterLink>
    </nav>
  </aside>
</template>
