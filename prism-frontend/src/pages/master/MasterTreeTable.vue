<script setup lang="ts">
import { computed } from 'vue'
import Skeleton from 'primevue/skeleton'
import TreeTable from 'primevue/treetable'
import type { TreeNode } from 'primevue/treenode'
import EmptyState from '@/components/common/EmptyState.vue'
import ListPaginationFooter from '@/components/common/ListPaginationFooter.vue'
import TableReloadShell from '@/components/common/TableReloadShell.vue'

interface SortEvent {
  sortField?: unknown
  sortOrder?: unknown
}

type SortOrder = 'asc' | 'desc'
type ExpandedKeys = Record<string, boolean>

const props = withDefaults(
  defineProps<{
    value: TreeNode[]
    loading?: boolean
    total: number
    page: number
    limit: number
    sortField?: string
    sortOrder?: SortOrder
    expandedKeys?: ExpandedKeys
  }>(),
  {
    loading: false,
  },
)

const emit = defineEmits<{
  'update:page': [value: number]
  'update:limit': [value: number]
  'update:expandedKeys': [value: ExpandedKeys]
  'node-expand': [value: TreeNode]
  sort: [value: { sort: string; order: SortOrder }]
}>()

const tableSortOrder = computed(() => {
  if (!props.sortOrder) return undefined
  return props.sortOrder === 'asc' ? 1 : -1
})
const skeletonRows = computed(() => Array.from({ length: props.limit }, (_, index) => index))
const initialLoading = computed(() => props.loading && props.value.length === 0)
const refreshingRows = computed(() => props.loading && props.value.length > 0)
const tablePt = {
  thead: {
    class: 'bg-surface-50 text-left text-xs font-semibold uppercase tracking-wide text-surface-500',
  },
  headerCell: {
    class: 'px-4 py-3 text-xs font-semibold uppercase tracking-wide text-surface-500',
  },
  columnHeaderContent: {
    class: 'gap-2',
  },
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

function handleExpandedKeys(value: ExpandedKeys) {
  emit('update:expandedKeys', value)
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
      <div class="overflow-x-auto">
        <TreeTable
          :value="value"
          lazy
          :expanded-keys="expandedKeys"
          sort-mode="single"
          removable-sort
          resizable-columns
          column-resize-mode="fit"
          :sort-field="sortField"
          :sort-order="tableSortOrder"
          :pt="tablePt"
          class="min-w-[48rem] rounded-lg border border-surface-200"
          @update:expandedKeys="handleExpandedKeys"
          @node-expand="(node) => emit('node-expand', node)"
          @sort="handleSort"
        >
          <slot />
        </TreeTable>
      </div>
    </TableReloadShell>

    <ListPaginationFooter
      :page="page"
      :limit="limit"
      :total="total"
      @update:page="(value) => emit('update:page', value)"
      @update:limit="(value) => { emit('update:limit', value); emit('update:page', 1) }"
    />
  </div>
</template>
