import { computed, ref } from 'vue'

export type SortOrder = 'asc' | 'desc'

export function usePagination() {
  const page = ref(1)
  const limit = ref(20)
  const sort = ref('created_at')
  const order = ref<SortOrder>('desc')

  const queryParams = computed(() => ({
    page: page.value,
    limit: limit.value,
    sort: sort.value,
    order: order.value,
  }))

  function setPage(n: number) {
    page.value = Math.max(1, n)
  }

  function nextPage() {
    page.value += 1
  }

  function prevPage() {
    setPage(page.value - 1)
  }

  function resetPage() {
    page.value = 1
  }

  return {
    page,
    limit,
    sort,
    order,
    queryParams,
    setPage,
    nextPage,
    prevPage,
    resetPage,
  }
}
