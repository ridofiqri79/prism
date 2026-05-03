<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue'
import Message from 'primevue/message'
import Select from 'primevue/select'
import CofinancingNetworkTable from '@/components/dashboard/CofinancingNetworkTable.vue'
import CurrencyExposureChart from '@/components/dashboard/CurrencyExposureChart.vue'
import DashboardFilterBar from '@/components/dashboard/DashboardFilterBar.vue'
import LenderCertaintyChart from '@/components/dashboard/LenderCertaintyChart.vue'
import LenderConversionTable from '@/components/dashboard/LenderConversionTable.vue'
import MetricCardComponent from '@/components/dashboard/MetricCard.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useDashboardStore } from '@/stores/dashboard.store'
import type { LenderFinancingMixParams, LenderType, MetricCard } from '@/types/dashboard.types'

const dashboard = useDashboardStore()

const filters = reactive<{
  lender_type: LenderType | null
  lender_id: string | null
  currency: string | null
  period_id: string | null
  publish_year: number | null
  budget_year: number | null
}>({
  lender_type: null,
  lender_id: null,
  currency: null,
  period_id: null,
  publish_year: null,
  budget_year: null,
})

const lenderTypeOptions = computed(() => [
  { label: 'Semua type', value: null },
  { label: 'Bilateral', value: 'Bilateral' },
  { label: 'Multilateral', value: 'Multilateral' },
  { label: 'KSA', value: 'KSA' },
])

const lenderOptions = computed(() =>
  (dashboard.filterOptions.lender ?? []).map((item) => ({
    label: item.label,
    value: item.key ?? item.id ?? '',
  })),
)

const currencyOptions = computed(() =>
  (dashboard.filterOptions.currency ?? []).map((item) => ({
    label: item.label,
    value: item.key ?? item.id ?? '',
  })),
)

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

const budgetYearOptions = computed(() =>
  (dashboard.filterOptions.budget_year ?? [])
    .map((item) => ({
      label: item.label,
      value: Number(item.key),
    }))
    .filter((item) => Number.isFinite(item.value)),
)

const mix = computed(() => dashboard.lenderFinancingMix)

const cards = computed<MetricCard[]>(() => {
  const summary = mix.value?.summary
  return [
    {
      key: 'bilateral_usd',
      label: 'Bilateral',
      value: summary?.bilateral_usd ?? 0,
      unit: 'USD',
      category: 'Loan Agreement',
    },
    {
      key: 'multilateral_usd',
      label: 'Multilateral',
      value: summary?.multilateral_usd ?? 0,
      unit: 'USD',
      category: 'Loan Agreement',
    },
    {
      key: 'ksa_usd',
      label: 'KSA',
      value: summary?.ksa_usd ?? 0,
      unit: 'USD',
      category: 'Loan Agreement',
    },
    {
      key: 'total_lenders',
      label: 'Active Lenders',
      value: summary?.total_lenders ?? 0,
      unit: 'lender',
      category: 'lender',
    },
    {
      key: 'cofinancing_projects',
      label: 'Cofinancing Projects',
      value: summary?.cofinancing_projects ?? 0,
      unit: 'project',
      category: 'project',
    },
  ]
})

function buildParams(): LenderFinancingMixParams {
  return {
    lender_type: filters.lender_type ?? undefined,
    lender_id: filters.lender_id ?? undefined,
    currency: filters.currency ?? undefined,
    period_id: filters.period_id ?? undefined,
    publish_year: filters.publish_year ?? undefined,
    budget_year: filters.budget_year ?? undefined,
  }
}

async function loadMix() {
  await dashboard.fetchLenderFinancingMix(buildParams())
}

async function loadInitialData() {
  await Promise.all([dashboard.fetchFilterOptions(), loadMix()])
}

function clearFilters() {
  filters.lender_type = null
  filters.lender_id = null
  filters.currency = null
  filters.period_id = null
  filters.publish_year = null
  filters.budget_year = null
  void loadMix()
}

onMounted(() => {
  void loadInitialData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      title="Lender & Financing Mix"
      subtitle="Profil lender, certainty ladder, conversion, cofinancing, dan currency exposure berdasarkan data tersimpan."
    />

    <DashboardFilterBar
      :loading="dashboard.lenderFinancingMixLoading"
      grid-class="grid gap-4 md:grid-cols-2 xl:grid-cols-[11rem_1fr_9rem_1fr_10rem_10rem]"
      @apply="loadMix"
      @reset="clearFilters"
    >
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Lender Type</span>
          <Select
            v-model="filters.lender_type"
            :options="lenderTypeOptions"
            option-label="label"
            option-value="value"
            class="w-full"
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
          <span class="text-sm font-medium text-surface-700">Currency</span>
          <Select
            v-model="filters.currency"
            :options="currencyOptions"
            option-label="label"
            option-value="value"
            show-clear
            filter
            class="w-full"
            placeholder="Semua"
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
    </DashboardFilterBar>

    <Message v-if="dashboard.lenderFinancingMixError" severity="error" icon="pi pi-exclamation-triangle">
      {{ dashboard.lenderFinancingMixError }}
    </Message>

    <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-5">
      <MetricCardComponent v-for="card in cards" :key="card.key" :card="card" />
    </section>

    <LenderCertaintyChart :data="mix?.certainty_ladder ?? []" />

    <section class="grid gap-4 xl:grid-cols-2">
      <CurrencyExposureChart :data="mix?.currency_exposure ?? []" />
      <CofinancingNetworkTable
        :items="mix?.cofinancing_items ?? []"
        :loading="dashboard.lenderFinancingMixLoading"
      />
    </section>

    <LenderConversionTable
      :items="mix?.lender_conversion ?? []"
      :loading="dashboard.lenderFinancingMixLoading"
    />
  </section>
</template>
