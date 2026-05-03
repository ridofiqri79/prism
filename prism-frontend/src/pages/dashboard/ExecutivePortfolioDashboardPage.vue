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
import type { DashboardFilterParams, DashboardQuarter } from '@/types/dashboard.types'

const dashboard = useDashboardStore()

const filters = reactive<{
  period_id: string | null
  publish_year: number | null
  budget_year: number | null
  quarter: DashboardQuarter | null
}>({
  period_id: null,
  publish_year: null,
  budget_year: null,
  quarter: null,
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

const budgetYearOptions = computed(() =>
  (dashboard.filterOptions.budget_year ?? []).map((item) => ({
    label: item.label,
    value: Number(item.key),
  })),
)

const quarterOptions = computed(() => [
  { label: 'Semua Triwulan', value: null },
  ...(dashboard.filterOptions.quarter ?? []).map((item) => ({
    label: item.label,
    value: (item.key ?? item.label) as DashboardQuarter,
  })),
])

const portfolio = computed(() => dashboard.executivePortfolio)

function buildParams(): DashboardFilterParams {
  return {
    period_id: filters.period_id ?? undefined,
    publish_year: filters.publish_year ?? undefined,
    budget_year: filters.budget_year ?? undefined,
    quarter: filters.quarter ?? undefined,
  }
}

async function loadPortfolio() {
  await dashboard.fetchExecutivePortfolio(buildParams())
}

async function loadInitialData() {
  await Promise.all([dashboard.fetchFilterOptions(), loadPortfolio()])
}

function clearFilters() {
  filters.period_id = null
  filters.publish_year = null
  filters.budget_year = null
  filters.quarter = null
  void loadPortfolio()
}

onMounted(() => {
  void loadInitialData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      title="Executive Portfolio"
      subtitle="Kontrol portofolio nasional dari pipeline hingga realisasi monitoring."
    />

    <DashboardFilterBar
      :loading="dashboard.loading"
      grid-class="grid gap-4 md:grid-cols-2 xl:grid-cols-[1fr_11rem_11rem_10rem]"
      @apply="loadPortfolio"
      @reset="clearFilters"
    >
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
          <span class="text-sm font-medium text-surface-700">Budget Year</span>
          <Select
            v-model="filters.budget_year"
            :options="budgetYearOptions"
            option-label="label"
            option-value="value"
            show-clear
            class="w-full"
            placeholder="Semua"
          />
        </label>
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Quarter</span>
          <Select
            v-model="filters.quarter"
            :options="quarterOptions"
            option-label="label"
            option-value="value"
            class="w-full"
          />
        </label>
    </DashboardFilterBar>

    <Message v-if="dashboard.error" severity="error" icon="pi pi-exclamation-triangle">
      {{ dashboard.error }}
    </Message>

    <InsightCallout v-if="portfolio?.insights.length" :insights="portfolio.insights" />

    <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-6">
      <MetricCard v-for="card in portfolio?.cards ?? []" :key="card.key" :card="card" />
    </section>

    <StageFunnelChart :data="portfolio?.funnel ?? []" />

    <section class="grid gap-4 xl:grid-cols-2">
      <TopBreakdownTable title="Top 10 K/L" :items="portfolio?.top_institutions ?? []" />
      <TopBreakdownTable title="Top 10 Lenders" :items="portfolio?.top_lenders ?? []" />
    </section>

    <RiskItemTable :items="portfolio?.risk_items ?? []" />
  </section>
</template>
