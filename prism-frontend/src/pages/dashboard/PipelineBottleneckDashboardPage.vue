<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import Button from 'primevue/button'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Select from 'primevue/select'
import BottleneckWorklistTable from '@/components/dashboard/BottleneckWorklistTable.vue'
import PipelineStageTabs from '@/components/dashboard/PipelineStageTabs.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useDashboardStore } from '@/stores/dashboard.store'
import type {
  PipelineBottleneckParams,
  PipelineBottleneckSort,
  PipelineBottleneckStage,
  SortOrder,
} from '@/types/dashboard.types'

const dashboard = useDashboardStore()

const activeStage = ref<PipelineBottleneckStage | null>(null)
const filters = reactive<{
  period_id: string | null
  publish_year: number | null
  institution_id: string | null
  lender_id: string | null
  min_age_days: number | null
  search: string
}>({
  period_id: null,
  publish_year: null,
  institution_id: null,
  lender_id: null,
  min_age_days: null,
  search: '',
})

const tableState = reactive<{
  page: number
  limit: number
  sort: PipelineBottleneckSort
  order: SortOrder
}>({
  page: 1,
  limit: 20,
  sort: 'age_days',
  order: 'desc',
})

const periodOptions = computed(() =>
  (dashboard.filterOptions.period ?? []).map((item) => ({
    label: item.label,
    value: item.key ?? item.id ?? '',
  })),
)

const publishYearOptions = computed(() =>
  (dashboard.filterOptions.publish_year ?? [])
    .map((item) => ({
      label: item.label,
      value: Number(item.key),
    }))
    .filter((item) => Number.isFinite(item.value)),
)

const institutionOptions = computed(() =>
  (dashboard.filterOptions.institution ?? []).map((item) => ({
    label: item.label,
    value: item.key ?? item.id ?? '',
  })),
)

const lenderOptions = computed(() =>
  (dashboard.filterOptions.lender ?? []).map((item) => ({
    label: item.label,
    value: item.key ?? item.id ?? '',
  })),
)

const pipeline = computed(() => dashboard.pipelineBottleneck)
const meta = computed(() => dashboard.pipelineBottleneckMeta)

function buildParams(): PipelineBottleneckParams {
  const search = filters.search.trim()

  return {
    stage: activeStage.value ?? undefined,
    period_id: filters.period_id ?? undefined,
    publish_year: filters.publish_year ?? undefined,
    institution_id: filters.institution_id ?? undefined,
    lender_id: filters.lender_id ?? undefined,
    min_age_days: filters.min_age_days ?? undefined,
    search: search || undefined,
    page: tableState.page,
    limit: tableState.limit,
    sort: tableState.sort,
    order: tableState.order,
  }
}

async function loadBottleneck(resetPage = false) {
  if (resetPage) tableState.page = 1
  await dashboard.fetchPipelineBottleneck(buildParams())
}

async function loadInitialData() {
  await Promise.all([dashboard.fetchFilterOptions(), loadBottleneck()])
}

function setStage(stage: PipelineBottleneckStage | null) {
  activeStage.value = stage
  void loadBottleneck(true)
}

function applyFilters() {
  void loadBottleneck(true)
}

function clearFilters() {
  activeStage.value = null
  filters.period_id = null
  filters.publish_year = null
  filters.institution_id = null
  filters.lender_id = null
  filters.min_age_days = null
  filters.search = ''
  tableState.page = 1
  tableState.limit = 20
  tableState.sort = 'age_days'
  tableState.order = 'desc'
  void loadBottleneck()
}

function handlePage(payload: { page: number; limit: number }) {
  tableState.page = payload.page
  tableState.limit = payload.limit
  void loadBottleneck()
}

function handleSort(payload: { sort: PipelineBottleneckSort; order: SortOrder }) {
  tableState.sort = payload.sort
  tableState.order = payload.order
  tableState.page = 1
  void loadBottleneck()
}

onMounted(() => {
  void loadInitialData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      title="Pipeline & Bottleneck"
      subtitle="Worklist proyek yang tertahan di Blue Book, Letter of Intent, Green Book, Daftar Kegiatan, Loan Agreement, dan monitoring."
    />

    <PipelineStageTabs
      :stages="pipeline?.stage_summary ?? []"
      :active-stage="activeStage"
      @update:active-stage="setStage"
    />

    <section class="rounded-lg border border-surface-200 bg-white p-4">
      <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-[1fr_10rem_1fr_1fr_10rem_auto_auto]">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Search</span>
          <InputText
            v-model="filters.search"
            class="w-full"
            placeholder="Nama proyek atau kode"
            @keyup.enter="applyFilters"
          />
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Period</span>
          <Select
            v-model="filters.period_id"
            :options="periodOptions"
            option-label="label"
            option-value="value"
            show-clear
            filter
            class="w-full"
            placeholder="Semua"
          />
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Institution</span>
          <Select
            v-model="filters.institution_id"
            :options="institutionOptions"
            option-label="label"
            option-value="value"
            show-clear
            filter
            class="w-full"
            placeholder="Semua K/L"
          />
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Lender</span>
          <Select
            v-model="filters.lender_id"
            :options="lenderOptions"
            option-label="label"
            option-value="value"
            show-clear
            filter
            class="w-full"
            placeholder="Semua lender"
          />
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">GB Year</span>
          <Select
            v-model="filters.publish_year"
            :options="publishYearOptions"
            option-label="label"
            option-value="value"
            show-clear
            class="w-full"
            placeholder="Semua"
          />
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Min age</span>
          <InputNumber
            v-model="filters.min_age_days"
            input-class="w-full"
            class="w-full"
            :min="0"
            :use-grouping="false"
            placeholder="Hari"
          />
        </label>

        <div class="flex items-end gap-2">
          <Button
            icon="pi pi-filter"
            label="Terapkan"
            :loading="dashboard.pipelineLoading"
            @click="applyFilters"
          />
          <Button icon="pi pi-times" label="Reset" severity="secondary" outlined @click="clearFilters" />
        </div>
      </div>
    </section>

    <Message v-if="dashboard.pipelineError" severity="error" icon="pi pi-exclamation-triangle">
      {{ dashboard.pipelineError }}
    </Message>

    <BottleneckWorklistTable
      :items="pipeline?.items ?? []"
      :loading="dashboard.pipelineLoading"
      :meta="meta"
      :sort="tableState.sort"
      :order="tableState.order"
      @page="handlePage"
      @sort="handleSort"
    />
  </section>
</template>
