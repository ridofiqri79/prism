<script setup lang="ts">
import { computed } from 'vue'
import { FunnelChart } from 'echarts/charts'
import { LegendComponent, TooltipComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import type { StageMetric } from '@/types/dashboard.types'

use([FunnelChart, LegendComponent, TooltipComponent, CanvasRenderer])

const props = defineProps<{
  data: StageMetric[]
}>()

const usdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
})

const chartRows = computed(() =>
  props.data.map((item) => ({
    name: item.label,
    value: item.project_count,
    amount: item.amount_usd,
  })),
)

const option = computed(() => ({
  color: ['#2563eb', '#0f766e', '#16a34a', '#ca8a04', '#dc2626'],
  tooltip: {
    trigger: 'item',
    formatter: (params: { name?: string; value?: number; data?: { amount?: number } }) => {
      const amount = params.data?.amount ?? 0
      return [
        `<strong>${params.name ?? ''}</strong>`,
        `Proyek: ${params.value ?? 0}`,
        `Nilai: ${usdFormatter.format(amount)}`,
      ].join('<br/>')
    },
  },
  legend: {
    bottom: 0,
    type: 'scroll',
  },
  series: [
    {
      name: 'Portfolio Funnel',
      type: 'funnel',
      left: '5%',
      top: 16,
      bottom: 52,
      width: '90%',
      minSize: '28%',
      maxSize: '100%',
      sort: 'none',
      gap: 4,
      label: {
        show: true,
        formatter: (params: { name?: string; value?: number }) =>
          `${params.name ?? ''}\n${params.value ?? 0} proyek`,
      },
      labelLine: {
        length: 12,
      },
      itemStyle: {
        borderColor: '#fff',
        borderWidth: 1,
      },
      data: chartRows.value,
    },
  ],
}))
</script>

<template>
  <section class="rounded-lg border border-surface-200 bg-white p-4">
    <div class="mb-3">
      <h2 class="text-lg font-semibold text-surface-950">Portfolio Funnel</h2>
      <p class="text-sm text-surface-500">Blue Book sampai Monitoring tanpa penghitungan ulang di komponen.</p>
    </div>
    <VChart :option="option" autoresize class="h-[24rem] w-full" />
  </section>
</template>
