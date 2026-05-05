<script setup lang="ts">
import { computed } from 'vue'
import { ScatterChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import type { KLPortfolioPerformanceItem } from '@/types/dashboard.types'

use([ScatterChart, GridComponent, TooltipComponent, CanvasRenderer])

const props = defineProps<{
  items: KLPortfolioPerformanceItem[]
}>()

const usdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
})

const chartData = computed(() =>
  props.items.map((item) => ({
    name: item.institution_name,
    value: [item.la_commitment_usd, item.pipeline_usd, item.risk_count],
    item,
  })),
)

const option = computed(() => ({
  color: ['#2563eb'],
  tooltip: {
    trigger: 'item',
    formatter: (params: { data?: { name?: string; item?: KLPortfolioPerformanceItem } }) => {
      const item = params.data?.item
      if (!item) return ''
      return [
        `<strong>${params.data?.name ?? ''}</strong>`,
        `LA: ${usdFormatter.format(item.la_commitment_usd)}`,
        `Pipeline: ${usdFormatter.format(item.pipeline_usd)}`,
        `Risk: ${item.risk_count}`,
      ].join('<br/>')
    },
  },
  grid: { left: 24, right: 24, top: 16, bottom: 28, containLabel: true },
  xAxis: {
    type: 'value',
    name: 'LA Commitment',
    axisLabel: {
      formatter: (value: number) => `$${Math.round(value / 1_000_000)}M`,
    },
  },
  yAxis: {
    type: 'value',
    name: 'Pipeline',
    axisLabel: {
      formatter: (value: number) => `$${Math.round(value / 1_000_000)}M`,
    },
  },
  series: [
    {
      name: 'K/L Performance',
      type: 'scatter',
      data: chartData.value,
      symbolSize: (value: number[]) => Math.max(10, Math.min(34, 10 + (value[2] ?? 0) * 3)),
      emphasis: {
        focus: 'self',
      },
    },
  ],
}))
</script>

<template>
  <section class="rounded-lg border border-surface-200 bg-white p-4">
    <div class="mb-3">
      <h2 class="text-lg font-semibold text-surface-950">LA Commitment vs Pipeline</h2>
      <p class="text-sm text-surface-500">Bubble size follows risk count.</p>
    </div>
    <VChart :option="option" autoresize class="h-80 w-full" />
  </section>
</template>
