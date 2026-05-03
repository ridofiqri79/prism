import http from '@/services/http'
import type { ApiResponse } from '@/types/api.types'
import type {
  DashboardFilterOptions,
  DashboardFilterParams,
  DataQualityGovernanceDashboard,
  DataQualityGovernanceParams,
  DashboardSummary,
  DashboardSummaryApiResponse,
  ExecutivePortfolioDashboard,
  GreenBookReadinessDashboard,
  GreenBookReadinessParams,
  JourneyResponse,
  KLPortfolioPerformanceDashboard,
  KLPortfolioPerformanceParams,
  LADisbursementDashboard,
  LADisbursementParams,
  LenderFinancingMixDashboard,
  LenderFinancingMixParams,
  MonitoringSummary,
  MonitoringSummaryApiResponse,
  PipelineBottleneckApiResponse,
  PipelineBottleneckParams,
  StageMetric,
  TimeSeriesPoint,
} from '@/types/dashboard.types'

function normalizeSummary(data: DashboardSummaryApiResponse): DashboardSummary {
  const realized =
    data.total_realized_usd ??
    data.total_realisasi_usd ??
    data.realized_disbursement_usd ??
    0

  return {
    total_bb_projects: data.total_bb_projects ?? 0,
    total_gb_projects: data.total_gb_projects ?? 0,
    total_loan_agreements: data.total_loan_agreements ?? 0,
    total_amount_usd:
      data.total_amount_usd ??
      data.la_commitment_usd ??
      data.gb_pipeline_usd ??
      data.bb_pipeline_usd ??
      0,
    total_realized_usd: realized,
    total_realisasi_usd: realized,
    overall_absorption_pct: data.overall_absorption_pct ?? data.absorption_pct ?? 0,
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

export const DashboardService = {
  async getSummary(params?: DashboardFilterParams) {
    const response = await http.get<ApiResponse<DashboardSummaryApiResponse>>('/dashboard/summary', {
      params,
    })
    return normalizeSummary(response.data.data)
  },

  async getStageFunnel(params?: DashboardFilterParams) {
    const response = await http.get<ApiResponse<StageMetric[]>>('/dashboard/stage-funnel', {
      params,
    })
    return response.data.data
  },

  async getMonitoringRollup(params?: DashboardFilterParams) {
    const response = await http.get<ApiResponse<TimeSeriesPoint[]>>('/dashboard/monitoring-rollup', {
      params,
    })
    return response.data.data
  },

  async getMonitoringSummary(params?: DashboardFilterParams) {
    const response = await http.get<ApiResponse<MonitoringSummaryApiResponse>>(
      '/dashboard/monitoring-summary',
      { params },
    )
    return normalizeMonitoringSummary(response.data.data)
  },

  async getFilterOptions() {
    const response = await http.get<ApiResponse<DashboardFilterOptions>>('/dashboard/filter-options')
    return response.data.data
  },

  async getExecutivePortfolio(params?: DashboardFilterParams) {
    const response = await http.get<ApiResponse<ExecutivePortfolioDashboard>>(
      '/dashboard/executive-portfolio',
      { params },
    )
    return response.data.data
  },

  async getPipelineBottleneck(params?: PipelineBottleneckParams) {
    const response = await http.get<PipelineBottleneckApiResponse>(
      '/dashboard/pipeline-bottleneck',
      { params },
    )
    return response.data
  },

  async getGreenBookReadiness(params?: GreenBookReadinessParams) {
    const response = await http.get<ApiResponse<GreenBookReadinessDashboard>>(
      '/dashboard/green-book-readiness',
      { params },
    )
    return response.data.data
  },

  async getLenderFinancingMix(params?: LenderFinancingMixParams) {
    const response = await http.get<ApiResponse<LenderFinancingMixDashboard>>(
      '/dashboard/lender-financing-mix',
      { params },
    )
    return response.data.data
  },

  async getKLPortfolioPerformance(params?: KLPortfolioPerformanceParams) {
    const response = await http.get<ApiResponse<KLPortfolioPerformanceDashboard>>(
      '/dashboard/kl-portfolio-performance',
      { params },
    )
    return response.data.data
  },

  async getLADisbursement(params?: LADisbursementParams) {
    const response = await http.get<ApiResponse<LADisbursementDashboard>>(
      '/dashboard/la-disbursement',
      { params },
    )
    return response.data.data
  },

  async getDataQualityGovernance(params?: DataQualityGovernanceParams) {
    const response = await http.get<ApiResponse<DataQualityGovernanceDashboard>>(
      '/dashboard/data-quality-governance',
      { params },
    )
    return response.data.data
  },

  async getJourney(bbProjectId: string) {
    const response = await http.get<ApiResponse<JourneyResponse>>(`/projects/${bbProjectId}/journey`)
    return response.data.data
  },
}
