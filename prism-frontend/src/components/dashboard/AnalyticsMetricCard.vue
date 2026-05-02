<script setup lang="ts">
import { computed } from 'vue'
import Skeleton from 'primevue/skeleton'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import AnalyticsDrilldownButton from '@/components/dashboard/AnalyticsDrilldownButton.vue'
import AnalyticsStatusBadge from '@/components/dashboard/AnalyticsStatusBadge.vue'
import type { AnalyticsMoneyMetric, DashboardDrilldownQuery } from '@/types/dashboard.types'

const props = withDefaults(
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

const isHighEmphasis = computed(() => props.metric.emphasis === 'high')
const fullValue = computed(() => {
  if (props.metric.format === 'currency') {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: props.metric.unit || 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(props.metric.value)
  }

  if (props.metric.format === 'percent') return formatPercent(props.metric.value)

  return formatNumber(props.metric.value)
})

function formatNumber(value: number) {
  return new Intl.NumberFormat('id-ID').format(value)
}

function formatPercent(value: number) {
  return `${value.toFixed(1)}%`
}
</script>

<template>
  <article
    class="flex min-h-36 flex-col justify-between rounded-lg border bg-white p-4 transition-shadow"
    :class="
      isHighEmphasis
        ? 'border-primary-200 shadow-sm ring-1 ring-primary-100'
        : 'border-surface-200'
    "
  >
    <div class="space-y-2">
      <div class="flex items-start justify-between gap-3">
        <div class="flex min-w-0 items-start gap-1.5">
          <p class="text-sm font-medium leading-snug text-surface-600">{{ metric.label }}</p>
          <i
            v-if="metric.hint"
            v-tooltip.top="metric.hint"
            class="pi pi-info-circle mt-0.5 shrink-0 text-xs text-surface-400"
            tabindex="0"
            :aria-label="metric.hint"
          />
        </div>
        <AnalyticsStatusBadge
          v-if="metric.severity"
          :value="metric.severity"
          :severity="metric.severity"
        />
      </div>

      <Skeleton v-if="loading" width="9rem" height="2rem" />
      <p
        v-else
        class="break-words font-semibold leading-tight text-surface-950"
        :class="isHighEmphasis ? 'text-3xl' : 'text-2xl'"
        :title="fullValue"
      >
        <CurrencyDisplay
          v-if="metric.format === 'currency'"
          :amount="metric.value"
          :currency="metric.unit || 'USD'"
          compact
        />
        <span v-else-if="metric.format === 'percent'">{{ formatPercent(metric.value) }}</span>
        <span v-else>{{ formatNumber(metric.value) }}</span>
      </p>
    </div>

    <div v-if="metric.drilldown" class="mt-4">
      <AnalyticsDrilldownButton
        :drilldown="metric.drilldown"
        :label="metric.actionLabel || 'Lihat detail'"
        @open="emit('drilldown', $event)"
      />
    </div>
  </article>
</template>
