<script setup lang="ts">
import { computed } from 'vue'
import { BarChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import DashboardChartCard from '@/components/dashboard/DashboardChartCard.vue'
import type { LenderCertaintyPoint, LenderCertaintyStage } from '@/types/dashboard.types'

use([BarChart, GridComponent, LegendComponent, TooltipComponent, CanvasRenderer])

const props = defineProps<{
  data: LenderCertaintyPoint[]
}>()

const stageLabels: Record<LenderCertaintyStage, string> = {
  LENDER_INDICATION: 'Indication',
  LOI: 'Letter of Intent',
  GB_FUNDING_SOURCE: 'Green Book',
  DK_FINANCING: 'Daftar Kegiatan',
  LA: 'Loan Agreement',
}

const stageOrder = Object.keys(stageLabels) as LenderCertaintyStage[]

const usdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
})

const lenders = computed(() => {
  const totals = new Map<string, { name: string; amount: number }>()
  for (const item of props.data) {
    const current = totals.get(item.lender_id) ?? { name: item.lender_name, amount: 0 }
    current.amount += item.amount_usd
    totals.set(item.lender_id, current)
  }
  return Array.from(totals.values())
    .sort((a, b) => b.amount - a.amount)
    .slice(0, 12)
    .map((item) => item.name)
})

const option = computed(() => ({
  color: ['#2563eb', '#0f766e', '#16a34a', '#ca8a04', '#dc2626'],
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'shadow' },
    formatter: (params: unknown) => {
      const rows = Array.isArray(params)
        ? (params as Array<{ marker?: string; seriesName?: string; value?: number }>)
        : []
      return rows
        .filter((row) => (row.value ?? 0) > 0)
        .map((row) => `${row.marker ?? ''}${row.seriesName ?? ''}: ${usdFormatter.format(row.value ?? 0)}`)
        .join('<br/>')
    },
  },
  legend: { bottom: 0, type: 'scroll' },
  grid: { left: 24, right: 16, top: 20, bottom: 72, containLabel: true },
  xAxis: {
    type: 'category',
    data: lenders.value,
    axisLabel: { interval: 0, rotate: lenders.value.length > 5 ? 28 : 0 },
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
    stack: 'certainty',
    emphasis: { focus: 'series' },
    data: lenders.value.map((lenderName) => {
      const item = props.data.find((row) => row.lender_name === lenderName && row.stage === stage)
      return item?.amount_usd ?? 0
    }),
  })),
}))
</script>

<template>
  <DashboardChartCard
    title="Certainty Ladder by Lender"
    subtitle="Indication sampai legal commitment, dipisahkan per sumber data."
    :empty="data.length === 0"
    empty-title="Tidak ada certainty ladder"
    empty-message="Ubah filter lender atau tahun untuk melihat data."
  >
    <VChart :option="option" autoresize class="h-[24rem] w-full" />
  </DashboardChartCard>
</template>
