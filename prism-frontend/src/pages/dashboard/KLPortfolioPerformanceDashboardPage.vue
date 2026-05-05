<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue'
import Message from 'primevue/message'
import Select from 'primevue/select'
import DashboardFilterBar from '@/components/dashboard/DashboardFilterBar.vue'
import KLPerformanceRadar from '@/components/dashboard/KLPerformanceRadar.vue'
import KLPerformanceTable from '@/components/dashboard/KLPerformanceTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useDashboardStore } from '@/stores/dashboard.store'
import type {
  InstitutionRole,
  KLPortfolioPerformanceParams,
  KLPortfolioSortBy,
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

const filters = reactive<{
  institution_id: string | null
  institution_role: InstitutionRole | null
  period_id: string | null
  publish_year: number | null
  sort_by: KLPortfolioSortBy
}>({
  institution_id: null,
  institution_role: null,
  period_id: null,
  publish_year: null,
  sort_by: 'pipeline_usd',
})

const usdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
})

const institutionOptions = computed(() =>
  (dashboard.filterOptions.institution ?? []).map((item) => ({
    label: item.label,
    value: item.key ?? item.id ?? '',
  })),
)

const institutionRoleOptions = computed(() => [
  { label: 'Semua role', value: null },
  { label: 'Executing Agency', value: 'Executing Agency' },
  { label: 'Implementing Agency', value: 'Implementing Agency' },
])

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

const sortOptions = computed(() => [
  { label: 'Pipeline USD', value: 'pipeline_usd' },
  { label: 'LA Commitment USD', value: 'la_commitment_usd' },
  { label: 'Risk Count', value: 'risk_count' },
])

const performance = computed(() => dashboard.klPortfolioPerformance)
const summary = computed(() => performance.value?.summary)

const kpiCards = computed(() => [
  {
    key: 'total',
    label: 'Total K/L',
    title: String(summary.value?.total_institutions ?? 0),
    detail: 'institution',
  },
  {
    key: 'top-exposure',
    label: 'Top Exposure',
    title: summary.value?.top_exposure_institution || '-',
    detail: usdFormatter.format(summary.value?.top_exposure_usd ?? 0),
  },
  {
    key: 'total-risk',
    label: 'Total Risk',
    title: String(summary.value?.total_institution_risk_count ?? 0),
    detail: 'risk item',
  },
  {
    key: 'highest-risk',
    label: 'Highest Risk',
    title: summary.value?.highest_risk_institution || '-',
    detail: `${summary.value?.highest_risk_count ?? 0} risk`,
  },
])

function buildParams(): KLPortfolioPerformanceParams {
  return {
    institution_id: filters.institution_id ?? undefined,
    institution_role: filters.institution_role ?? undefined,
    period_id: filters.period_id ?? undefined,
    publish_year: filters.publish_year ?? undefined,
    sort_by: filters.sort_by,
  }
}

async function loadPerformance() {
  await dashboard.fetchKLPortfolioPerformance(buildParams())
}

async function loadInitialData() {
  await Promise.all([dashboard.fetchFilterOptions(), loadPerformance()])
}

function clearFilters() {
  filters.institution_id = null
  filters.institution_role = null
  filters.period_id = null
  filters.publish_year = null
  filters.sort_by = 'pipeline_usd'
  void loadPerformance()
}

onMounted(() => {
  void loadInitialData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      v-if="!props.embedded"
      title="K/L Portfolio Performance"
      subtitle="Perbandingan portfolio dan kinerja K/L/Badan lintas Blue Book, Green Book, Daftar Kegiatan, dan Loan Agreement."
    />

    <DashboardFilterBar
      :loading="dashboard.klPortfolioPerformanceLoading"
      grid-class="grid gap-4 md:grid-cols-2 xl:grid-cols-[1fr_12rem_1fr_10rem_12rem]"
      @apply="loadPerformance"
      @reset="clearFilters"
    >
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
          <span class="text-sm font-medium text-surface-700">Role</span>
          <Select
            v-model="filters.institution_role"
            :options="institutionRoleOptions"
            option-label="label"
            option-value="value"
            class="w-full"
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
            placeholder="Semua period"
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
          <span class="text-sm font-medium text-surface-700">Sort</span>
          <Select
            v-model="filters.sort_by"
            :options="sortOptions"
            option-label="label"
            option-value="value"
            class="w-full"
          />
        </label>
    </DashboardFilterBar>

    <Message v-if="dashboard.klPortfolioPerformanceError" severity="error" icon="pi pi-exclamation-triangle">
      {{ dashboard.klPortfolioPerformanceError }}
    </Message>

    <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <article v-for="card in kpiCards" :key="card.key" class="rounded-lg border border-surface-200 bg-white p-4">
        <p class="text-sm font-medium text-surface-600">{{ card.label }}</p>
        <p class="mt-3 break-words text-xl font-semibold text-surface-950">{{ card.title }}</p>
        <p class="mt-1 text-sm text-surface-500">{{ card.detail }}</p>
      </article>
    </section>

    <KLPerformanceRadar :items="performance?.items ?? []" />

    <KLPerformanceTable
      :items="performance?.items ?? []"
      :loading="dashboard.klPortfolioPerformanceLoading"
    />
  </section>
</template>
