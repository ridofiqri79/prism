import http from '@/services/http'
import type { ApiResponse } from '@/types/api.types'
import type {
  DashboardBlueBookDistribution,
  DashboardDaftarKegiatanDistribution,
  DashboardGreenBookDistribution,
  DashboardLoanAgreementDistribution,
  DashboardStageOverview,
} from '@/types/dashboard.types'

interface DashboardDistributionParams {
  period_ids?: string[]
}

export const DashboardService = {
  async getStageOverview(params?: DashboardDistributionParams) {
    const response = await http.get<ApiResponse<DashboardStageOverview>>(
      '/dashboard/stage-overview',
      { params },
    )

    return response.data.data
  },

  async getBlueBookDistribution(params?: DashboardDistributionParams) {
    const response = await http.get<ApiResponse<DashboardBlueBookDistribution>>(
      '/dashboard/blue-book-distribution',
      { params },
    )

    return response.data.data
  },

  async getGreenBookDistribution(params?: DashboardDistributionParams) {
    const response = await http.get<ApiResponse<DashboardGreenBookDistribution>>(
      '/dashboard/green-book-distribution',
      { params },
    )

    return response.data.data
  },

  async getDaftarKegiatanDistribution(params?: DashboardDistributionParams) {
    const response = await http.get<ApiResponse<DashboardDaftarKegiatanDistribution>>(
      '/dashboard/daftar-kegiatan-distribution',
      { params },
    )

    return response.data.data
  },

  async getLoanAgreementDistribution(params?: DashboardDistributionParams) {
    const response = await http.get<ApiResponse<DashboardLoanAgreementDistribution>>(
      '/dashboard/loan-agreement-distribution',
      { params },
    )

    return response.data.data
  },
}
