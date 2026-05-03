import { ref } from 'vue'
import { defineStore } from 'pinia'
import { DashboardService } from '@/services/dashboard.service'
import type { PaginationMeta } from '@/types/api.types'
import type {
  DashboardFilterOptions,
  DashboardFilterParams,
  DashboardSummary,
  DataQualityGovernanceDashboard,
  DataQualityGovernanceParams,
  ExecutivePortfolioDashboard,
  GreenBookReadinessDashboard,
  GreenBookReadinessParams,
  KLPortfolioPerformanceDashboard,
  KLPortfolioPerformanceParams,
  LADisbursementDashboard,
  LADisbursementParams,
  LenderFinancingMixDashboard,
  LenderFinancingMixParams,
  PipelineBottleneckParams,
  PipelineBottleneckResponse,
  StageMetric,
  TimeSeriesPoint,
} from '@/types/dashboard.types'

export const useDashboardStore = defineStore('dashboard', () => {
  const summary = ref<DashboardSummary | null>(null)
  const stageFunnel = ref<StageMetric[]>([])
  const monitoringRollup = ref<TimeSeriesPoint[]>([])
  const executivePortfolio = ref<ExecutivePortfolioDashboard | null>(null)
  const pipelineBottleneck = ref<PipelineBottleneckResponse | null>(null)
  const pipelineBottleneckMeta = ref<PaginationMeta | null>(null)
  const greenBookReadiness = ref<GreenBookReadinessDashboard | null>(null)
  const lenderFinancingMix = ref<LenderFinancingMixDashboard | null>(null)
  const klPortfolioPerformance = ref<KLPortfolioPerformanceDashboard | null>(null)
  const laDisbursement = ref<LADisbursementDashboard | null>(null)
  const dataQualityGovernance = ref<DataQualityGovernanceDashboard | null>(null)
  const filterOptions = ref<DashboardFilterOptions>({})
  const loading = ref(false)
  const summaryLoading = ref(false)
  const stageFunnelLoading = ref(false)
  const monitoringRollupLoading = ref(false)
  const pipelineLoading = ref(false)
  const greenBookReadinessLoading = ref(false)
  const lenderFinancingMixLoading = ref(false)
  const klPortfolioPerformanceLoading = ref(false)
  const laDisbursementLoading = ref(false)
  const dataQualityGovernanceLoading = ref(false)
  const filterOptionsLoading = ref(false)
  const error = ref<string | null>(null)
  const summaryError = ref<string | null>(null)
  const stageFunnelError = ref<string | null>(null)
  const monitoringRollupError = ref<string | null>(null)
  const pipelineError = ref<string | null>(null)
  const greenBookReadinessError = ref<string | null>(null)
  const lenderFinancingMixError = ref<string | null>(null)
  const klPortfolioPerformanceError = ref<string | null>(null)
  const laDisbursementError = ref<string | null>(null)
  const dataQualityGovernanceError = ref<string | null>(null)

  async function fetchSummary(params?: DashboardFilterParams) {
    summaryLoading.value = true
    summaryError.value = null
    try {
      summary.value = await DashboardService.getSummary(params)
      return summary.value
    } catch {
      summaryError.value = 'Gagal memuat dashboard summary'
      summary.value = null
      return null
    } finally {
      summaryLoading.value = false
    }
  }

  async function fetchStageFunnel(params?: DashboardFilterParams) {
    stageFunnelLoading.value = true
    stageFunnelError.value = null
    try {
      stageFunnel.value = await DashboardService.getStageFunnel(params)
      return stageFunnel.value
    } catch {
      stageFunnelError.value = 'Gagal memuat stage funnel'
      stageFunnel.value = []
      return []
    } finally {
      stageFunnelLoading.value = false
    }
  }

  async function fetchMonitoringRollup(params?: DashboardFilterParams) {
    monitoringRollupLoading.value = true
    monitoringRollupError.value = null
    try {
      monitoringRollup.value = await DashboardService.getMonitoringRollup(params)
      return monitoringRollup.value
    } catch {
      monitoringRollupError.value = 'Gagal memuat monitoring rollup'
      monitoringRollup.value = []
      return []
    } finally {
      monitoringRollupLoading.value = false
    }
  }

  async function fetchExecutivePortfolio(params?: DashboardFilterParams) {
    loading.value = true
    error.value = null
    try {
      executivePortfolio.value = await DashboardService.getExecutivePortfolio(params)
      return executivePortfolio.value
    } catch {
      error.value = 'Gagal memuat Executive Portfolio'
      executivePortfolio.value = null
      return null
    } finally {
      loading.value = false
    }
  }

  async function fetchFilterOptions() {
    filterOptionsLoading.value = true
    try {
      filterOptions.value = await DashboardService.getFilterOptions()
      return filterOptions.value
    } finally {
      filterOptionsLoading.value = false
    }
  }

  async function fetchPipelineBottleneck(params?: PipelineBottleneckParams) {
    pipelineLoading.value = true
    pipelineError.value = null
    try {
      const response = await DashboardService.getPipelineBottleneck(params)
      pipelineBottleneck.value = response.data
      pipelineBottleneckMeta.value = response.meta
      return response.data
    } catch {
      pipelineError.value = 'Gagal memuat Pipeline & Bottleneck'
      pipelineBottleneck.value = null
      pipelineBottleneckMeta.value = null
      return null
    } finally {
      pipelineLoading.value = false
    }
  }

  async function fetchGreenBookReadiness(params?: GreenBookReadinessParams) {
    greenBookReadinessLoading.value = true
    greenBookReadinessError.value = null
    try {
      greenBookReadiness.value = await DashboardService.getGreenBookReadiness(params)
      return greenBookReadiness.value
    } catch {
      greenBookReadinessError.value = 'Gagal memuat Green Book Readiness'
      greenBookReadiness.value = null
      return null
    } finally {
      greenBookReadinessLoading.value = false
    }
  }

  async function fetchLenderFinancingMix(params?: LenderFinancingMixParams) {
    lenderFinancingMixLoading.value = true
    lenderFinancingMixError.value = null
    try {
      lenderFinancingMix.value = await DashboardService.getLenderFinancingMix(params)
      return lenderFinancingMix.value
    } catch {
      lenderFinancingMixError.value = 'Gagal memuat Lender & Financing Mix'
      lenderFinancingMix.value = null
      return null
    } finally {
      lenderFinancingMixLoading.value = false
    }
  }

  async function fetchKLPortfolioPerformance(params?: KLPortfolioPerformanceParams) {
    klPortfolioPerformanceLoading.value = true
    klPortfolioPerformanceError.value = null
    try {
      klPortfolioPerformance.value = await DashboardService.getKLPortfolioPerformance(params)
      return klPortfolioPerformance.value
    } catch {
      klPortfolioPerformanceError.value = 'Gagal memuat K/L Portfolio Performance'
      klPortfolioPerformance.value = null
      return null
    } finally {
      klPortfolioPerformanceLoading.value = false
    }
  }

  async function fetchLADisbursement(params?: LADisbursementParams) {
    laDisbursementLoading.value = true
    laDisbursementError.value = null
    try {
      laDisbursement.value = await DashboardService.getLADisbursement(params)
      return laDisbursement.value
    } catch {
      laDisbursementError.value = 'Gagal memuat Loan Agreement & Disbursement'
      laDisbursement.value = null
      return null
    } finally {
      laDisbursementLoading.value = false
    }
  }

  async function fetchDataQualityGovernance(params?: DataQualityGovernanceParams) {
    dataQualityGovernanceLoading.value = true
    dataQualityGovernanceError.value = null
    try {
      dataQualityGovernance.value = await DashboardService.getDataQualityGovernance(params)
      return dataQualityGovernance.value
    } catch {
      dataQualityGovernanceError.value = 'Gagal memuat Data Quality & Governance'
      dataQualityGovernance.value = null
      return null
    } finally {
      dataQualityGovernanceLoading.value = false
    }
  }

  function $reset() {
    summary.value = null
    stageFunnel.value = []
    monitoringRollup.value = []
    executivePortfolio.value = null
    pipelineBottleneck.value = null
    pipelineBottleneckMeta.value = null
    greenBookReadiness.value = null
    lenderFinancingMix.value = null
    klPortfolioPerformance.value = null
    laDisbursement.value = null
    dataQualityGovernance.value = null
    filterOptions.value = {}
    loading.value = false
    summaryLoading.value = false
    stageFunnelLoading.value = false
    monitoringRollupLoading.value = false
    pipelineLoading.value = false
    greenBookReadinessLoading.value = false
    lenderFinancingMixLoading.value = false
    klPortfolioPerformanceLoading.value = false
    laDisbursementLoading.value = false
    dataQualityGovernanceLoading.value = false
    filterOptionsLoading.value = false
    error.value = null
    summaryError.value = null
    stageFunnelError.value = null
    monitoringRollupError.value = null
    pipelineError.value = null
    greenBookReadinessError.value = null
    lenderFinancingMixError.value = null
    klPortfolioPerformanceError.value = null
    laDisbursementError.value = null
    dataQualityGovernanceError.value = null
  }

  return {
    summary,
    stageFunnel,
    monitoringRollup,
    executivePortfolio,
    pipelineBottleneck,
    pipelineBottleneckMeta,
    greenBookReadiness,
    lenderFinancingMix,
    klPortfolioPerformance,
    laDisbursement,
    dataQualityGovernance,
    filterOptions,
    loading,
    summaryLoading,
    stageFunnelLoading,
    monitoringRollupLoading,
    pipelineLoading,
    greenBookReadinessLoading,
    lenderFinancingMixLoading,
    klPortfolioPerformanceLoading,
    laDisbursementLoading,
    dataQualityGovernanceLoading,
    filterOptionsLoading,
    error,
    summaryError,
    stageFunnelError,
    monitoringRollupError,
    pipelineError,
    greenBookReadinessError,
    lenderFinancingMixError,
    klPortfolioPerformanceError,
    laDisbursementError,
    dataQualityGovernanceError,
    fetchSummary,
    fetchStageFunnel,
    fetchMonitoringRollup,
    fetchExecutivePortfolio,
    fetchPipelineBottleneck,
    fetchGreenBookReadiness,
    fetchLenderFinancingMix,
    fetchKLPortfolioPerformance,
    fetchLADisbursement,
    fetchDataQualityGovernance,
    fetchFilterOptions,
    $reset,
  }
})
