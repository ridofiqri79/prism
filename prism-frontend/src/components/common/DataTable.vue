<script setup lang="ts">
import { computed } from 'vue'
import PrimeColumn from 'primevue/column'
import PrimeDataTable from 'primevue/datatable'
import Paginator from 'primevue/paginator'
import Skeleton from 'primevue/skeleton'
import EmptyState from '@/components/common/EmptyState.vue'

export interface ColumnDef {
  field: string
  header: string
  sortable?: boolean
}

interface PageEvent {
  page: number
  rows: number
}

interface SortEvent {
  sortField?: unknown
  sortOrder?: unknown
}

const props = withDefaults(
  defineProps<{
    data: Record<string, unknown>[]
    columns: ColumnDef[]
    loading?: boolean
    total: number
    page: number
    limit: number
  }>(),
  {
    loading: false,
  },
)

const emit = defineEmits<{
  'update:page': [value: number]
  'update:limit': [value: number]
  sort: [value: { sort: string; order: 'asc' | 'desc' }]
}>()

const first = computed(() => (props.page - 1) * props.limit)
const skeletonRows = computed(() => Array.from({ length: props.limit }, (_, index) => index))

function handlePage(event: PageEvent) {
  emit('update:page', event.page + 1)
  emit('update:limit', event.rows)
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
    <div v-if="loading" class="overflow-hidden rounded-lg border border-surface-200 bg-white">
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

    <PrimeDataTable
      v-else
      :value="data"
      lazy
      striped-rows
      data-key="id"
      class="overflow-hidden rounded-lg border border-surface-200"
      @sort="handleSort"
    >
      <PrimeColumn
        v-for="column in columns"
        :key="column.field"
        :field="column.field"
        :header="column.header"
        :sortable="column.sortable"
      >
        <template #body="{ data: row }">
          <slot name="body-row" :row="row" :column="column">
            {{ row[column.field] }}
          </slot>
        </template>
      </PrimeColumn>
    </PrimeDataTable>

    <Paginator
      :first="first"
      :rows="limit"
      :total-records="total"
      :rows-per-page-options="[10, 20, 50, 100]"
      @page="handlePage"
    />
  </div>
</template>
