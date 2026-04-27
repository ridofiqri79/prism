<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import Button from 'primevue/button'
import Select from 'primevue/select'
import Tag from 'primevue/tag'
import AbsorptionBar from '@/components/monitoring/AbsorptionBar.vue'
import MonitoringChart from '@/components/monitoring/MonitoringChart.vue'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import PageHeader from '@/components/common/PageHeader.vue'
import SummaryCard from '@/components/common/SummaryCard.vue'
import LenderSelect from '@/components/forms/LenderSelect.vue'
import { DashboardService } from '@/services/dashboard.service'
import type {
  DashboardFilterParams,
  DashboardQuarter,
  DashboardSummary,
  MonitoringSummary,
} from '@/types/dashboard.types'
import type { MonitoringDisbursement } from '@/types/monitoring.types'

const currentYear = new Date().getFullYear()
const summary = ref<DashboardSummary | null>(null)
const monitoringSummary = ref<MonitoringSummary | null>(null)
const loading = ref(false)
const filters = reactive<{
  budget_year: number | null
  quarter: DashboardQuarter | null
  lender_id: string | null
}>({
  budget_year: currentYear,
  quarter: null,
  lender_id: null,
})
const yearOptions = Array.from({ length: 9 }, (_, index) => {
  const year = currentYear - 4 + index
  return { label: String(year), value: year }
})
const quarterOptions: Array<{ label: string; value: DashboardQuarter | null }> = [
  { label: 'Semua Triwulan', value: null },
  { label: 'TW1 (Apr-Jun)', value: 'TW1' },
  { label: 'TW2 (Jul-Sep)', value: 'TW2' },
  { label: 'TW3 (Okt-Des)', value: 'TW3' },
  { label: 'TW4 (Jan-Mar)', value: 'TW4' },
]

const cards = computed(() => {
  const data = summary.value

  return [
    { label: 'Total BB Projects', value: data?.total_bb_projects ?? 0, format: 'number' as const },
    { label: 'Total GB Projects', value: data?.total_gb_projects ?? 0, format: 'number' as const },
    { label: 'Total Loan Agreements', value: data?.total_loan_agreements ?? 0, format: 'number' as const },
    { label: 'Total Amount', value: data?.total_amount_usd ?? 0, unit: 'USD', format: 'currency' as const },
    { label: 'Total Realisasi', value: data?.total_realized_usd ?? 0, unit: 'USD', format: 'currency' as const },
    { label: 'Overall Absorption', value: data?.overall_absorption_pct ?? 0, format: 'percent' as const },
    { label: 'Active Monitoring', value: data?.active_monitoring ?? 0, format: 'number' as const },
  ]
})

const chartData = computed<MonitoringDisbursement[]>(() => {
  const data = monitoringSummary.value
  if (!data) return []

  return [
    {
      id: 'dashboard-summary',
      loan_agreement_id: 'dashboard',
      budget_year: data.budget_year ?? data.tahun_anggaran ?? filters.budget_year ?? currentYear,
      quarter: data.quarter ?? data.triwulan ?? filters.quarter ?? 'TW1',
      exchange_rate_usd_idr: 0,
      exchange_rate_la_idr: 0,
      planned_la: 0,
      planned_usd: data.total_planned_usd,
      planned_idr: 0,
      realized_la: 0,
      realized_usd: data.total_realized_usd,
      realized_idr: 0,
      absorption_pct: data.absorption_pct,
      komponen: [],
    },
  ]
})

function buildParams(): DashboardFilterParams {
  return {
    budget_year: filters.budget_year ?? undefined,
    quarter: filters.quarter ?? undefined,
    lender_id: filters.lender_id ?? undefined,
  }
}

async function loadSummary() {
  summary.value = await DashboardService.getSummary()
}

async function loadMonitoringSummary() {
  monitoringSummary.value = await DashboardService.getMonitoringSummary(buildParams())
}

async function loadData() {
  loading.value = true
  try {
    await Promise.all([loadSummary(), loadMonitoringSummary()])
  } finally {
    loading.value = false
  }
}

async function applyFilters() {
  loading.value = true
  try {
    await loadMonitoringSummary()
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  void loadData()
})
</script>

<template>
  <section class="space-y-6">
    <PageHeader title="Dashboard" subtitle="Ringkasan monitoring pinjaman luar negeri" />

    <section class="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
      <SummaryCard
        v-for="card in cards"
        :key="card.label"
        :label="card.label"
        :value="card.value"
        :unit="card.unit"
        :format="card.format"
      />
    </section>

    <section class="rounded-lg border border-surface-200 bg-white p-4">
      <div class="grid gap-4 md:grid-cols-[12rem_14rem_1fr_auto]">
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Budget Year</span>
          <Select
            v-model="filters.budget_year"
            :options="yearOptions"
            option-label="label"
            option-value="value"
            show-clear
            class="w-full"
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
        <label class="block space-y-2">
          <span class="text-sm font-medium text-surface-700">Lender</span>
          <LenderSelect v-model="filters.lender_id" placeholder="Semua lender" />
        </label>
        <div class="flex items-end">
          <Button label="Terapkan Filter" icon="pi pi-filter" :loading="loading" @click="applyFilters" />
        </div>
      </div>
    </section>

    <section class="grid gap-4 xl:grid-cols-[24rem_1fr]">
      <div class="rounded-lg border border-surface-200 bg-white p-5">
        <p class="text-sm text-surface-500">Monitoring Overview</p>
        <p class="mt-2 text-4xl font-semibold text-surface-950">
          {{ (monitoringSummary?.absorption_pct ?? 0).toFixed(1) }}%
        </p>
        <div class="mt-4">
          <AbsorptionBar :pct="monitoringSummary?.absorption_pct ?? 0" />
        </div>
        <dl class="mt-6 grid gap-4 text-sm">
          <div class="flex items-center justify-between gap-3">
            <dt class="text-surface-500">Rencana USD</dt>
            <dd class="font-semibold text-surface-900">
              <CurrencyDisplay :amount="monitoringSummary?.total_planned_usd ?? 0" currency="USD" compact />
            </dd>
          </div>
          <div class="flex items-center justify-between gap-3">
            <dt class="text-surface-500">Realisasi USD</dt>
            <dd class="font-semibold text-surface-900">
              <CurrencyDisplay :amount="monitoringSummary?.total_realized_usd ?? 0" currency="USD" compact />
            </dd>
          </div>
        </dl>
      </div>

      <MonitoringChart :data="chartData" />
    </section>

    <section class="overflow-hidden rounded-lg border border-surface-200 bg-white">
      <div class="border-b border-surface-200 p-4">
        <h2 class="text-lg font-semibold text-surface-950">Breakdown by Lender</h2>
      </div>
      <div class="overflow-auto">
        <table class="w-full min-w-[52rem] text-left text-sm">
          <thead class="bg-surface-50 text-xs uppercase tracking-wide text-surface-500">
            <tr>
              <th class="px-4 py-3">Lender</th>
              <th class="px-4 py-3">Type</th>
              <th class="px-4 py-3">Rencana USD</th>
              <th class="px-4 py-3">Realisasi USD</th>
              <th class="px-4 py-3">Absorption</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-surface-100">
            <tr v-if="(monitoringSummary?.by_lender.length ?? 0) === 0">
              <td colspan="5" class="px-4 py-6 text-center text-surface-500">
                Tidak ada data monitoring untuk filter ini.
              </td>
            </tr>
            <tr v-for="item in monitoringSummary?.by_lender ?? []" :key="item.lender.id">
              <td class="px-4 py-3 font-medium text-surface-900">{{ item.lender.name }}</td>
              <td class="px-4 py-3"><Tag :value="item.lender.type" severity="info" rounded /></td>
              <td class="px-4 py-3"><CurrencyDisplay :amount="item.planned_usd" currency="USD" /></td>
              <td class="px-4 py-3"><CurrencyDisplay :amount="item.realized_usd" currency="USD" /></td>
              <td class="px-4 py-3"><AbsorptionBar :pct="item.absorption_pct" /></td>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </section>
</template>
