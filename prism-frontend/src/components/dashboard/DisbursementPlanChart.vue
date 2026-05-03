<script setup lang="ts">
import { computed } from 'vue'
import { BarChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import DashboardChartCard from '@/components/dashboard/DashboardChartCard.vue'
import type { GreenBookDisbursementYear } from '@/types/dashboard.types'

use([BarChart, GridComponent, TooltipComponent, CanvasRenderer])

const props = withDefaults(
  defineProps<{
    data?: GreenBookDisbursementYear[]
    loading?: boolean
  }>(),
  {
    data: () => [],
    loading: false,
  },
)

const usdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
})

function toNumber(value: number | string | null | undefined) {
  const numericValue = Number(value)
  return Number.isFinite(numericValue) ? numericValue : 0
}

const chartRows = computed(() =>
  props.data
    .map((item) => ({
      year: item.year,
      amount_usd: toNumber(item.amount_usd),
    }))
    .filter((item) => Number.isFinite(Number(item.year)))
    .sort((a, b) => Number(a.year) - Number(b.year)),
)

const hasData = computed(() => chartRows.value.length > 0)

const option = computed(() => ({
  color: ['#2563eb'],
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'shadow' },
    formatter: (params: unknown) => {
      const items = Array.isArray(params) ? (params as Array<{ axisValue?: string; value?: number }>) : []
      const item = items[0]
      return [`<strong>${item?.axisValue ?? ''}</strong>`, usdFormatter.format(item?.value ?? 0)].join(
        '<br/>',
      )
    },
  },
  grid: { left: 24, right: 16, top: 16, bottom: 24, containLabel: true },
  xAxis: {
    type: 'category',
    data: chartRows.value.map((item) => String(item.year)),
  },
  yAxis: {
    type: 'value',
    axisLabel: {
      formatter: (value: number) => `$${Math.round(value / 1_000_000)}M`,
    },
  },
  series: [
    {
      name: 'Disbursement Plan',
      type: 'bar',
      data: chartRows.value.map((item) => item.amount_usd),
      barMaxWidth: 42,
    },
  ],
}))
</script>

<template>
  <DashboardChartCard
    title="Disbursement Plan by Year"
    subtitle="Total rencana per proyek per tahun, bukan per lender."
    :loading="loading"
    :empty="!hasData"
    empty-title="Belum ada disbursement plan"
    empty-message="Disbursement plan belum tersedia untuk filter ini."
  >
    <div class="h-80 min-h-0 w-full overflow-hidden">
      <VChart :option="option" autoresize class="h-full w-full" />
    </div>
  </DashboardChartCard>
</template>
