<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import Button from 'primevue/button'
import Column from 'primevue/column'
import DataTable from 'primevue/datatable'
import Tag from 'primevue/tag'
import type { PaginationMeta } from '@/types/api.types'
import type {
  PipelineBottleneckItem,
  PipelineBottleneckSort,
  SortOrder,
} from '@/types/dashboard.types'

const props = defineProps<{
  items: PipelineBottleneckItem[]
  loading: boolean
  meta: PaginationMeta | null
  sort: PipelineBottleneckSort
  order: SortOrder
}>()

const emit = defineEmits<{
  page: [payload: { page: number; limit: number }]
  sort: [payload: { sort: PipelineBottleneckSort; order: SortOrder }]
}>()

const router = useRouter()
const hasJourneyRoute = computed(() => router.hasRoute('project-journey'))
const hasLoanAgreementRoute = computed(() => router.hasRoute('loan-agreement-detail'))
const first = computed(() => ((props.meta?.page ?? 1) - 1) * (props.meta?.limit ?? 20))
const sortOrderValue = computed(() => (props.order === 'asc' ? 1 : -1))

const sortableFields = new Set<PipelineBottleneckSort>([
  'stage',
  'project_name',
  'amount_usd',
  'age_days',
])

const usdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
})

const dateFormatter = new Intl.DateTimeFormat('id-ID', {
  day: '2-digit',
  month: 'short',
  year: 'numeric',
})

function isSortableField(value: string): value is PipelineBottleneckSort {
  return sortableFields.has(value as PipelineBottleneckSort)
}

function riskLabel(ageDays: number) {
  if (ageDays > 180) return 'High'
  if (ageDays >= 90) return 'Medium'
  return 'Low'
}

function riskTone(ageDays: number) {
  if (ageDays > 180) return 'danger'
  if (ageDays >= 90) return 'warn'
  return 'success'
}

function formatDate(value?: string) {
  if (!value) return '-'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '-'
  return dateFormatter.format(date)
}

function handlePage(event: { page?: number; rows?: number }) {
  emit('page', {
    page: (event.page ?? 0) + 1,
    limit: event.rows ?? props.meta?.limit ?? 20,
  })
}

function handleSort(event: { sortField?: unknown; sortOrder?: number | null }) {
  if (typeof event.sortField !== 'string' || !isSortableField(event.sortField)) return

  emit('sort', {
    sort: event.sortField,
    order: event.sortOrder === 1 ? 'asc' : 'desc',
  })
}
</script>

<template>
  <section class="rounded-lg border border-surface-200 bg-white p-4">
    <div class="mb-3 flex flex-col gap-1 md:flex-row md:items-end md:justify-between">
      <div>
        <h2 class="text-lg font-semibold text-surface-950">Worklist bottleneck</h2>
        <p class="text-sm text-surface-500">Daftar proyek yang perlu ditindaklanjuti per tahapan pipeline.</p>
      </div>
      <p class="text-sm text-surface-500">
        {{ meta?.total ?? 0 }} item
      </p>
    </div>

    <DataTable
      :value="items"
      :loading="loading"
      :first="first"
      :rows="meta?.limit ?? 20"
      :total-records="meta?.total ?? 0"
      :sort-field="sort"
      :sort-order="sortOrderValue"
      lazy
      paginator
      size="small"
      scrollable
      scroll-height="34rem"
      @page="handlePage"
      @sort="handleSort"
    >
      <Column field="project_name" header="Project" sortable>
        <template #body="{ data }">
          <div class="min-w-72">
            <p class="font-medium text-surface-950">{{ data.project_name }}</p>
            <p class="mt-1 text-xs text-surface-500">{{ data.code || '-' }}</p>
          </div>
        </template>
      </Column>

      <Column field="stage" header="Stage" sortable class="w-64">
        <template #body="{ data }">
          <div class="space-y-1">
            <Tag :value="data.stage_label" severity="secondary" />
            <p class="text-xs text-surface-500">{{ data.current_stage }}</p>
          </div>
        </template>
      </Column>

      <Column field="age_days" header="Age" sortable class="w-36">
        <template #body="{ data }">
          <div class="space-y-1">
            <Tag :value="riskLabel(data.age_days)" :severity="riskTone(data.age_days)" />
            <p class="text-xs text-surface-500">{{ data.age_days }} hari</p>
          </div>
        </template>
      </Column>

      <Column field="amount_usd" header="Amount" sortable class="w-44">
        <template #body="{ data }">
          {{ usdFormatter.format(data.amount_usd ?? 0) }}
        </template>
      </Column>

      <Column field="institution_name" header="K/L" class="w-64">
        <template #body="{ data }">
          {{ data.institution_name || '-' }}
        </template>
      </Column>

      <Column header="Lender" class="w-56">
        <template #body="{ data }">
          <div class="flex flex-wrap gap-1">
            <Tag
              v-for="lender in data.lender_names"
              :key="lender"
              :value="lender"
              severity="info"
            />
            <span v-if="!data.lender_names.length" class="text-sm text-surface-400">-</span>
          </div>
        </template>
      </Column>

      <Column field="recommended_action" header="Recommended action" class="w-80">
        <template #body="{ data }">
          <p class="text-sm text-surface-700">{{ data.recommended_action }}</p>
          <p class="mt-1 text-xs text-surface-500">Tanggal acuan: {{ formatDate(data.relevant_at) }}</p>
        </template>
      </Column>

      <Column header="Open" class="w-24">
        <template #body="{ data }">
          <RouterLink
            v-if="data.journey_bb_project_id && hasJourneyRoute"
            :to="{ name: 'project-journey', params: { bbProjectId: data.journey_bb_project_id } }"
          >
            <Button icon="pi pi-arrow-right" text rounded aria-label="Open journey" />
          </RouterLink>
          <RouterLink
            v-else-if="
              data.reference_type === 'loan_agreement' &&
              data.project_id &&
              hasLoanAgreementRoute
            "
            :to="{ name: 'loan-agreement-detail', params: { id: data.project_id } }"
          >
            <Button icon="pi pi-arrow-right" text rounded aria-label="Open detail" />
          </RouterLink>
          <Button v-else icon="pi pi-minus" text rounded disabled aria-label="No route" />
        </template>
      </Column>

      <template #empty>
        <div class="py-6 text-center text-sm text-surface-500">Tidak ada bottleneck untuk filter ini.</div>
      </template>
    </DataTable>
  </section>
</template>
