import { ref } from 'vue'
import { defineStore } from 'pinia'
import { DashboardService } from '@/services/dashboard.service'
import type { PaginationMeta } from '@/types/api.types'
import type {
  DashboardFilterOptions,
  DashboardFilterParams,
  ExecutivePortfolioDashboard,
  PipelineBottleneckParams,
  PipelineBottleneckResponse,
} from '@/types/dashboard.types'

export const useDashboardStore = defineStore('dashboard', () => {
  const executivePortfolio = ref<ExecutivePortfolioDashboard | null>(null)
  const pipelineBottleneck = ref<PipelineBottleneckResponse | null>(null)
  const pipelineBottleneckMeta = ref<PaginationMeta | null>(null)
  const filterOptions = ref<DashboardFilterOptions>({})
  const loading = ref(false)
  const pipelineLoading = ref(false)
  const filterOptionsLoading = ref(false)
  const error = ref<string | null>(null)
  const pipelineError = ref<string | null>(null)

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

  function $reset() {
    executivePortfolio.value = null
    pipelineBottleneck.value = null
    pipelineBottleneckMeta.value = null
    filterOptions.value = {}
    loading.value = false
    pipelineLoading.value = false
    filterOptionsLoading.value = false
    error.value = null
    pipelineError.value = null
  }

  return {
    executivePortfolio,
    pipelineBottleneck,
    pipelineBottleneckMeta,
    filterOptions,
    loading,
    pipelineLoading,
    filterOptionsLoading,
    error,
    pipelineError,
    fetchExecutivePortfolio,
    fetchPipelineBottleneck,
    fetchFilterOptions,
    $reset,
  }
})
