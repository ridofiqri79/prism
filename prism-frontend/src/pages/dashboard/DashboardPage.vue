<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Button from 'primevue/button'
import SelectButton from 'primevue/selectbutton'
import PageHeader from '@/components/common/PageHeader.vue'
import AnalyticsBreakdownTable from '@/components/dashboard/AnalyticsBreakdownTable.vue'
import AnalyticsChartPanel from '@/components/dashboard/AnalyticsChartPanel.vue'
import AnalyticsEmptyState from '@/components/dashboard/AnalyticsEmptyState.vue'
import AnalyticsMatrixTable from '@/components/dashboard/AnalyticsMatrixTable.vue'
import AnalyticsMetricGrid from '@/components/dashboard/AnalyticsMetricGrid.vue'
import DashboardAnalyticsFilterBar from '@/components/dashboard/DashboardAnalyticsFilterBar.vue'
import {
  useDashboardAnalytics,
  type DashboardAnalyticsSectionKey,
} from '@/composables/useDashboardAnalytics'
import { useMasterStore } from '@/stores/master.store'
import type {
  AnalyticsBreakdownTableColumn,
  AnalyticsBreakdownTableRow,
  AnalyticsMoneyMetric,
  DashboardAbsorptionRankedItem,
  DashboardAnalyticsPipelineStage,
  DashboardDrilldownQuery,
  DashboardLoanAgreementRiskItem,
  DashboardLenderProportionStage,
  DashboardYearlyItem,
} from '@/types/dashboard.types'

type DashboardAnalyticsTab =
  | 'portfolio'
  | 'institutions'
  | 'lenders'
  | 'absorption'
  | 'yearly'
  | 'risks'
type AbsorptionLevel = 'institution' | 'project' | 'lender'
type TooltipParam = {
  dataIndex: number
  seriesName: string
  value: number
  marker: string
  name: string
  axisValue: string
}

const masterStore = useMasterStore()
const analytics = useDashboardAnalytics()
const activeTab = ref<DashboardAnalyticsTab>('portfolio')
const absorptionLevel = ref<AbsorptionLevel>('institution')
const matrixTopN = ref(10)

const tabs: Array<{
  key: DashboardAnalyticsTab
  label: string
  sections: DashboardAnalyticsSectionKey[]
}> = [
  { key: 'portfolio', label: 'Portfolio', sections: ['overview'] },
  { key: 'institutions', label: 'Kementerian/Lembaga', sections: ['institutions'] },
  { key: 'lenders', label: 'Lender', sections: ['lenders', 'lenderProportion'] },
  { key: 'absorption', label: 'Penyerapan', sections: ['absorption'] },
  { key: 'yearly', label: 'Tahunan', sections: ['yearly'] },
  { key: 'risks', label: 'Risiko & Data Quality', sections: ['risks'] },
]
const absorptionLevelOptions: Array<{ label: string; value: AbsorptionLevel }> = [
  { label: 'Kementerian/Lembaga', value: 'institution' },
  { label: 'Project', value: 'project' },
  { label: 'Lender', value: 'lender' },
]

const portfolioColumns: AnalyticsBreakdownTableColumn[] = [
  { key: 'stage', label: 'Stage' },
  { key: 'project_count', label: 'Project', kind: 'number', align: 'right' },
  { key: 'total_loan_usd', label: 'Nilai Pinjaman USD', kind: 'currency', align: 'right' },
  { key: 'drilldown', label: 'Aksi', kind: 'drilldown', align: 'center' },
]
const institutionColumns: AnalyticsBreakdownTableColumn[] = [
  { key: 'institution', label: 'Kementerian/Lembaga' },
  { key: 'project_count', label: 'Project', kind: 'number', align: 'right' },
  { key: 'assignment_count', label: 'Assignment', kind: 'number', align: 'right' },
  { key: 'loan_agreement_count', label: 'Loan Agreement', kind: 'number', align: 'right' },
  { key: 'monitoring_count', label: 'Monitoring', kind: 'number', align: 'right' },
  { key: 'agreement_amount_usd', label: 'Nilai Pinjaman USD', kind: 'currency', align: 'right' },
  { key: 'planned_usd', label: 'Rencana USD', kind: 'currency', align: 'right' },
  { key: 'realized_usd', label: 'Realisasi USD', kind: 'currency', align: 'right' },
  { key: 'absorption_pct', label: 'Penyerapan', kind: 'absorption', align: 'left' },
  { key: 'BB', label: 'Blue Book', kind: 'number', align: 'right' },
  { key: 'GB', label: 'Green Book', kind: 'number', align: 'right' },
  { key: 'DK', label: 'Daftar Kegiatan', kind: 'number', align: 'right' },
  { key: 'LA', label: 'Loan Agreement Stage', kind: 'number', align: 'right' },
  { key: 'Monitoring', label: 'Monitoring Stage', kind: 'number', align: 'right' },
  { key: 'drilldown', label: 'Aksi', kind: 'drilldown', align: 'center' },
]
const lenderColumns: AnalyticsBreakdownTableColumn[] = [
  { key: 'lender', label: 'Lender' },
  { key: 'type', label: 'Tipe', kind: 'badge', align: 'center' },
  { key: 'loan_agreement_count', label: 'Loan Agreement', kind: 'number', align: 'right' },
  { key: 'project_count', label: 'Project Coverage', kind: 'number', align: 'right' },
  { key: 'institution_count', label: 'Kementerian/Lembaga Coverage', kind: 'number', align: 'right' },
  { key: 'monitoring_count', label: 'Monitoring', kind: 'number', align: 'right' },
  { key: 'agreement_amount_usd', label: 'Nilai Pinjaman USD', kind: 'currency', align: 'right' },
  { key: 'planned_usd', label: 'Rencana USD', kind: 'currency', align: 'right' },
  { key: 'realized_usd', label: 'Realisasi USD', kind: 'currency', align: 'right' },
  { key: 'absorption_pct', label: 'Penyerapan', kind: 'absorption', align: 'left' },
  { key: 'drilldown', label: 'Aksi', kind: 'drilldown', align: 'center' },
]
const lenderProportionColumns: AnalyticsBreakdownTableColumn[] = [
  { key: 'stage', label: 'Stage' },
  { key: 'type', label: 'Tipe', kind: 'badge', align: 'center' },
  { key: 'project_count', label: 'Project', kind: 'number', align: 'right' },
  { key: 'lender_count', label: 'Lender', kind: 'number', align: 'right' },
  { key: 'amount_usd', label: 'Nilai USD', kind: 'currency', align: 'right' },
  { key: 'share_pct', label: 'Proporsi', kind: 'percent', align: 'right' },
  { key: 'drilldown', label: 'Aksi', kind: 'drilldown', align: 'center' },
]
const absorptionColumns: AnalyticsBreakdownTableColumn[] = [
  { key: 'rank', label: 'Rank', kind: 'number', align: 'right' },
  { key: 'name', label: 'Nama' },
  { key: 'planned_usd', label: 'Rencana USD', kind: 'currency', align: 'right' },
  { key: 'realized_usd', label: 'Realisasi USD', kind: 'currency', align: 'right' },
  { key: 'variance_usd', label: 'Variance USD', kind: 'currency', align: 'right' },
  { key: 'absorption_pct', label: 'Penyerapan', kind: 'absorption', align: 'left' },
  { key: 'status', label: 'Status', kind: 'badge', align: 'center' },
  { key: 'drilldown', label: 'Aksi', kind: 'drilldown', align: 'center' },
]
const yearlyColumns: AnalyticsBreakdownTableColumn[] = [
  { key: 'period', label: 'Periode' },
  { key: 'planned_usd', label: 'Rencana USD', kind: 'currency', align: 'right' },
  { key: 'realized_usd', label: 'Realisasi USD', kind: 'currency', align: 'right' },
  { key: 'absorption_pct', label: 'Penyerapan', kind: 'absorption', align: 'left' },
  { key: 'loan_agreement_count', label: 'Loan Agreement', kind: 'number', align: 'right' },
  { key: 'project_count', label: 'Project Aktif', kind: 'number', align: 'right' },
  { key: 'drilldown', label: 'Aksi', kind: 'drilldown', align: 'center' },
]
const riskWatchlistColumns: AnalyticsBreakdownTableColumn[] = [
  { key: 'category', label: 'Kategori' },
  { key: 'project', label: 'Project' },
  { key: 'institution', label: 'Kementerian/Lembaga' },
  { key: 'lender', label: 'Lender' },
  { key: 'amount_usd', label: 'Nilai USD', kind: 'currency', align: 'right' },
  { key: 'absorption_pct', label: 'Penyerapan', kind: 'absorption', align: 'left' },
  { key: 'severity', label: 'Status', kind: 'badge', align: 'center' },
  { key: 'drilldown', label: 'Aksi', kind: 'drilldown', align: 'center' },
]
const pipelineBottleneckColumns: AnalyticsBreakdownTableColumn[] = [
  { key: 'label', label: 'Kategori' },
  { key: 'stage', label: 'Stage' },
  { key: 'count', label: 'Jumlah', kind: 'number', align: 'right' },
  { key: 'amount_usd', label: 'Nilai USD', kind: 'currency', align: 'right' },
  { key: 'oldest_date', label: 'Tanggal Terlama' },
  { key: 'severity', label: 'Severity', kind: 'badge', align: 'center' },
  { key: 'drilldown', label: 'Aksi', kind: 'drilldown', align: 'center' },
]
const dataQualityColumns: AnalyticsBreakdownTableColumn[] = [
  { key: 'label', label: 'Kelengkapan data' },
  { key: 'stage', label: 'Stage' },
  { key: 'count', label: 'Jumlah', kind: 'number', align: 'right' },
  { key: 'severity', label: 'Severity', kind: 'badge', align: 'center' },
  { key: 'drilldown', label: 'Aksi', kind: 'drilldown', align: 'center' },
]

const portfolioMetrics = computed<AnalyticsMoneyMetric[]>(() => {
  const overview = analytics.overview.value
  const portfolio = overview?.portfolio

  return [
    metric('project_count', 'Project logical', portfolio?.project_count ?? 0, 'number', overview?.drilldown),
    metric('assignment_count', 'Assignment Kementerian/Lembaga', portfolio?.assignment_count ?? 0),
    metric(
      'pipeline_loan',
      'Pipeline loan USD',
      portfolio?.total_pipeline_loan_usd ?? 0,
      'currency',
    ),
    metric(
      'agreement_amount',
      'Loan Agreement USD',
      portfolio?.total_agreement_amount_usd ?? 0,
      'currency',
    ),
    metric('planned', 'Rencana USD', portfolio?.total_planned_usd ?? 0, 'currency'),
    metric('realized', 'Realisasi USD', portfolio?.total_realized_usd ?? 0, 'currency'),
    metric('absorption', 'Penyerapan', portfolio?.absorption_pct ?? 0, 'percent'),
  ]
})
const institutionMetrics = computed<AnalyticsMoneyMetric[]>(() => {
  const summary = analytics.institutions.value?.summary
  const drilldown = analytics.institutions.value?.drilldown

  return [
    metric('institution_count', 'Kementerian/Lembaga', summary?.institution_count ?? 0, 'number', drilldown),
    metric('project_count', 'Project logical', summary?.project_count ?? 0, 'number', drilldown),
    metric('assignment_count', 'Assignment', summary?.assignment_count ?? 0),
    metric(
      'agreement_amount',
      'Loan Agreement USD',
      summary?.total_agreement_amount_usd ?? 0,
      'currency',
    ),
    metric('planned', 'Rencana USD', summary?.total_planned_usd ?? 0, 'currency'),
    metric('realized', 'Realisasi USD', summary?.total_realized_usd ?? 0, 'currency'),
    metric('absorption', 'Penyerapan', summary?.absorption_pct ?? 0, 'percent'),
  ]
})
const lenderMetrics = computed<AnalyticsMoneyMetric[]>(() => {
  const summary = analytics.lenders.value?.summary
  const drilldown = analytics.lenders.value?.drilldown

  return [
    metric('lender_count', 'Lender legal binding', summary?.lender_count ?? 0, 'number', drilldown),
    metric('loan_agreement_count', 'Loan Agreement', summary?.loan_agreement_count ?? 0, 'number', drilldown),
    metric(
      'agreement_amount',
      'Loan Agreement USD',
      summary?.total_agreement_amount_usd ?? 0,
      'currency',
    ),
    metric('planned', 'Rencana USD', summary?.total_planned_usd ?? 0, 'currency'),
    metric('realized', 'Realisasi USD', summary?.total_realized_usd ?? 0, 'currency'),
    metric('absorption', 'Penyerapan', summary?.absorption_pct ?? 0, 'percent'),
  ]
})
const absorptionMetrics = computed<AnalyticsMoneyMetric[]>(() => {
  const summary = analytics.absorption.value?.summary
  const drilldown = analytics.absorption.value?.drilldown

  return [
    metric('planned', 'Rencana USD', summary?.planned_usd ?? 0, 'currency', drilldown),
    metric('realized', 'Realisasi USD', summary?.realized_usd ?? 0, 'currency', drilldown),
    metric('absorption', 'Penyerapan', summary?.absorption_pct ?? 0, 'percent', drilldown),
  ]
})
const yearlyMetrics = computed<AnalyticsMoneyMetric[]>(() => {
  const summary = analytics.yearly.value?.summary
  const drilldown = analytics.yearly.value?.drilldown

  return [
    metric('planned', 'Rencana USD', summary?.planned_usd ?? 0, 'currency', drilldown),
    metric('realized', 'Realisasi USD', summary?.realized_usd ?? 0, 'currency', drilldown),
    metric('absorption', 'Penyerapan', summary?.absorption_pct ?? 0, 'percent', drilldown),
    metric('loan_agreement_count', 'Loan Agreement', summary?.loan_agreement_count ?? 0),
    metric('project_count', 'Project aktif', summary?.project_count ?? 0),
  ]
})
const riskMetrics = computed<AnalyticsMoneyMetric[]>(() =>
  (analytics.risks.value?.risk_cards ?? []).map((card) => ({
    key: card.code,
    label: riskCardLabel(card.code, card.label),
    value: card.amount_usd ?? card.count,
    format: card.amount_usd !== undefined ? 'currency' : 'number',
    unit: card.amount_usd !== undefined ? 'USD' : undefined,
    severity:
      card.severity === 'warning' ? 'warning' : card.severity === 'danger' ? 'danger' : 'info',
    drilldown: card.drilldown,
  })),
)
const dataQualityMetrics = computed<AnalyticsMoneyMetric[]>(() =>
  (analytics.risks.value?.data_quality ?? []).map((item) => ({
    key: `${item.code}-${item.stage}`,
    label: `${item.label} - ${item.stage}`,
    value: item.count,
    format: 'number',
    severity:
      item.severity === 'warning' ? 'warning' : item.severity === 'danger' ? 'danger' : 'info',
    drilldown: item.drilldown,
  })),
)

const portfolioRows = computed<AnalyticsBreakdownTableRow[]>(() =>
  (analytics.overview.value?.pipeline_funnel ?? []).map((item) => ({
    id: item.stage,
    cells: {
      stage: stageLabel(item.stage),
      project_count: item.project_count,
      total_loan_usd: item.total_loan_usd,
    },
    drilldown: item.drilldown,
  })),
)
const institutionRows = computed<AnalyticsBreakdownTableRow[]>(() =>
  (analytics.institutions.value?.items ?? []).map((item) => ({
    id: item.institution.id,
    severity: absorptionSeverity(item.absorption_pct),
    cells: {
      institution: institutionLabel(item.institution),
      project_count: item.project_count,
      assignment_count: item.assignment_count,
      loan_agreement_count: item.loan_agreement_count,
      monitoring_count: item.monitoring_count,
      agreement_amount_usd: item.agreement_amount_usd,
      planned_usd: item.planned_usd,
      realized_usd: item.realized_usd,
      absorption_pct: item.absorption_pct,
      BB: item.pipeline_breakdown.BB,
      GB: item.pipeline_breakdown.GB,
      DK: item.pipeline_breakdown.DK,
      LA: item.pipeline_breakdown.LA,
      Monitoring: item.pipeline_breakdown.Monitoring,
    },
    drilldown: item.drilldown,
  })),
)
const lenderRows = computed<AnalyticsBreakdownTableRow[]>(() =>
  (analytics.lenders.value?.items ?? []).map((item) => ({
    id: item.lender.id,
    severity: lenderSeverity(item.lender.type),
    cells: {
      lender: lenderLabel(item.lender),
      type: item.lender.type,
      loan_agreement_count: item.loan_agreement_count,
      project_count: item.project_count,
      institution_count: item.institution_count,
      monitoring_count: item.monitoring_count,
      agreement_amount_usd: item.agreement_amount_usd,
      planned_usd: item.planned_usd,
      realized_usd: item.realized_usd,
      absorption_pct: item.absorption_pct,
    },
    drilldown: item.drilldown,
  })),
)
const lenderProportionRows = computed<AnalyticsBreakdownTableRow[]>(() =>
  (analytics.lenderProportion.value?.by_stage ?? []).flatMap((stage) =>
    stage.items.map((item) => ({
      id: `${stage.stage}-${item.type}`,
      severity: lenderSeverity(item.type),
      cells: {
        stage: stage.stage,
        type: item.type,
        project_count: item.project_count,
        lender_count: item.lender_count,
        amount_usd: item.amount_usd,
        share_pct: item.share_pct,
      },
      drilldown: item.drilldown,
    })),
  ),
)
const activeAbsorptionItems = computed<DashboardAbsorptionRankedItem[]>(() => {
  const data = analytics.absorption.value

  if (!data) return []
  if (absorptionLevel.value === 'project') return data.by_project
  if (absorptionLevel.value === 'lender') return data.by_lender

  return data.by_institution
})
const activeAbsorptionLabel = computed(() => {
  if (absorptionLevel.value === 'project') return 'Project'
  if (absorptionLevel.value === 'lender') return 'Lender'

  return 'Kementerian/Lembaga'
})
const absorptionRows = computed<AnalyticsBreakdownTableRow[]>(() =>
  absorptionRankRows(activeAbsorptionItems.value),
)
const lowAbsorptionRows = computed<AnalyticsBreakdownTableRow[]>(() =>
  absorptionRankRows(
    activeAbsorptionItems.value.filter((item) => item.status === 'low').slice(0, 8),
  ),
)
const yearlyItems = computed<DashboardYearlyItem[]>(() =>
  [...(analytics.yearly.value?.items ?? [])].sort(
    (left, right) =>
      left.budget_year - right.budget_year || quarterIndex(left.quarter) - quarterIndex(right.quarter),
  ),
)
const yearlyRows = computed<AnalyticsBreakdownTableRow[]>(() =>
  yearlyItems.value.map((item) => ({
    id: `${item.budget_year}-${item.quarter}`,
    severity: absorptionSeverity(item.absorption_pct),
    cells: {
      period: `${item.budget_year} ${item.quarter}`,
      planned_usd: item.planned_usd,
      realized_usd: item.realized_usd,
      absorption_pct: item.absorption_pct,
      loan_agreement_count: item.loan_agreement_count,
      project_count: item.project_count,
    },
    drilldown: item.drilldown,
  })),
)
const riskWatchlistRows = computed<AnalyticsBreakdownTableRow[]>(() => [
  ...riskRowsFor(
    'Penyerapan rendah',
    analytics.risks.value?.watchlists.low_absorption_projects ?? [],
  ),
  ...riskRowsFor(
    'Loan Agreement efektif tanpa monitoring',
    analytics.risks.value?.watchlists.effective_without_monitoring ?? [],
  ),
  ...riskRowsFor('Mendekati closing date', analytics.risks.value?.watchlists.closing_risks ?? []),
  ...riskRowsFor('Loan Agreement diperpanjang', analytics.risks.value?.watchlists.extended_loans ?? []),
])
const pipelineBottleneckRows = computed<AnalyticsBreakdownTableRow[]>(() =>
  (analytics.risks.value?.watchlists.pipeline_bottlenecks ?? []).map((item) => ({
    id: `bottleneck-${item.stage}`,
    severity: item.severity,
    cells: {
      label: item.label,
      stage: stageLabel(item.stage),
      count: item.project_count,
      amount_usd: item.total_loan_usd,
      oldest_date: item.oldest_date || '-',
      severity: item.severity,
    },
    drilldown: item.drilldown,
  })),
)
const dataQualityRows = computed<AnalyticsBreakdownTableRow[]>(() =>
  (analytics.risks.value?.data_quality ?? []).map((item) => ({
    id: `data-quality-${item.code}-${item.stage}`,
    severity: item.severity,
    cells: {
      label: item.label,
      stage: item.stage,
      count: item.count,
      severity: item.severity,
    },
    drilldown: item.drilldown,
  })),
)

const pipelineFunnelChartOption = computed(() => {
  const rows = analytics.overview.value?.pipeline_funnel ?? []
  const labels = rows.map((item) => stageLabel(item.stage))

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params: unknown) => {
        const item = tooltipItems(params)[0]
        const row = rows[item?.dataIndex ?? 0]

        return [
          `<strong>${row ? stageLabel(row.stage) : ''}</strong>`,
          `Project: ${formatNumber(row?.project_count ?? 0)}`,
          `Nilai pinjaman: ${formatUSD(row?.total_loan_usd ?? 0)}`,
        ].join('<br/>')
      },
    },
    grid: { left: 128, right: 24, top: 16, bottom: 24, containLabel: true },
    xAxis: { type: 'value' },
    yAxis: { type: 'category', data: labels },
    series: [
      {
        name: 'Project',
        type: 'bar',
        data: rows.map((item) => item.project_count),
        barMaxWidth: 22,
        itemStyle: { color: '#2563eb', borderRadius: [0, 5, 5, 0] },
      },
    ],
  }
})
const institutionProjectChartOption = computed(() =>
  horizontalBarOption(
    topInstitutionsByProject.value.map((item) => shortLabel(institutionLabel(item.institution))),
    topInstitutionsByProject.value.map((item) => item.project_count),
    'Project',
    '#0f766e',
    formatNumber,
  ),
)
const institutionAbsorptionChartOption = computed(() =>
  horizontalBarOption(
    topInstitutionsByAbsorption.value.map((item) => shortLabel(institutionLabel(item.institution))),
    topInstitutionsByAbsorption.value.map((item) => item.absorption_pct),
    'Penyerapan',
    '#0284c7',
    (value) => `${value.toFixed(1)}%`,
    100,
  ),
)
const lenderPerformanceChartOption = computed(() =>
  horizontalBarOption(
    topLendersByAmount.value.map((item) => shortLabel(lenderLabel(item.lender))),
    topLendersByAmount.value.map((item) => item.agreement_amount_usd),
    'Loan Agreement USD',
    '#7c3aed',
    formatUSD,
    undefined,
    formatAxisUSD,
  ),
)
const absorptionChartOption = computed(() => {
  const rows = [...activeAbsorptionItems.value]
    .sort((left, right) => left.absorption_pct - right.absorption_pct)
    .slice(0, 10)

  return horizontalBarOption(
    rows.map((item) => shortLabel(item.name)),
    rows.map((item) => item.absorption_pct),
    'Penyerapan',
    '#dc2626',
    (value) => `${value.toFixed(1)}%`,
    100,
  )
})
const yearlyTrendChartOption = computed(() => {
  const labels = yearlyItems.value.map((item) => `${item.budget_year} ${item.quarter}`)

  return {
    color: ['#2563eb', '#16a34a', '#f59e0b'],
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params: unknown) => {
        const item = tooltipItems(params)[0]
        const row = yearlyItems.value[item?.dataIndex ?? 0]

        return [
          `<strong>${row ? `${row.budget_year} ${row.quarter}` : ''}</strong>`,
          `Rencana: ${formatUSD(row?.planned_usd ?? 0)}`,
          `Realisasi: ${formatUSD(row?.realized_usd ?? 0)}`,
          `Penyerapan: ${(row?.absorption_pct ?? 0).toFixed(1)}%`,
        ].join('<br/>')
      },
    },
    legend: { top: 0 },
    grid: { left: 56, right: 56, top: 48, bottom: 56, containLabel: true },
    xAxis: { type: 'category', data: labels, axisLabel: { rotate: labels.length > 6 ? 30 : 0 } },
    yAxis: [
      { type: 'value', axisLabel: { formatter: (value: number) => formatAxisUSD(value) } },
      { type: 'value', max: 100, axisLabel: { formatter: (value: number) => `${value}%` } },
    ],
    series: [
      {
        name: 'Rencana USD',
        type: 'bar',
        data: yearlyItems.value.map((item) => item.planned_usd),
        barMaxWidth: 28,
      },
      {
        name: 'Realisasi USD',
        type: 'bar',
        data: yearlyItems.value.map((item) => item.realized_usd),
        barMaxWidth: 28,
      },
      {
        name: 'Penyerapan',
        type: 'line',
        yAxisIndex: 1,
        data: yearlyItems.value.map((item) => item.absorption_pct),
        smooth: true,
      },
    ],
  }
})
const lenderProportionChartOption = computed(() => {
  const stages = analytics.lenderProportion.value?.by_stage ?? []
  const labels = stages.map((stage) => stage.stage)
  const lenderTypes = ['Bilateral', 'Multilateral', 'KSA'] as const
  const colors: Record<(typeof lenderTypes)[number], string> = {
    Bilateral: '#2563eb',
    Multilateral: '#64748b',
    KSA: '#d97706',
  }

  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params: unknown) => {
        const rows = tooltipItems(params)
        const stage = stages[rows[0]?.dataIndex ?? 0]

        return [
          `<strong>${stage?.stage ?? ''}</strong>`,
          ...rows.map((row) => {
            const item = stage?.items.find((candidate) => candidate.type === row.seriesName)
            const count = item?.project_count ?? 0
            const amount = item?.amount_usd ?? 0

            return `${row.marker}${row.seriesName}: ${row.value.toFixed(1)}% | Project ${formatNumber(count)} | ${formatUSD(amount)}`
          }),
        ].join('<br/>')
      },
    },
    legend: { top: 0 },
    grid: { left: 160, right: 24, top: 48, bottom: 24, containLabel: true },
    xAxis: { type: 'value', max: 100, axisLabel: { formatter: (value: number) => `${value}%` } },
    yAxis: { type: 'category', data: labels },
    series: lenderTypes.map((type) => ({
      name: type,
      type: 'bar',
      stack: 'share',
      data: stages.map((stage) => stage.items.find((item) => item.type === type)?.share_pct ?? 0),
      barMaxWidth: 24,
      itemStyle: { color: colors[type] },
    })),
  }
})

const topInstitutionsByProject = computed(() =>
  [...(analytics.institutions.value?.items ?? [])]
    .sort((left, right) => right.project_count - left.project_count)
    .slice(0, 10),
)
const topInstitutionsByAbsorption = computed(() =>
  [...(analytics.institutions.value?.items ?? [])]
    .filter((item) => item.planned_usd > 0)
    .sort((left, right) => right.absorption_pct - left.absorption_pct)
    .slice(0, 10),
)
const topLendersByAmount = computed(() =>
  [...(analytics.lenders.value?.items ?? [])]
    .sort((left, right) => right.agreement_amount_usd - left.agreement_amount_usd)
    .slice(0, 10),
)
const yearlyContext = computed(() => {
  const year = analytics.appliedFilters.budget_year
  const quarter = analytics.appliedFilters.quarter

  if (year && quarter) return `Filter aktif: ${year} ${quarter}.`
  if (year) return `Filter tahun aktif: ${year}; endpoint menampilkan semua triwulan pada tahun tersebut.`
  if (quarter) return `Filter triwulan aktif: ${quarter}; data ditampilkan lintas tahun sesuai filter.`

  return 'Tanpa filter tahun/triwulan; data ditampilkan berurutan dari tahun dan triwulan yang tersedia.'
})
const lenderProportionAmountNotes = computed(() =>
  (analytics.lenderProportion.value?.by_stage ?? [])
    .filter((stage) => stageTotalAmount(stage) === 0 && stageTotalProjects(stage) > 0)
    .map((stage) => stage.stage),
)
const currentTabSections = computed(
  () => tabs.find((tab) => tab.key === activeTab.value)?.sections ?? [],
)

function metric(
  key: string,
  label: string,
  value: number,
  format: AnalyticsMoneyMetric['format'] = 'number',
  drilldown?: DashboardDrilldownQuery,
): AnalyticsMoneyMetric {
  return {
    key,
    label,
    value,
    format,
    unit: format === 'currency' ? 'USD' : undefined,
    drilldown,
  }
}

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

function lenderSeverity(type: string) {
  if (type === 'KSA') return 'warning'
  if (type === 'Multilateral') return 'secondary'
  return 'info'
}

function absorptionSeverity(value: number) {
  if (value < 50) return 'danger'
  if (value >= 90) return 'success'
  return 'info'
}

function absorptionStatusLabel(value: string) {
  if (value === 'low') return 'Rendah'
  if (value === 'high') return 'Tinggi'
  return 'Normal'
}

function absorptionRankRows(rows: DashboardAbsorptionRankedItem[]): AnalyticsBreakdownTableRow[] {
  return rows.map((item) => ({
    id: `${activeAbsorptionLabel.value}-${item.id}`,
    severity: item.status === 'low' ? 'danger' : item.status === 'high' ? 'success' : 'info',
    cells: {
      rank: item.rank ?? 0,
      name: item.name,
      planned_usd: item.planned_usd,
      realized_usd: item.realized_usd,
      variance_usd: item.variance_usd,
      absorption_pct: item.absorption_pct,
      status: absorptionStatusLabel(String(item.status)),
    },
    drilldown: item.drilldown,
  }))
}

function riskRowsFor(
  category: string,
  items: DashboardLoanAgreementRiskItem[],
): AnalyticsBreakdownTableRow[] {
  return items.map((item) => ({
    id: `${category}-${item.loan_agreement_id}-${item.risk_code}-${item.budget_year ?? ''}-${item.quarter ?? ''}`,
    severity: item.severity,
    cells: {
      category,
      project: item.project_name || '-',
      institution: item.institution ? institutionLabel(item.institution) : '-',
      lender: lenderLabel(item.lender),
      amount_usd: riskAmountUSD(item),
      absorption_pct: riskHasAbsorption(item) ? item.absorption_pct : null,
      severity: severityLabel(item.severity),
    },
    drilldown: item.drilldown,
  }))
}

function riskAmountUSD(item: DashboardLoanAgreementRiskItem) {
  if (item.risk_code === 'LOW_ABSORPTION') return item.planned_usd
  if (item.agreement_amount_usd !== undefined) return item.agreement_amount_usd
  return null
}

function riskHasAbsorption(item: DashboardLoanAgreementRiskItem) {
  return item.risk_code === 'LOW_ABSORPTION' || item.risk_code === 'CLOSING_RISK'
}

function severityLabel(value: string) {
  if (value === 'danger') return 'Tinggi'
  if (value === 'warning') return 'Perlu perhatian'
  if (value === 'success') return 'Selesai'
  if (value === 'info') return 'Info'
  return value || '-'
}

function riskCardLabel(code: string, fallback: string) {
  const labels: Record<string, string> = {
    LOW_ABSORPTION: 'Penyerapan rendah',
    EFFECTIVE_WITHOUT_MONITORING: 'Loan Agreement efektif tanpa monitoring',
    CLOSING_RISK: 'Mendekati closing date',
    EXTENDED_LOAN: 'Loan Agreement diperpanjang',
    PIPELINE_BOTTLENECK: 'Belum berlanjut ke tahap berikutnya',
    DATA_QUALITY: 'Kelengkapan data',
  }

  return labels[code] ?? fallback
}

function institutionLabel(institution: { name: string; short_name?: string | null }) {
  return institution.short_name ? `${institution.name} (${institution.short_name})` : institution.name
}

function lenderLabel(lender: { name: string; short_name?: string | null }) {
  return lender.short_name ? `${lender.name} (${lender.short_name})` : lender.name
}

function sectionLoading(sections: DashboardAnalyticsSectionKey[]) {
  return sections.some((section) => analytics.loading[section])
}

function sectionError(sections: DashboardAnalyticsSectionKey[]) {
  return sections.map((section) => analytics.errors[section]).find(Boolean)
}

function handleDrilldown(drilldown: DashboardDrilldownQuery) {
  void analytics.openDrilldown(drilldown)
}

function retryActiveSections() {
  currentTabSections.value.forEach((section) => {
    void analytics.fetchSection(section)
  })
}

function quarterIndex(quarter: string) {
  const order: Record<string, number> = { TW1: 1, TW2: 2, TW3: 3, TW4: 4 }

  return order[quarter] ?? 99
}

function recordFrom(value: unknown): Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value) ? value as Record<string, unknown> : {}
}

function numberFrom(value: unknown) {
  if (typeof value === 'number' && Number.isFinite(value)) return value
  if (typeof value === 'string' && value.trim()) {
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : 0
  }

  return 0
}

function stringFrom(value: unknown) {
  return typeof value === 'string' ? value : ''
}

function tooltipItems(params: unknown): TooltipParam[] {
  const rawItems = Array.isArray(params) ? params : [params]

  return rawItems.map((item) => {
    const raw = recordFrom(item)
    const rawValue = raw.value
    const value = Array.isArray(rawValue) ? numberFrom(rawValue[1]) : numberFrom(rawValue)

    return {
      dataIndex: numberFrom(raw.dataIndex),
      seriesName: stringFrom(raw.seriesName),
      value,
      marker: stringFrom(raw.marker),
      name: stringFrom(raw.name),
      axisValue: stringFrom(raw.axisValue),
    }
  })
}

function formatNumber(value: number) {
  return new Intl.NumberFormat('id-ID').format(value)
}

function formatUSD(value: number) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value)
}

function formatAxisUSD(value: number) {
  const abs = Math.abs(value)

  if (abs >= 1_000_000_000) return `USD ${(value / 1_000_000_000).toFixed(1)}B`
  if (abs >= 1_000_000) return `USD ${(value / 1_000_000).toFixed(1)}M`
  if (abs >= 1_000) return `USD ${(value / 1_000).toFixed(1)}K`

  return `USD ${value}`
}

function shortLabel(value: string) {
  return value.length > 34 ? `${value.slice(0, 31)}...` : value
}

function horizontalBarOption(
  labels: string[],
  values: number[],
  seriesName: string,
  color: string,
  valueFormatter: (value: number) => string,
  max?: number,
  axisFormatter: (value: number) => string = valueFormatter,
) {
  return {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params: unknown) => {
        const item = tooltipItems(params)[0]

        return `<strong>${item?.axisValue || item?.name || ''}</strong><br/>${seriesName}: ${valueFormatter(item?.value ?? 0)}`
      },
    },
    grid: { left: 168, right: 24, top: 16, bottom: 24, containLabel: true },
    xAxis: {
      type: 'value',
      max,
      axisLabel: { formatter: (value: number) => (max === 100 ? `${value}%` : axisFormatter(value)) },
    },
    yAxis: { type: 'category', data: labels },
    series: [
      {
        name: seriesName,
        type: 'bar',
        data: values,
        barMaxWidth: 22,
        itemStyle: { color, borderRadius: [0, 5, 5, 0] },
      },
    ],
  }
}

function stageTotalAmount(stage: DashboardLenderProportionStage) {
  return stage.items.reduce((sum, item) => sum + item.amount_usd, 0)
}

function stageTotalProjects(stage: DashboardLenderProportionStage) {
  return stage.items.reduce((sum, item) => sum + item.project_count, 0)
}

onMounted(() => {
  void analytics.initialize()
})
</script>

<template>
  <section class="space-y-5">
    <PageHeader
      title="Dashboard Analytics"
      subtitle="Ringkasan portfolio pinjaman luar negeri, Kementerian/Lembaga, lender, dan penyerapan."
    />

    <DashboardAnalyticsFilterBar
      :model-value="analytics.draftFilters"
      :lenders="masterStore.lenders"
      :institutions="masterStore.institutions"
      :regions="masterStore.regions"
      :program-titles="masterStore.programTitles"
      :loading="analytics.anyLoading.value"
      :loading-options="analytics.loadingMasterData.value"
      @update:model-value="analytics.updateDraftFilters"
      @apply="analytics.applyFilters"
      @reset="analytics.resetFilters"
    />

    <nav
      class="flex gap-2 overflow-x-auto rounded-lg border border-surface-200 bg-white p-2"
      aria-label="Analytics tabs"
    >
      <Button
        v-for="tab in tabs"
        :key="tab.key"
        :label="tab.label"
        :severity="activeTab === tab.key ? undefined : 'secondary'"
        :outlined="activeTab !== tab.key"
        size="small"
        @click="activeTab = tab.key"
      />
    </nav>

    <section
      v-if="sectionError(currentTabSections)"
      class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700"
    >
      <span>{{ sectionError(currentTabSections) }}</span>
      <Button
        label="Coba lagi"
        icon="pi pi-refresh"
        severity="danger"
        size="small"
        outlined
        @click="retryActiveSections"
      />
    </section>

    <section v-if="activeTab === 'portfolio'" class="space-y-4">
      <AnalyticsMetricGrid
        :metrics="portfolioMetrics"
        :loading="sectionLoading(['overview'])"
        @drilldown="handleDrilldown"
      />
      <div class="grid gap-4 xl:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]">
        <AnalyticsChartPanel
          title="Funnel Pipeline"
          description="Jumlah project per stage memakai latest snapshot default."
          :option="pipelineFunnelChartOption"
          :loading="sectionLoading(['overview'])"
          :empty="portfolioRows.length === 0"
        />
        <AnalyticsBreakdownTable
          title="Pipeline Portfolio"
          description="Blue Book, Green Book, Daftar Kegiatan, Loan Agreement, dan Monitoring ditampilkan sebagai stage eksplisit."
          :columns="portfolioColumns"
          :rows="portfolioRows"
          :loading="sectionLoading(['overview'])"
          @drilldown="handleDrilldown"
        />
      </div>
    </section>

    <section v-else-if="activeTab === 'institutions'" class="space-y-4">
      <AnalyticsMetricGrid
        :metrics="institutionMetrics"
        :loading="sectionLoading(['institutions'])"
        @drilldown="handleDrilldown"
      />
      <div class="grid gap-4 xl:grid-cols-2">
        <AnalyticsChartPanel
          title="Top 10 Kementerian/Lembaga by Project"
          description="Project count deduplicated; assignment count tetap overlap-aware di tabel."
          :option="institutionProjectChartOption"
          :loading="sectionLoading(['institutions'])"
          :empty="topInstitutionsByProject.length === 0"
        />
        <AnalyticsChartPanel
          title="Top Penyerapan Kementerian/Lembaga"
          description="Hanya menghitung baris dengan rencana USD lebih dari 0."
          :option="institutionAbsorptionChartOption"
          :loading="sectionLoading(['institutions'])"
          :empty="topInstitutionsByAbsorption.length === 0"
        />
      </div>
      <AnalyticsBreakdownTable
        title="Performa Kementerian/Lembaga"
        description="Nama panjang dipertahankan di tabel, tanpa menampilkan UUID."
        :columns="institutionColumns"
        :rows="institutionRows"
        :loading="sectionLoading(['institutions'])"
        @drilldown="handleDrilldown"
      />
    </section>

    <section v-else-if="activeTab === 'lenders'" class="space-y-4">
      <AnalyticsMetricGrid
        :metrics="lenderMetrics"
        :loading="sectionLoading(['lenders'])"
        @drilldown="handleDrilldown"
      />
      <AnalyticsChartPanel
        title="Top Lender by Loan Agreement USD"
        description="Basis lender pada performa legal adalah Loan Agreement; KSA dipisahkan sebagai tipe sendiri."
        :option="lenderPerformanceChartOption"
        :loading="sectionLoading(['lenders'])"
        :empty="topLendersByAmount.length === 0"
      />
      <AnalyticsBreakdownTable
        title="Performa Lender Legal"
        description="Tidak memakai lender indication sebagai performa legal binding."
        :columns="lenderColumns"
        :rows="lenderRows"
        :loading="sectionLoading(['lenders'])"
        @drilldown="handleDrilldown"
      />
      <AnalyticsMatrixTable
        v-model:top-n="matrixTopN"
        :items="analytics.lenders.value?.lender_institution_matrix ?? []"
        :loading="sectionLoading(['lenders'])"
        @drilldown="handleDrilldown"
      />
      <div class="grid gap-4 xl:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]">
        <AnalyticsChartPanel
          title="Proporsi Lender per Stage"
          description="Lender Indication, Green Book Funding Source, Loan Agreement, dan Monitoring Realization tidak digabung tanpa label."
          :option="lenderProportionChartOption"
          :loading="sectionLoading(['lenderProportion'])"
          :empty="lenderProportionRows.length === 0"
        />
        <AnalyticsBreakdownTable
          title="Detail Proporsi Lender"
          :columns="lenderProportionColumns"
          :rows="lenderProportionRows"
          :loading="sectionLoading(['lenderProportion'])"
          @drilldown="handleDrilldown"
        />
      </div>
      <p
        v-if="lenderProportionAmountNotes.length > 0"
        class="rounded-lg border border-amber-200 bg-amber-50 p-3 text-sm text-amber-800"
      >
        Stage tanpa nilai USD memakai count sebagai konteks: {{ lenderProportionAmountNotes.join(', ') }}.
      </p>
    </section>

    <section v-else-if="activeTab === 'absorption'" class="space-y-4">
      <AnalyticsMetricGrid
        :metrics="absorptionMetrics"
        :loading="sectionLoading(['absorption'])"
        @drilldown="handleDrilldown"
      />
      <div class="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-surface-200 bg-white p-3">
        <div>
          <h2 class="text-base font-semibold text-surface-950">Level Penyerapan</h2>
          <p class="text-sm text-surface-500">Rencana 0 tetap ditampilkan sebagai 0% dari backend.</p>
        </div>
        <SelectButton
          v-model="absorptionLevel"
          :options="absorptionLevelOptions"
          option-label="label"
          option-value="value"
          :allow-empty="false"
        />
      </div>
      <div class="grid gap-4 xl:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]">
        <AnalyticsChartPanel
          :title="`Top Low Absorption ${activeAbsorptionLabel}`"
          description="Urutan dari penyerapan terendah agar anomali mudah dipindai."
          :option="absorptionChartOption"
          :loading="sectionLoading(['absorption'])"
          :empty="activeAbsorptionItems.length === 0"
        />
        <AnalyticsBreakdownTable
          title="Daftar Penyerapan Rendah"
          :columns="absorptionColumns"
          :rows="lowAbsorptionRows"
          :loading="sectionLoading(['absorption'])"
          empty-title="Tidak ada penyerapan rendah"
          empty-description="Tidak ada baris berstatus rendah untuk level dan filter aktif."
          @drilldown="handleDrilldown"
        />
      </div>
      <AnalyticsBreakdownTable
        :title="`Ranking Penyerapan ${activeAbsorptionLabel}`"
        :columns="absorptionColumns"
        :rows="absorptionRows"
        :loading="sectionLoading(['absorption'])"
        @drilldown="handleDrilldown"
      />
    </section>

    <section v-else-if="activeTab === 'yearly'" class="space-y-4">
      <AnalyticsMetricGrid
        :metrics="yearlyMetrics"
        :loading="sectionLoading(['yearly'])"
        @drilldown="handleDrilldown"
      />
      <p class="rounded-lg border border-surface-200 bg-white p-3 text-sm text-surface-600">
        {{ yearlyContext }}
      </p>
      <AnalyticsChartPanel
        title="Trend Planned vs Realized"
        description="Grouped bar untuk rencana/realisasi dan line untuk penyerapan."
        :option="yearlyTrendChartOption"
        :loading="sectionLoading(['yearly'])"
        :empty="yearlyRows.length === 0"
      />
      <AnalyticsBreakdownTable
        title="Detail Performa Tahunan"
        :columns="yearlyColumns"
        :rows="yearlyRows"
        :loading="sectionLoading(['yearly'])"
        @drilldown="handleDrilldown"
      />
    </section>

    <section v-else-if="activeTab === 'risks'" class="space-y-4">
      <AnalyticsMetricGrid
        v-if="riskMetrics.length > 0 || sectionLoading(['risks'])"
        :metrics="riskMetrics"
        :loading="sectionLoading(['risks'])"
        @drilldown="handleDrilldown"
      />
      <AnalyticsEmptyState
        v-else
        title="Tidak ada risk card"
        description="Endpoint risks mengembalikan daftar kosong untuk filter aktif."
      />
      <AnalyticsBreakdownTable
        title="Risk Watchlist"
        description="Loan Agreement, monitoring, dan risiko closing ditampilkan dengan stage legal yang eksplisit."
        :columns="riskWatchlistColumns"
        :rows="riskWatchlistRows"
        :loading="sectionLoading(['risks'])"
        empty-title="Tidak ada risiko aktif"
        empty-description="Tidak ada watchlist Loan Agreement atau monitoring untuk filter aktif."
        @drilldown="handleDrilldown"
      />
      <AnalyticsBreakdownTable
        title="Belum berlanjut ke tahap berikutnya"
        description="Project pipeline memakai latest snapshot default dan tidak menghitung revisi historis ganda."
        :columns="pipelineBottleneckColumns"
        :rows="pipelineBottleneckRows"
        :loading="sectionLoading(['risks'])"
        empty-title="Tidak ada bottleneck pipeline"
        empty-description="Tidak ada project yang tertahan pada filter aktif."
        @drilldown="handleDrilldown"
      />
      <AnalyticsMetricGrid
        v-if="dataQualityMetrics.length > 0 || sectionLoading(['risks'])"
        :metrics="dataQualityMetrics"
        :loading="sectionLoading(['risks'])"
        @drilldown="handleDrilldown"
      />
      <AnalyticsBreakdownTable
        title="Kelengkapan data"
        description="Klik tiap isu untuk membuka workspace target dengan filter yang masih bisa diubah."
        :columns="dataQualityColumns"
        :rows="dataQualityRows"
        :loading="sectionLoading(['risks'])"
        empty-title="Tidak ada isu kelengkapan data"
        empty-description="Data quality tidak menemukan isu untuk filter aktif."
        @drilldown="handleDrilldown"
      />
    </section>
  </section>
</template>
