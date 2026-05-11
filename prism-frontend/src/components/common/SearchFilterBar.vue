<script setup lang="ts">
import { computed, ref, useSlots } from 'vue'
import Button from 'primevue/button'
import IconField from 'primevue/iconfield'
import InputIcon from 'primevue/inputicon'
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

const slots = useSlots()
const filterPanelOpen = ref(props.initiallyOpen)
const searchValue = computed({
  get: () => props.search,
  set: (value: string) => emit('update:search', value),
})
const visibleFilterCount = computed(() => props.filterCount ?? props.activeFilters.length)
const filterButtonBadge = computed(() =>
  visibleFilterCount.value > 0 ? String(visibleFilterCount.value) : undefined,
)
const hasActiveFilters = computed(() => props.activeFilters.length > 0)
const showFilterButton = computed(() => !props.hideFilterButton && Boolean(slots.filters))

function applyFilters() {
  emit('apply')
  filterPanelOpen.value = false
}
</script>

<template>
  <section
    class="overflow-hidden rounded-lg border border-primary-100/80 bg-white shadow-sm shadow-surface-200/60"
  >
    <div class="p-4">
      <div class="flex flex-col gap-3 lg:flex-row lg:items-center">
        <label class="min-w-0 flex-1">
          <span class="sr-only">{{ searchPlaceholder }}</span>
          <IconField icon-position="left">
            <InputIcon class="pi pi-search text-sm text-surface-400" />
            <InputText
              v-model="searchValue"
              :placeholder="searchPlaceholder"
              :aria-label="searchPlaceholder"
              class="h-11 w-full"
            />
          </IconField>
        </label>

        <div class="flex shrink-0 flex-wrap items-center gap-2 sm:flex-nowrap">
          <Button
            v-if="showFilterButton"
            type="button"
            label="Filter"
            :icon="filterPanelOpen ? 'pi pi-chevron-up' : 'pi pi-filter'"
            :badge="filterButtonBadge"
            badge-severity="secondary"
            severity="secondary"
            outlined
            class="h-11 w-full shrink-0 sm:w-auto"
            :aria-expanded="filterPanelOpen"
            @click="filterPanelOpen = !filterPanelOpen"
          />
          <template v-if="$slots.actions">
            <span class="hidden h-7 w-px shrink-0 bg-surface-200 sm:block" aria-hidden="true" />
            <slot name="actions" />
          </template>
        </div>
      </div>

      <div v-if="hasActiveFilters" class="mt-3 flex flex-col gap-2 sm:flex-row sm:items-start">
        <span class="shrink-0 pt-1 text-xs font-semibold uppercase tracking-wide text-surface-400">
          Filter aktif
        </span>
        <div class="flex min-w-0 flex-wrap gap-2">
          <button
            v-for="filter in activeFilters"
            :key="filter.key"
            type="button"
            class="inline-flex max-w-full items-center gap-2 rounded-full border border-primary-100 bg-primary-50 px-3 py-1.5 text-xs font-semibold text-primary-800 transition-colors hover:border-prism-gold hover:bg-prism-gold/15"
            @click="filter.removable === false ? undefined : emit('remove', filter.key)"
          >
            <span class="shrink-0 text-primary-700">{{ filter.label }}</span>
            <span v-if="filter.value" class="min-w-0 truncate text-surface-700">{{
              filter.value
            }}</span>
            <i
              v-if="filter.removable !== false"
              class="pi pi-times text-[0.65rem] text-surface-500"
            />
          </button>
        </div>
      </div>
    </div>

    <Transition
      enter-active-class="transition duration-150 ease-out"
      enter-from-class="-translate-y-1 opacity-0"
      enter-to-class="translate-y-0 opacity-100"
      leave-active-class="transition duration-100 ease-in"
      leave-from-class="translate-y-0 opacity-100"
      leave-to-class="-translate-y-1 opacity-0"
    >
      <div v-if="filterPanelOpen && $slots.filters" class="border-t border-primary-100 bg-surface-0">
        <div class="space-y-4 p-4">
          <div class="flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
            <p class="text-xs font-semibold uppercase tracking-wide text-primary-700">
              Filter lanjutan
            </p>
            <span v-if="visibleFilterCount > 0" class="text-xs font-medium text-surface-500">
              {{ visibleFilterCount }} filter diterapkan
            </span>
          </div>
          <div class="grid min-w-0 gap-3 md:grid-cols-2 xl:grid-cols-6">
            <slot name="filters" />
          </div>
        </div>

        <div
          class="flex flex-col gap-2 border-t border-surface-100 bg-surface-50/60 px-4 py-3 sm:flex-row sm:justify-end"
        >
          <Button
            type="button"
            label="Reset"
            icon="pi pi-filter-slash"
            severity="secondary"
            outlined
            @click="emit('reset')"
          />
          <Button type="button" label="Terapkan" icon="pi pi-check" @click="applyFilters" />
        </div>
      </div>
    </Transition>
  </section>
</template>
