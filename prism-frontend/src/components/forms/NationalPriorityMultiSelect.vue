<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import MultiSelect from 'primevue/multiselect'
import { useMasterStore } from '@/stores/master.store'
import type { NationalPriority } from '@/types/master.types'

type NationalPriorityOption = NationalPriority & {
  display_label: string
}

const props = withDefaults(
  defineProps<{
    modelValue: string[]
    periodId?: string
    extraOptions?: NationalPriority[]
    placeholder?: string
    disabled?: boolean
  }>(),
  {
    periodId: undefined,
    extraOptions: () => [],
    placeholder: 'Pilih prioritas nasional',
    disabled: false,
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
}>()

const masterStore = useMasterStore()
const cachedOptions = ref<NationalPriority[]>([])
let searchTimer: ReturnType<typeof window.setTimeout> | undefined

const selectedValues = computed({
  get: () => props.modelValue,
  set: (value: string[]) => emit('update:modelValue', value),
})

const mergedOptions = computed<NationalPriority[]>(() => {
  const byId = new Map<string, NationalPriority>()

  for (const priority of [
    ...cachedOptions.value,
    ...masterStore.nationalPriorities,
    ...props.extraOptions,
  ]) {
    byId.set(priority.id, priority)
  }

  return [...byId.values()]
})

const options = computed<NationalPriorityOption[]>(() =>
  mergedOptions.value
    .filter((priority) => !props.periodId || priority.period_id === props.periodId)
    .map((priority) => ({
      ...priority,
      display_label: priority.period?.name ? `${priority.title} (${priority.period.name})` : priority.title,
    })),
)

function lookupParams(search?: string) {
  return {
    limit: 50,
    search: search?.trim() || undefined,
    period_id: props.periodId,
    sort: 'title',
    order: 'asc' as const,
  }
}

function loadOptions(search?: string, force = true) {
  void masterStore.fetchNationalPriorities(force, lookupParams(search))
}

function cachePriorities(priorities: NationalPriority[]) {
  if (priorities.length === 0) {
    return
  }

  const byId = new Map(cachedOptions.value.map((priority) => [priority.id, priority]))

  for (const priority of priorities) {
    byId.set(priority.id, priority)
  }

  cachedOptions.value = [...byId.values()]
}

function scheduleSearch(search: string) {
  if (searchTimer) {
    window.clearTimeout(searchTimer)
  }

  searchTimer = window.setTimeout(() => loadOptions(search, true), 250)
}

function optionTitle(option: Record<string, unknown>) {
  return String(option.title ?? '')
}

function optionPeriodName(option: Record<string, unknown>) {
  const period = option.period

  if (period && typeof period === 'object' && 'name' in period) {
    return String((period as { name?: unknown }).name ?? '')
  }

  return ''
}

watch(
  () => masterStore.nationalPriorities,
  (priorities) => cachePriorities(priorities),
  { immediate: true },
)

watch(
  () => props.extraOptions,
  (priorities) => cachePriorities(priorities),
  { immediate: true },
)

onMounted(() => {
  void masterStore.fetchNationalPriorities(false, lookupParams())
})

onBeforeUnmount(() => {
  if (searchTimer) {
    window.clearTimeout(searchTimer)
  }
})
</script>

<template>
  <MultiSelect
    v-model="selectedValues"
    :options="options"
    option-label="display_label"
    option-value="id"
    :placeholder="placeholder"
    :disabled="disabled"
    filter
    filter-placeholder="Cari prioritas nasional"
    display="chip"
    scroll-height="18rem"
    class="w-full"
    @show="loadOptions(undefined, true)"
    @filter="scheduleSearch($event.value)"
  >
    <template #option="{ option }">
      <span class="block min-w-0">
        <span class="block truncate font-medium">{{ optionTitle(option) }}</span>
        <span v-if="optionPeriodName(option)" class="block truncate text-xs text-surface-500">
          {{ optionPeriodName(option) }}
        </span>
      </span>
    </template>
  </MultiSelect>
</template>
