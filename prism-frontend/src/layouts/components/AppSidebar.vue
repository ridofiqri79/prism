<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth.store'
import { usePermission } from '@/composables/usePermission'

interface NavigationItem {
  label: string
  to: string
  icon: string
  module?: string
  adminOnly?: boolean
}

const route = useRoute()
const auth = useAuthStore()
const { can } = usePermission()

const masterItems = computed<NavigationItem[]>(() =>
  [
    { label: 'Countries', to: '/master/countries', icon: 'pi pi-globe', module: 'country' },
    { label: 'Lenders', to: '/master/lenders', icon: 'pi pi-building-columns', module: 'lender' },
    { label: 'Institutions', to: '/master/institutions', icon: 'pi pi-sitemap', module: 'institution' },
    { label: 'Regions', to: '/master/regions', icon: 'pi pi-map', module: 'region' },
    {
      label: 'Program Titles',
      to: '/master/program-titles',
      icon: 'pi pi-book',
      module: 'program_title',
    },
    {
      label: 'Bappenas Partners',
      to: '/master/bappenas-partners',
      icon: 'pi pi-users',
      module: 'bappenas_partner',
    },
    { label: 'Periods', to: '/master/periods', icon: 'pi pi-calendar', module: 'period' },
    {
      label: 'National Priorities',
      to: '/master/national-priorities',
      icon: 'pi pi-flag',
      module: 'national_priority',
    },
  ].filter((item) => can(item.module ?? '', 'read')),
)

const items = computed<NavigationItem[]>(() =>
  [
    { label: 'Dashboard', to: '/dashboard', icon: 'pi pi-home' },
    { label: 'Blue Books', to: '/blue-books', icon: 'pi pi-folder-open', module: 'bb_project' },
    { label: 'Green Books', to: '/green-books', icon: 'pi pi-folder', module: 'gb_project' },
    {
      label: 'Daftar Kegiatan',
      to: '/daftar-kegiatan',
      icon: 'pi pi-list',
      module: 'daftar_kegiatan',
    },
    {
      label: 'Loan Agreements',
      to: '/loan-agreements',
      icon: 'pi pi-file-edit',
      module: 'loan_agreement',
    },
    { label: 'Journey', to: '/journey', icon: 'pi pi-share-alt', module: 'bb_project' },
    { label: 'Users', to: '/users', icon: 'pi pi-user', adminOnly: true },
  ].filter((item) => {
    if (item.adminOnly) return auth.user?.role === 'ADMIN'
    if (!item.module) return true
    return can(item.module, 'read')
  }),
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
      <p class="mt-1 text-sm text-surface-500">Project Loan Integrated Monitoring System</p>
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

      <div v-if="masterItems.length > 0" class="pt-4">
        <p class="px-3 pb-2 text-xs font-semibold uppercase tracking-[0.16em] text-surface-400">
          Master Data
        </p>
        <RouterLink
          v-for="item in masterItems"
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
      </div>
    </nav>
  </aside>
</template>
