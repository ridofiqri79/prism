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
import DashboardPipelineFunnel from '@/components/dashboard/DashboardPipelineFunnel.vue'
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
  group: string
  label: string
  sections: DashboardAnalyticsSectionKey[]
}> = [
  { key: 'portfolio', group: 'Ringkasan', label: 'Portfolio', sections: ['overview'] },
  { key: 'institutions', group: 'Entitas', label: 'Kementerian/Lembaga', sections: ['institutions'] },
  { key: 'lenders', group: 'Entitas', label: 'Lender', sections: ['lenders', 'lenderProportion'] },
  { key: 'absorption', group: 'Metrik', label: 'Penyerapan', sections: ['absorption'] },
  { key: 'yearly', group: 'Waktu', label: 'Tahunan', sections: ['yearly'] },
  { key: 'risks', group: 'Risiko & Kualitas', label: 'Watchlist', sections: ['risks'] },
]
const tabGroups = computed(() => {
  const groups: Array<{ label: string; tabs: typeof tabs }> = []

  tabs.forEach((tab) => {
    const group = groups.find((item) => item.label === tab.group)

    if (group) {
      group.tabs.push(tab)
    } else {
      groups.push({ label: tab.group, tabs: [tab] })
    }
  })

  return groups
})
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
    metric(
      'pipeline_loan',
      'Nilai pinjaman pipeline',
      portfolio?.total_pipeline_loan_usd ?? 0,
      'currency',
      undefined,
      'Total nilai pinjaman dari proyek yang masih berada di pipeline perencanaan.',
      'high',
    ),
    metric(
      'agreement_amount',
      'Loan Agreement USD',
      portfolio?.total_agreement_amount_usd ?? 0,
      'currency',
      undefined,
      'Nilai pinjaman yang sudah memiliki Loan Agreement.',
      'high',
    ),
    metric(
      'project_count',
      'Total proyek',
      portfolio?.project_count ?? 0,
      'number',
      overview?.drilldown,
      'Proyek dihitung unik lintas revisi.',
    ),
    metric(
      'assignment_count',
      'Assignment Kementerian/Lembaga',
      portfolio?.assignment_count ?? 0,
      'number',
      undefined,
      'Assignment dapat lebih besar dari total proyek jika satu proyek punya lebih dari satu Kementerian/Lembaga.',
    ),
    metric(
      'planned',
      'Rencana monitoring USD',
      portfolio?.total_planned_usd ?? 0,
      'currency',
      undefined,
      'Total rencana monitoring disbursement untuk filter aktif.',
    ),
    metric(
      'realized',
      'Realisasi monitoring USD',
      portfolio?.total_realized_usd ?? 0,
      'currency',
      undefined,
      'Total realisasi monitoring disbursement untuk filter aktif.',
    ),
    metric(
      'absorption',
      'Penyerapan',
      portfolio?.absorption_pct ?? 0,
      'percent',
      undefined,
      'Realisasi dibagi rencana. Jika rencana 0, backend mengembalikan 0%.',
    ),
  ]
})
const institutionMetrics = computed<AnalyticsMoneyMetric[]>(() => {
  const summary = analytics.institutions.value?.summary
  const drilldown = analytics.institutions.value?.drilldown

  return [
    metric(
      'institution_count',
      'Kementerian/Lembaga',
      summary?.institution_count ?? 0,
      'number',
      drilldown,
      'Jumlah Kementerian/Lembaga yang memiliki proyek pada filter aktif.',
    ),
    metric(
      'project_count',
      'Total proyek',
      summary?.project_count ?? 0,
      'number',
      drilldown,
      'Proyek dihitung unik lintas revisi.',
    ),
    metric(
      'assignment_count',
      'Assignment',
      summary?.assignment_count ?? 0,
      'number',
      undefined,
      'Assignment menghitung relasi proyek ke Kementerian/Lembaga dan dapat overlap.',
    ),
    metric(
      'agreement_amount',
      'Loan Agreement USD',
      summary?.total_agreement_amount_usd ?? 0,
      'currency',
      undefined,
      'Nilai Loan Agreement untuk Kementerian/Lembaga pada filter aktif.',
    ),
    metric('planned', 'Rencana USD', summary?.total_planned_usd ?? 0, 'currency', undefined, 'Total rencana monitoring.'),
    metric('realized', 'Realisasi USD', summary?.total_realized_usd ?? 0, 'currency', undefined, 'Total realisasi monitoring.'),
    metric('absorption', 'Penyerapan', summary?.absorption_pct ?? 0, 'percent', undefined, 'Realisasi dibagi rencana.'),
  ]
})
const lenderMetrics = computed<AnalyticsMoneyMetric[]>(() => {
  const summary = analytics.lenders.value?.summary
  const drilldown = analytics.lenders.value?.drilldown

  return [
    metric(
      'lender_count',
      'Lender dengan Loan Agreement',
      summary?.lender_count ?? 0,
      'number',
      drilldown,
      'Jumlah lender yang sudah memiliki Loan Agreement pada filter aktif.',
    ),
    metric(
      'loan_agreement_count',
      'Loan Agreement',
      summary?.loan_agreement_count ?? 0,
      'number',
      drilldown,
      'Jumlah Loan Agreement yang cocok dengan filter aktif.',
    ),
    metric(
      'agreement_amount',
      'Loan Agreement USD',
      summary?.total_agreement_amount_usd ?? 0,
      'currency',
      undefined,
      'Nilai pinjaman dari Loan Agreement.',
    ),
    metric('planned', 'Rencana USD', summary?.total_planned_usd ?? 0, 'currency', undefined, 'Total rencana monitoring.'),
    metric('realized', 'Realisasi USD', summary?.total_realized_usd ?? 0, 'currency', undefined, 'Total realisasi monitoring.'),
    metric('absorption', 'Penyerapan', summary?.absorption_pct ?? 0, 'percent', undefined, 'Realisasi dibagi rencana.'),
  ]
})
const absorptionMetrics = computed<AnalyticsMoneyMetric[]>(() => {
  const summary = analytics.absorption.value?.summary
  const drilldown = analytics.absorption.value?.drilldown

  return [
    metric('planned', 'Rencana USD', summary?.planned_usd ?? 0, 'currency', drilldown, 'Total rencana monitoring.'),
    metric('realized', 'Realisasi USD', summary?.realized_usd ?? 0, 'currency', drilldown, 'Total realisasi monitoring.'),
    metric('absorption', 'Penyerapan', summary?.absorption_pct ?? 0, 'percent', drilldown, 'Realisasi dibagi rencana.'),
  ]
})
const yearlyMetrics = computed<AnalyticsMoneyMetric[]>(() => {
  const summary = analytics.yearly.value?.summary
  const drilldown = analytics.yearly.value?.drilldown

  return [
    metric('planned', 'Rencana USD', summary?.planned_usd ?? 0, 'currency', drilldown, 'Total rencana monitoring pada periode yang tampil.'),
    metric('realized', 'Realisasi USD', summary?.realized_usd ?? 0, 'currency', drilldown, 'Total realisasi monitoring pada periode yang tampil.'),
    metric('absorption', 'Penyerapan', summary?.absorption_pct ?? 0, 'percent', drilldown, 'Realisasi dibagi rencana.'),
    metric('loan_agreement_count', 'Loan Agreement', summary?.loan_agreement_count ?? 0, 'number', undefined, 'Jumlah Loan Agreement pada periode yang tampil.'),
    metric('project_count', 'Project aktif', summary?.project_count ?? 0, 'number', undefined, 'Jumlah proyek aktif pada periode yang tampil.'),
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
const activeAbsorptionChartItems = computed(() =>
  activeAbsorptionItems.value.filter((item) => item.planned_usd > 0),
)
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
  const rows = [...activeAbsorptionChartItems.value]
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
const loanAgreementLenderProportionStage = computed(
  () =>
    (analytics.lenderProportion.value?.by_stage ?? []).find(
      (stage) => stage.stage === 'Loan Agreement',
    ) ?? null,
)
const loanAgreementLenderProportionItems = computed(
  () =>
    loanAgreementLenderProportionStage.value?.items.filter(
      (item) => item.share_pct > 0 || item.amount_usd > 0 || item.project_count > 0,
    ) ?? [],
)
const loanAgreementLenderPieUsesProjectCount = computed(
  () =>
    loanAgreementLenderProportionItems.value.length > 0 &&
    loanAgreementLenderProportionItems.value.every((item) => item.share_pct <= 0) &&
    loanAgreementLenderProportionItems.value.some((item) => item.project_count > 0),
)
const loanAgreementLenderPieChartOption = computed(() => {
  const items = loanAgreementLenderProportionItems.value
  const useProjectCount = loanAgreementLenderPieUsesProjectCount.value

  return {
    color: ['#2563eb', '#64748b', '#d97706'],
    tooltip: {
      trigger: 'item',
      formatter: (params: unknown) => {
        const raw = recordFrom(params)
        const data = recordFrom(raw.data)
        const type = stringFrom(data.name)
        const item = items.find((candidate) => candidate.type === type)

        return [
          `<strong>${type}</strong>`,
          `${useProjectCount ? 'Proporsi jumlah project' : 'Proporsi nilai USD'}: ${numberFrom(raw.percent).toFixed(1)}%`,
          `Project: ${formatNumber(item?.project_count ?? 0)}`,
          `Nilai: ${formatUSD(item?.amount_usd ?? 0)}`,
        ].join('<br/>')
      },
    },
    legend: { orient: 'vertical', left: 0, top: 'middle' },
    series: [
      {
        name: 'Loan Agreement',
        type: 'pie',
        radius: ['44%', '72%'],
        center: ['60%', '50%'],
        avoidLabelOverlap: true,
        label: { formatter: '{b}: {d}%' },
        data: items.map((item) => ({
          name: item.type,
          value: useProjectCount ? item.project_count : item.share_pct,
        })),
      },
    ],
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
const activeResultCount = computed(() => analytics.overview.value?.portfolio.project_count)
const activeFilterSummary = computed(() => {
  const filters = analytics.appliedFilters
  const parts: string[] = []

  if (filters.budget_year) parts.push(`TA ${filters.budget_year}`)
  if (filters.quarter) parts.push(filters.quarter)
  if (filters.lender_ids.length > 0) parts.push(`${filters.lender_ids.length} lender`)
  if (filters.institution_ids.length > 0) parts.push(`${filters.institution_ids.length} Kementerian/Lembaga`)

  return parts.length > 0 ? parts.join(', ') : 'filter aktif'
})

function metric(
  key: string,
  label: string,
  value: number,
  format: AnalyticsMoneyMetric['format'] = 'number',
  drilldown?: DashboardDrilldownQuery,
  hint?: string,
  emphasis: AnalyticsMoneyMetric['emphasis'] = 'normal',
  actionLabel?: string,
): AnalyticsMoneyMetric {
  return {
    key,
    label,
    value,
    format,
    unit: format === 'currency' ? 'USD' : undefined,
    hint,
    emphasis,
    actionLabel,
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
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
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
      subtitle="Ringkasan portfolio, entitas, penyerapan, dan risiko berdasarkan filter aktif."
    />

    <DashboardAnalyticsFilterBar
      :model-value="analytics.draftFilters"
      :applied-filters="analytics.appliedFilters"
      :result-count="activeResultCount"
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
      class="flex gap-3 overflow-x-auto rounded-lg border border-surface-200 bg-white p-2"
      aria-label="Analytics tabs"
    >
      <div v-for="group in tabGroups" :key="group.label" class="flex shrink-0 items-center gap-2">
        <span class="px-1 text-[0.65rem] font-semibold uppercase text-surface-400">
          {{ group.label }}
        </span>
        <Button
          v-for="tab in group.tabs"
          :key="tab.key"
          :label="tab.label"
          :severity="activeTab === tab.key ? undefined : 'secondary'"
          :outlined="activeTab !== tab.key"
          size="small"
          @click="activeTab = tab.key"
        />
      </div>
    </nav>

    <p class="rounded-lg border border-surface-200 bg-white p-3 text-sm text-surface-600">
      Angka 0 berarti nilai tercatat nol. Panel kosong berarti tidak ada baris untuk
      {{ activeFilterSummary }}.
    </p>

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
        <DashboardPipelineFunnel
          :rows="analytics.overview.value?.pipeline_funnel ?? []"
          :loading="sectionLoading(['overview'])"
          :empty="portfolioRows.length === 0"
        />
        <AnalyticsBreakdownTable
          title="Pipeline Portfolio"
          description="Setiap tahap ditampilkan terpisah agar proyek yang sudah lanjut tetap bisa ditelusuri dari tahap asal."
          :columns="portfolioColumns"
          :rows="portfolioRows"
          :loading="sectionLoading(['overview'])"
          empty-title="Tidak ada proyek pada portfolio"
          :empty-description="`Tidak ada proyek untuk ${activeFilterSummary}. Longgarkan filter atau reset untuk melihat seluruh portfolio.`"
          @drilldown="handleDrilldown"
        >
          <template #empty-actions>
            <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
          </template>
        </AnalyticsBreakdownTable>
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
          title="Kementerian/Lembaga by Project"
          description="Jumlah proyek unik; detail assignment tersedia di tabel."
          :option="institutionProjectChartOption"
          :loading="sectionLoading(['institutions'])"
          :empty="topInstitutionsByProject.length === 0"
          :height="topInstitutionsByProject.length < 3 ? '12rem' : '20rem'"
          empty-title="Tidak ada Kementerian/Lembaga"
          :empty-description="`Tidak ada Kementerian/Lembaga untuk ${activeFilterSummary}. Coba kosongkan filter Lender atau tahun.`"
        >
          <template #empty-actions>
            <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
          </template>
        </AnalyticsChartPanel>
        <AnalyticsChartPanel
          title="Top Penyerapan Kementerian/Lembaga"
          description="Hanya menampilkan entitas yang sudah memiliki rencana monitoring."
          :option="institutionAbsorptionChartOption"
          :loading="sectionLoading(['institutions'])"
          :empty="topInstitutionsByAbsorption.length === 0"
          :height="topInstitutionsByAbsorption.length < 3 ? '12rem' : '20rem'"
          empty-title="Belum ada data penyerapan"
          empty-description="Tidak ada rencana monitoring pada Kementerian/Lembaga untuk filter aktif."
        >
          <template #empty-actions>
            <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
          </template>
        </AnalyticsChartPanel>
      </div>
      <AnalyticsBreakdownTable
        title="Performa Kementerian/Lembaga"
        description="Membandingkan cakupan proyek, nilai Loan Agreement, dan realisasi monitoring per Kementerian/Lembaga."
        :columns="institutionColumns"
        :rows="institutionRows"
        :loading="sectionLoading(['institutions'])"
        empty-title="Tidak ada Kementerian/Lembaga"
        :empty-description="`Tidak ada baris Kementerian/Lembaga untuk ${activeFilterSummary}.`"
        @drilldown="handleDrilldown"
      >
        <template #empty-actions>
          <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
        </template>
      </AnalyticsBreakdownTable>
    </section>

    <section v-else-if="activeTab === 'lenders'" class="space-y-4">
      <AnalyticsMetricGrid
        :metrics="lenderMetrics"
        :loading="sectionLoading(['lenders'])"
        @drilldown="handleDrilldown"
      />
      <AnalyticsChartPanel
        title="Top Lender by Loan Agreement USD"
        description="Menampilkan lender yang sudah tercatat pada Loan Agreement."
        :option="lenderPerformanceChartOption"
        :loading="sectionLoading(['lenders'])"
        :empty="topLendersByAmount.length === 0"
        :height="topLendersByAmount.length < 3 ? '12rem' : '20rem'"
        empty-title="Tidak ada Loan Agreement untuk lender"
        :empty-description="`Tidak ada lender dengan Loan Agreement untuk ${activeFilterSummary}. Proyek mungkin masih berada di tahap Daftar Kegiatan atau sebelumnya.`"
      >
        <template #empty-actions>
          <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
        </template>
      </AnalyticsChartPanel>
      <AnalyticsBreakdownTable
        title="Performa Lender Legal"
        description="Cakupan lender dihitung dari Loan Agreement dan monitoring yang sudah tercatat."
        :columns="lenderColumns"
        :rows="lenderRows"
        :loading="sectionLoading(['lenders'])"
        empty-title="Tidak ada lender legal"
        :empty-description="`Tidak ada Loan Agreement untuk lender pada ${activeFilterSummary}.`"
        @drilldown="handleDrilldown"
      >
        <template #empty-actions>
          <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
        </template>
      </AnalyticsBreakdownTable>
      <AnalyticsMatrixTable
        v-model:top-n="matrixTopN"
        :items="analytics.lenders.value?.lender_institution_matrix ?? []"
        :loading="sectionLoading(['lenders'])"
        @drilldown="handleDrilldown"
      />
      <div class="grid gap-4 xl:grid-cols-2">
        <AnalyticsChartPanel
          title="Pie Proporsi Lender Loan Agreement"
          description="Proporsi lender pada stage Loan Agreement."
          :option="loanAgreementLenderPieChartOption"
          :loading="sectionLoading(['lenderProportion'])"
          :empty="loanAgreementLenderProportionItems.length === 0"
          empty-title="Tidak ada proporsi Loan Agreement"
          empty-description="Belum ada lender pada stage Loan Agreement untuk filter aktif."
        >
          <template #empty-actions>
            <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
          </template>
        </AnalyticsChartPanel>
        <AnalyticsChartPanel
          title="Proporsi Lender per Stage"
          :option="lenderProportionChartOption"
          :loading="sectionLoading(['lenderProportion'])"
          :empty="lenderProportionRows.length === 0"
          empty-title="Tidak ada proporsi lender"
          description="Membedakan lender indikatif, funding source, agreement, dan monitoring."
          :empty-description="`Tidak ada data proporsi lender untuk ${activeFilterSummary}.`"
        >
          <template #empty-actions>
            <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
          </template>
        </AnalyticsChartPanel>
      </div>
      <AnalyticsBreakdownTable
        title="Detail Proporsi Lender"
        :columns="lenderProportionColumns"
        :rows="lenderProportionRows"
        :loading="sectionLoading(['lenderProportion'])"
        empty-title="Tidak ada detail proporsi lender"
        :empty-description="`Tidak ada proporsi lender untuk ${activeFilterSummary}.`"
        @drilldown="handleDrilldown"
      >
        <template #empty-actions>
          <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
        </template>
      </AnalyticsBreakdownTable>
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
          <p class="text-sm text-surface-500">Jika belum ada rencana monitoring, penyerapan ditampilkan 0% sesuai aturan backend.</p>
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
          :title="`Penyerapan Terendah ${activeAbsorptionLabel}`"
          description="Urutan dari penyerapan terendah agar anomali mudah dipindai."
          :option="absorptionChartOption"
          :loading="sectionLoading(['absorption'])"
          :empty="activeAbsorptionChartItems.length === 0"
          :height="activeAbsorptionChartItems.length < 3 ? '12rem' : '20rem'"
          empty-title="Belum ada rencana monitoring"
          :empty-description="`Tidak ada baris dengan rencana USD lebih dari 0 untuk ${activeFilterSummary}. Angka penyerapan 0% di KPI dapat berarti belum ada rencana, bukan performa rendah.`"
        >
          <template #empty-actions>
            <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
          </template>
        </AnalyticsChartPanel>
        <AnalyticsBreakdownTable
          title="Daftar Penyerapan Rendah"
          :columns="absorptionColumns"
          :rows="lowAbsorptionRows"
          :loading="sectionLoading(['absorption'])"
          empty-title="Tidak ada penyerapan rendah"
          :empty-description="`Tidak ada baris berstatus rendah untuk ${activeAbsorptionLabel} pada ${activeFilterSummary}.`"
          @drilldown="handleDrilldown"
        >
          <template #empty-actions>
            <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
          </template>
        </AnalyticsBreakdownTable>
      </div>
      <AnalyticsBreakdownTable
        :title="`Ranking Penyerapan ${activeAbsorptionLabel}`"
        :columns="absorptionColumns"
        :rows="absorptionRows"
        :loading="sectionLoading(['absorption'])"
        empty-title="Tidak ada ranking penyerapan"
        :empty-description="`Belum ada data penyerapan untuk ${activeAbsorptionLabel} pada ${activeFilterSummary}.`"
        @drilldown="handleDrilldown"
      >
        <template #empty-actions>
          <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
        </template>
      </AnalyticsBreakdownTable>
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
        title="Tren Rencana vs Realisasi"
        description="Rencana dan realisasi ditampilkan per tahun anggaran dan triwulan."
        :option="yearlyTrendChartOption"
        :loading="sectionLoading(['yearly'])"
        :empty="yearlyRows.length === 0"
        empty-title="Tidak ada tren tahunan"
        :empty-description="`Tidak ada monitoring tahunan untuk ${activeFilterSummary}.`"
      >
        <template #empty-actions>
          <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
        </template>
      </AnalyticsChartPanel>
      <AnalyticsBreakdownTable
        title="Detail Performa Tahunan"
        :columns="yearlyColumns"
        :rows="yearlyRows"
        :loading="sectionLoading(['yearly'])"
        empty-title="Tidak ada data tahunan"
        :empty-description="`Tidak ada baris monitoring tahunan untuk ${activeFilterSummary}.`"
        @drilldown="handleDrilldown"
      >
        <template #empty-actions>
          <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
        </template>
      </AnalyticsBreakdownTable>
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
        :description="`Tidak ada kartu risiko untuk ${activeFilterSummary}.`"
      >
        <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
      </AnalyticsEmptyState>
      <AnalyticsBreakdownTable
        title="Risk Watchlist"
        description="Memantau Loan Agreement, monitoring, dan risiko closing yang perlu ditindaklanjuti."
        :columns="riskWatchlistColumns"
        :rows="riskWatchlistRows"
        :loading="sectionLoading(['risks'])"
        empty-title="Tidak ada risiko aktif"
        :empty-description="`Tidak ada watchlist Loan Agreement atau monitoring untuk ${activeFilterSummary}.`"
        @drilldown="handleDrilldown"
      >
        <template #empty-actions>
          <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
        </template>
      </AnalyticsBreakdownTable>
      <AnalyticsBreakdownTable
        title="Belum berlanjut ke tahap berikutnya"
        description="Menunjukkan proyek yang belum masuk ke tahap lanjutan pada filter aktif."
        :columns="pipelineBottleneckColumns"
        :rows="pipelineBottleneckRows"
        :loading="sectionLoading(['risks'])"
        empty-title="Tidak ada bottleneck pipeline"
        :empty-description="`Tidak ada proyek yang tertahan untuk ${activeFilterSummary}.`"
        @drilldown="handleDrilldown"
      >
        <template #empty-actions>
          <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
        </template>
      </AnalyticsBreakdownTable>
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
        :empty-description="`Tidak ada isu kelengkapan data untuk ${activeFilterSummary}.`"
        @drilldown="handleDrilldown"
      >
        <template #empty-actions>
          <Button label="Reset filter" icon="pi pi-refresh" size="small" outlined @click="analytics.resetFilters" />
        </template>
      </AnalyticsBreakdownTable>
    </section>
  </section>
</template>
