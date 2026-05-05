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
  LenderFinancingMixDashboard,
  LenderFinancingMixParams,
  PipelineBottleneckApiResponse,
  PipelineBottleneckParams,
  StageMetric,
} from '@/types/dashboard.types'

function normalizeSummary(data: DashboardSummaryApiResponse): DashboardSummary {
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
