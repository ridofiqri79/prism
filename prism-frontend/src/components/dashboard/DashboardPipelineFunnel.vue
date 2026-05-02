<script setup lang="ts">
import { computed } from 'vue'
import Skeleton from 'primevue/skeleton'
import AnalyticsEmptyState from '@/components/dashboard/AnalyticsEmptyState.vue'
import type {
  DashboardAnalyticsPipelineFunnelItem,
  DashboardAnalyticsPipelineStage,
} from '@/types/dashboard.types'

const props = withDefaults(
  defineProps<{
    rows: DashboardAnalyticsPipelineFunnelItem[]
    loading?: boolean
    empty?: boolean
  }>(),
  {
    loading: false,
    empty: false,
  },
)

const colors = ['#2563eb', '#0f766e', '#d97706', '#7c3aed', '#475569']
const viewBoxWidth = 744
const stageHeight = 48
const stageGap = 10
const topPadding = 30
const bottomPadding = 30
const labelX = 26
const valueX = 680
const centerX = 398
const maxWidth = 420
const minWidth = 76

const numberFormatter = new Intl.NumberFormat('id-ID')
const usdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  notation: 'compact',
  maximumFractionDigits: 1,
})

const maxProjectCount = computed(() =>
  Math.max(...props.rows.map((row) => row.project_count), 1),
)
const viewBoxHeight = computed(() =>
  topPadding + props.rows.length * stageHeight + Math.max(props.rows.length - 1, 0) * stageGap + bottomPadding,
)
const funnelRows = computed(() => {
  const baseRows = props.rows.map((row, index) => {
    const ratio = row.project_count / maxProjectCount.value
    const visualRatio = row.project_count > 0 ? Math.max(Math.sqrt(ratio), 0.22) : 0.16
    const width = Math.max(minWidth, maxWidth * visualRatio)
    const y = topPadding + index * (stageHeight + stageGap)

    return {
      ...row,
      color: colors[index % colors.length],
      label: stageLabel(row.stage),
      projectText: `${formatNumber(row.project_count)} project`,
      amountText: formatUSD(row.total_loan_usd),
      shareText: `${Math.round(ratio * 100)}%`,
      width,
      y,
      isZero: row.project_count <= 0,
    }
  })

  return baseRows.map((row, index) => {
    const nextWidth = baseRows[index + 1]?.width ?? Math.max(minWidth, row.width * 0.78)
    const topLeft = centerX - row.width / 2
    const topRight = centerX + row.width / 2
    const bottomLeft = centerX - nextWidth / 2
    const bottomRight = centerX + nextWidth / 2

    return {
      ...row,
      path: [
        `M ${topLeft} ${row.y}`,
        `L ${topRight} ${row.y}`,
        `L ${bottomRight} ${row.y + stageHeight}`,
        `L ${bottomLeft} ${row.y + stageHeight}`,
        'Z',
      ].join(' '),
      midY: row.y + stageHeight / 2,
    }
  })
})

function stageLabel(stage: DashboardAnalyticsPipelineStage) {
  const labels: Record<DashboardAnalyticsPipelineStage, string> = {
    BB: 'Blue Book',
    GB: 'Green Book',
    DK: 'Daftar Kegiatan',
    LA: 'Loan Agreement',
    Monitoring: 'Monitoring',
  }

  return labels[stage]
}

function formatNumber(value: number) {
  return numberFormatter.format(value)
}

function formatUSD(value: number) {
  if (value <= 0) return 'USD 0'
  return usdFormatter.format(value)
}
</script>

<template>
  <section class="overflow-hidden rounded-lg border border-surface-200 bg-white">
    <div class="border-b border-surface-200 p-4">
      <h2 class="text-base font-semibold text-surface-950">Funnel Pipeline</h2>
      <p class="mt-1 text-sm text-surface-500">
        Total project aktual per stage; project yang sudah lanjut tetap dihitung di stage asal.
      </p>
    </div>

    <div v-if="loading" class="space-y-3 p-4">
      <Skeleton height="2rem" />
      <Skeleton height="14rem" />
    </div>

    <div v-else-if="empty" class="p-4">
      <AnalyticsEmptyState
        title="Tidak ada data"
        description="Data funnel kosong untuk filter aktif."
      />
    </div>

    <div v-else class="p-4">
      <div class="pipeline-funnel overflow-x-auto rounded-lg bg-surface-50 px-3 py-4">
        <svg
          class="min-w-[44rem] w-full"
          :viewBox="`0 0 ${viewBoxWidth} ${viewBoxHeight}`"
          role="img"
          aria-label="Diagram funnel pipeline project"
        >
          <defs>
            <filter id="pipeline-funnel-shadow" x="-8%" y="-32%" width="116%" height="164%">
              <feDropShadow dx="0" dy="8" stdDeviation="6" flood-color="#0f172a" flood-opacity="0.14" />
            </filter>
          </defs>

          <g v-for="(row, index) in funnelRows" :key="row.stage">
            <title>
              {{ row.label }} - {{ row.projectText }} - Nilai pinjaman {{ row.amountText }}
            </title>

            <text :x="labelX" :y="row.midY - 5" class="pipeline-funnel__stage">
              {{ row.label }}
            </text>
            <text :x="labelX" :y="row.midY + 13" class="pipeline-funnel__amount">
              {{ row.amountText }}
            </text>

            <path
              :d="row.path"
              :fill="row.color"
              :class="{ 'pipeline-funnel__segment--empty': row.isZero }"
              class="pipeline-funnel__segment"
              filter="url(#pipeline-funnel-shadow)"
            />
            <text :x="centerX" :y="row.midY - 1" class="pipeline-funnel__value">
              {{ row.projectText }}
            </text>
            <text :x="valueX" :y="row.midY - 5" class="pipeline-funnel__count">
              {{ row.shareText }}
            </text>
            <text :x="valueX" :y="row.midY + 13" class="pipeline-funnel__count-label">
              dari terbesar
            </text>
          </g>
        </svg>
      </div>
    </div>
  </section>
</template>

<style scoped>
.pipeline-funnel {
  scrollbar-width: thin;
}

.pipeline-funnel__segment {
  stroke: #ffffff;
  stroke-linejoin: round;
  stroke-width: 3px;
}

.pipeline-funnel__segment--empty {
  opacity: 0.5;
}

.pipeline-funnel__stage {
  fill: #0f172a;
  font-size: 14px;
  font-weight: 700;
}

.pipeline-funnel__amount {
  fill: #64748b;
  font-size: 12px;
  font-weight: 500;
}

.pipeline-funnel__value {
  dominant-baseline: middle;
  fill: #ffffff;
  font-size: 14px;
  font-weight: 800;
  paint-order: stroke;
  stroke: rgb(15 23 42 / 0.24);
  stroke-width: 2px;
  text-anchor: middle;
}

.pipeline-funnel__count {
  fill: #0f172a;
  font-size: 16px;
  font-weight: 800;
  text-anchor: end;
}

.pipeline-funnel__count-label {
  fill: #64748b;
  font-size: 11px;
  font-weight: 600;
  text-anchor: end;
}
</style>
