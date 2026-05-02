<script setup lang="ts">
import AnalyticsMetricCard from '@/components/dashboard/AnalyticsMetricCard.vue'
import type { AnalyticsMoneyMetric, DashboardDrilldownQuery } from '@/types/dashboard.types'

withDefaults(
  defineProps<{
    metrics: AnalyticsMoneyMetric[]
    loading?: boolean
  }>(),
  {
    loading: false,
  },
)

const emit = defineEmits<{
  drilldown: [DashboardDrilldownQuery]
}>()
</script>

<template>
  <section class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
    <AnalyticsMetricCard
      v-for="metric in metrics"
      :key="metric.key"
      :metric="metric"
      :loading="loading"
      @drilldown="emit('drilldown', $event)"
    />
  </section>
</template>
