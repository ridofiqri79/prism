<script setup lang="ts">
import { computed } from 'vue'
import { SankeyChart } from 'echarts/charts'
import { TooltipComponent } from 'echarts/components'
import { use } from 'echarts/core'
import { LabelLayout } from 'echarts/features'
import { CanvasRenderer } from 'echarts/renderers'
import VChart from 'vue-echarts'
import CurrencyDisplay from '@/components/common/CurrencyDisplay.vue'
import type {
  DKProjectJourney,
  GBProjectJourney,
  JourneyFlowLink,
  JourneyFlowNode,
  JourneyFlowStage,
  JourneyResponse,
  JourneyStageState,
  LAJourney,
  MonitoringSummaryItem,
} from '@/types/journey.types'

use([SankeyChart, TooltipComponent, LabelLayout, CanvasRenderer])

const props = defineProps<{
  journey: JourneyResponse
}>()

type FlowBuildResult = {
  nodes: JourneyFlowNode[]
  links: JourneyFlowLink[]
}

type FlowTooltipParam = {
  dataType?: string
  data?: unknown
  marker?: string
  name?: string
}

type LinkOptions = {
  pending?: boolean
}

const usdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  notation: 'compact',
  maximumFractionDigits: 2,
})

const stagePalette: Record<
  JourneyFlowStage,
  {
    filled: string
    pending: string
    label: string
  }
> = {
  'blue-book': {
    filled: '#2563eb',
    pending: '#bfdbfe',
    label: '#0f172a',
  },
  'green-book': {
    filled: '#16a34a',
    pending: '#bbf7d0',
    label: '#0f172a',
  },
  'daftar-kegiatan': {
    filled: '#f97316',
    pending: '#fed7aa',
    label: '#0f172a',
  },
  'loan-agreement': {
    filled: '#64748b',
    pending: '#cbd5e1',
    label: '#334155',
  },
  monitoring: {
    filled: '#14b8a6',
    pending: '#cbd5e1',
    label: '#334155',
  },
}

const flow = computed(() => buildFlow(props.journey))
const flowNodes = computed(() => flow.value.nodes)
const flowLinks = computed(() => flow.value.links)
const chartHeight = computed(() => `${Math.min(760, Math.max(380, flowNodes.value.length * 34))}px`)

const totalFlowUsd = computed(() =>
  props.journey.gb_projects.reduce((sum, project) => sum + latestLoanUsdForGreenBook(project), 0),
)

const chartOption = computed(() => ({
  tooltip: {
    trigger: 'item',
    confine: true,
    formatter: (param: unknown) => {
      if (!isTooltipParam(param)) return ''
      if (isFlowLink(param.data)) {
        const value =
          param.data.rawValue > 0
            ? `Loan USD: ${usdFormatter.format(param.data.rawValue)}`
            : 'Loan USD belum tersedia'

        return [
          `<strong>${escapeHtml(param.data.label)}</strong>`,
          `${escapeHtml(param.data.source)} -> ${escapeHtml(param.data.target)}`,
          value,
        ].join('<br/>')
      }

      return escapeHtml(param.name || '')
    },
  },
  series: [
    {
      type: 'sankey',
      data: flowNodes.value,
      links: flowLinks.value,
      left: 16,
      right: 180,
      top: 20,
      bottom: 20,
      nodeWidth: 14,
      nodeGap: 14,
      draggable: false,
      layoutIterations: 48,
      label: {
        color: '#334155',
        fontSize: 11,
      },
      itemStyle: {
        borderWidth: 0,
        borderRadius: 3,
      },
      lineStyle: {
        color: 'gradient',
        curveness: 0.46,
        opacity: 0.32,
      },
      emphasis: {
        focus: 'adjacency',
        lineStyle: {
          opacity: 0.5,
        },
      },
    },
  ],
}))

function buildFlow(journey: JourneyResponse): FlowBuildResult {
  const nodes = new Map<string, JourneyFlowNode>()
  const links: JourneyFlowLink[] = []
  const blueBookNode = `Blue Book\n${journey.bb_project.bb_code}`

  ensureNode(
    nodes,
    blueBookNode,
    journey.bb_project.has_newer_revision ? 'warning' : 'completed',
    'blue-book',
  )

  if (journey.gb_projects.length === 0) {
    const pendingGreenBookNode = 'Green Book\nBelum ada'
    ensureNode(nodes, pendingGreenBookNode, 'pending', 'green-book')
    links.push(makeLink(blueBookNode, pendingGreenBookNode, 0, 'Blue Book ke Green Book'))
    return { nodes: Array.from(nodes.values()), links }
  }

  for (const greenBookProject of journey.gb_projects) {
    const greenBookNode = `Green Book\n${greenBookProject.gb_code}`
    const greenBookAmount = latestLoanUsdForGreenBook(greenBookProject)
    ensureNode(
      nodes,
      greenBookNode,
      greenBookProject.has_newer_revision ? 'warning' : 'completed',
      'green-book',
    )
    links.push(makeLink(blueBookNode, greenBookNode, greenBookAmount, 'Blue Book ke Green Book'))

    if (greenBookProject.dk_projects.length === 0) {
      const pendingDkNode = `Daftar Kegiatan\nBelum ada untuk ${greenBookProject.gb_code}`
      ensureNode(nodes, pendingDkNode, 'pending', 'daftar-kegiatan')
      links.push(
        makeLink(greenBookNode, pendingDkNode, greenBookAmount, 'Green Book ke Daftar Kegiatan', {
          pending: true,
        }),
      )
      continue
    }

    const dkAmounts = distributeFlow(
      greenBookAmount,
      greenBookProject.dk_projects,
      flowWeightForDk,
    )

    for (const [dkIndex, dkProject] of greenBookProject.dk_projects.entries()) {
      const dkNode = `Daftar Kegiatan\n${shortLabel(dkLabel(dkProject), 42)}`
      const dkAmount = dkAmounts[dkIndex] ?? 0
      ensureNode(nodes, dkNode, 'completed', 'daftar-kegiatan')
      links.push(makeLink(greenBookNode, dkNode, dkAmount, 'Green Book ke Daftar Kegiatan'))

      const loanAgreements = loanAgreementsForDk(dkProject)
      if (loanAgreements.length === 0) {
        const pendingLoanAgreementNode = `Loan Agreement\nBelum ada untuk ${shortLabel(
          dkLabel(dkProject),
          28,
        )}`
        ensureNode(nodes, pendingLoanAgreementNode, 'pending', 'loan-agreement')
        links.push(
          makeLink(dkNode, pendingLoanAgreementNode, dkAmount, 'Daftar Kegiatan ke Loan Agreement', {
            pending: true,
          }),
        )
        continue
      }

      const loanAgreementAmounts = distributeFlow(
        dkAmount,
        loanAgreements,
        flowWeightForLoanAgreement,
      )
      for (const [loanAgreementIndex, loanAgreement] of loanAgreements.entries()) {
        appendLoanAgreementFlow(
          nodes,
          links,
          dkNode,
          loanAgreementAmounts[loanAgreementIndex] ?? 0,
          loanAgreement,
        )
      }
    }
  }

  return { nodes: Array.from(nodes.values()), links }
}

function appendLoanAgreementFlow(
  nodes: Map<string, JourneyFlowNode>,
  links: JourneyFlowLink[],
  dkNode: string,
  flowAmount: number,
  loanAgreement: LAJourney,
) {
  const loanAgreementNode = `Loan Agreement\n${loanAgreement.loan_code}`
  ensureNode(
    nodes,
    loanAgreementNode,
    loanAgreement.is_extended ? 'extended' : 'completed',
    'loan-agreement',
  )
  links.push(
    makeLink(dkNode, loanAgreementNode, flowAmount, 'Daftar Kegiatan ke Loan Agreement'),
  )

  if (loanAgreement.monitoring.length === 0) {
    const pendingMonitoringNode = `Monitoring Disbursement\nBelum ada untuk ${loanAgreement.loan_code}`
    ensureNode(nodes, pendingMonitoringNode, 'pending', 'monitoring')
    links.push(
      makeLink(loanAgreementNode, pendingMonitoringNode, flowAmount, 'Loan Agreement ke Monitoring', {
        pending: true,
      }),
    )
    return
  }

  const monitoringAmounts = distributeFlow(
    flowAmount,
    loanAgreement.monitoring,
    flowWeightForMonitoring,
  )
  for (const [monitoringIndex, monitoring] of loanAgreement.monitoring.entries()) {
    const monitoringNode = `Monitoring\n${monitoring.quarter} ${monitoring.budget_year}`
    ensureNode(
      nodes,
      monitoringNode,
      monitoring.absorption_pct >= 80 ? 'completed' : 'warning',
      'monitoring',
    )
    links.push(
      makeLink(
        loanAgreementNode,
        monitoringNode,
        monitoringAmounts[monitoringIndex] ?? 0,
        `${monitoring.quarter} ${monitoring.budget_year} - ${monitoring.absorption_pct.toFixed(
          1,
        )}% penyerapan`,
      ),
    )
  }
}

function ensureNode(
  nodes: Map<string, JourneyFlowNode>,
  name: string,
  state: JourneyStageState,
  stage: JourneyFlowStage,
) {
  if (nodes.has(name)) return
  nodes.set(name, {
    name,
    itemStyle: {
      color: nodeColor(stage, state),
      borderColor: attentionBorderColor(state),
      borderWidth: state === 'warning' || state === 'extended' ? 2 : 0,
    },
    label: {
      color: labelColor(stage, state),
      fontWeight: state === 'pending' ? 400 : 600,
    },
  })
}

function makeLink(
  source: string,
  target: string,
  rawValue: number,
  label: string,
  options: LinkOptions = {},
): JourneyFlowLink {
  return {
    source,
    target,
    rawValue,
    value: normalizeFlowValue(rawValue),
    label,
    lineStyle: {
      opacity: options.pending ? 0.16 : rawValue > 0 ? 0.34 : 0.14,
    },
  }
}

function normalizeFlowValue(value: number) {
  if (!Number.isFinite(value) || value <= 0) return 1
  return Math.max(1, value / 1_000_000)
}

function latestLoanUsdForGreenBook(project: GBProjectJourney) {
  if (typeof project.latest_loan_usd === 'number' && Number.isFinite(project.latest_loan_usd)) {
    return Math.max(0, project.latest_loan_usd)
  }
  return currentLoanUsdForGreenBook(project)
}

function currentLoanUsdForGreenBook(project: GBProjectJourney) {
  return project.funding_sources.reduce((sum, source) => sum + safeFlowWeight(source.loan_usd), 0)
}

function flowWeightForDk(project: DKProjectJourney) {
  return loanAgreementsForDk(project).reduce(
    (sum, loanAgreement) => sum + flowWeightForLoanAgreement(loanAgreement),
    0,
  )
}

function flowWeightForLoanAgreement(loanAgreement: LAJourney) {
  return safeFlowWeight(loanAgreement.allocation_usd)
}

function flowWeightForMonitoring(monitoring: MonitoringSummaryItem) {
  return safeFlowWeight(monitoring.realized_usd > 0 ? monitoring.realized_usd : monitoring.planned_usd)
}

function distributeFlow<T>(total: number, items: T[], weightFor: (item: T) => number) {
  const safeTotal = safeFlowWeight(total)
  if (items.length === 0) return []
  if (safeTotal <= 0) return items.map(() => 0)

  const weights = items.map((item) => safeFlowWeight(weightFor(item)))
  const totalWeight = weights.reduce((sum, weight) => sum + weight, 0)
  if (totalWeight <= 0) return items.map(() => safeTotal / items.length)

  return weights.map((weight) => (safeTotal * weight) / totalWeight)
}

function safeFlowWeight(value: number | null | undefined) {
  return typeof value === 'number' && Number.isFinite(value) && value > 0 ? value : 0
}

function loanAgreementsForDk(project: DKProjectJourney) {
  return project.loan_agreements ?? []
}

function dkLabel(project: DKProjectJourney) {
  return (
    project.daftar_kegiatan?.letter_number ||
    project.daftar_kegiatan?.subject ||
    project.project_name ||
    project.id
  )
}

function shortLabel(value: string, maxLength: number) {
  if (value.length <= maxLength) return value
  return `${value.slice(0, Math.max(0, maxLength - 1))}...`
}

function nodeColor(stage: JourneyFlowStage, state: JourneyStageState) {
  if (state === 'pending') return stagePalette[stage].pending
  return stagePalette[stage].filled
}

function labelColor(stage: JourneyFlowStage, state: JourneyStageState) {
  if (state === 'pending') return '#64748b'
  return stagePalette[stage].label
}

function attentionBorderColor(state: JourneyStageState) {
  if (state === 'extended') return '#f59e0b'
  if (state === 'warning') return '#fb923c'
  return 'transparent'
}

function isTooltipParam(value: unknown): value is FlowTooltipParam {
  return typeof value === 'object' && value !== null
}

function isFlowLink(value: unknown): value is JourneyFlowLink {
  if (typeof value !== 'object' || value === null) return false
  const record = value as Record<string, unknown>
  return (
    typeof record.source === 'string' &&
    typeof record.target === 'string' &&
    typeof record.rawValue === 'number' &&
    typeof record.label === 'string'
  )
}

function escapeHtml(value: string) {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;')
}
</script>

<template>
  <section class="rounded-lg border border-surface-200 bg-white p-4">
    <div class="mb-4 flex flex-wrap items-start justify-between gap-3">
      <div>
        <h2 class="text-base font-semibold text-surface-950">Alur Visual</h2>
        <p class="text-sm text-surface-500">
          Warna node mengikuti tahap, sementara ketebalan garis mengikuti Loan USD Green Book terbaru.
        </p>
      </div>
      <div class="flex flex-wrap gap-2 text-xs">
        <span
          class="inline-flex items-center gap-1 rounded-full bg-blue-50 px-2 py-1 text-blue-700"
        >
          <span class="h-2 w-2 rounded-full bg-blue-600" />
          Blue Book
        </span>
        <span
          class="inline-flex items-center gap-1 rounded-full bg-green-50 px-2 py-1 text-green-700"
        >
          <span class="h-2 w-2 rounded-full bg-green-600" />
          Green Book
        </span>
        <span
          class="inline-flex items-center gap-1 rounded-full bg-orange-50 px-2 py-1 text-orange-700"
        >
          <span class="h-2 w-2 rounded-full bg-orange-500" />
          Daftar Kegiatan
        </span>
        <span
          class="inline-flex items-center gap-1 rounded-full bg-amber-50 px-2 py-1 text-amber-700"
        >
          <span class="h-2 w-2 rounded-full border-2 border-amber-500 bg-white" />
          Perlu perhatian
        </span>
      </div>
    </div>

    <div class="mb-3 grid gap-3 text-sm md:grid-cols-3">
      <div class="rounded-lg border border-surface-100 bg-surface-50 p-3">
        <p class="text-xs font-medium uppercase tracking-wide text-surface-500">Node</p>
        <p class="mt-1 text-lg font-semibold text-surface-900">{{ flowNodes.length }}</p>
      </div>
      <div class="rounded-lg border border-surface-100 bg-surface-50 p-3">
        <p class="text-xs font-medium uppercase tracking-wide text-surface-500">Relasi</p>
        <p class="mt-1 text-lg font-semibold text-surface-900">{{ flowLinks.length }}</p>
      </div>
      <div class="rounded-lg border border-surface-100 bg-surface-50 p-3">
        <p class="text-xs font-medium uppercase tracking-wide text-surface-500">Loan USD Green Book</p>
        <p class="mt-1 text-lg font-semibold text-surface-900">
          <CurrencyDisplay :amount="totalFlowUsd" currency="USD" compact />
        </p>
      </div>
    </div>

    <div class="overflow-x-auto rounded-lg border border-surface-100 bg-surface-50">
      <VChart
        :option="chartOption"
        autoresize
        class="min-w-[58rem] w-full"
        :style="{ height: chartHeight }"
      />
    </div>
  </section>
</template>
