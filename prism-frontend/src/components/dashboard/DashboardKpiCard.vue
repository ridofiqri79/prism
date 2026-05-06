<script setup lang="ts">
import type { DashboardKpi, DashboardMetricTone } from '@/types/dashboard-flow.types'

defineProps<{
  item: DashboardKpi
}>()

function kpiToneClass(tone: DashboardMetricTone = 'neutral') {
  const toneClasses: Record<DashboardMetricTone, string> = {
    neutral: 'text-surface-600',
    good: 'text-prism-green-dark',
    warning: 'text-amber-600',
    danger: 'text-red-600',
  }

  return toneClasses[tone]
}
</script>

<template>
  <article class="rounded-lg border border-surface-200 bg-white px-4 py-4 shadow-sm">
    <p class="text-[11px] font-semibold uppercase text-surface-400">{{ item.label }}</p>
    <p class="mt-2 text-2xl font-semibold leading-none text-surface-950">
      {{ item.value }}
      <span v-if="item.unit" class="text-sm font-medium text-surface-500">{{ item.unit }}</span>
    </p>
    <p class="mt-2 text-xs font-semibold" :class="kpiToneClass(item.tone)">
      {{ item.delta }}
    </p>
  </article>
</template>
