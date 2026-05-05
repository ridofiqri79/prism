<script setup lang="ts">
import { computed, onMounted } from 'vue'
import InputNumber from 'primevue/inputnumber'
import InputText from 'primevue/inputtext'
import Message from 'primevue/message'
import Select from 'primevue/select'
import BottleneckWorklistTable from '@/components/dashboard/BottleneckWorklistTable.vue'
import DashboardFilterBar from '@/components/dashboard/DashboardFilterBar.vue'
import PipelineStageTabs from '@/components/dashboard/PipelineStageTabs.vue'
import { useListControls } from '@/composables/useListControls'
import PageHeader from '@/components/common/PageHeader.vue'
import { useDashboardStore } from '@/stores/dashboard.store'
import type {
  PipelineBottleneckParams,
  PipelineBottleneckSort,
  PipelineBottleneckStage,
  SortOrder,
} from '@/types/dashboard.types'

const dashboard = useDashboardStore()

const props = withDefaults(
  defineProps<{
    embedded?: boolean
  }>(),
  {
    embedded: false,
  },
)

const listControls = useListControls({
  initialFilters: {
    stage: null as PipelineBottleneckStage | null,
    period_id: null as string | null,
    publish_year: null as number | null,
    institution_id: null as string | null,
    lender_id: null as string | null,
    min_age_days: null as number | null,
  },
  initialLimit: 20,
  initialSort: 'age_days',
  initialOrder: 'desc',
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
const activeStage = computed(() => listControls.appliedFilters.stage)
const currentSort = computed(() => listControls.sort.value as PipelineBottleneckSort)
const currentOrder = computed(() => listControls.order.value)

function buildParams(): PipelineBottleneckParams {
  const params = listControls.buildParams()

  return {
    stage: params.stage as PipelineBottleneckStage | undefined,
    period_id: params.period_id as string | undefined,
    publish_year: params.publish_year as number | undefined,
    institution_id: params.institution_id as string | undefined,
    lender_id: params.lender_id as string | undefined,
    min_age_days: params.min_age_days as number | undefined,
    search: params.search as string | undefined,
    page: params.page as number,
    limit: params.limit as number,
    sort: (params.sort as PipelineBottleneckSort | undefined) ?? 'age_days',
    order: (params.order as SortOrder | undefined) ?? 'desc',
  }
}

async function loadBottleneck() {
  await dashboard.fetchPipelineBottleneck(buildParams())
}

async function loadInitialData() {
  await Promise.all([dashboard.fetchFilterOptions(), loadBottleneck()])
}

function setStage(stage: PipelineBottleneckStage | null) {
  listControls.draftFilters.stage = stage
  listControls.applyFilters()
  void loadBottleneck()
}

function applyFilters() {
  listControls.debouncedSearch.value = listControls.search.value.trim()
  listControls.applyFilters()
  void loadBottleneck()
}

function clearFilters() {
  listControls.resetFilters()
  listControls.setLimit(20)
  listControls.setSort({ sort: 'age_days', order: 'desc' })
  void loadBottleneck()
}

function handlePage(payload: { page: number; limit: number }) {
  listControls.page.value = payload.page
  listControls.limit.value = payload.limit
  void loadBottleneck()
}

function handleSort(payload: { sort: PipelineBottleneckSort; order: SortOrder }) {
  listControls.setSort(payload)
  void loadBottleneck()
}

onMounted(() => {
  void loadInitialData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      v-if="!props.embedded"
      title="Pipeline & Bottleneck"
      subtitle="Worklist proyek yang tertahan di Blue Book, Letter of Intent, Green Book, Daftar Kegiatan, dan Loan Agreement."
    />

    <PipelineStageTabs
      :stages="pipeline?.stage_summary ?? []"
      :active-stage="activeStage"
      @update:active-stage="setStage"
    />

    <DashboardFilterBar
      :loading="dashboard.pipelineLoading"
      grid-class="grid gap-4 md:grid-cols-2 xl:grid-cols-[1fr_10rem_1fr_1fr_10rem_10rem]"
      @apply="applyFilters"
      @reset="clearFilters"
    >
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Search</span>
          <InputText
            v-model="listControls.search.value"
            class="w-full"
            placeholder="Nama proyek atau kode"
            @keyup.enter="applyFilters"
          />
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Period</span>
          <Select
            v-model="listControls.draftFilters.period_id"
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
            v-model="listControls.draftFilters.institution_id"
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
            v-model="listControls.draftFilters.lender_id"
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
            v-model="listControls.draftFilters.publish_year"
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
            v-model="listControls.draftFilters.min_age_days"
            input-class="w-full"
            class="w-full"
            :min="0"
            :use-grouping="false"
            placeholder="Hari"
          />
        </label>
    </DashboardFilterBar>

    <Message v-if="dashboard.pipelineError" severity="error" icon="pi pi-exclamation-triangle">
      {{ dashboard.pipelineError }}
    </Message>

    <BottleneckWorklistTable
      :items="pipeline?.items ?? []"
      :loading="dashboard.pipelineLoading"
      :meta="meta"
      :sort="currentSort"
      :order="currentOrder"
      @page="handlePage"
      @sort="handleSort"
    />
  </section>
</template>
