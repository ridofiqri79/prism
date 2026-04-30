import http from '@/services/http'
import type { ApiResponse, PaginatedResponse } from '@/types/api.types'
import type {
  BBProject,
  BBProjectHistoryItem,
  BBProjectListParams,
  BBProjectPayload,
  BlueBook,
  BlueBookPayload,
  LoI,
  LoIPayload,
} from '@/types/blue-book.types'
import type { ListParams, MasterImportSummary } from '@/types/master.types'

export const BlueBookService = {
  async getBlueBooks(params?: ListParams) {
    const response = await http.get<PaginatedResponse<BlueBook>>('/blue-books', { params })
    return response.data
  },

  async getBlueBook(id: string) {
    const response = await http.get<ApiResponse<BlueBook>>(`/blue-books/${id}`)
    return response.data.data
  },

  async createBlueBook(data: BlueBookPayload) {
    const response = await http.post<ApiResponse<BlueBook>>('/blue-books', data)
    return response.data.data
  },

  async updateBlueBook(id: string, data: BlueBookPayload) {
    const response = await http.put<ApiResponse<BlueBook>>(`/blue-books/${id}`, data)
    return response.data.data
  },

  async deleteBlueBook(id: string) {
    await http.delete(`/blue-books/${id}`)
  },

  async getProjects(blueBookId: string, params?: BBProjectListParams) {
    const response = await http.get<PaginatedResponse<BBProject>>(
      `/blue-books/${blueBookId}/projects`,
      { params },
    )
    return response.data
  },

  async getProject(blueBookId: string, id: string) {
    const response = await http.get<ApiResponse<BBProject>>(
      `/blue-books/${blueBookId}/projects/${id}`,
    )
    return response.data.data
  },

  async getBBProjectHistory(id: string) {
    const response = await http.get<ApiResponse<BBProjectHistoryItem[]>>(
      `/bb-projects/${id}/history`,
    )
    return response.data.data
  },

  async createProject(blueBookId: string, data: BBProjectPayload) {
    const response = await http.post<ApiResponse<BBProject>>(
      `/blue-books/${blueBookId}/projects`,
      data,
    )
    return response.data.data
  },

  async updateProject(blueBookId: string, id: string, data: BBProjectPayload) {
    const response = await http.put<ApiResponse<BBProject>>(
      `/blue-books/${blueBookId}/projects/${id}`,
      data,
    )
    return response.data.data
  },

  async deleteProject(blueBookId: string, id: string) {
    await http.delete(`/blue-books/${blueBookId}/projects/${id}`)
  },

  async downloadImportTemplate(blueBookId: string) {
    const response = await http.get<Blob>(`/blue-books/${blueBookId}/import-projects/template`, {
      responseType: 'blob',
    })

    return response.data
  },

  async previewImportProjects(blueBookId: string, file: File) {
    const formData = new FormData()
    formData.append('file', file)

    const response = await http.post<ApiResponse<MasterImportSummary>>(
      `/blue-books/${blueBookId}/import-projects/preview`,
      formData,
    )

    return response.data.data
  },

  async executeImportProjects(blueBookId: string, file: File) {
    const formData = new FormData()
    formData.append('file', file)

    const response = await http.post<ApiResponse<MasterImportSummary>>(
      `/blue-books/${blueBookId}/import-projects/execute`,
      formData,
    )

    return response.data.data
  },

  async getLoI(projectId: string) {
    const response = await http.get<ApiResponse<LoI[]>>(`/bb-projects/${projectId}/loi`)
    return response.data.data
  },

  async createLoI(projectId: string, data: LoIPayload) {
    const response = await http.post<ApiResponse<LoI>>(`/bb-projects/${projectId}/loi`, data)
    return response.data.data
  },

  async deleteLoI(projectId: string, id: string) {
    await http.delete(`/bb-projects/${projectId}/loi/${id}`)
  },
}
