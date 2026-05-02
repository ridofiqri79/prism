<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import Button from 'primevue/button'
import PageHeader from '@/components/common/PageHeader.vue'
import AnalyticsBreakdownTable from '@/components/dashboard/AnalyticsBreakdownTable.vue'
import AnalyticsEmptyState from '@/components/dashboard/AnalyticsEmptyState.vue'
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
  DashboardAnalyticsPipelineStage,
  DashboardDrilldownQuery,
} from '@/types/dashboard.types'

type DashboardAnalyticsTab =
  | 'portfolio'
  | 'institutions'
  | 'lenders'
  | 'absorption'
  | 'yearly'
  | 'risks'

const masterStore = useMasterStore()
const analytics = useDashboardAnalytics()
const activeTab = ref<DashboardAnalyticsTab>('portfolio')
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

const portfolioColumns: AnalyticsBreakdownTableColumn[] = [
  { key: 'stage', label: 'Stage' },
  { key: 'project_count', label: 'Project', kind: 'number', align: 'right' },
  { key: 'total_loan_usd', label: 'Nilai Pinjaman USD', kind: 'currency', align: 'right' },
  { key: 'drilldown', label: 'Drilldown', kind: 'drilldown', align: 'center' },
]
const institutionColumns: AnalyticsBreakdownTableColumn[] = [
  { key: 'institution', label: 'Kementerian/Lembaga' },
  { key: 'project_count', label: 'Project', kind: 'number', align: 'right' },
  { key: 'assignment_count', label: 'Assignment', kind: 'number', align: 'right' },
  { key: 'agreement_amount_usd', label: 'Agreement USD', kind: 'currency', align: 'right' },
  { key: 'absorption_pct', label: 'Penyerapan', kind: 'percent', align: 'right' },
  { key: 'drilldown', label: 'Drilldown', kind: 'drilldown', align: 'center' },
]
const lenderColumns: AnalyticsBreakdownTableColumn[] = [
  { key: 'lender', label: 'Lender' },
  { key: 'type', label: 'Tipe', kind: 'badge', align: 'center' },
  { key: 'project_count', label: 'Project', kind: 'number', align: 'right' },
  { key: 'agreement_amount_usd', label: 'Agreement USD', kind: 'currency', align: 'right' },
  { key: 'absorption_pct', label: 'Penyerapan', kind: 'percent', align: 'right' },
  { key: 'drilldown', label: 'Drilldown', kind: 'drilldown', align: 'center' },
]
const lenderProportionColumns: AnalyticsBreakdownTableColumn[] = [
  { key: 'stage', label: 'Stage' },
  { key: 'type', label: 'Tipe', kind: 'badge', align: 'center' },
  { key: 'project_count', label: 'Project', kind: 'number', align: 'right' },
  { key: 'lender_count', label: 'Lender', kind: 'number', align: 'right' },
  { key: 'amount_usd', label: 'Nilai USD', kind: 'currency', align: 'right' },
  { key: 'share_pct', label: 'Proporsi', kind: 'percent', align: 'right' },
  { key: 'drilldown', label: 'Drilldown', kind: 'drilldown', align: 'center' },
]
const absorptionColumns: AnalyticsBreakdownTableColumn[] = [
  { key: 'name', label: 'Nama' },
  { key: 'dimension', label: 'Dimensi', kind: 'badge', align: 'center' },
  { key: 'planned_usd', label: 'Rencana USD', kind: 'currency', align: 'right' },
  { key: 'realized_usd', label: 'Realisasi USD', kind: 'currency', align: 'right' },
  { key: 'absorption_pct', label: 'Penyerapan', kind: 'percent', align: 'right' },
  { key: 'status', label: 'Status', kind: 'badge', align: 'center' },
  { key: 'drilldown', label: 'Drilldown', kind: 'drilldown', align: 'center' },
]
const yearlyColumns: AnalyticsBreakdownTableColumn[] = [
  { key: 'period', label: 'Periode' },
  { key: 'planned_usd', label: 'Rencana USD', kind: 'currency', align: 'right' },
  { key: 'realized_usd', label: 'Realisasi USD', kind: 'currency', align: 'right' },
  { key: 'absorption_pct', label: 'Penyerapan', kind: 'percent', align: 'right' },
  { key: 'loan_agreement_count', label: 'Loan Agreement', kind: 'number', align: 'right' },
  { key: 'project_count', label: 'Project', kind: 'number', align: 'right' },
  { key: 'drilldown', label: 'Drilldown', kind: 'drilldown', align: 'center' },
]
const riskColumns: AnalyticsBreakdownTableColumn[] = [
  { key: 'label', label: 'Risk/Data Quality' },
  { key: 'stage', label: 'Stage' },
  { key: 'count', label: 'Jumlah', kind: 'number', align: 'right' },
  { key: 'severity', label: 'Severity', kind: 'badge', align: 'center' },
  { key: 'drilldown', label: 'Drilldown', kind: 'drilldown', align: 'center' },
]

const portfolioMetrics = computed<AnalyticsMoneyMetric[]>(() => {
  const portfolio = analytics.overview.value?.portfolio

  return [
    metric('project_count', 'Project logical', portfolio?.project_count ?? 0),
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

  return [
    metric('institution_count', 'Kementerian/Lembaga', summary?.institution_count ?? 0),
    metric('project_count', 'Project logical', summary?.project_count ?? 0),
    metric('assignment_count', 'Assignment', summary?.assignment_count ?? 0),
    metric(
      'agreement_amount',
      'Agreement USD',
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

  return [
    metric('lender_count', 'Lender legal', summary?.lender_count ?? 0),
    metric('loan_agreement_count', 'Loan Agreement', summary?.loan_agreement_count ?? 0),
    metric(
      'agreement_amount',
      'Agreement USD',
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

  return [
    metric('planned', 'Rencana USD', summary?.planned_usd ?? 0, 'currency'),
    metric('realized', 'Realisasi USD', summary?.realized_usd ?? 0, 'currency'),
    metric('absorption', 'Penyerapan', summary?.absorption_pct ?? 0, 'percent'),
  ]
})
const yearlyMetrics = computed<AnalyticsMoneyMetric[]>(() => {
  const summary = analytics.yearly.value?.summary

  return [
    metric('planned', 'Rencana USD', summary?.planned_usd ?? 0, 'currency'),
    metric('realized', 'Realisasi USD', summary?.realized_usd ?? 0, 'currency'),
    metric('absorption', 'Penyerapan', summary?.absorption_pct ?? 0, 'percent'),
    metric('loan_agreement_count', 'Loan Agreement', summary?.loan_agreement_count ?? 0),
    metric('project_count', 'Project', summary?.project_count ?? 0),
  ]
})
const riskMetrics = computed<AnalyticsMoneyMetric[]>(() =>
  (analytics.risks.value?.risk_cards ?? []).map((card) => ({
    key: card.code,
    label: card.label,
    value: card.amount_usd ?? card.count,
    format: card.amount_usd !== undefined ? 'currency' : 'number',
    unit: card.amount_usd !== undefined ? 'USD' : undefined,
    severity:
      card.severity === 'warning' ? 'warning' : card.severity === 'danger' ? 'danger' : 'info',
    drilldown: card.drilldown,
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
    cells: {
      institution: item.institution.short_name
        ? `${item.institution.name} (${item.institution.short_name})`
        : item.institution.name,
      project_count: item.project_count,
      assignment_count: item.assignment_count,
      agreement_amount_usd: item.agreement_amount_usd,
      absorption_pct: item.absorption_pct,
    },
    drilldown: item.drilldown,
  })),
)
const lenderRows = computed<AnalyticsBreakdownTableRow[]>(() =>
  (analytics.lenders.value?.items ?? []).map((item) => ({
    id: item.lender.id,
    severity: lenderSeverity(item.lender.type),
    cells: {
      lender: item.lender.short_name
        ? `${item.lender.name} (${item.lender.short_name})`
        : item.lender.name,
      type: item.lender.type,
      project_count: item.project_count,
      agreement_amount_usd: item.agreement_amount_usd,
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
const absorptionRows = computed<AnalyticsBreakdownTableRow[]>(() => [
  ...absorptionRankRows(analytics.absorption.value?.by_institution ?? [], 'Kementerian/Lembaga'),
  ...absorptionRankRows(analytics.absorption.value?.by_project ?? [], 'Project'),
  ...absorptionRankRows(analytics.absorption.value?.by_lender ?? [], 'Lender'),
])
const yearlyRows = computed<AnalyticsBreakdownTableRow[]>(() =>
  (analytics.yearly.value?.items ?? []).map((item) => ({
    id: `${item.budget_year}-${item.quarter}`,
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
const riskRows = computed<AnalyticsBreakdownTableRow[]>(() => [
  ...(analytics.risks.value?.watchlists.pipeline_bottlenecks ?? []).map((item) => ({
    id: `bottleneck-${item.stage}`,
    severity: item.severity,
    cells: {
      label: item.label,
      stage: stageLabel(item.stage),
      count: item.project_count,
      severity: item.severity,
    },
    drilldown: item.drilldown,
  })),
  ...(analytics.risks.value?.data_quality ?? []).map((item) => ({
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
])

function metric(
  key: string,
  label: string,
  value: number,
  format: AnalyticsMoneyMetric['format'] = 'number',
): AnalyticsMoneyMetric {
  return {
    key,
    label,
    value,
    format,
    unit: format === 'currency' ? 'USD' : undefined,
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

function absorptionRankRows(
  rows: NonNullable<typeof analytics.absorption.value>['by_institution'],
  dimension: string,
) {
  return rows.map((item) => ({
    id: `${dimension}-${item.id}`,
    severity: item.status === 'low' ? 'danger' : item.status === 'high' ? 'success' : 'info',
    cells: {
      name: item.name,
      dimension,
      planned_usd: item.planned_usd,
      realized_usd: item.realized_usd,
      absorption_pct: item.absorption_pct,
      status: item.status,
    },
    drilldown: item.drilldown,
  }))
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

onMounted(() => {
  void analytics.initialize()
})
</script>

<template>
  <section class="space-y-5">
    <PageHeader
      title="Dashboard Analytics"
      subtitle="Ringkasan portfolio pinjaman luar negeri, monitoring, risiko, dan data quality."
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
      v-if="sectionError(tabs.find((tab) => tab.key === activeTab)?.sections ?? [])"
      class="rounded-lg border border-red-200 bg-red-50 p-4 text-sm text-red-700"
    >
      {{ sectionError(tabs.find((tab) => tab.key === activeTab)?.sections ?? []) }}
    </section>

    <section v-if="activeTab === 'portfolio'" class="space-y-4">
      <AnalyticsMetricGrid
        :metrics="portfolioMetrics"
        :loading="sectionLoading(['overview'])"
        @drilldown="handleDrilldown"
      />
      <AnalyticsBreakdownTable
        title="Pipeline Portfolio"
        description="Default memakai latest snapshot agar revisi tidak double-count."
        :columns="portfolioColumns"
        :rows="portfolioRows"
        :loading="sectionLoading(['overview'])"
        @drilldown="handleDrilldown"
      />
    </section>

    <section v-else-if="activeTab === 'institutions'" class="space-y-4">
      <AnalyticsMetricGrid
        :metrics="institutionMetrics"
        :loading="sectionLoading(['institutions'])"
        @drilldown="handleDrilldown"
      />
      <AnalyticsBreakdownTable
        title="Distribusi Kementerian/Lembaga"
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
      <AnalyticsBreakdownTable
        title="Performa Lender Legal"
        description="Basis lender pada tab ini adalah Loan Agreement dan Monitoring."
        :columns="lenderColumns"
        :rows="lenderRows"
        :loading="sectionLoading(['lenders'])"
        @drilldown="handleDrilldown"
      />
      <AnalyticsBreakdownTable
        title="Proporsi Lender per Stage"
        description="Stage lender ditampilkan eksplisit agar indikasi, funding source, dan legal agreement tidak tercampur."
        :columns="lenderProportionColumns"
        :rows="lenderProportionRows"
        :loading="sectionLoading(['lenderProportion'])"
        @drilldown="handleDrilldown"
      />
    </section>

    <section v-else-if="activeTab === 'absorption'" class="space-y-4">
      <AnalyticsMetricGrid
        :metrics="absorptionMetrics"
        :loading="sectionLoading(['absorption'])"
        @drilldown="handleDrilldown"
      />
      <AnalyticsBreakdownTable
        title="Penyerapan"
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
      <AnalyticsBreakdownTable
        title="Performa Tahunan"
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
        title="Risiko & Data Quality"
        :columns="riskColumns"
        :rows="riskRows"
        :loading="sectionLoading(['risks'])"
        @drilldown="handleDrilldown"
      />
    </section>
  </section>
</template>
