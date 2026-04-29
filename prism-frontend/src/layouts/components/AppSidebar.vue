<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
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

interface NavigationGroup {
  section: string
  label: string
  icon: string
  items: NavigationItem[]
}

const SIDEBAR_COLLAPSED_KEY = 'prism-sidebar-collapsed'

const route = useRoute()
const auth = useAuthStore()
const { can } = usePermission()
const isAdmin = computed(() => auth.user?.role === 'ADMIN')
const searchInput = ref<HTMLInputElement | null>(null)
const searchQuery = ref('')
const isCollapsed = ref(
  typeof window !== 'undefined' && window.localStorage.getItem(SIDEBAR_COLLAPSED_KEY) === 'true',
)
const expandedGroups = ref<Record<string, boolean>>({
  'Tahapan Perencanaan': true,
  'Master Data': true,
  Administrasi: true,
})

const accountLabel = computed(() => auth.user?.email || auth.user?.username || 'PRISM')
const accountMeta = computed(() => (auth.user ? `Akun ${auth.user.role}` : 'Belum ada sesi aktif'))
const hasSearch = computed(() => searchQuery.value.trim().length > 0)

watch(isCollapsed, (value) => {
  if (typeof window === 'undefined') return
  window.localStorage.setItem(SIDEBAR_COLLAPSED_KEY, String(value))
})

onMounted(() => {
  window.addEventListener('keydown', handleSearchShortcut)
})

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleSearchShortcut)
})

function canAccessItem(item: NavigationItem) {
  if (item.adminOnly) return isAdmin.value
  if (!item.module) return true
  return can(item.module, 'read')
}

function filterNavigation(items: NavigationItem[]) {
  return items.filter(canAccessItem)
}

const primaryItems = computed<NavigationItem[]>(() =>
  filterNavigation([
    { label: 'Dashboard', to: '/dashboard', icon: 'pi pi-home' },
    { label: 'Project', to: '/projects', icon: 'pi pi-table', module: 'bb_project' },
    { label: 'Perjalanan Proyek', to: '/journey', icon: 'pi pi-sitemap', module: 'bb_project' },
  ]),
)

const planningDocumentItems = computed<NavigationItem[]>(() =>
  filterNavigation([
    { label: 'Blue Book', to: '/blue-books', icon: 'pi pi-folder-open', module: 'blue_book' },
    { label: 'Green Book', to: '/green-books', icon: 'pi pi-folder', module: 'green_book' },
    {
      label: 'Daftar Kegiatan',
      to: '/daftar-kegiatan',
      icon: 'pi pi-list',
      module: 'daftar_kegiatan',
    },
    {
      label: 'Loan Agreement',
      to: '/loan-agreements',
      icon: 'pi pi-file-edit',
      module: 'loan_agreement',
    },
  ]),
)

const referenceItems = computed<NavigationItem[]>(() =>
  filterNavigation([
    { label: 'Negara', to: '/master/countries', icon: 'pi pi-globe', module: 'country' },
    { label: 'Lender', to: '/master/lenders', icon: 'pi pi-building-columns', module: 'lender' },
    { label: 'Currency', to: '/master/currencies', icon: 'pi pi-dollar', module: 'currency' },
    { label: 'Instansi', to: '/master/institutions', icon: 'pi pi-sitemap', module: 'institution' },
    { label: 'Wilayah', to: '/master/regions', icon: 'pi pi-map', module: 'region' },
    {
      label: 'Judul Program',
      to: '/master/program-titles',
      icon: 'pi pi-book',
      module: 'program_title',
    },
    {
      label: 'Mitra Bappenas',
      to: '/master/bappenas-partners',
      icon: 'pi pi-users',
      module: 'bappenas_partner',
    },
    { label: 'Periode', to: '/master/periods', icon: 'pi pi-calendar', module: 'period' },
    {
      label: 'Prioritas Nasional',
      to: '/master/national-priorities',
      icon: 'pi pi-flag',
      module: 'national_priority',
    },
  ]),
)

const adminAccessItems = computed<NavigationItem[]>(() =>
  filterNavigation([
    { label: 'Pengguna', to: '/users', icon: 'pi pi-user', adminOnly: true },
    {
      label: 'Impor Data',
      to: '/master/import-data',
      icon: 'pi pi-file-import',
      adminOnly: true,
    },
  ]),
)

const navigationGroups = computed<NavigationGroup[]>(() =>
  [
    {
      section: 'Perencanaan',
      label: 'Tahapan Perencanaan',
      icon: 'pi pi-compass',
      items: planningDocumentItems.value,
    },
    { section: 'Data', label: 'Master Data', icon: 'pi pi-database', items: referenceItems.value },
    {
      section: 'Sistem',
      label: 'Administrasi',
      icon: 'pi pi-shield',
      items: adminAccessItems.value,
    },
  ].filter((group) => group.items.length > 0),
)

const visiblePrimaryItems = computed(() => filterItemsBySearch(primaryItems.value))

const visibleNavigationGroups = computed<NavigationGroup[]>(() => {
  const query = normalizedSearchQuery()

  return navigationGroups.value
    .map((group) => {
      const groupMatches =
        group.label.toLowerCase().includes(query) || group.section.toLowerCase().includes(query)
      const items = groupMatches ? group.items : filterItemsBySearch(group.items)

      return { ...group, items }
    })
    .filter((group) => group.items.length > 0)
})

function isActive(path: string) {
  return route.path === path || route.path.startsWith(`${path}/`)
}

function toggleSidebar() {
  isCollapsed.value = !isCollapsed.value
}

function normalizedSearchQuery() {
  return searchQuery.value.trim().toLowerCase()
}

function filterItemsBySearch(items: NavigationItem[]) {
  const query = normalizedSearchQuery()

  if (!query) return items

  return items.filter((item) => item.label.toLowerCase().includes(query))
}

function isGroupExpanded(group: NavigationGroup) {
  if (hasSearch.value) return true
  if (group.items.some((item) => isActive(item.to))) return true
  return expandedGroups.value[group.label] ?? true
}

function toggleGroup(group: NavigationGroup) {
  expandedGroups.value = {
    ...expandedGroups.value,
    [group.label]: !isGroupExpanded(group),
  }
}

function handleSearchShortcut(event: KeyboardEvent) {
  if (!(event.ctrlKey || event.metaKey) || event.key.toLowerCase() !== 'k') return

  event.preventDefault()
  isCollapsed.value = false
  window.setTimeout(() => searchInput.value?.focus(), 0)
}
</script>

<template>
  <aside
    class="prism-sidebar hidden shrink-0 border-r transition-[width] duration-200 lg:flex lg:flex-col"
    :class="isCollapsed ? 'w-[68px]' : 'w-[260px]'"
  >
    <div
      class="prism-sidebar-header flex items-center border-b"
      :class="isCollapsed ? 'px-2' : 'px-4'"
    >
      <div
        class="flex w-full"
        :class="isCollapsed ? 'justify-center' : 'items-center gap-3'"
      >
        <div class="flex min-w-0 items-center gap-3" :class="isCollapsed ? 'justify-center' : ''">
          <img
            src="/prism-logo.png"
            alt=""
            class="prism-sidebar-logo shrink-0 rounded-md object-contain p-0.5"
            :class="isCollapsed ? 'h-5 w-5' : 'h-7 w-7'"
          />
          <div v-if="!isCollapsed" class="min-w-0">
            <p class="prism-sidebar-title truncate text-sm font-medium">{{ accountLabel }}</p>
            <p class="prism-sidebar-muted truncate text-xs">{{ accountMeta }}</p>
          </div>
        </div>
      </div>
    </div>

    <div v-if="!isCollapsed" class="px-4 py-3">
      <label class="sr-only" for="sidebar-search">Cari menu</label>
      <div class="relative">
        <i
          class="prism-sidebar-search-icon pi pi-search pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-xs"
        />
        <input
          id="sidebar-search"
          ref="searchInput"
          v-model="searchQuery"
          type="search"
          class="prism-sidebar-search-input h-9 w-full rounded-lg border px-9 pr-16 text-sm outline-none transition"
          placeholder="Quick search..."
        />
        <span
          class="prism-sidebar-kbd pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-[11px]"
        >
          Ctrl K
        </span>
      </div>
    </div>

    <nav class="flex-1 overflow-y-auto pb-4" :class="isCollapsed ? 'px-2 py-3' : 'px-3'">
      <div class="space-y-1">
        <RouterLink
          v-for="item in visiblePrimaryItems"
          :key="item.to"
          :to="item.to"
          class="prism-sidebar-item flex min-h-8 items-center rounded-lg text-sm transition-colors"
          :class="[
            isActive(item.to) ? 'is-active' : 'font-medium',
            isCollapsed ? 'justify-center px-0 py-2.5' : 'gap-3 px-3 py-1.5',
          ]"
          :aria-label="item.label"
          :title="isCollapsed ? item.label : undefined"
        >
          <i :class="[item.icon, 'shrink-0 text-[13px]']" />
          <span v-if="!isCollapsed" class="truncate">{{ item.label }}</span>
        </RouterLink>
      </div>

      <section
        v-for="group in visibleNavigationGroups"
        :key="group.label"
        :class="isCollapsed ? 'mt-3' : 'mt-5'"
      >
        <p v-if="!isCollapsed" class="prism-sidebar-section mb-2 px-3 text-xs font-medium">
          {{ group.section }}
        </p>

        <div class="space-y-1">
          <button
            type="button"
            class="prism-sidebar-item flex min-h-8 w-full items-center rounded-lg text-sm font-medium transition-colors"
            :class="[isCollapsed ? 'justify-center px-0 py-2.5' : 'gap-3 px-3 py-1.5 text-left']"
            :aria-expanded="isGroupExpanded(group)"
            :aria-label="group.label"
            :title="isCollapsed ? group.label : undefined"
            @click="toggleGroup(group)"
          >
            <i :class="[group.icon, 'shrink-0 text-[13px]']" />
            <span v-if="!isCollapsed" class="min-w-0 flex-1 truncate">{{ group.label }}</span>
            <i
              v-if="!isCollapsed"
              :class="[
                isGroupExpanded(group) ? 'pi pi-angle-down' : 'pi pi-angle-right',
                'shrink-0 text-[11px]',
              ]"
            />
          </button>

          <div
            v-if="!isCollapsed && isGroupExpanded(group)"
            class="prism-sidebar-child-list ml-[18px] border-l py-1 pl-2"
          >
            <RouterLink
              v-for="item in group.items"
              :key="item.to"
              :to="item.to"
              class="prism-sidebar-item flex min-h-8 items-center gap-2 rounded-lg px-2.5 py-1.5 text-sm transition-colors"
              :class="
                isActive(item.to) ? 'is-active' : 'font-medium'
              "
              :aria-label="item.label"
            >
              <i :class="[item.icon, 'shrink-0 text-[13px]']" />
              <span class="truncate">{{ item.label }}</span>
            </RouterLink>
          </div>
        </div>
      </section>
    </nav>

    <div class="prism-sidebar-footer border-t p-3" :class="isCollapsed ? 'flex justify-center' : ''">
      <button
        type="button"
        class="prism-sidebar-icon-button flex h-9 w-9 shrink-0 items-center justify-center rounded-lg transition-colors"
        :class="isCollapsed ? '' : 'ml-auto'"
        :aria-label="isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'"
        :title="isCollapsed ? 'Expand sidebar' : 'Collapse sidebar'"
        @click="toggleSidebar"
      >
        <i class="pi pi-bars text-sm" />
      </button>
    </div>
  </aside>
</template>
