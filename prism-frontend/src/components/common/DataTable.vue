<script setup lang="ts">
import { computed, type CSSProperties } from 'vue'
import PrimeColumn from 'primevue/column'
import PrimeDataTable from 'primevue/datatable'
import Skeleton from 'primevue/skeleton'
import EmptyState from '@/components/common/EmptyState.vue'
import ListPaginationFooter from '@/components/common/ListPaginationFooter.vue'
import TableReloadShell from '@/components/common/TableReloadShell.vue'
import { isCurrencyColumn, primeTablePt, tableCellClasses } from '@/utils/table-styles'

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

const tableSortOrder = computed(() => {
  if (!props.sortOrder) return undefined
  return props.sortOrder === 'asc' ? 1 : -1
})
const skeletonRows = computed(() => Array.from({ length: props.limit }, (_, index) => index))
const initialLoading = computed(() => props.loading && props.data.length === 0)
const refreshingRows = computed(() => props.loading && props.data.length > 0)
const tableMinWidth = computed(() => {
  const total = props.columns.reduce((sum, column) => sum + defaultColumnMinWidthRem(column), 0)

  return `${Math.max(total, 42)}rem`
})
const tableStyle = computed<CSSProperties>(() => ({
  minWidth: tableMinWidth.value,
  tableLayout: 'auto',
  width: '100%',
}))
const tablePt = primeTablePt

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
    minWidth: column.minWidth ?? `${defaultColumnMinWidthRem(column)}rem`,
    overflowWrap: column.nowrap ? 'normal' : 'anywhere',
    textAlign: resolvedColumnAlign(column),
    verticalAlign: 'top',
    whiteSpace: column.nowrap ? 'nowrap' : 'normal',
    width: column.width,
    wordBreak: column.nowrap ? 'normal' : 'break-word',
  }
}

function resolvedColumnAlign(column: ColumnDef) {
  if (column.field === 'actions') return column.align ?? 'right'

  return column.align ?? (isColumnCurrency(column) ? 'right' : 'left')
}

function isColumnCurrency(column: ColumnDef) {
  return isCurrencyColumn(column.field, column.header)
}

function columnClass(column: ColumnDef) {
  return tableCellClasses({
    align: resolvedColumnAlign(column),
    currency: isColumnCurrency(column),
    nowrap: column.nowrap,
  })
}

function defaultColumnMinWidthRem(column: ColumnDef) {
  return column.field === 'actions' ? 14 : 8
}

function columnHeaderStyle(column: ColumnDef): CSSProperties {
  return {
    ...columnCellStyle(column),
    overflowWrap: 'normal',
    whiteSpace: 'nowrap',
    wordBreak: 'normal',
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

    <EmptyState v-else-if="data.length === 0" compact />

    <TableReloadShell v-else :refreshing="refreshingRows">
      <div class="overflow-x-auto">
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
          :pt="tablePt"
          class="prism-data-table w-full rounded-lg border border-surface-200"
          @sort="handleSort"
        >
          <PrimeColumn
            v-for="column in columns"
            :key="column.field"
            :field="column.field"
            :header="column.header"
            :sortable="column.sortable"
            :header-class="columnClass(column)"
            :body-class="columnClass(column)"
            :style="columnCellStyle(column)"
            :header-style="columnHeaderStyle(column)"
            :body-style="columnCellStyle(column)"
          >
            <template #body="{ data: row }">
              <slot name="body-row" :row="row" :column="column">
                {{ row[column.field] }}
              </slot>
            </template>
          </PrimeColumn>
        </PrimeDataTable>
      </div>
    </TableReloadShell>

    <ListPaginationFooter
      :page="page"
      :limit="limit"
      :total="total"
      @update:page="emit('update:page', $event)"
      @update:limit="emit('update:limit', $event)"
    />
  </div>
</template>
