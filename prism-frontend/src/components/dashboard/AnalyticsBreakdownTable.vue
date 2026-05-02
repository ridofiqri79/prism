<script setup lang="ts">
import Skeleton from 'primevue/skeleton'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import AnalyticsDrilldownButton from '@/components/dashboard/AnalyticsDrilldownButton.vue'
import AnalyticsEmptyState from '@/components/dashboard/AnalyticsEmptyState.vue'
import AnalyticsStatusBadge from '@/components/dashboard/AnalyticsStatusBadge.vue'
import type {
  AnalyticsBreakdownTableColumn,
  AnalyticsBreakdownTableRow,
  DashboardDrilldownQuery,
} from '@/types/dashboard.types'

withDefaults(
  defineProps<{
    title: string
    description?: string
    columns: AnalyticsBreakdownTableColumn[]
    rows: AnalyticsBreakdownTableRow[]
    loading?: boolean
    emptyTitle?: string
    emptyDescription?: string
  }>(),
  {
    description: undefined,
    loading: false,
    emptyTitle: 'Tidak ada data',
    emptyDescription: 'Data kosong untuk filter aktif.',
  },
)

const emit = defineEmits<{
  drilldown: [DashboardDrilldownQuery]
}>()

function formatNumber(value: string | number | null | undefined) {
  if (typeof value === 'number') return new Intl.NumberFormat('id-ID').format(value)
  return value || '-'
}

function formatPercent(value: string | number | null | undefined) {
  if (typeof value === 'number') return `${value.toFixed(1)}%`
  return value || '-'
}

function normalizedPercent(value: string | number | null | undefined) {
  const numeric = typeof value === 'number' && Number.isFinite(value) ? value : 0

  return Math.min(100, Math.max(0, numeric))
}

function absorptionClass(value: string | number | null | undefined) {
  const numeric = typeof value === 'number' && Number.isFinite(value) ? value : 0

  if (numeric < 50) return 'bg-red-500'
  if (numeric >= 90) return 'bg-emerald-500'
  return 'bg-sky-500'
}

function cellClass(column: AnalyticsBreakdownTableColumn) {
  if (
    column.align === 'right' ||
    column.kind === 'number' ||
    column.kind === 'currency' ||
    column.kind === 'percent'
  ) {
    return 'text-right'
  }
  if (column.align === 'center' || column.kind === 'badge' || column.kind === 'drilldown') {
    return 'text-center'
  }
  return 'text-left'
}
</script>

<template>
  <section class="overflow-hidden rounded-lg border border-surface-200 bg-white">
    <div class="border-b border-surface-200 p-4">
      <h2 class="text-base font-semibold text-surface-950">{{ title }}</h2>
      <p v-if="description" class="mt-1 text-sm text-surface-500">{{ description }}</p>
    </div>

    <div v-if="loading" class="space-y-3 p-4">
      <Skeleton v-for="row in 4" :key="row" height="2.25rem" />
    </div>

    <div v-else-if="rows.length === 0" class="p-4">
      <AnalyticsEmptyState :title="emptyTitle" :description="emptyDescription" />
    </div>

    <div v-else class="overflow-x-auto">
      <table class="w-full min-w-max text-left text-sm">
        <thead class="bg-surface-50 text-xs uppercase text-surface-500">
          <tr>
            <th
              v-for="column in columns"
              :key="column.key"
              class="px-4 py-3 font-semibold"
              :class="cellClass(column)"
            >
              {{ column.label }}
            </th>
          </tr>
        </thead>
        <tbody class="divide-y divide-surface-100">
          <tr v-for="row in rows" :key="row.id" class="align-top hover:bg-surface-50/70">
            <td
              v-for="column in columns"
              :key="column.key"
              class="px-4 py-3 text-surface-700"
              :class="cellClass(column)"
            >
              <CurrencyDisplay
                v-if="column.kind === 'currency'"
                :amount="Number(row.cells[column.key] ?? 0)"
                currency="USD"
              />
              <span v-else-if="column.kind === 'percent'">{{
                formatPercent(row.cells[column.key])
              }}</span>
              <span v-else-if="column.kind === 'number'">{{
                formatNumber(row.cells[column.key])
              }}</span>
              <AnalyticsStatusBadge
                v-else-if="column.kind === 'badge'"
                :value="String(row.cells[column.key] || '-')"
                :severity="row.severity"
              />
              <div v-else-if="column.kind === 'absorption'" class="min-w-36 space-y-1">
                <div class="h-2 overflow-hidden rounded-full bg-surface-100">
                  <div
                    class="h-full rounded-full"
                    :class="absorptionClass(row.cells[column.key])"
                    :style="{ width: `${normalizedPercent(row.cells[column.key])}%` }"
                  />
                </div>
                <span class="block text-xs font-semibold text-surface-700">
                  {{ formatPercent(row.cells[column.key]) }}
                </span>
              </div>
              <AnalyticsDrilldownButton
                v-else-if="column.kind === 'drilldown'"
                :drilldown="row.drilldown"
                label="Buka"
                @open="emit('drilldown', $event)"
              />
              <span v-else class="block max-w-[28rem] whitespace-normal leading-relaxed">
                {{ row.cells[column.key] || '-' }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </section>
</template>
