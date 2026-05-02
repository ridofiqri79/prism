import http from '@/services/http'
import type { ApiResponse } from '@/types/api.types'
import type {
  AnalyticsStageBreakdown,
  DashboardAbsorptionAnalytics,
  DashboardAbsorptionRankedItem,
  DashboardAnalyticsFilterParams,
  DashboardAnalyticsInsight,
  DashboardAnalyticsLenderRef,
  DashboardAnalyticsOverview,
  DashboardAnalyticsPipelineFunnelItem,
  DashboardAnalyticsPipelineStage,
  DashboardAnalyticsSeverity,
  DashboardDataQualityItem,
  DashboardDrilldownQuery,
  DashboardExtendedLoanBreakdown,
  DashboardFilterParams,
  DashboardInstitutionAnalytics,
  DashboardInstitutionAnalyticsItem,
  DashboardInstitutionAnalyticsSummary,
  DashboardLenderAnalytics,
  DashboardLenderAnalyticsItem,
  DashboardLenderInstitutionMatrixItem,
  DashboardLenderProportionAnalytics,
  DashboardLenderProportionItem,
  DashboardLoanAgreementRiskItem,
  DashboardPipelineBottleneckItem,
  DashboardQuarter,
  DashboardRiskAnalytics,
  DashboardRiskCard,
  DashboardRiskSummary,
  DashboardRiskThresholds,
  DashboardSummary,
  DashboardSummaryApiResponse,
  JourneyResponse,
  MonitoringSummary,
  MonitoringSummaryApiResponse,
} from '@/types/dashboard.types'
import type { InstitutionLevel, LenderType } from '@/types/master.types'

type UnknownRecord = Record<string, unknown>

const emptyDrilldown: DashboardDrilldownQuery = {
  target: 'projects',
  query: {},
}

function isRecord(value: unknown): value is UnknownRecord {
  return typeof value === 'object' && value !== null && !Array.isArray(value)
}

function recordFrom(value: unknown): UnknownRecord {
  return isRecord(value) ? value : {}
}

function arrayFrom<T>(value: unknown, mapper: (item: UnknownRecord) => T): T[] {
  if (!Array.isArray(value)) return []

  return value.map((item) => mapper(recordFrom(item)))
}

function numberFrom(value: unknown): number {
  if (typeof value === 'number' && Number.isFinite(value)) return value
  if (typeof value === 'string' && value.trim() !== '') {
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : 0
  }
  return 0
}

function stringFrom(value: unknown): string {
  return typeof value === 'string' ? value : ''
}

function optionalString(value: unknown): string | undefined {
  const text = stringFrom(value)
  return text ? text : undefined
}

function quarterFrom(value: unknown): DashboardQuarter | undefined {
  return value === 'TW1' || value === 'TW2' || value === 'TW3' || value === 'TW4'
    ? value
    : undefined
}

function lenderTypeFrom(value: unknown): LenderType {
  return value === 'Bilateral' || value === 'Multilateral' || value === 'KSA' ? value : 'Bilateral'
}

function pipelineStageFrom(value: unknown): DashboardAnalyticsPipelineStage {
  return value === 'BB' ||
    value === 'GB' ||
    value === 'DK' ||
    value === 'LA' ||
    value === 'Monitoring'
    ? value
    : 'BB'
}

function severityFrom(value: unknown): DashboardAnalyticsSeverity | string {
  return stringFrom(value) || 'secondary'
}

function normalizeQuery(value: unknown): Record<string, string[]> {
  const raw = recordFrom(value)
  const query: Record<string, string[]> = {}

  Object.entries(raw).forEach(([key, item]) => {
    if (Array.isArray(item)) {
      query[key] = item.map(String).filter(Boolean)
      return
    }

    if (item !== undefined && item !== null && String(item) !== '') {
      query[key] = [String(item)]
    }
  })

  return query
}

function normalizeDrilldown(value: unknown, fallbackTarget = 'projects'): DashboardDrilldownQuery {
  const raw = recordFrom(value)

  return {
    target: stringFrom(raw.target) || fallbackTarget,
    query: normalizeQuery(raw.query),
  }
}

function normalizeLenderRef(value: unknown): DashboardAnalyticsLenderRef {
  const raw = recordFrom(value)

  return {
    id: stringFrom(raw.id),
    name: stringFrom(raw.name),
    short_name: optionalString(raw.short_name),
    type: lenderTypeFrom(raw.type),
  }
}

function normalizeInstitutionRef(value: unknown) {
  const raw = recordFrom(value)

  return {
    id: stringFrom(raw.id),
    name: stringFrom(raw.name),
    short_name: optionalString(raw.short_name),
    level: stringFrom(raw.level) as InstitutionLevel,
  }
}

function normalizeStageBreakdown(value: unknown): AnalyticsStageBreakdown {
  const raw = recordFrom(value)

  return {
    BB: numberFrom(raw.BB),
    GB: numberFrom(raw.GB),
    DK: numberFrom(raw.DK),
    LA: numberFrom(raw.LA),
    Monitoring: numberFrom(raw.Monitoring),
  }
}

function normalizeSummary(data: DashboardSummaryApiResponse): DashboardSummary {
  const realized = data.total_realized_usd ?? data.total_realisasi_usd ?? 0

  return {
    total_bb_projects: data.total_bb_projects ?? 0,
    total_gb_projects: data.total_gb_projects ?? 0,
    total_loan_agreements: data.total_loan_agreements ?? 0,
    total_amount_usd: data.total_amount_usd ?? 0,
    total_realized_usd: realized,
    total_realisasi_usd: realized,
    overall_absorption_pct: data.overall_absorption_pct ?? 0,
    active_monitoring: data.active_monitoring ?? 0,
  }
}

function normalizeMonitoringSummary(data: MonitoringSummaryApiResponse): MonitoringSummary {
  const planned = data.total_planned_usd ?? data.total_rencana_usd ?? 0
  const realized = data.total_realized_usd ?? data.total_realisasi_usd ?? 0

  return {
    budget_year: data.budget_year ?? data.tahun_anggaran,
    tahun_anggaran: data.tahun_anggaran ?? data.budget_year,
    quarter: data.quarter ?? data.triwulan,
    triwulan: data.triwulan ?? data.quarter,
    total_planned_usd: planned,
    total_rencana_usd: planned,
    total_realized_usd: realized,
    total_realisasi_usd: realized,
    absorption_pct: data.absorption_pct ?? 0,
    by_lender: (data.by_lender ?? []).map((item) => {
      const lenderPlanned = item.planned_usd ?? item.rencana_usd ?? 0
      const lenderRealized = item.realized_usd ?? item.realisasi_usd ?? 0

      return {
        lender: item.lender,
        planned_usd: lenderPlanned,
        rencana_usd: lenderPlanned,
        realized_usd: lenderRealized,
        realisasi_usd: lenderRealized,
        absorption_pct: item.absorption_pct ?? 0,
      }
    }),
  }
}

function normalizeOverview(data: unknown): DashboardAnalyticsOverview {
  const raw = recordFrom(data)
  const portfolio = recordFrom(raw.portfolio)

  return {
    portfolio: {
      project_count: numberFrom(portfolio.project_count),
      assignment_count: numberFrom(portfolio.assignment_count),
      total_pipeline_loan_usd: numberFrom(portfolio.total_pipeline_loan_usd),
      total_agreement_amount_usd: numberFrom(portfolio.total_agreement_amount_usd),
      total_planned_usd: numberFrom(portfolio.total_planned_usd),
      total_realized_usd: numberFrom(portfolio.total_realized_usd),
      absorption_pct: numberFrom(portfolio.absorption_pct),
    },
    pipeline_funnel: arrayFrom(
      raw.pipeline_funnel,
      (item): DashboardAnalyticsPipelineFunnelItem => ({
        stage: pipelineStageFrom(item.stage),
        project_count: numberFrom(item.project_count),
        total_loan_usd: numberFrom(item.total_loan_usd),
        drilldown: normalizeDrilldown(item.drilldown),
      }),
    ),
    top_insights: arrayFrom(
      raw.top_insights,
      (item): DashboardAnalyticsInsight => ({
        key: stringFrom(item.key),
        label: stringFrom(item.label),
        value: numberFrom(item.value),
        severity: severityFrom(item.severity),
        drilldown: normalizeDrilldown(item.drilldown),
      }),
    ),
    drilldown: normalizeDrilldown(raw.drilldown),
  }
}

function normalizeInstitutionSummary(value: unknown): DashboardInstitutionAnalyticsSummary {
  const raw = recordFrom(value)

  return {
    institution_count: numberFrom(raw.institution_count),
    project_count: numberFrom(raw.project_count),
    assignment_count: numberFrom(raw.assignment_count),
    total_agreement_amount_usd: numberFrom(raw.total_agreement_amount_usd),
    total_planned_usd: numberFrom(raw.total_planned_usd),
    total_realized_usd: numberFrom(raw.total_realized_usd),
    absorption_pct: numberFrom(raw.absorption_pct),
  }
}

function normalizeInstitutionAnalytics(data: unknown): DashboardInstitutionAnalytics {
  const raw = recordFrom(data)

  return {
    summary: normalizeInstitutionSummary(raw.summary),
    items: arrayFrom(
      raw.items,
      (item): DashboardInstitutionAnalyticsItem => ({
        institution: normalizeInstitutionRef(item.institution),
        project_count: numberFrom(item.project_count),
        assignment_count: numberFrom(item.assignment_count),
        loan_agreement_count: numberFrom(item.loan_agreement_count),
        monitoring_count: numberFrom(item.monitoring_count),
        agreement_amount_usd: numberFrom(item.agreement_amount_usd),
        planned_usd: numberFrom(item.planned_usd),
        realized_usd: numberFrom(item.realized_usd),
        absorption_pct: numberFrom(item.absorption_pct),
        pipeline_breakdown: normalizeStageBreakdown(item.pipeline_breakdown),
        drilldown: normalizeDrilldown(item.drilldown),
      }),
    ),
    drilldown: normalizeDrilldown(raw.drilldown),
  }
}

function normalizeLenderSummary(value: unknown) {
  const raw = recordFrom(value)

  return {
    lender_count: numberFrom(raw.lender_count),
    loan_agreement_count: numberFrom(raw.loan_agreement_count),
    total_agreement_amount_usd: numberFrom(raw.total_agreement_amount_usd),
    total_planned_usd: numberFrom(raw.total_planned_usd),
    total_realized_usd: numberFrom(raw.total_realized_usd),
    absorption_pct: numberFrom(raw.absorption_pct),
  }
}

function normalizeLenderAnalytics(data: unknown): DashboardLenderAnalytics {
  const raw = recordFrom(data)

  return {
    summary: normalizeLenderSummary(raw.summary),
    items: arrayFrom(
      raw.items,
      (item): DashboardLenderAnalyticsItem => ({
        lender: normalizeLenderRef(item.lender),
        loan_agreement_count: numberFrom(item.loan_agreement_count),
        project_count: numberFrom(item.project_count),
        institution_count: numberFrom(item.institution_count),
        monitoring_count: numberFrom(item.monitoring_count),
        agreement_amount_usd: numberFrom(item.agreement_amount_usd),
        planned_usd: numberFrom(item.planned_usd),
        realized_usd: numberFrom(item.realized_usd),
        absorption_pct: numberFrom(item.absorption_pct),
        drilldown: normalizeDrilldown(item.drilldown),
      }),
    ),
    lender_institution_matrix: arrayFrom(
      raw.lender_institution_matrix,
      (item): DashboardLenderInstitutionMatrixItem => ({
        institution: normalizeInstitutionRef(item.institution),
        lender: normalizeLenderRef(item.lender),
        project_count: numberFrom(item.project_count),
        loan_agreement_count: numberFrom(item.loan_agreement_count),
        monitoring_count: numberFrom(item.monitoring_count),
        agreement_amount_usd: numberFrom(item.agreement_amount_usd),
        planned_usd: numberFrom(item.planned_usd),
        realized_usd: numberFrom(item.realized_usd),
        absorption_pct: numberFrom(item.absorption_pct),
        drilldown: normalizeDrilldown(item.drilldown),
      }),
    ),
    drilldown: normalizeDrilldown(raw.drilldown),
  }
}

function normalizeAbsorptionItem(item: UnknownRecord): DashboardAbsorptionRankedItem {
  return {
    rank: numberFrom(item.rank) || undefined,
    id: stringFrom(item.id),
    name: stringFrom(item.name),
    dimension: stringFrom(item.dimension),
    planned_usd: numberFrom(item.planned_usd),
    realized_usd: numberFrom(item.realized_usd),
    absorption_pct: numberFrom(item.absorption_pct),
    variance_usd: numberFrom(item.variance_usd),
    status: stringFrom(item.status) || 'normal',
    drilldown: normalizeDrilldown(item.drilldown, 'monitoring'),
  }
}

function normalizeAbsorptionAnalytics(data: unknown): DashboardAbsorptionAnalytics {
  const raw = recordFrom(data)
  const summary = recordFrom(raw.summary)

  return {
    summary: {
      planned_usd: numberFrom(summary.planned_usd),
      realized_usd: numberFrom(summary.realized_usd),
      absorption_pct: numberFrom(summary.absorption_pct),
    },
    by_institution: arrayFrom(raw.by_institution, normalizeAbsorptionItem),
    by_project: arrayFrom(raw.by_project, normalizeAbsorptionItem),
    by_lender: arrayFrom(raw.by_lender, normalizeAbsorptionItem),
    drilldown: normalizeDrilldown(raw.drilldown, 'monitoring'),
  }
}

function normalizeYearlyAnalytics(data: unknown) {
  const raw = recordFrom(data)
  const summary = recordFrom(raw.summary)

  return {
    summary: {
      planned_usd: numberFrom(summary.planned_usd),
      realized_usd: numberFrom(summary.realized_usd),
      absorption_pct: numberFrom(summary.absorption_pct),
      loan_agreement_count: numberFrom(summary.loan_agreement_count),
      project_count: numberFrom(summary.project_count),
    },
    items: arrayFrom(raw.items, (item) => ({
      budget_year: numberFrom(item.budget_year),
      quarter: quarterFrom(item.quarter) ?? 'TW1',
      planned_usd: numberFrom(item.planned_usd),
      realized_usd: numberFrom(item.realized_usd),
      absorption_pct: numberFrom(item.absorption_pct),
      loan_agreement_count: numberFrom(item.loan_agreement_count),
      project_count: numberFrom(item.project_count),
      drilldown: normalizeDrilldown(item.drilldown, 'monitoring'),
    })),
    drilldown: normalizeDrilldown(raw.drilldown, 'monitoring'),
  }
}

function normalizeLenderProportion(data: unknown): DashboardLenderProportionAnalytics {
  const raw = recordFrom(data)

  return {
    by_stage: arrayFrom(raw.by_stage, (stage) => ({
      stage: stringFrom(stage.stage),
      items: arrayFrom(
        stage.items,
        (item): DashboardLenderProportionItem => ({
          type: lenderTypeFrom(item.type),
          project_count: numberFrom(item.project_count),
          lender_count: numberFrom(item.lender_count),
          amount_usd: numberFrom(item.amount_usd),
          share_pct: numberFrom(item.share_pct),
          drilldown: normalizeDrilldown(item.drilldown),
        }),
      ),
    })),
    drilldown: normalizeDrilldown(raw.drilldown),
  }
}

function normalizeRiskSummary(value: unknown): DashboardRiskSummary {
  const raw = recordFrom(value)

  return {
    low_absorption_count: numberFrom(raw.low_absorption_count),
    effective_without_monitoring_count: numberFrom(raw.effective_without_monitoring_count),
    closing_risk_count: numberFrom(raw.closing_risk_count),
    extended_loan_count: numberFrom(raw.extended_loan_count),
    data_quality_issue_count: numberFrom(raw.data_quality_issue_count),
    bottleneck_project_count: numberFrom(raw.bottleneck_project_count),
  }
}

function normalizeRiskThresholds(value: unknown): DashboardRiskThresholds {
  const raw = recordFrom(value)

  return {
    low_absorption_threshold: numberFrom(raw.low_absorption_threshold),
    closing_months_threshold: numberFrom(raw.closing_months_threshold),
    closing_absorption_threshold: numberFrom(raw.closing_absorption_threshold),
    stale_monitoring_quarters: numberFrom(raw.stale_monitoring_quarters),
  }
}

function normalizeLoanAgreementRiskItem(item: UnknownRecord): DashboardLoanAgreementRiskItem {
  return {
    risk_code: stringFrom(item.risk_code),
    risk_label: optionalString(item.risk_label),
    severity: severityFrom(item.severity),
    project_id: stringFrom(item.project_id),
    project_name: stringFrom(item.project_name),
    loan_agreement_id: stringFrom(item.loan_agreement_id),
    loan_code: stringFrom(item.loan_code),
    lender: normalizeLenderRef(item.lender),
    institution: isRecord(item.institution) ? normalizeInstitutionRef(item.institution) : undefined,
    effective_date: optionalString(item.effective_date),
    original_closing_date: optionalString(item.original_closing_date),
    closing_date: optionalString(item.closing_date),
    budget_year:
      item.budget_year === undefined || item.budget_year === null
        ? undefined
        : numberFrom(item.budget_year),
    quarter: quarterFrom(item.quarter),
    planned_usd: numberFrom(item.planned_usd),
    realized_usd: numberFrom(item.realized_usd),
    absorption_pct: numberFrom(item.absorption_pct),
    agreement_amount_usd: numberFrom(item.agreement_amount_usd),
    days_since_effective: numberFrom(item.days_since_effective),
    days_to_closing: numberFrom(item.days_to_closing),
    months_to_closing: numberFrom(item.months_to_closing),
    extension_days: numberFrom(item.extension_days),
    stale_quarters: numberFrom(item.stale_quarters),
    monitoring_status: optionalString(item.monitoring_status),
    drilldown: normalizeDrilldown(item.drilldown, 'monitoring'),
  }
}

function normalizeRiskAnalytics(data: unknown): DashboardRiskAnalytics {
  const raw = recordFrom(data)
  const watchlists = recordFrom(raw.watchlists)
  const extendedInsight = recordFrom(raw.extended_loan_insight)

  return {
    summary: normalizeRiskSummary(raw.summary),
    thresholds: normalizeRiskThresholds(raw.thresholds),
    risk_cards: arrayFrom(
      raw.risk_cards,
      (item): DashboardRiskCard => ({
        code: stringFrom(item.code),
        label: stringFrom(item.label),
        count: numberFrom(item.count),
        severity: severityFrom(item.severity),
        amount_usd: item.amount_usd === undefined ? undefined : numberFrom(item.amount_usd),
        drilldown: normalizeDrilldown(item.drilldown),
      }),
    ),
    watchlists: {
      low_absorption_projects: arrayFrom(
        watchlists.low_absorption_projects,
        normalizeLoanAgreementRiskItem,
      ),
      effective_without_monitoring: arrayFrom(
        watchlists.effective_without_monitoring,
        normalizeLoanAgreementRiskItem,
      ),
      closing_risks: arrayFrom(watchlists.closing_risks, normalizeLoanAgreementRiskItem),
      extended_loans: arrayFrom(watchlists.extended_loans, normalizeLoanAgreementRiskItem),
      pipeline_bottlenecks: arrayFrom(
        watchlists.pipeline_bottlenecks,
        (item): DashboardPipelineBottleneckItem => ({
          stage: pipelineStageFrom(item.stage),
          label: stringFrom(item.label),
          project_count: numberFrom(item.project_count),
          total_loan_usd: numberFrom(item.total_loan_usd),
          oldest_date: optionalString(item.oldest_date),
          severity: severityFrom(item.severity),
          drilldown: normalizeDrilldown(item.drilldown),
        }),
      ),
    },
    extended_loan_insight: {
      count: numberFrom(extendedInsight.count),
      amount_usd: numberFrom(extendedInsight.amount_usd),
      average_extension_days: numberFrom(extendedInsight.average_extension_days),
      by_lender: arrayFrom(extendedInsight.by_lender, normalizeExtendedLoanBreakdown),
      by_institution: arrayFrom(extendedInsight.by_institution, normalizeExtendedLoanBreakdown),
    },
    data_quality: arrayFrom(
      raw.data_quality,
      (item): DashboardDataQualityItem => ({
        code: stringFrom(item.code),
        label: stringFrom(item.label),
        stage: stringFrom(item.stage),
        severity: severityFrom(item.severity),
        count: numberFrom(item.count),
        drilldown: normalizeDrilldown(item.drilldown),
      }),
    ),
    drilldown: normalizeDrilldown(raw.drilldown),
  }
}

function normalizeExtendedLoanBreakdown(item: UnknownRecord): DashboardExtendedLoanBreakdown {
  return {
    dimension: stringFrom(item.dimension),
    entity:
      stringFrom(item.dimension) === 'institution'
        ? normalizeInstitutionRef(item.entity)
        : normalizeLenderRef(item.entity),
    loan_agreement_count: numberFrom(item.loan_agreement_count),
    amount_usd: numberFrom(item.amount_usd),
    average_extension_days: numberFrom(item.average_extension_days),
    drilldown: normalizeDrilldown(item.drilldown, 'loan_agreements'),
  }
}

async function getAnalytics<T>(
  endpoint: string,
  params: DashboardAnalyticsFilterParams | undefined,
  normalize: (data: unknown) => T,
) {
  const response = await http.get<ApiResponse<unknown>>(endpoint, { params })

  return normalize(response.data.data)
}

export const DashboardService = {
  async getSummary() {
    const response = await http.get<ApiResponse<DashboardSummaryApiResponse>>('/dashboard/summary')
    return normalizeSummary(response.data.data)
  },

  async getMonitoringSummary(params?: DashboardFilterParams) {
    const response = await http.get<ApiResponse<MonitoringSummaryApiResponse>>(
      '/dashboard/monitoring-summary',
      { params },
    )
    return normalizeMonitoringSummary(response.data.data)
  },

  async getJourney(bbProjectId: string) {
    const response = await http.get<ApiResponse<JourneyResponse>>(
      `/projects/${bbProjectId}/journey`,
    )
    return response.data.data
  },

  getAnalyticsOverview(params?: DashboardAnalyticsFilterParams) {
    return getAnalytics('/dashboard/analytics/overview', params, normalizeOverview)
  },

  getAnalyticsInstitutions(params?: DashboardAnalyticsFilterParams) {
    return getAnalytics('/dashboard/analytics/institutions', params, normalizeInstitutionAnalytics)
  },

  getAnalyticsLenders(params?: DashboardAnalyticsFilterParams) {
    return getAnalytics('/dashboard/analytics/lenders', params, normalizeLenderAnalytics)
  },

  getAnalyticsAbsorption(params?: DashboardAnalyticsFilterParams) {
    return getAnalytics('/dashboard/analytics/absorption', params, normalizeAbsorptionAnalytics)
  },

  getAnalyticsYearly(params?: DashboardAnalyticsFilterParams) {
    return getAnalytics('/dashboard/analytics/yearly', params, normalizeYearlyAnalytics)
  },

  getAnalyticsLenderProportion(params?: DashboardAnalyticsFilterParams) {
    return getAnalytics('/dashboard/analytics/lender-proportion', params, normalizeLenderProportion)
  },

  getAnalyticsRisks(params?: DashboardAnalyticsFilterParams) {
    return getAnalytics('/dashboard/analytics/risks', params, normalizeRiskAnalytics)
  },
}
