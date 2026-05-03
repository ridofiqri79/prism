<script setup lang="ts">
import { computed } from 'vue'
import { BarChart, LineChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import DashboardChartCard from '@/components/dashboard/DashboardChartCard.vue'
import type { LADisbursementTrendPoint } from '@/types/dashboard.types'

use([BarChart, LineChart, GridComponent, LegendComponent, TooltipComponent, CanvasRenderer])

const props = defineProps<{
  data: LADisbursementTrendPoint[]
}>()

const usdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
})

const option = computed(() => ({
  color: ['#2563eb', '#059669', '#ca8a04'],
  tooltip: {
    trigger: 'axis',
    valueFormatter: (value: number | string) =>
      typeof value === 'number' ? usdFormatter.format(value) : String(value),
  },
  legend: {
    top: 0,
  },
  grid: {
    left: 8,
    right: 8,
    top: 42,
    bottom: 8,
    containLabel: true,
  },
  xAxis: {
    type: 'category',
    data: props.data.map((item) => item.period),
  },
  yAxis: [
    {
      type: 'value',
      axisLabel: {
        formatter: (value: number) => usdFormatter.format(value).replace('.00', ''),
      },
    },
    {
      type: 'value',
      min: 0,
      max: 100,
      axisLabel: {
        formatter: '{value}%',
      },
    },
  ],
  series: [
    {
      name: 'Planned',
      type: 'bar',
      data: props.data.map((item) => item.planned_usd),
      barMaxWidth: 32,
    },
    {
      name: 'Realized',
      type: 'bar',
      data: props.data.map((item) => item.realized_usd),
      barMaxWidth: 32,
    },
    {
      name: 'Absorption',
      type: 'line',
      yAxisIndex: 1,
      smooth: true,
      data: props.data.map((item) => item.absorption_pct),
      tooltip: {
        valueFormatter: (value: number) => `${value.toFixed(2)}%`,
      },
    },
  ],
}))
</script>

<template>
  <DashboardChartCard
    title="Planned vs Realized Trend"
    subtitle="Monitoring disbursement per budget year and quarter."
    :empty="data.length === 0"
    empty-title="Tidak ada trend monitoring"
    empty-message="Tidak ada trend monitoring untuk filter ini."
  >
    <VChart :option="option" autoresize class="h-[24rem] w-full" />
  </DashboardChartCard>
</template>
