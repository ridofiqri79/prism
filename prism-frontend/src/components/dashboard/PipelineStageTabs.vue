<script setup lang="ts">
import { computed } from 'vue'
import type { PipelineBottleneckStage, PipelineStageSummary } from '@/types/dashboard.types'

const props = defineProps<{
  stages: PipelineStageSummary[]
  activeStage: PipelineBottleneckStage | null
}>()

const emit = defineEmits<{
  'update:activeStage': [value: PipelineBottleneckStage | null]
}>()

const usdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  maximumFractionDigits: 0,
  notation: 'compact',
})

const total = computed(() =>
  props.stages.reduce(
    (acc, stage) => ({
      count: acc.count + stage.project_count,
      amount: acc.amount + stage.amount_usd,
    }),
    { count: 0, amount: 0 },
  ),
)

function stageLabel(stage: PipelineStageSummary) {
  const labels: Record<PipelineBottleneckStage, string> = {
    BB_NO_LENDER: 'Blue Book tanpa lender',
    INDICATION_NO_LOI: 'Indikasi tanpa Letter of Intent',
    LOI_NO_GB: 'Letter of Intent tanpa Green Book',
    GB_NO_DK: 'Green Book tanpa Daftar Kegiatan',
    DK_NO_LA: 'Daftar Kegiatan tanpa Loan Agreement',
    LA_NOT_EFFECTIVE: 'Loan Agreement belum efektif',
    EFFECTIVE_NO_MONITORING: 'Efektif tanpa monitoring',
  }

  return labels[stage.stage] ?? stage.label
}

function tabClass(isActive: boolean) {
  return [
    'min-h-24 rounded-lg border p-4 text-left transition-colors',
    isActive
      ? 'border-primary-500 bg-primary-50 text-primary-950'
      : 'border-surface-200 bg-white text-surface-800 hover:border-primary-300 hover:bg-surface-50',
  ]
}
</script>

<template>
  <section class="grid gap-3 md:grid-cols-2 xl:grid-cols-4">
    <button type="button" :class="tabClass(activeStage === null)" @click="emit('update:activeStage', null)">
      <span class="block text-sm font-medium">Semua bottleneck</span>
      <span class="mt-2 block text-2xl font-semibold">{{ total.count }}</span>
      <span class="mt-1 block text-xs text-surface-500">{{ usdFormatter.format(total.amount) }}</span>
    </button>

    <button
      v-for="stage in stages"
      :key="stage.stage"
      type="button"
      :class="tabClass(activeStage === stage.stage)"
      @click="emit('update:activeStage', stage.stage)"
    >
      <span class="block text-sm font-medium">{{ stageLabel(stage) }}</span>
      <span class="mt-2 block text-2xl font-semibold">{{ stage.project_count }}</span>
      <span class="mt-1 block text-xs text-surface-500">
        {{ usdFormatter.format(stage.amount_usd) }} - avg {{ stage.avg_age_days.toFixed(0) }} hari
      </span>
    </button>
  </section>
</template>
