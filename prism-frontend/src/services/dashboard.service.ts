import http from '@/services/http'
import type { ApiResponse } from '@/types/api.types'
import type {
  DashboardBlueBookDistribution,
  DashboardDaftarKegiatanDistribution,
  DashboardGreenBookDistribution,
  DashboardLoanAgreementDistribution,
} from '@/types/dashboard.types'

export const DashboardService = {
  async getBlueBookDistribution() {
    const response = await http.get<ApiResponse<DashboardBlueBookDistribution>>(
      '/dashboard/blue-book-distribution',
    )

    return response.data.data
  },

  async getGreenBookDistribution() {
    const response = await http.get<ApiResponse<DashboardGreenBookDistribution>>(
      '/dashboard/green-book-distribution',
    )

    return response.data.data
  },

  async getDaftarKegiatanDistribution() {
    const response = await http.get<ApiResponse<DashboardDaftarKegiatanDistribution>>(
      '/dashboard/daftar-kegiatan-distribution',
    )

    return response.data.data
  },

  async getLoanAgreementDistribution() {
    const response = await http.get<ApiResponse<DashboardLoanAgreementDistribution>>(
      '/dashboard/loan-agreement-distribution',
    )

    return response.data.data
  },
}
