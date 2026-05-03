<script setup lang="ts">
import { computed } from 'vue'
import { BarChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import type { CurrencyExposureItem } from '@/types/dashboard.types'

use([BarChart, GridComponent, LegendComponent, TooltipComponent, CanvasRenderer])

const props = defineProps<{
  data: CurrencyExposureItem[]
}>()

const stageLabels: Record<string, string> = {
  GB_FUNDING_SOURCE: 'Green Book',
  DK_FINANCING: 'Daftar Kegiatan',
  LA: 'Loan Agreement',
}

const stageOrder = ['GB_FUNDING_SOURCE', 'DK_FINANCING', 'LA']

const usdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
})

const numberFormatter = new Intl.NumberFormat('en-US', {
  maximumFractionDigits: 0,
})

const currencies = computed(() =>
  Array.from(new Set(props.data.map((item) => item.currency))).sort((a, b) => a.localeCompare(b)),
)

const option = computed(() => ({
  color: ['#2563eb', '#0f766e', '#ca8a04'],
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'shadow' },
    formatter: (params: unknown) => {
      const rows = Array.isArray(params)
        ? (params as Array<{ axisValue?: string; marker?: string; seriesName?: string; value?: number }>)
        : []
      const currency = rows[0]?.axisValue ?? ''
      const body = rows
        .filter((row) => (row.value ?? 0) > 0)
        .map((row) => `${row.marker ?? ''}${row.seriesName ?? ''}: ${usdFormatter.format(row.value ?? 0)}`)
      const original = props.data
        .filter((item) => item.currency === currency)
        .reduce((sum, item) => sum + item.amount_original, 0)
      return [`<strong>${currency}</strong>`, ...body, `Original: ${numberFormatter.format(original)}`].join(
        '<br/>',
      )
    },
  },
  legend: { bottom: 0 },
  grid: { left: 24, right: 16, top: 20, bottom: 56, containLabel: true },
  xAxis: {
    type: 'category',
    data: currencies.value,
  },
  yAxis: {
    type: 'value',
    axisLabel: {
      formatter: (value: number) => `$${Math.round(value / 1_000_000)}M`,
    },
  },
  series: stageOrder.map((stage) => ({
    name: stageLabels[stage],
    type: 'bar',
    stack: 'currency',
    data: currencies.value.map((currency) => {
      const rows = props.data.filter((item) => item.currency === currency && item.stage === stage)
      return rows.reduce((sum, item) => sum + item.amount_usd, 0)
    }),
  })),
}))
</script>

<template>
  <section class="rounded-lg border border-surface-200 bg-white p-4">
    <div class="mb-3">
      <h2 class="text-lg font-semibold text-surface-950">Currency Exposure</h2>
      <p class="text-sm text-surface-500">Menggunakan nilai original dan USD yang tersimpan, tanpa kalkulasi kurs.</p>
    </div>
    <VChart :option="option" autoresize class="h-80 w-full" />
  </section>
</template>
