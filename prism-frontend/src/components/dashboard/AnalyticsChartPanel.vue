<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { BarChart, LineChart, PieChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { SVGRenderer } from 'echarts/renderers'
import type { EChartsOption } from 'echarts'
import Skeleton from 'primevue/skeleton'
import VChart from 'vue-echarts'
import AnalyticsEmptyState from '@/components/dashboard/AnalyticsEmptyState.vue'

use([BarChart, LineChart, PieChart, GridComponent, LegendComponent, TooltipComponent, SVGRenderer])

const props = withDefaults(
  defineProps<{
    title: string
    description?: string
    option: unknown
    loading?: boolean
    empty?: boolean
    emptyTitle?: string
    emptyDescription?: string
    height?: string
  }>(),
  {
    description: undefined,
    loading: false,
    empty: false,
    emptyTitle: 'Tidak ada data',
    emptyDescription: 'Data chart kosong untuk filter aktif.',
    height: '20rem',
  },
)

const chartOption = computed(() => props.option as EChartsOption)
const chartStyle = computed(() => ({ height: props.height }))
const chartInitOptions = { renderer: 'svg' as const }
const chartReady = ref(false)
let frameId: number | null = null

function cancelScheduledFrame() {
  if (frameId !== null) {
    window.cancelAnimationFrame(frameId)
    frameId = null
  }
}

function scheduleChartMount() {
  cancelScheduledFrame()
  chartReady.value = false

  if (props.loading || props.empty) return

  void nextTick(() => {
    frameId = window.requestAnimationFrame(() => {
      frameId = window.requestAnimationFrame(() => {
        chartReady.value = true
        frameId = null
      })
    })
  })
}

onMounted(scheduleChartMount)
onBeforeUnmount(cancelScheduledFrame)
watch([() => props.loading, () => props.empty], scheduleChartMount)
</script>

<template>
  <section class="overflow-hidden rounded-lg border border-surface-200 bg-white">
    <div class="border-b border-surface-200 p-4">
      <h2 class="text-base font-semibold text-surface-950">{{ title }}</h2>
      <p v-if="description" class="mt-1 text-sm text-surface-500">{{ description }}</p>
    </div>

    <div v-if="loading" class="space-y-3 p-4">
      <Skeleton height="2rem" />
      <Skeleton height="14rem" />
    </div>

    <div v-else-if="empty" class="p-4">
      <AnalyticsEmptyState :title="emptyTitle" :description="emptyDescription" />
    </div>

    <div v-else class="p-4">
      <Skeleton v-if="!chartReady" class="analytics-chart w-full" :style="chartStyle" />
      <VChart
        v-else
        :option="chartOption"
        :init-options="chartInitOptions"
        autoresize
        class="analytics-chart w-full"
        :style="chartStyle"
      />
    </div>
  </section>
</template>

<style scoped>
.analytics-chart {
  min-height: 16rem;
}
</style>
