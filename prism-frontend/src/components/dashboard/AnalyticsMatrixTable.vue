<script setup lang="ts">
import { computed } from 'vue'
import Select from 'primevue/select'
import Skeleton from 'primevue/skeleton'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import AnalyticsDrilldownButton from '@/components/dashboard/AnalyticsDrilldownButton.vue'
import AnalyticsEmptyState from '@/components/dashboard/AnalyticsEmptyState.vue'
import AnalyticsStatusBadge from '@/components/dashboard/AnalyticsStatusBadge.vue'
import type {
  DashboardAnalyticsLenderRef,
  DashboardDrilldownQuery,
  DashboardLenderInstitutionMatrixItem,
} from '@/types/dashboard.types'

const props = withDefaults(
  defineProps<{
    items: DashboardLenderInstitutionMatrixItem[]
    topN: number
    loading?: boolean
  }>(),
  {
    loading: false,
  },
)

const emit = defineEmits<{
  'update:topN': [number]
  drilldown: [DashboardDrilldownQuery]
}>()

const topNOptions = [
  { label: 'Top 5 lender', value: 5 },
  { label: 'Top 10 lender', value: 10 },
  { label: 'Top 15 lender', value: 15 },
]

const topLenders = computed(() => {
  const totals = new Map<string, { lender: DashboardAnalyticsLenderRef; amount: number; projects: number }>()

  props.items.forEach((item) => {
    const current = totals.get(item.lender.id) ?? {
      lender: item.lender,
      amount: 0,
      projects: 0,
    }

    current.amount += item.agreement_amount_usd
    current.projects += item.project_count
    totals.set(item.lender.id, current)
  })

  return [...totals.values()]
    .sort((left, right) => right.amount - left.amount || right.projects - left.projects)
    .slice(0, props.topN)
    .map((item) => item.lender)
})

const institutionRows = computed(() => {
  const rows = new Map<string, { name: string; shortName?: string | null; items: DashboardLenderInstitutionMatrixItem[] }>()
  const visibleLenderIds = new Set(topLenders.value.map((lender) => lender.id))

  props.items
    .filter((item) => visibleLenderIds.has(item.lender.id))
    .forEach((item) => {
      const current = rows.get(item.institution.id) ?? {
        name: item.institution.name,
        shortName: item.institution.short_name,
        items: [],
      }

      current.items.push(item)
      rows.set(item.institution.id, current)
    })

  return [...rows.entries()]
    .map(([id, row]) => ({
      id,
      ...row,
      projectTotal: row.items.reduce((sum, item) => sum + item.project_count, 0),
      amountTotal: row.items.reduce((sum, item) => sum + item.agreement_amount_usd, 0),
    }))
    .sort((left, right) => right.amountTotal - left.amountTotal || right.projectTotal - left.projectTotal)
})

function lenderLabel(lender: DashboardAnalyticsLenderRef) {
  return lender.short_name || lender.name
}

function institutionLabel(row: { name: string; shortName?: string | null }) {
  return row.shortName ? `${row.name} (${row.shortName})` : row.name
}

function lenderSeverity(type: string) {
  if (type === 'KSA') return 'warning'
  if (type === 'Multilateral') return 'secondary'
  return 'info'
}

function cellFor(institutionId: string, lenderId: string) {
  return props.items.find(
    (item) => item.institution.id === institutionId && item.lender.id === lenderId,
  )
}

function formatPercent(value: number) {
  return `${value.toFixed(1)}%`
}
</script>

<template>
  <section class="overflow-hidden rounded-lg border border-surface-200 bg-white">
    <div class="flex flex-wrap items-start justify-between gap-3 border-b border-surface-200 p-4">
      <div>
        <h2 class="text-base font-semibold text-surface-950">
          Matrix Lender per Kementerian/Lembaga
        </h2>
        <p class="mt-1 text-sm text-surface-500">
          Menunjukkan lender yang sudah masuk Loan Agreement atau monitoring pada tiap Kementerian/Lembaga.
        </p>
      </div>
      <Select
        :model-value="topN"
        :options="topNOptions"
        option-label="label"
        option-value="value"
        class="w-44"
        @update:model-value="emit('update:topN', Number($event))"
      />
    </div>

    <div v-if="loading" class="space-y-3 p-4">
      <Skeleton v-for="row in 4" :key="row" height="2.5rem" />
    </div>

    <div v-else-if="items.length === 0" class="p-4">
      <AnalyticsEmptyState
        title="Matrix kosong"
        description="Tidak ada kombinasi lender dan Kementerian/Lembaga untuk filter aktif. Coba hapus filter lender atau tahun."
      />
    </div>

    <div v-else class="overflow-x-auto">
      <table class="w-full min-w-max text-left text-sm">
        <thead class="bg-surface-50 text-xs uppercase text-surface-500">
          <tr>
            <th class="sticky left-0 z-10 bg-surface-50 px-4 py-3 font-semibold">
              Kementerian/Lembaga
            </th>
            <th
              v-for="lender in topLenders"
              :key="lender.id"
              class="min-w-56 px-4 py-3 text-left font-semibold"
            >
              <div class="space-y-1">
                <span class="block normal-case text-surface-700" :title="lender.name">
                  {{ lenderLabel(lender) }}
                </span>
                <AnalyticsStatusBadge :value="lender.type" :severity="lenderSeverity(lender.type)" />
              </div>
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-surface-100">
          <tr v-for="row in institutionRows" :key="row.id" class="align-top hover:bg-surface-50/70">
            <th
              class="sticky left-0 z-10 max-w-80 bg-white px-4 py-3 text-left font-medium text-surface-900"
              :title="institutionLabel(row)"
            >
              <span class="line-clamp-2">{{ institutionLabel(row) }}</span>
            </th>
            <td v-for="lender in topLenders" :key="lender.id" class="px-4 py-3">
              <div v-if="cellFor(row.id, lender.id)" class="space-y-2">
                <div class="flex items-center justify-between gap-3">
                  <span class="text-xs text-surface-500">Project</span>
                  <span class="font-semibold text-surface-950">
                    {{ cellFor(row.id, lender.id)?.project_count }}
                  </span>
                </div>
                <div class="text-xs text-surface-500">
                  <CurrencyDisplay
                    :amount="cellFor(row.id, lender.id)?.agreement_amount_usd ?? 0"
                    currency="USD"
                  />
                </div>
                <div class="flex items-center justify-between gap-3 text-xs">
                  <span class="text-surface-500">Penyerapan</span>
                  <span class="font-semibold text-surface-800">
                    {{ formatPercent(cellFor(row.id, lender.id)?.absorption_pct ?? 0) }}
                  </span>
                </div>
                <AnalyticsDrilldownButton
                  :drilldown="cellFor(row.id, lender.id)?.drilldown"
                  label="Lihat detail"
                  @open="emit('drilldown', $event)"
                />
              </div>
              <span v-else class="text-surface-400">-</span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>
