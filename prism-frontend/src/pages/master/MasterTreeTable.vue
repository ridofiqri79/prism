<script setup lang="ts">
import { computed } from 'vue'
import Paginator from 'primevue/paginator'
import Skeleton from 'primevue/skeleton'
import TreeTable from 'primevue/treetable'
import type { TreeNode } from 'primevue/treenode'
import EmptyState from '@/components/common/EmptyState.vue'
import TableReloadShell from '@/components/common/TableReloadShell.vue'

interface PageEvent {
  page: number
  rows: number
}

interface SortEvent {
  sortField?: unknown
  sortOrder?: unknown
}

type SortOrder = 'asc' | 'desc'

const props = withDefaults(
  defineProps<{
    value: TreeNode[]
    loading?: boolean
    total: number
    page: number
    limit: number
    sortField?: string
    sortOrder?: SortOrder
  }>(),
  {
    loading: false,
  },
)

const emit = defineEmits<{
  'update:page': [value: number]
  'update:limit': [value: number]
  sort: [value: { sort: string; order: SortOrder }]
}>()

const first = computed(() => (props.page - 1) * props.limit)
const tableSortOrder = computed(() => {
  if (!props.sortOrder) return undefined
  return props.sortOrder === 'asc' ? 1 : -1
})
const skeletonRows = computed(() => Array.from({ length: props.limit }, (_, index) => index))
const initialLoading = computed(() => props.loading && props.value.length === 0)
const refreshingRows = computed(() => props.loading && props.value.length > 0)

function handlePage(event: PageEvent) {
  if (event.rows !== props.limit) {
    emit('update:limit', event.rows)
    return
  }

  emit('update:page', event.page + 1)
}

function handleSort(event: SortEvent) {
  if (typeof event.sortField !== 'string' || event.sortOrder === 0) {
    return
  }

  emit('sort', {
    sort: event.sortField,
    order: event.sortOrder === 1 ? 'asc' : 'desc',
  })
}
</script>

<template>
  <div class="space-y-4">
    <div v-if="initialLoading" class="overflow-hidden rounded-lg border border-surface-200 bg-white">
      <div
        v-for="row in skeletonRows"
        :key="row"
        class="grid grid-cols-4 gap-4 border-b border-surface-100 p-4 last:border-b-0"
      >
        <Skeleton height="1.5rem" />
        <Skeleton height="1.5rem" />
        <Skeleton height="1.5rem" />
        <Skeleton height="1.5rem" />
      </div>
    </div>

    <EmptyState v-else-if="value.length === 0" />

    <TableReloadShell v-else :refreshing="refreshingRows">
      <TreeTable
        :value="value"
        sort-mode="single"
        removable-sort
        resizable-columns
        column-resize-mode="fit"
        :sort-field="sortField"
        :sort-order="tableSortOrder"
        class="overflow-hidden rounded-lg border border-surface-200"
        @sort="handleSort"
      >
        <slot />
      </TreeTable>
    </TableReloadShell>

    <Paginator
      :first="first"
      :rows="limit"
      :total-records="total"
      :rows-per-page-options="[10, 20, 50, 100]"
      @page="handlePage"
    />
  </div>
</template>
