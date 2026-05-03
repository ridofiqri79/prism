import http from '@/services/http'
import type { ApiResponse, PaginatedResponse } from '@/types/api.types'
import type {
  MonitoringApiResponse,
  MonitoringDisbursement,
  MonitoringLoanAgreementListParams,
  MonitoringLoanAgreementReference,
  MonitoringListParams,
  MonitoringPayload,
} from '@/types/monitoring.types'
import type { MasterImportSummary } from '@/types/master.types'

function normalizeMonitoring(item: MonitoringApiResponse): MonitoringDisbursement {
  return {
    ...item,
    absorption_pct: item.absorption_pct ?? item.penyerapan_pct ?? 0,
    komponen: item.komponen ?? [],
  }
}

export const MonitoringService = {
  async getLoanAgreementReferences(params?: MonitoringLoanAgreementListParams) {
    const response = await http.get<PaginatedResponse<MonitoringLoanAgreementReference>>(
      '/monitoring/loan-agreements',
      { params },
    )
    return response.data
  },

  async getMonitorings(loanAgreementId: string, params?: MonitoringListParams) {
    const response = await http.get<PaginatedResponse<MonitoringApiResponse>>(
      `/loan-agreements/${loanAgreementId}/monitoring`,
      { params },
    )

    return {
      ...response.data,
      data: response.data.data.map(normalizeMonitoring),
    }
  },

  async getMonitoring(loanAgreementId: string, id: string) {
    const response = await http.get<ApiResponse<MonitoringApiResponse>>(
      `/loan-agreements/${loanAgreementId}/monitoring/${id}`,
    )
    return normalizeMonitoring(response.data.data)
  },

  async createMonitoring(loanAgreementId: string, data: MonitoringPayload) {
    const response = await http.post<ApiResponse<MonitoringApiResponse>>(
      `/loan-agreements/${loanAgreementId}/monitoring`,
      data,
    )
    return normalizeMonitoring(response.data.data)
  },

  async updateMonitoring(loanAgreementId: string, id: string, data: MonitoringPayload) {
    const response = await http.put<ApiResponse<MonitoringApiResponse>>(
      `/loan-agreements/${loanAgreementId}/monitoring/${id}`,
      data,
    )
    return normalizeMonitoring(response.data.data)
  },

  async deleteMonitoring(loanAgreementId: string, id: string) {
    await http.delete(`/loan-agreements/${loanAgreementId}/monitoring/${id}`)
  },

  async downloadImportTemplate() {
    const response = await http.get<Blob>('/monitoring/import/template', {
      responseType: 'blob',
    })
    return response.data
  },

  async previewImport(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    const response = await http.post<ApiResponse<MasterImportSummary>>(
      '/monitoring/import/preview',
      formData,
    )
    return response.data.data
  },

  async executeImport(file: File) {
    const formData = new FormData()
    formData.append('file', file)
    const response = await http.post<ApiResponse<MasterImportSummary>>(
      '/monitoring/import/execute',
      formData,
    )
    return response.data.data
  },
}
