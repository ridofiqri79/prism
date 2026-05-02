<script setup lang="ts">
import { computed } from 'vue'
import { BarChart, LineChart } from 'echarts/charts'
import { GridComponent, LegendComponent, TooltipComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { CanvasRenderer } from 'echarts/renderers'
import type { EChartsOption } from 'echarts'
import Skeleton from 'primevue/skeleton'
import VChart from 'vue-echarts'
import AnalyticsEmptyState from '@/components/dashboard/AnalyticsEmptyState.vue'

use([BarChart, LineChart, GridComponent, LegendComponent, TooltipComponent, CanvasRenderer])

const props = withDefaults(
  defineProps<{
    title: string
    description?: string
    option: unknown
    loading?: boolean
    empty?: boolean
    emptyTitle?: string
    emptyDescription?: string
    heightClass?: string
  }>(),
  {
    description: undefined,
    loading: false,
    empty: false,
    emptyTitle: 'Tidak ada data',
    emptyDescription: 'Data chart kosong untuk filter aktif.',
    heightClass: 'h-80',
  },
)

const chartOption = computed(() => props.option as EChartsOption)
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
      <VChart :option="chartOption" autoresize class="w-full" :class="heightClass" />
    </div>
  </section>
</template>
