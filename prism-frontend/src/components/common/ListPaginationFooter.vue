<script setup lang="ts">
import { computed } from 'vue'
import Paginator from 'primevue/paginator'
import type {
  PaginatorPassThroughMethodOptions,
  PaginatorPassThroughOptions,
} from 'primevue/paginator'

interface PageEvent {
  page: number
  rows: number
}

const props = withDefaults(
  defineProps<{
    page: number
    limit: number
    total: number
    rowsPerPageOptions?: number[]
  }>(),
  {
    rowsPerPageOptions: () => [10, 20, 50, 100],
  },
)

const emit = defineEmits<{
  'update:page': [value: number]
  'update:limit': [value: number]
}>()

const first = computed(() => (props.page - 1) * props.limit)
const pageStart = computed(() => (props.total > 0 ? first.value + 1 : 0))
const pageEnd = computed(() => Math.min(first.value + props.limit, props.total))
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
    emit('update:page', 1)
    return
  }

  emit('update:page', event.page + 1)
}
</script>

<template>
  <div class="rounded-lg border border-surface-200 bg-white px-3 py-3 shadow-sm shadow-surface-200/50">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <p class="text-sm font-medium text-surface-600 sm:whitespace-nowrap">
        <template v-if="total > 0">
          Menampilkan
          <span class="font-semibold text-surface-900">{{ pageStart }}-{{ pageEnd }}</span>
          dari
          <span class="font-semibold text-surface-900">{{ total }}</span>
          data
        </template>
        <template v-else>
          Menampilkan <span class="font-semibold text-surface-900">0</span> dari
          <span class="font-semibold text-surface-900">0</span> data
        </template>
      </p>
      <Paginator
        v-if="total > 0"
        :first="first"
        :rows="limit"
        :total-records="total"
        :rows-per-page-options="rowsPerPageOptions"
        :template="paginatorTemplate"
        current-page-report-template="{currentPage} / {totalPages}"
        :pt="paginatorPt"
        @page="handlePage"
      />
      <p v-else class="text-sm font-semibold text-surface-500">Tidak ada halaman</p>
    </div>
  </div>
</template>
