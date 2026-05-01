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
  JourneyResponse,
  JourneyStageState,
  LAJourney,
} from '@/types/dashboard.types'

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

const usdFormatter = new Intl.NumberFormat('en-US', {
  style: 'currency',
  currency: 'USD',
  notation: 'compact',
  maximumFractionDigits: 2,
})

const flow = computed(() => buildFlow(props.journey))
const flowNodes = computed(() => flow.value.nodes)
const flowLinks = computed(() => flow.value.links)
const chartHeight = computed(() => `${Math.min(760, Math.max(380, flowNodes.value.length * 34))}px`)

const totalFlowUsd = computed(() =>
  props.journey.gb_projects.reduce((sum, project) => sum + fundingTotalForGreenBook(project), 0),
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
            ? `Nilai: ${usdFormatter.format(param.data.rawValue)}`
            : 'Nilai USD belum tersedia'

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
        opacity: 0.24,
      },
      emphasis: {
        focus: 'adjacency',
        lineStyle: {
          opacity: 0.46,
        },
      },
    },
  ],
}))

function buildFlow(journey: JourneyResponse): FlowBuildResult {
  const nodes = new Map<string, JourneyFlowNode>()
  const links: JourneyFlowLink[] = []
  const blueBookNode = `Blue Book\n${journey.bb_project.bb_code}`

  ensureNode(nodes, blueBookNode, journey.bb_project.has_newer_revision ? 'warning' : 'completed')

  if (journey.gb_projects.length === 0) {
    const pendingGreenBookNode = 'Green Book\nBelum ada'
    ensureNode(nodes, pendingGreenBookNode, 'pending')
    links.push(makeLink(blueBookNode, pendingGreenBookNode, 0, 'Blue Book ke Green Book'))
    return { nodes: Array.from(nodes.values()), links }
  }

  for (const greenBookProject of journey.gb_projects) {
    const greenBookNode = `Green Book\n${greenBookProject.gb_code}`
    const greenBookAmount = fundingTotalForGreenBook(greenBookProject)
    ensureNode(nodes, greenBookNode, greenBookProject.has_newer_revision ? 'warning' : 'completed')
    links.push(makeLink(blueBookNode, greenBookNode, greenBookAmount, 'Blue Book ke Green Book'))

    if (greenBookProject.dk_projects.length === 0) {
      const pendingDkNode = `Daftar Kegiatan\nBelum ada untuk ${greenBookProject.gb_code}`
      ensureNode(nodes, pendingDkNode, 'pending')
      links.push(makeLink(greenBookNode, pendingDkNode, 0, 'Green Book ke Daftar Kegiatan'))
      continue
    }

    for (const dkProject of greenBookProject.dk_projects) {
      const dkNode = `Daftar Kegiatan\n${shortLabel(dkLabel(dkProject), 42)}`
      const dkAmount = amountForDk(greenBookProject, dkProject)
      ensureNode(nodes, dkNode, 'completed')
      links.push(makeLink(greenBookNode, dkNode, dkAmount, 'Green Book ke Daftar Kegiatan'))

      if (!dkProject.loan_agreement) {
        const pendingLoanAgreementNode = `Loan Agreement\nBelum ada untuk ${shortLabel(
          dkLabel(dkProject),
          28,
        )}`
        ensureNode(nodes, pendingLoanAgreementNode, 'pending')
        links.push(
          makeLink(dkNode, pendingLoanAgreementNode, 0, 'Daftar Kegiatan ke Loan Agreement'),
        )
        continue
      }

      appendLoanAgreementFlow(nodes, links, dkNode, dkAmount, dkProject.loan_agreement)
    }
  }

  return { nodes: Array.from(nodes.values()), links }
}

function appendLoanAgreementFlow(
  nodes: Map<string, JourneyFlowNode>,
  links: JourneyFlowLink[],
  dkNode: string,
  fallbackAmount: number,
  loanAgreement: LAJourney,
) {
  const loanAgreementNode = `Loan Agreement\n${loanAgreement.loan_code}`
  const loanAgreementAmount = loanAgreement.amount_usd ?? fallbackAmount
  ensureNode(nodes, loanAgreementNode, loanAgreement.is_extended ? 'extended' : 'completed')
  links.push(
    makeLink(dkNode, loanAgreementNode, loanAgreementAmount, 'Daftar Kegiatan ke Loan Agreement'),
  )

  if (loanAgreement.monitoring.length === 0) {
    const pendingMonitoringNode = `Monitoring Disbursement\nBelum ada untuk ${loanAgreement.loan_code}`
    ensureNode(nodes, pendingMonitoringNode, 'pending')
    links.push(
      makeLink(loanAgreementNode, pendingMonitoringNode, 0, 'Loan Agreement ke Monitoring'),
    )
    return
  }

  for (const monitoring of loanAgreement.monitoring) {
    const monitoringNode = `Monitoring\n${monitoring.quarter} ${monitoring.budget_year}`
    const monitoringAmount =
      monitoring.realized_usd > 0 ? monitoring.realized_usd : monitoring.planned_usd
    ensureNode(nodes, monitoringNode, monitoring.absorption_pct >= 80 ? 'completed' : 'warning')
    links.push(
      makeLink(
        loanAgreementNode,
        monitoringNode,
        monitoringAmount,
        `${monitoring.quarter} ${monitoring.budget_year} - ${monitoring.absorption_pct.toFixed(
          1,
        )}% penyerapan`,
      ),
    )
  }
}

function ensureNode(nodes: Map<string, JourneyFlowNode>, name: string, state: JourneyStageState) {
  if (nodes.has(name)) return
  nodes.set(name, {
    name,
    itemStyle: {
      color: nodeColor(state),
    },
    label: {
      color: labelColor(state),
      fontWeight: state === 'pending' ? 400 : 600,
    },
  })
}

function makeLink(
  source: string,
  target: string,
  rawValue: number,
  label: string,
): JourneyFlowLink {
  return {
    source,
    target,
    rawValue,
    value: normalizeFlowValue(rawValue),
    label,
    lineStyle: {
      opacity: rawValue > 0 ? 0.28 : 0.14,
    },
  }
}

function normalizeFlowValue(value: number) {
  if (!Number.isFinite(value) || value <= 0) return 1
  return Math.max(1, value / 1_000_000)
}

function fundingTotalForGreenBook(project: GBProjectJourney) {
  return project.funding_sources.reduce(
    (sum, source) =>
      sum + (source.loan_usd ?? 0) + (source.grant_usd ?? 0) + (source.local_usd ?? 0),
    0,
  )
}

function amountForDk(greenBookProject: GBProjectJourney, dkProject: DKProjectJourney) {
  if ((dkProject.loan_agreement?.amount_usd ?? 0) > 0) {
    return dkProject.loan_agreement?.amount_usd ?? 0
  }

  const greenBookAmount = fundingTotalForGreenBook(greenBookProject)
  if (greenBookAmount <= 0) return 0
  return greenBookAmount / Math.max(1, greenBookProject.dk_projects.length)
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

function nodeColor(state: JourneyStageState) {
  if (state === 'extended') return '#f59e0b'
  if (state === 'warning') return '#f97316'
  if (state === 'completed') return '#1fa06f'
  return '#cbd5e1'
}

function labelColor(state: JourneyStageState) {
  if (state === 'pending') return '#64748b'
  return '#0f172a'
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
          Ketebalan garis mengikuti nilai USD bila tersedia, lalu tetap menampilkan gap tahap.
        </p>
      </div>
      <div class="flex flex-wrap gap-2 text-xs">
        <span
          class="inline-flex items-center gap-1 rounded-full bg-emerald-50 px-2 py-1 text-emerald-700"
        >
          <span class="h-2 w-2 rounded-full bg-emerald-500" />
          Lengkap
        </span>
        <span
          class="inline-flex items-center gap-1 rounded-full bg-orange-50 px-2 py-1 text-orange-700"
        >
          <span class="h-2 w-2 rounded-full bg-orange-500" />
          Revisi/gap perhatian
        </span>
        <span
          class="inline-flex items-center gap-1 rounded-full bg-surface-100 px-2 py-1 text-surface-600"
        >
          <span class="h-2 w-2 rounded-full bg-surface-300" />
          Belum ada
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
        <p class="text-xs font-medium uppercase tracking-wide text-surface-500">Nilai Green Book</p>
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
