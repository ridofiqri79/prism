<script setup lang="ts">
import { computed } from 'vue'
import { PieChart } from 'echarts/charts'
import { LegendComponent, TooltipComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import DashboardChartCard from '@/components/dashboard/DashboardChartCard.vue'
import type { GreenBookFundingAllocation } from '@/types/dashboard.types'

use([PieChart, LegendComponent, TooltipComponent, CanvasRenderer])

const props = withDefaults(
  defineProps<{
    allocation?: GreenBookFundingAllocation
    loading?: boolean
  }>(),
  {
    allocation: () => ({
      services: 0,
      constructions: 0,
      goods: 0,
      trainings: 0,
      other: 0,
    }),
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

const rows = computed(() => [
  { name: 'Services', value: toNumber(props.allocation.services) },
  { name: 'Constructions', value: toNumber(props.allocation.constructions) },
  { name: 'Goods', value: toNumber(props.allocation.goods) },
  { name: 'Trainings', value: toNumber(props.allocation.trainings) },
  { name: 'Other', value: toNumber(props.allocation.other) },
])

const hasAllocation = computed(() => rows.value.some((item) => item.value > 0))

const option = computed(() => ({
  color: ['#2563eb', '#0f766e', '#ca8a04', '#7c3aed', '#64748b'],
  tooltip: {
    trigger: 'item',
    formatter: (params: { name?: string; value?: number; percent?: number }) =>
      [
        `<strong>${params.name ?? ''}</strong>`,
        usdFormatter.format(params.value ?? 0),
        `${(params.percent ?? 0).toFixed(1)}%`,
      ].join('<br/>'),
  },
  legend: {
    bottom: 0,
    type: 'scroll',
  },
  series: [
    {
      name: 'Funding Allocation',
      type: 'pie',
      radius: ['48%', '72%'],
      center: ['50%', '44%'],
      avoidLabelOverlap: true,
      label: {
        show: false,
      },
      emphasis: {
        label: {
          show: true,
          formatter: '{b}\n{d}%',
          fontWeight: 600,
        },
      },
      data: rows.value,
    },
  ],
}))
</script>

<template>
  <DashboardChartCard
    title="Funding Allocation"
    subtitle="Alokasi mengikuti relasi funding allocation ke activity Green Book."
    :loading="loading"
    :empty="!hasAllocation"
    empty-title="Belum ada allocation"
    empty-message="Funding allocation belum tersedia untuk filter ini."
  >
    <div class="h-80 min-h-0 w-full overflow-hidden">
      <VChart :option="option" autoresize class="h-full w-full" />
    </div>
  </DashboardChartCard>
</template>
