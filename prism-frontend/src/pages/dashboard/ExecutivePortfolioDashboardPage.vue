<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue'
import Message from 'primevue/message'
import Select from 'primevue/select'
import DashboardFilterBar from '@/components/dashboard/DashboardFilterBar.vue'
import InsightCallout from '@/components/dashboard/InsightCallout.vue'
import MetricCard from '@/components/dashboard/MetricCard.vue'
import RiskItemTable from '@/components/dashboard/RiskItemTable.vue'
import StageFunnelChart from '@/components/dashboard/StageFunnelChart.vue'
import TopBreakdownTable from '@/components/dashboard/TopBreakdownTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useDashboardStore } from '@/stores/dashboard.store'
import type {
  DashboardFilterParams,
  MetricCard as DashboardMetricCard,
  StageMetric,
} from '@/types/dashboard.types'

const dashboard = useDashboardStore()

const props = withDefaults(
  defineProps<{
    embedded?: boolean
    autoload?: boolean
  }>(),
  {
    embedded: false,
    autoload: true,
  },
)

const filters = reactive<{
  period_id: string | null
  publish_year: number | null
}>({
  period_id: null,
  publish_year: null,
})

const periodOptions = computed(() =>
  (dashboard.filterOptions.period ?? []).map((item) => ({
    label: item.label,
    value: item.key ?? item.id ?? '',
  })),
)

const publishYearOptions = computed(() =>
  (dashboard.filterOptions.publish_year ?? []).map((item) => ({
    label: item.label,
    value: Number(item.key),
  })),
)

const portfolio = computed(() => dashboard.executivePortfolio)

const projectStages = [
  {
    stage: 'BB',
    label: 'Proyek Blue Book',
    hint: 'Jumlah proyek pada tahapan Blue Book',
  },
  {
    stage: 'GB',
    label: 'Proyek Green Book',
    hint: 'Jumlah proyek pada tahapan Green Book',
  },
  {
    stage: 'DK',
    label: 'Proyek Daftar Kegiatan',
    hint: 'Jumlah proyek yang sudah masuk Daftar Kegiatan',
  },
  {
    stage: 'LA',
    label: 'Proyek Loan Agreement',
    hint: 'Jumlah proyek yang sudah memiliki Loan Agreement',
  },
] as const

const financialCardLabels: Record<string, string> = {
  dk_financing_usd: 'Nilai Pembiayaan Daftar Kegiatan',
  la_commitment_usd: 'Komitmen Loan Agreement',
}

function findStageMetric(funnel: StageMetric[], stage: string) {
  return funnel.find((item) => item.stage === stage)
}

const projectMetricCards = computed<DashboardMetricCard[]>(() => {
  const funnel = portfolio.value?.funnel ?? []

  return projectStages.map((stage) => {
    const row = findStageMetric(funnel, stage.stage)

    return {
      key: `project_${stage.stage.toLowerCase()}`,
      label: stage.label,
      value: row?.project_count ?? 0,
      unit: 'project',
      category: 'project',
      hint: stage.hint,
    }
  })
})

const financialMetricCards = computed<DashboardMetricCard[]>(() =>
  (portfolio.value?.cards ?? [])
    .filter((card) => Object.keys(financialCardLabels).includes(card.key))
    .map((card) => ({
      ...card,
      label: financialCardLabels[card.key] ?? card.label,
      category: 'commitment',
    })),
)

function buildParams(): DashboardFilterParams {
  return {
    period_id: filters.period_id ?? undefined,
    publish_year: filters.publish_year ?? undefined,
  }
}

async function loadPortfolio() {
  await dashboard.fetchExecutivePortfolio(buildParams())
}

async function loadInitialData() {
  await Promise.all([
    dashboard.fetchFilterOptions(),
    loadPortfolio(),
    dashboard.fetchPipelineBottleneck({ page: 1, limit: 12, sort: 'age_days', order: 'desc' }),
  ])
}

function clearFilters() {
  filters.period_id = null
  filters.publish_year = null
  void loadPortfolio()
}

onMounted(() => {
  if (props.autoload) {
    void loadInitialData()
  }
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      v-if="!props.embedded"
      title="Ringkasan Eksekutif"
      subtitle="Kontrol portofolio nasional dari alur perencanaan hingga komitmen legal."
    />

    <DashboardFilterBar
      :loading="dashboard.loading"
      grid-class="grid gap-4 md:grid-cols-2 xl:grid-cols-[1fr_11rem]"
      @apply="loadPortfolio"
      @reset="clearFilters"
    >
      <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Periode Blue Book</span>
        <Select
          v-model="filters.period_id"
          :options="periodOptions"
          option-label="label"
          option-value="value"
          show-clear
          filter
          class="w-full"
          placeholder="Semua periode"
        />
      </label>
      <label class="block space-y-2">
        <span class="text-sm font-medium text-surface-700">Tahun Green Book</span>
        <Select
          v-model="filters.publish_year"
          :options="publishYearOptions"
          option-label="label"
          option-value="value"
          show-clear
          class="w-full"
          placeholder="Semua tahun"
        />
      </label>
    </DashboardFilterBar>

    <Message v-if="dashboard.error" severity="error" icon="pi pi-exclamation-triangle">
      {{ dashboard.error }}
    </Message>

    <InsightCallout v-if="portfolio?.insights.length" :insights="portfolio.insights" />

    <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <MetricCard v-for="card in projectMetricCards" :key="card.key" :card="card" />
    </section>

    <section class="grid gap-4 md:grid-cols-2">
      <MetricCard v-for="card in financialMetricCards" :key="card.key" :card="card" />
    </section>

    <StageFunnelChart
      :data="portfolio?.funnel ?? []"
      :top-institutions="portfolio?.top_institutions ?? []"
      :top-lenders="portfolio?.top_lenders ?? []"
      :pipeline-summary="dashboard.pipelineBottleneck?.stage_summary ?? []"
      :bottleneck-items="dashboard.pipelineBottleneck?.items ?? []"
      :risk-items="portfolio?.risk_items ?? []"
    />

    <section class="grid gap-4 xl:grid-cols-2">
      <TopBreakdownTable title="10 K/L Teratas" :items="portfolio?.top_institutions ?? []" />
      <TopBreakdownTable title="10 Lender Teratas" :items="portfolio?.top_lenders ?? []" />
    </section>

    <RiskItemTable :items="portfolio?.risk_items ?? []" />
  </section>
</template>
