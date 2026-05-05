<script setup lang="ts">
import { computed, ref } from 'vue'
import DashboardChartCard from '@/components/dashboard/DashboardChartCard.vue'
import type {
  BreakdownItem,
  PipelineBottleneckItem,
  PipelineStageSummary,
  RiskItem,
  StageMetric,
} from '@/types/dashboard.types'

const props = withDefaults(
  defineProps<{
    data: StageMetric[]
    topInstitutions?: BreakdownItem[]
    topLenders?: BreakdownItem[]
    pipelineSummary?: PipelineStageSummary[]
    bottleneckItems?: PipelineBottleneckItem[]
    riskItems?: RiskItem[]
  }>(),
  {
    topInstitutions: () => [],
    topLenders: () => [],
    pipelineSummary: () => [],
    bottleneckItems: () => [],
    riskItems: () => [],
  },
)

const stageOrder = ['BB', 'GB', 'DK', 'LA'] as const
type FlowStage = (typeof stageOrder)[number]

const stageDefinitions: Record<
  FlowStage,
  {
    title: string
    shortTitle: string
    amountLabel: string
    nextLabel: string
    color: string
    detailTitle: string
    riskAccent: 'green' | 'amber' | 'red'
  }
> = {
  BB: {
    title: 'Blue Book',
    shortTitle: 'Blue Book',
    amountLabel: 'Nilai indikatif',
    nextLabel: 'Green Book',
    color: '#2563eb',
    detailTitle: 'Breakdown by K/L',
    riskAccent: 'green',
  },
  GB: {
    title: 'Green Book',
    shortTitle: 'Green Book',
    amountLabel: 'Nilai pipeline',
    nextLabel: 'Daftar Kegiatan',
    color: '#2fa7d8',
    detailTitle: 'Breakdown by lender',
    riskAccent: 'green',
  },
  DK: {
    title: 'Daftar Kegiatan',
    shortTitle: 'Daftar Kegiatan',
    amountLabel: 'Nilai pembiayaan',
    nextLabel: 'Loan Agreement',
    color: '#35b99f',
    detailTitle: 'Penyebab tertahan',
    riskAccent: 'red',
  },
  LA: {
    title: 'Loan Agreement',
    shortTitle: 'Loan Agreement',
    amountLabel: 'Komitmen legal',
    nextLabel: 'Tahap final',
    color: '#fb923c',
    detailTitle: 'Catatan',
    riskAccent: 'amber',
  },
}

const countFormatter = new Intl.NumberFormat('id-ID', {
  maximumFractionDigits: 0,
})

const compactNumberFormatter = new Intl.NumberFormat('id-ID', {
  maximumFractionDigits: 2,
})

const percentFormatter = new Intl.NumberFormat('id-ID', {
  maximumFractionDigits: 1,
  minimumFractionDigits: 1,
})

const lastSyncFormatter = new Intl.DateTimeFormat('id-ID', {
  hour: '2-digit',
  minute: '2-digit',
  timeZoneName: 'short',
})

const expandedStages = ref<FlowStage[]>([])

function metricForStage(stage: FlowStage) {
  return props.data.find((item) => item.stage === stage)
}

function amountValue(metric?: StageMetric) {
  return metric?.amount_usd ?? 0
}

function projectCount(metric?: StageMetric) {
  return Math.max(metric?.project_count ?? 0, 0)
}

function formatAmount(amount: number) {
  const absAmount = Math.abs(amount)

  if (absAmount >= 1_000_000_000_000) {
    return `Rp ${compactNumberFormatter.format(amount / 1_000_000_000_000)} T`
  }

  if (absAmount >= 1_000_000_000) {
    return `Rp ${compactNumberFormatter.format(amount / 1_000_000_000)} M`
  }

  if (absAmount >= 1_000_000) {
    return `Rp ${compactNumberFormatter.format(amount / 1_000_000)} Jt`
  }

  if (absAmount > 0) {
    return `Rp ${compactNumberFormatter.format(amount)}`
  }

  return 'Rp 0'
}

function formatSignedAmount(amount: number) {
  const sign = amount > 0 ? '+' : amount < 0 ? '-' : ''
  return `${sign}${formatAmount(Math.abs(amount))}`
}

function formatPercent(value: number) {
  return `${percentFormatter.format(value)}%`
}

function rowPercent(stage: FlowStage) {
  const current = projectCount(metricForStage(stage))
  const base = Math.max(projectCount(metricForStage('BB')), 1)

  return (current / base) * 100
}

function conversionPercent(stage: FlowStage) {
  const index = stageOrder.indexOf(stage)
  const nextStage = stageOrder[index + 1]

  if (!nextStage) return null

  const current = projectCount(metricForStage(stage))
  const next = projectCount(metricForStage(nextStage))

  if (current === 0) return 0

  return (next / current) * 100
}

function dropOffForStage(stage: FlowStage) {
  const index = stageOrder.indexOf(stage)
  const nextStage = stageOrder[index + 1]

  if (!nextStage) return null

  const currentMetric = metricForStage(stage)
  const nextMetric = metricForStage(nextStage)
  const projectDelta = projectCount(nextMetric) - projectCount(currentMetric)
  const amountDelta = amountValue(nextMetric) - amountValue(currentMetric)

  return {
    projectDelta,
    amountDelta,
    isBottleneck: projectDelta < 0,
  }
}

function isExpanded(stage: FlowStage) {
  return expandedStages.value.includes(stage)
}

function toggleStage(stage: FlowStage) {
  expandedStages.value = isExpanded(stage)
    ? expandedStages.value.filter((item) => item !== stage)
    : [...expandedStages.value, stage]
}

function breakdownItemsForStage(stage: FlowStage) {
  if (stage === 'BB') return props.topInstitutions.slice(0, 4)
  if (stage === 'GB') return props.topLenders.slice(0, 4)

  const relevantSummary = props.pipelineSummary
    .filter((item) => (stage === 'DK' ? item.stage === 'DK_NO_LA' : item.stage === 'LA_NOT_EFFECTIVE'))
    .map((item) => ({
      label: item.label,
      item_count: item.project_count,
      amount_usd: item.amount_usd,
    }))

  if (relevantSummary.length > 0) return relevantSummary

  return props.pipelineSummary
    .slice(0, 4)
    .map((item) => ({ label: item.label, item_count: item.project_count, amount_usd: item.amount_usd }))
}

function topProjectsForStage(stage: FlowStage) {
  const stageMap: Record<FlowStage, string[]> = {
    BB: ['BB_NO_LENDER', 'INDICATION_NO_LOI', 'LOI_NO_GB'],
    GB: ['GB_NO_DK'],
    DK: ['DK_NO_LA'],
    LA: ['LA_NOT_EFFECTIVE'],
  }
  const relevantStages = stageMap[stage]
  const bottlenecks = props.bottleneckItems.filter((item) =>
    relevantStages.includes(item.current_stage),
  )

  if (bottlenecks.length > 0) {
    return bottlenecks.slice(0, 3).map((item) => ({
      title: item.project_name,
      meta: item.institution_name || item.stage_label,
      hint: `${countFormatter.format(item.age_days)} hari`,
      amount: item.amount_usd,
    }))
  }

  return props.riskItems.slice(0, 3).map((item) => ({
    title: item.title,
    meta: item.risk_type || item.severity,
    hint: item.description || 'Perlu tindak lanjut',
    amount: item.amount_usd ?? 0,
  }))
}

function averageAgeForStage(stage: FlowStage) {
  const projects = topProjectsForStage(stage)
  const sourceItems = props.bottleneckItems.filter((item) =>
    projects.some((project) => project.title === item.project_name),
  )

  if (sourceItems.length === 0) return 0

  const totalAge = sourceItems.reduce((sum, item) => sum + item.age_days, 0)
  return Math.round(totalAge / sourceItems.length)
}

function stuckCountForStage(stage: FlowStage) {
  const threshold = stage === 'DK' ? 180 : 120

  return props.bottleneckItems.filter((item) => item.age_days > threshold).length
}

function riskLabelForStage(stage: FlowStage) {
  const accent = stageDefinitions[stage].riskAccent

  if (accent === 'red') return 'Tinggi'
  if (accent === 'amber') return 'Sedang'
  return 'Rendah'
}

function riskClassForStage(stage: FlowStage) {
  const accent = stageDefinitions[stage].riskAccent

  if (accent === 'red') return 'text-red-600'
  if (accent === 'amber') return 'text-orange-500'
  return 'text-emerald-600'
}

const flowRows = computed(() =>
  stageOrder.map((stage, index) => {
    const definition = stageDefinitions[stage]
    const metric = metricForStage(stage)
    const count = projectCount(metric)
    const percent = rowPercent(stage)
    const conversion = conversionPercent(stage)

    return {
      stage,
      title: definition.title,
      displayIndex: index + 1,
      color: definition.color,
      count,
      amount: amountValue(metric),
      percent,
      conversion,
      nextLabel: definition.nextLabel,
      amountLabel: definition.amountLabel,
      barWidth: `${Math.min(Math.max(percent, count > 0 ? 8 : 6), 100)}%`,
    }
  }),
)

const legendItems = computed(() =>
  stageOrder.map((stage) => ({
    key: stage,
    label: stageDefinitions[stage].shortTitle,
    color: stageDefinitions[stage].color,
  })),
)

const hasFlowData = computed(() => flowRows.value.some((row) => row.count > 0))

const chartAriaLabel = computed(() =>
  flowRows.value.map((row) => `${row.title}: ${row.count} proyek`).join(', '),
)

const lastSyncLabel = computed(() => lastSyncFormatter.format(new Date()))
</script>

<template>
  <DashboardChartCard
    title="Sankey Alur Proyek Perencanaan"
    :empty="!hasFlowData"
    empty-title="Tidak ada data alur proyek"
    empty-message="Ubah filter untuk melihat alur proyek."
  >
    <template #actions>
      <div class="flex flex-wrap items-center gap-x-4 gap-y-2 text-xs text-surface-700">
        <span
          v-for="item in legendItems"
          :key="item.key"
          class="inline-flex items-center gap-1.5"
        >
          <span class="h-2 w-2 rounded-sm" :style="{ backgroundColor: item.color }" />
          {{ item.label }}
        </span>
      </div>
    </template>

    <div
      role="img"
      :aria-label="chartAriaLabel"
      class="overflow-hidden rounded-lg border border-surface-100 bg-surface-0"
    >
      <div
        class="grid grid-cols-[9.75rem_minmax(26rem,1fr)_10.5rem] border-b border-surface-100 px-3 py-3 font-mono text-[0.65rem] uppercase text-slate-500"
      >
        <span>Tahap</span>
        <span>Volume &amp; Nilai</span>
        <span class="text-right">Konversi -> Tahap Berikut</span>
      </div>

      <div class="overflow-x-auto">
        <div class="min-w-[58rem]">
          <template v-for="row in flowRows" :key="row.stage">
            <button
              type="button"
              class="grid w-full grid-cols-[9.75rem_minmax(26rem,1fr)_10.5rem] items-stretch bg-surface-50/70 text-left transition hover:bg-surface-100/70 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
              :aria-expanded="isExpanded(row.stage)"
              @click="toggleStage(row.stage)"
            >
              <div class="border-r border-surface-200 px-3 py-6">
                <p class="font-mono text-[0.65rem] uppercase text-slate-500">
                  Tahap {{ String(row.displayIndex).padStart(2, '0') }}
                </p>
                <div class="mt-1 flex items-center gap-2">
                  <span
                    class="h-2 w-2 rounded-sm"
                    :style="{ backgroundColor: row.color }"
                  />
                  <span class="text-sm font-semibold text-surface-950">{{ row.title }}</span>
                </div>
              </div>

              <div class="border-r border-surface-200 px-4 py-4">
                <div class="h-16 rounded-lg bg-white">
                  <div
                    class="flex h-16 min-w-[4.75rem] items-center justify-between rounded-lg px-4 text-white shadow-sm"
                    :style="{ width: row.barWidth, backgroundColor: row.color }"
                  >
                    <div>
                      <p class="text-2xl font-bold leading-none">
                        {{ countFormatter.format(row.count) }}
                        <span class="text-xs font-medium">proyek</span>
                      </p>
                      <p class="mt-1 font-mono text-[0.7rem]">{{ formatAmount(row.amount) }}</p>
                    </div>
                    <p class="font-mono text-[0.65rem]">
                      {{ row.stage === 'BB' ? '100% of pipeline' : formatPercent(row.percent) }}
                    </p>
                  </div>
                </div>
              </div>

              <div class="flex flex-col items-end justify-center px-3 py-4 text-right">
                <template v-if="row.conversion !== null">
                  <p
                    class="font-mono text-lg font-semibold"
                    :class="row.conversion === 0 ? 'text-red-600' : 'text-surface-950'"
                  >
                    {{ formatPercent(row.conversion) }}
                  </p>
                  <p class="mt-1 font-mono text-[0.65rem] uppercase text-slate-500">
                    -> {{ row.nextLabel }}
                  </p>
                  <p
                    class="mt-1 text-xs"
                    :class="row.conversion === 0 ? 'text-red-600' : 'text-slate-700'"
                  >
                    {{ countFormatter.format(Math.max(row.count - (metricForStage(stageOrder[stageOrder.indexOf(row.stage) + 1])?.project_count ?? 0), 0)) }}
                    proyek tertahan
                  </p>
                </template>
                <template v-else>
                  <p class="h-px w-4 bg-surface-950" />
                  <p class="mt-3 font-mono text-[0.65rem] uppercase text-slate-500">Tahap final</p>
                  <p class="mt-1 text-xs text-slate-700">target Q3: 12 proyek</p>
                </template>
              </div>
            </button>

            <div v-if="isExpanded(row.stage)" class="bg-white px-3 py-3">
              <div class="grid gap-4 rounded-lg border border-surface-200 p-4 lg:grid-cols-[1.1fr_0.7fr_0.75fr]">
                <section>
                  <h3 class="font-mono text-[0.65rem] uppercase text-slate-500">
                    {{ stageDefinitions[row.stage].detailTitle }}
                  </h3>
                  <div v-if="breakdownItemsForStage(row.stage).length" class="mt-3 space-y-3">
                    <div
                      v-for="item in breakdownItemsForStage(row.stage)"
                      :key="item.id ?? item.key ?? item.label"
                    >
                      <div class="flex items-center justify-between gap-3 text-xs">
                        <span class="truncate text-surface-950">{{ item.label }}</span>
                        <span class="shrink-0 font-mono text-slate-700">
                          {{ countFormatter.format(item.item_count ?? 0) }}
                          <template v-if="item.amount_usd"> · {{ formatAmount(item.amount_usd) }}</template>
                        </span>
                      </div>
                      <div class="mt-1 h-1 rounded-full bg-surface-100">
                        <div
                          class="h-1 rounded-full"
                          :style="{
                            width: `${Math.min(Math.max(item.percentage ?? ((item.item_count ?? 0) / Math.max(row.count, 1)) * 100, 8), 100)}%`,
                            backgroundColor: row.color,
                          }"
                        />
                      </div>
                    </div>
                  </div>
                  <p v-else class="mt-3 text-xs text-slate-500">
                    Data detail belum tersedia untuk filter ini.
                  </p>
                </section>

                <section>
                  <h3 class="font-mono text-[0.65rem] uppercase text-slate-500">
                    {{ row.stage === 'LA' ? 'Pipeline LA - 90 hari' : 'Top 3 proyek' }}
                  </h3>
                  <div v-if="topProjectsForStage(row.stage).length" class="mt-3 space-y-2">
                    <div
                      v-for="project in topProjectsForStage(row.stage)"
                      :key="project.title"
                      class="rounded-md border border-surface-200 px-3 py-2"
                    >
                      <div class="flex items-start justify-between gap-3">
                        <p class="text-xs font-semibold text-surface-950">{{ project.title }}</p>
                        <span class="shrink-0 font-mono text-[0.65rem] text-slate-700">
                          {{ formatAmount(project.amount) }}
                        </span>
                      </div>
                      <p class="mt-1 truncate font-mono text-[0.65rem] text-slate-500">
                        {{ project.meta }} · {{ project.hint }}
                      </p>
                    </div>
                  </div>
                  <p v-else class="mt-3 text-xs text-slate-500">
                    Belum ada proyek tertahan yang cocok.
                  </p>
                </section>

                <section>
                  <h3 class="font-mono text-[0.65rem] uppercase text-slate-500">Aging &amp; Risk</h3>
                  <div class="mt-3 grid grid-cols-2 gap-2">
                    <div class="rounded-md border border-surface-200 p-3">
                      <p class="font-mono text-[0.65rem] uppercase text-slate-500">Avg. usia tahap</p>
                      <p class="mt-1 text-lg font-semibold text-surface-950">
                        {{ countFormatter.format(averageAgeForStage(row.stage)) }}
                        <span class="text-xs font-normal">hari</span>
                      </p>
                    </div>
                    <div class="rounded-md border border-surface-200 p-3">
                      <p class="font-mono text-[0.65rem] uppercase text-slate-500">Stuck &gt; 120 hari</p>
                      <p class="mt-1 text-lg font-semibold text-surface-950">
                        {{ countFormatter.format(stuckCountForStage(row.stage)) }}
                        <span class="text-xs font-normal">proyek</span>
                      </p>
                    </div>
                    <div class="rounded-md border border-surface-200 p-3">
                      <p class="font-mono text-[0.65rem] uppercase text-slate-500">
                        {{ row.stage === 'GB' ? 'Negotiation status' : 'Dokumen lengkap' }}
                      </p>
                      <p class="mt-1 text-lg font-semibold text-surface-950">
                        {{ formatPercent(Math.min(row.conversion ?? row.percent, 100)) }}
                      </p>
                    </div>
                    <div class="rounded-md border border-surface-200 p-3">
                      <p class="font-mono text-[0.65rem] uppercase text-slate-500">Risk score</p>
                      <p class="mt-1 text-lg font-semibold" :class="riskClassForStage(row.stage)">
                        {{ riskLabelForStage(row.stage) }}
                      </p>
                    </div>
                  </div>
                </section>
              </div>
            </div>

            <div
              v-if="dropOffForStage(row.stage)"
              class="relative bg-white py-2"
            >
              <div class="mx-[10.6rem] border-t border-dashed border-surface-200" />
              <div class="absolute left-1/2 top-1/2 -translate-x-1/2 -translate-y-1/2 rounded-full border border-surface-200 bg-white px-3 py-0.5 font-mono text-[0.65rem] uppercase text-slate-700">
                <span :class="dropOffForStage(row.stage)?.isBottleneck ? 'text-red-600' : 'text-emerald-600'">
                  {{ dropOffForStage(row.stage)?.projectDelta }}
                  proyek · {{ formatSignedAmount(dropOffForStage(row.stage)?.amountDelta ?? 0) }}
                </span>
                · {{ dropOffForStage(row.stage)?.isBottleneck ? 'drop-off' : 'gain' }}
              </div>
            </div>
          </template>
        </div>
      </div>

      <div class="flex flex-col gap-2 border-t border-surface-100 px-4 py-3 text-xs text-slate-600 md:flex-row md:items-center md:justify-between">
        <span>Sumber data: SIMRENAS · Direktorat PPLN, Bappenas</span>
        <span class="flex items-center gap-3">
          <span>Klik tahap untuk membuka detail</span>
          <span class="rounded border border-surface-200 px-2 py-0.5 font-mono text-[0.65rem]">
            Last sync {{ lastSyncLabel }}
          </span>
        </span>
      </div>
    </div>
  </DashboardChartCard>
</template>
