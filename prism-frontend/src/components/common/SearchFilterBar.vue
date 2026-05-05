<script setup lang="ts">
import { computed, ref } from 'vue'
import Button from 'primevue/button'
import InputText from 'primevue/inputtext'

export interface ActiveFilterPill {
  key: string
  label: string
  value?: string
  removable?: boolean
}

const props = withDefaults(
  defineProps<{
    search: string
    searchPlaceholder?: string
    activeFilters?: ActiveFilterPill[]
    filterCount?: number
    initiallyOpen?: boolean
    /** Hide the filter toggle button — use for search-only bars without a filter panel */
    hideFilterButton?: boolean
  }>(),
  {
    searchPlaceholder: 'Cari data',
    activeFilters: () => [],
    filterCount: undefined,
    initiallyOpen: false,
    hideFilterButton: false,
  },
)

const emit = defineEmits<{
  'update:search': [value: string]
  apply: []
  reset: []
  remove: [key: string]
}>()

const filterPanelOpen = ref(props.initiallyOpen)
const searchValue = computed({
  get: () => props.search,
  set: (value: string) => emit('update:search', value),
})
const visibleFilterCount = computed(() => props.filterCount ?? props.activeFilters.length)
const filterButtonLabel = computed(() =>
  visibleFilterCount.value > 0 ? `Filter (${visibleFilterCount.value})` : 'Filter',
)

function applyFilters() {
  emit('apply')
  filterPanelOpen.value = false
}
</script>

<template>
  <section class="rounded-lg border border-primary-100 bg-white p-4 shadow-sm shadow-surface-200/50">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
      <label class="relative min-w-0 flex-1">
        <span class="sr-only">{{ searchPlaceholder }}</span>
        <i class="pi pi-search pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-sm text-surface-400" />
        <InputText
          v-model="searchValue"
          :placeholder="searchPlaceholder"
          :aria-label="searchPlaceholder"
          class="w-full pl-10"
        />
      </label>

      <Button
        v-if="!hideFilterButton"
        type="button"
        :label="filterButtonLabel"
        :icon="filterPanelOpen ? 'pi pi-chevron-up' : 'pi pi-sliders-h'"
        severity="secondary"
        outlined
        class="h-12 shrink-0"
        :aria-expanded="filterPanelOpen"
        @click="filterPanelOpen = !filterPanelOpen"
      />
    </div>

    <div v-if="activeFilters.length > 0" class="mt-3 flex flex-wrap gap-2">
      <button
        v-for="filter in activeFilters"
        :key="filter.key"
        type="button"
        class="inline-flex max-w-full items-center gap-2 rounded-full border border-primary-100 bg-primary-50 px-3 py-1.5 text-xs font-semibold text-primary-800 transition-colors hover:border-prism-gold hover:bg-prism-gold/15"
        @click="filter.removable === false ? undefined : emit('remove', filter.key)"
      >
        <span class="shrink-0 text-primary-700">{{ filter.label }}</span>
        <span v-if="filter.value" class="min-w-0 truncate text-surface-700">{{ filter.value }}</span>
        <i v-if="filter.removable !== false" class="pi pi-times text-[0.65rem] text-surface-500" />
      </button>
    </div>

    <div v-if="filterPanelOpen" class="mt-4 border-t border-primary-100 pt-4">
      <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_auto] xl:items-end">
        <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-6">
          <slot name="filters" />
        </div>

        <div class="flex flex-wrap justify-end gap-2 xl:flex-nowrap">
          <Button type="button" label="Reset" icon="pi pi-filter-slash" severity="secondary" outlined @click="emit('reset')" />
          <Button type="button" label="Terapkan" icon="pi pi-check" @click="applyFilters" />
        </div>
      </div>
    </div>
  </section>
</template>
