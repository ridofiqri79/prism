<script setup lang="ts">
import { computed } from 'vue'
import { BarChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import type { LAComponentBreakdownItem } from '@/types/dashboard.types'

use([BarChart, GridComponent, LegendComponent, TooltipComponent, CanvasRenderer])

const props = defineProps<{
  data: LAComponentBreakdownItem[]
}>()

const usdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
})

const option = computed(() => ({
  color: ['#2563eb', '#059669'],
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
    type: 'value',
    axisLabel: {
      formatter: (value: number) => usdFormatter.format(value).replace('.00', ''),
    },
  },
  yAxis: {
    type: 'category',
    data: props.data.map((item) => item.component_name),
  },
  series: [
    {
      name: 'Planned',
      type: 'bar',
      data: props.data.map((item) => item.planned_usd),
      barMaxWidth: 26,
    },
    {
      name: 'Realized',
      type: 'bar',
      data: props.data.map((item) => item.realized_usd),
      barMaxWidth: 26,
    },
  ],
}))
</script>

<template>
  <section class="rounded-lg border border-surface-200 bg-white p-4">
    <div class="mb-3">
      <h2 class="text-lg font-semibold text-surface-950">Component Breakdown</h2>
      <p class="text-sm text-surface-500">Breakdown dari monitoring_komponen jika tersedia.</p>
    </div>
    <VChart v-if="data.length" :option="option" autoresize class="h-[24rem] w-full" />
    <div v-else class="flex h-[18rem] items-center justify-center text-sm text-surface-500">
      Tidak ada data komponen monitoring untuk filter ini.
    </div>
  </section>
</template>
