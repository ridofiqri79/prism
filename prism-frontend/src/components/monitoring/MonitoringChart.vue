<script setup lang="ts">
import { computed } from 'vue'
import { BarChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import type { MonitoringDisbursement, Quarter } from '@/types/monitoring.types'

use([BarChart, GridComponent, LegendComponent, TooltipComponent, CanvasRenderer])

const props = defineProps<{
  data: MonitoringDisbursement[]
}>()

const quarters: Quarter[] = ['TW1', 'TW2', 'TW3', 'TW4']
const chartData = computed(() =>
  quarters.map((quarter) => {
    const rows = props.data.filter((item) => item.quarter === quarter)
    const planned = rows.reduce((sum, item) => sum + item.planned_usd, 0)
    const realized = rows.reduce((sum, item) => sum + item.realized_usd, 0)
    const absorption = planned === 0 ? 0 : Math.round((realized / planned) * 1000) / 10

    return { quarter, planned, realized, absorption }
  }),
)

const option = computed(() => ({
  color: ['#2563eb', '#059669'],
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'shadow' },
    formatter: (params: unknown) => {
      const items = Array.isArray(params) ? (params as Array<{ axisValue?: string }>) : []
      const quarter = String(items[0]?.axisValue ?? '')
      const row = chartData.value.find((item) => item.quarter === quarter)
      const formatter = new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD' })

      return [
        `<strong>${quarter}</strong>`,
        `Rencana: ${formatter.format(row?.planned ?? 0)}`,
        `Realisasi: ${formatter.format(row?.realized ?? 0)}`,
        `Penyerapan: ${(row?.absorption ?? 0).toFixed(1)}%`,
      ].join('<br/>')
    },
  },
  legend: { top: 0 },
  grid: { left: 32, right: 16, top: 44, bottom: 24, containLabel: true },
  xAxis: {
    type: 'category',
    data: chartData.value.map((item) => item.quarter),
  },
  yAxis: {
    type: 'value',
    axisLabel: {
      formatter: (value: number) => `$${Math.round(value / 1_000_000)}M`,
    },
  },
  series: [
    {
      name: 'Planned',
      type: 'bar',
      data: chartData.value.map((item) => item.planned),
      barMaxWidth: 36,
    },
    {
      name: 'Realized',
      type: 'bar',
      data: chartData.value.map((item) => item.realized),
      barMaxWidth: 36,
    },
  ],
}))
</script>

<template>
  <div class="rounded-lg border border-surface-200 bg-white p-4">
    <div class="mb-4">
      <h2 class="text-lg font-semibold text-surface-950">Planned vs Realized per Triwulan</h2>
      <p class="text-sm text-surface-500">Nilai dalam USD, dikelompokkan berdasarkan TW1-TW4.</p>
    </div>
    <VChart :option="option" autoresize class="h-80 w-full" />
  </div>
</template>
