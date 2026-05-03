<script setup lang="ts">
import { computed, onMounted, reactive } from 'vue'
import { useRouter } from 'vue-router'
import Message from 'primevue/message'
import Select from 'primevue/select'
import ClosingRiskTable from '@/components/dashboard/ClosingRiskTable.vue'
import ComponentBreakdownChart from '@/components/dashboard/ComponentBreakdownChart.vue'
import DashboardFilterBar from '@/components/dashboard/DashboardFilterBar.vue'
import DisbursementTrendChart from '@/components/dashboard/DisbursementTrendChart.vue'
import MetricCard from '@/components/dashboard/MetricCard.vue'
import UnderDisbursementTable from '@/components/dashboard/UnderDisbursementTable.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import { useDashboardStore } from '@/stores/dashboard.store'
import type {
  DashboardQuarter,
  LADisbursementParams,
  LARiskLevel,
  MetricCard as MetricCardType,
} from '@/types/dashboard.types'

const dashboard = useDashboardStore()
const router = useRouter()

const filters = reactive<{
  budget_year: number | null
  quarter: DashboardQuarter | null
  lender_id: string | null
  institution_id: string | null
  is_extended: boolean | null
  closing_months: 3 | 6 | 12 | null
  risk_level: LARiskLevel | null
}>({
  budget_year: null,
  quarter: null,
  lender_id: null,
  institution_id: null,
  is_extended: null,
  closing_months: null,
  risk_level: null,
})

const budgetYearOptions = computed(() =>
  (dashboard.filterOptions.budget_year ?? [])
    .map((item) => ({
      label: item.label,
      value: Number(item.key),
    }))
    .filter((item) => Number.isFinite(item.value)),
)

const quarterOptions = computed(() =>
  (dashboard.filterOptions.quarter ?? []).map((item) => ({
    label: item.label,
    value: item.key as DashboardQuarter,
  })),
)

const lenderOptions = computed(() =>
  (dashboard.filterOptions.lender ?? []).map((item) => ({
    label: item.label,
    value: item.key ?? item.id ?? '',
  })),
)

const institutionOptions = computed(() =>
  (dashboard.filterOptions.institution ?? []).map((item) => ({
    label: item.label,
    value: item.key ?? item.id ?? '',
  })),
)

const extensionOptions = computed(() => [
  { label: 'Semua LA', value: null },
  { label: 'Extended only', value: true },
  { label: 'Original closing', value: false },
])

const closingMonthOptions = computed(() => [
  { label: 'Semua closing', value: null },
  { label: '3 bulan', value: 3 },
  { label: '6 bulan', value: 6 },
  { label: '12 bulan', value: 12 },
])

const riskLevelOptions = computed(() => [
  { label: 'Semua risk', value: null },
  { label: 'Low', value: 'low' },
  { label: 'Medium', value: 'medium' },
  { label: 'High', value: 'high' },
])

const disbursement = computed(() => dashboard.laDisbursement)
const summary = computed(() => disbursement.value?.summary)

const cards = computed<MetricCardType[]>(() => [
  {
    key: 'la_count',
    label: 'Loan Agreement Count',
    value: summary.value?.la_count ?? 0,
    unit: 'project',
    category: 'legal',
  },
  {
    key: 'commitment_usd',
    label: 'Commitment',
    value: summary.value?.commitment_usd ?? 0,
    unit: 'USD',
    category: 'legal',
  },
  {
    key: 'realized_usd',
    label: 'Realized',
    value: summary.value?.realized_usd ?? 0,
    unit: 'USD',
    category: 'monitoring',
  },
  {
    key: 'absorption_pct',
    label: 'Absorption',
    value: summary.value?.absorption_pct ?? 0,
    unit: 'percent',
    category: 'monitoring',
  },
  {
    key: 'undisbursed_usd',
    label: 'Undisbursed',
    value: summary.value?.undisbursed_usd ?? 0,
    unit: 'USD',
    category: 'balance',
  },
  {
    key: 'extended_count',
    label: 'Extended',
    value: summary.value?.extended_count ?? 0,
    unit: 'project',
    category: 'risk',
  },
])

function buildParams(): LADisbursementParams {
  return {
    budget_year: filters.budget_year ?? undefined,
    quarter: filters.quarter ?? undefined,
    lender_id: filters.lender_id ?? undefined,
    institution_id: filters.institution_id ?? undefined,
    is_extended: filters.is_extended ?? undefined,
    closing_months: filters.closing_months ?? undefined,
    risk_level: filters.risk_level ?? undefined,
  }
}

async function loadDashboard() {
  await dashboard.fetchLADisbursement(buildParams())
}

async function loadInitialData() {
  await Promise.all([dashboard.fetchFilterOptions(), loadDashboard()])
}

function clearFilters() {
  filters.budget_year = null
  filters.quarter = null
  filters.lender_id = null
  filters.institution_id = null
  filters.is_extended = null
  filters.closing_months = null
  filters.risk_level = null
  void loadDashboard()
}

function openLoanAgreement(id: string) {
  void router.push({ name: 'loan-agreement-detail', params: { id } })
}

onMounted(() => {
  void loadInitialData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader
      title="Loan Agreement & Disbursement"
      subtitle="Monitoring komitmen legal, efektivitas, closing risk, undisbursed balance, dan serapan per triwulan."
    />

    <DashboardFilterBar
      :loading="dashboard.laDisbursementLoading"
      grid-class="grid gap-4 md:grid-cols-2 xl:grid-cols-[10rem_8rem_1fr_1fr_11rem_10rem_10rem]"
      @apply="loadDashboard"
      @reset="clearFilters"
    >
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
            show-clear
            class="w-full"
            placeholder="Semua"
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
          <span class="text-sm font-medium text-surface-700">Extension</span>
          <Select
            v-model="filters.is_extended"
            :options="extensionOptions"
            option-label="label"
            option-value="value"
            class="w-full"
          />
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Closing</span>
          <Select
            v-model="filters.closing_months"
            :options="closingMonthOptions"
            option-label="label"
            option-value="value"
            class="w-full"
          />
        </label>

        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Risk</span>
          <Select
            v-model="filters.risk_level"
            :options="riskLevelOptions"
            option-label="label"
            option-value="value"
            class="w-full"
          />
        </label>
    </DashboardFilterBar>

    <Message v-if="dashboard.laDisbursementError" severity="error" icon="pi pi-exclamation-triangle">
      {{ dashboard.laDisbursementError }}
    </Message>

    <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-6">
      <MetricCard v-for="card in cards" :key="card.key" :card="card" />
    </section>

    <DisbursementTrendChart :data="disbursement?.quarterly_trend ?? []" />

    <div class="grid gap-6 xl:grid-cols-2">
      <ClosingRiskTable
        :items="disbursement?.closing_risks ?? []"
        :loading="dashboard.laDisbursementLoading"
        @open="openLoanAgreement"
      />
      <UnderDisbursementTable
        :items="disbursement?.under_disbursement_risks ?? []"
        :loading="dashboard.laDisbursementLoading"
        @open="openLoanAgreement"
      />
    </div>

    <ComponentBreakdownChart :data="disbursement?.component_breakdown ?? []" />
  </section>
</template>
