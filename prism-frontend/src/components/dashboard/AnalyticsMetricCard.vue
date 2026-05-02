<script setup lang="ts">
import Skeleton from 'primevue/skeleton'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import AnalyticsDrilldownButton from '@/components/dashboard/AnalyticsDrilldownButton.vue'
import AnalyticsStatusBadge from '@/components/dashboard/AnalyticsStatusBadge.vue'
import type { AnalyticsMoneyMetric, DashboardDrilldownQuery } from '@/types/dashboard.types'

withDefaults(
  defineProps<{
    metric: AnalyticsMoneyMetric
    loading?: boolean
  }>(),
  {
    loading: false,
  },
)

const emit = defineEmits<{
  drilldown: [DashboardDrilldownQuery]
}>()

function formatNumber(value: number) {
  return new Intl.NumberFormat('id-ID').format(value)
}

function formatPercent(value: number) {
  return `${value.toFixed(1)}%`
}
</script>

<template>
  <article
    class="flex min-h-36 flex-col justify-between rounded-lg border border-surface-200 bg-white p-4"
  >
    <div class="space-y-2">
      <div class="flex items-start justify-between gap-3">
        <p class="text-sm font-medium leading-snug text-surface-600">{{ metric.label }}</p>
        <AnalyticsStatusBadge
          v-if="metric.severity"
          :value="metric.severity"
          :severity="metric.severity"
        />
      </div>

      <Skeleton v-if="loading" width="9rem" height="2rem" />
      <p v-else class="break-words text-2xl font-semibold leading-tight text-surface-950">
        <CurrencyDisplay
          v-if="metric.format === 'currency'"
          :amount="metric.value"
          :currency="metric.unit || 'USD'"
        />
        <span v-else-if="metric.format === 'percent'">{{ formatPercent(metric.value) }}</span>
        <span v-else>{{ formatNumber(metric.value) }}</span>
      </p>
    </div>

    <div v-if="metric.drilldown" class="mt-4">
      <AnalyticsDrilldownButton
        :drilldown="metric.drilldown"
        label="Lihat"
        @open="emit('drilldown', $event)"
      />
    </div>
  </article>
</template>
