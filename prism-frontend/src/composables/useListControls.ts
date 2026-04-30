import { computed, onUnmounted, reactive, ref, watch } from 'vue'
import type { ActiveFilterPill } from '@/components/common/SearchFilterBar.vue'

export type ListSortOrder = 'asc' | 'desc'
export type ListParamValue = string | string[] | number | number[] | boolean | undefined

interface UseListControlsOptions<TFilters extends object> {
  initialFilters: TFilters
  initialPage?: number
  initialLimit?: number
  initialSort?: string
  initialOrder?: ListSortOrder
  searchDebounceMs?: number
  filterLabels?: Partial<Record<keyof TFilters & string, string>>
  formatFilterValue?: (key: keyof TFilters & string, value: TFilters[keyof TFilters]) => string
}

function cloneFilters<TFilters extends object>(value: TFilters): TFilters {
  return JSON.parse(JSON.stringify(value)) as TFilters
}

function resetObject<TFilters extends object>(target: TFilters, source: TFilters) {
  const mutableTarget = target as Record<string, unknown>
  for (const key of Object.keys(mutableTarget)) {
    delete mutableTarget[key]
  }

  Object.assign(mutableTarget, cloneFilters(source))
}

function isActiveFilterValue(value: unknown) {
  if (Array.isArray(value)) return value.length > 0
  if (typeof value === 'string') return value.trim().length > 0
  if (value === null || value === undefined) return false
  return true
}

function defaultFilterSummary(value: unknown) {
  if (Array.isArray(value)) {
    if (value.length <= 2) return value.join(', ')
    return `${value.slice(0, 2).join(', ')} +${value.length - 2}`
  }
  if (typeof value === 'boolean') return value ? 'Ya' : 'Tidak'
  return String(value)
}

export function useListControls<TFilters extends object>(
  options: UseListControlsOptions<TFilters>,
) {
  const initialFilters = cloneFilters(options.initialFilters)
  const page = ref(options.initialPage ?? 1)
  const limit = ref(options.initialLimit ?? 20)
  const sort = ref(options.initialSort ?? '')
  const order = ref<ListSortOrder>(options.initialOrder ?? 'asc')
  const search = ref('')
  const debouncedSearch = ref('')
  const draftFilters = reactive(cloneFilters(initialFilters)) as TFilters
  const appliedFilters = reactive(cloneFilters(initialFilters)) as TFilters
  let searchTimer: ReturnType<typeof window.setTimeout> | undefined

  const activeFilterPills = computed<ActiveFilterPill[]>(() =>
    Object.entries(appliedFilters)
      .filter(([, value]) => isActiveFilterValue(value))
      .map(([key, value]) => ({
        key,
        label: (options.filterLabels as Record<string, string> | undefined)?.[key] ?? key,
        value: options.formatFilterValue
          ? options.formatFilterValue(key as keyof TFilters & string, value as TFilters[keyof TFilters])
          : defaultFilterSummary(value),
      })),
  )

  const activeFilterCount = computed(() => activeFilterPills.value.length)

  function resetPage() {
    page.value = 1
  }

  function setLimit(value: number) {
    limit.value = value
    resetPage()
  }

  function setSort(value: { sort: string; order: ListSortOrder }) {
    sort.value = value.sort
    order.value = value.order
    resetPage()
  }

  function applyFilters() {
    resetObject(appliedFilters, draftFilters)
    resetPage()
  }

  function resetFilters() {
    search.value = ''
    debouncedSearch.value = ''
    resetObject(draftFilters, initialFilters)
    resetObject(appliedFilters, initialFilters)
    resetPage()
  }

  function removeFilter(key: string) {
    const emptyFilters = cloneFilters(initialFilters) as Record<string, unknown>
    const emptyValue = emptyFilters[key]

    ;(draftFilters as Record<string, unknown>)[key] = emptyValue
    ;(appliedFilters as Record<string, unknown>)[key] = emptyValue
    resetPage()
  }

  function buildParams(extra?: Record<string, ListParamValue>) {
    const params: Record<string, ListParamValue> = {
      page: page.value,
      limit: limit.value,
      ...extra,
    }

    if (sort.value) params.sort = sort.value
    if (sort.value) params.order = order.value
    if (debouncedSearch.value.trim()) params.search = debouncedSearch.value.trim()

    for (const [key, value] of Object.entries(appliedFilters)) {
      if (isActiveFilterValue(value)) {
        params[key] = value as ListParamValue
      }
    }

    return params
  }

  watch(search, (value) => {
    if (searchTimer) {
      window.clearTimeout(searchTimer)
    }

    searchTimer = window.setTimeout(() => {
      page.value = 1
      debouncedSearch.value = value.trim()
    }, options.searchDebounceMs ?? 300)
  })

  function dispose() {
    if (searchTimer) {
      window.clearTimeout(searchTimer)
      searchTimer = undefined
    }
  }

  onUnmounted(dispose)

  return {
    page,
    limit,
    sort,
    order,
    search,
    debouncedSearch,
    draftFilters,
    appliedFilters,
    activeFilterPills,
    activeFilterCount,
    resetPage,
    setLimit,
    setSort,
    applyFilters,
    resetFilters,
    removeFilter,
    buildParams,
    dispose,
  }
}
