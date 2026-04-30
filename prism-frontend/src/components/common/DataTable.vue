<script setup lang="ts">
import { computed, type CSSProperties } from 'vue'
import PrimeColumn from 'primevue/column'
import PrimeDataTable from 'primevue/datatable'
import Paginator from 'primevue/paginator'
import type {
  PaginatorPassThroughMethodOptions,
  PaginatorPassThroughOptions,
} from 'primevue/paginator'
import Skeleton from 'primevue/skeleton'
import EmptyState from '@/components/common/EmptyState.vue'
import TableReloadShell from '@/components/common/TableReloadShell.vue'

export interface ColumnDef {
  field: string
  header: string
  sortable?: boolean
  width?: string
  minWidth?: string
  maxWidth?: string
  align?: 'left' | 'center' | 'right'
  nowrap?: boolean
}

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
    data: Record<string, unknown>[]
    columns: ColumnDef[]
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
const pageStart = computed(() => (props.total > 0 ? first.value + 1 : 0))
const pageEnd = computed(() => Math.min(first.value + props.limit, props.total))
const tableSortOrder = computed(() => {
  if (!props.sortOrder) return undefined
  return props.sortOrder === 'asc' ? 1 : -1
})
const skeletonRows = computed(() => Array.from({ length: props.limit }, (_, index) => index))
const initialLoading = computed(() => props.loading && props.data.length === 0)
const refreshingRows = computed(() => props.loading && props.data.length > 0)
const tableStyle: CSSProperties = {
  tableLayout: 'fixed',
  width: '100%',
}
const paginatorTemplate = {
  '640px': 'PrevPageLink CurrentPageReport NextPageLink RowsPerPageDropdown',
  default: 'FirstPageLink PrevPageLink PageLinks NextPageLink LastPageLink RowsPerPageDropdown',
}
const paginatorNavButtonPt = ({ context }: PaginatorPassThroughMethodOptions<unknown>) => ({
  class: [
    'h-9 min-w-9 rounded-md border border-transparent text-surface-500 transition-colors',
    context.disabled
      ? 'cursor-not-allowed opacity-40'
      : 'hover:border-primary-100 hover:bg-primary-50 hover:text-primary',
  ],
})
const paginatorPt: PaginatorPassThroughOptions = {
  paginatorContainer: {
    class: 'w-full',
  },
  root: {
    class: 'border-0 bg-transparent p-0',
  },
  content: {
    class: 'flex flex-wrap items-center justify-center gap-1 sm:justify-end',
  },
  pages: {
    class: 'flex flex-wrap items-center justify-center gap-1',
  },
  first: paginatorNavButtonPt,
  prev: paginatorNavButtonPt,
  next: paginatorNavButtonPt,
  last: paginatorNavButtonPt,
  page: ({ context }: PaginatorPassThroughMethodOptions<unknown>) => ({
    class: [
      'h-9 min-w-9 rounded-md border text-sm font-semibold transition-colors',
      context.active
        ? 'border-primary bg-primary text-white shadow-sm'
        : 'border-transparent text-surface-600 hover:border-primary-100 hover:bg-primary-50 hover:text-primary',
    ],
  }),
  current: {
    class: 'px-2 text-sm font-semibold text-surface-600',
  },
  pcRowPerPageDropdown: {
    root: {
      class: 'ml-1 h-9 min-w-24 rounded-md border-surface-200',
    },
    label: {
      class: 'py-1.5 text-sm font-semibold text-surface-700',
    },
    dropdown: {
      class: 'w-8 text-surface-500',
    },
  },
}

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

function columnCellStyle(column: ColumnDef): CSSProperties {
  return {
    maxWidth: column.maxWidth,
    minWidth: column.minWidth,
    overflowWrap: column.nowrap ? 'normal' : 'anywhere',
    textAlign: column.align ?? 'left',
    verticalAlign: 'top',
    whiteSpace: column.nowrap ? 'nowrap' : 'normal',
    width: column.width,
    wordBreak: column.nowrap ? 'normal' : 'break-word',
  }
}
</script>

<template>
  <div class="space-y-4">
    <div v-if="initialLoading" class="overflow-hidden rounded-lg border border-surface-200 bg-white">
      <div
        v-for="row in skeletonRows"
        :key="row"
        class="grid gap-4 border-b border-surface-100 p-4 last:border-b-0"
        :style="{ gridTemplateColumns: `repeat(${Math.max(columns.length, 1)}, minmax(0, 1fr))` }"
      >
        <Skeleton v-for="column in columns" :key="column.field" height="1.5rem" />
      </div>
    </div>

    <EmptyState v-else-if="data.length === 0" />

    <TableReloadShell v-else :refreshing="refreshingRows">
      <PrimeDataTable
        :value="data"
        lazy
        striped-rows
        removable-sort
        resizable-columns
        column-resize-mode="fit"
        data-key="id"
        :sort-field="sortField"
        :sort-order="tableSortOrder"
        :table-style="tableStyle"
        class="w-full overflow-hidden rounded-lg border border-surface-200"
        @sort="handleSort"
      >
        <PrimeColumn
          v-for="column in columns"
          :key="column.field"
          :field="column.field"
          :header="column.header"
          :sortable="column.sortable"
          :style="columnCellStyle(column)"
          :header-style="columnCellStyle(column)"
          :body-style="columnCellStyle(column)"
        >
          <template #body="{ data: row }">
            <slot name="body-row" :row="row" :column="column">
              {{ row[column.field] }}
            </slot>
          </template>
        </PrimeColumn>
      </PrimeDataTable>
    </TableReloadShell>

    <div
      class="rounded-lg border border-surface-200 bg-white px-3 py-3 shadow-sm shadow-surface-200/50"
    >
      <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <p class="text-sm font-medium text-surface-600">
          Menampilkan
          <span class="font-semibold text-surface-900">{{ pageStart }}-{{ pageEnd }}</span>
          dari
          <span class="font-semibold text-surface-900">{{ total }}</span>
          data
        </p>
        <Paginator
          :first="first"
          :rows="limit"
          :total-records="total"
          :rows-per-page-options="[10, 20, 50, 100]"
          :template="paginatorTemplate"
          current-page-report-template="{currentPage} / {totalPages}"
          :pt="paginatorPt"
          @page="handlePage"
        />
      </div>
    </div>
  </div>
</template>
