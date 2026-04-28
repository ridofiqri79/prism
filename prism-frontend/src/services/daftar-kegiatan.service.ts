import http from '@/services/http'
import type { ApiResponse, PaginatedResponse } from '@/types/api.types'
import type {
  DaftarKegiatan,
  DaftarKegiatanPayload,
  DKProject,
  DKProjectPayload,
} from '@/types/daftar-kegiatan.types'
import type { ListParams, MasterImportSummary } from '@/types/master.types'

export const DaftarKegiatanService = {
  async getDaftarKegiatan(params?: ListParams) {
    const response = await http.get<PaginatedResponse<DaftarKegiatan>>('/daftar-kegiatan', {
      params,
    })
    return response.data
  },

  async getDK(id: string) {
    const response = await http.get<ApiResponse<DaftarKegiatan>>(`/daftar-kegiatan/${id}`)
    return response.data.data
  },

  async createDK(data: DaftarKegiatanPayload) {
    const response = await http.post<ApiResponse<DaftarKegiatan>>('/daftar-kegiatan', data)
    return response.data.data
  },

  async updateDK(id: string, data: DaftarKegiatanPayload) {
    const response = await http.put<ApiResponse<DaftarKegiatan>>(`/daftar-kegiatan/${id}`, data)
    return response.data.data
  },

  async deleteDK(id: string) {
    await http.delete(`/daftar-kegiatan/${id}`)
  },

  async downloadImportTemplate() {
    const response = await http.get<Blob>('/daftar-kegiatan/import/template', {
      responseType: 'blob',
    })

    return response.data
  },

  async previewImport(file: File) {
    const formData = new FormData()
    formData.append('file', file)

    const response = await http.post<ApiResponse<MasterImportSummary>>(
      '/daftar-kegiatan/import/preview',
      formData,
    )

    return response.data.data
  },

  async executeImport(file: File) {
    const formData = new FormData()
    formData.append('file', file)

    const response = await http.post<ApiResponse<MasterImportSummary>>(
      '/daftar-kegiatan/import/execute',
      formData,
    )

    return response.data.data
  },

  async getProjects(dkId: string, params?: ListParams) {
    const response = await http.get<PaginatedResponse<DKProject>>(
      `/daftar-kegiatan/${dkId}/projects`,
      { params },
    )
    return response.data
  },

  async getProject(dkId: string, id: string) {
    const response = await http.get<ApiResponse<DKProject>>(
      `/daftar-kegiatan/${dkId}/projects/${id}`,
    )
    return response.data.data
  },

  async createProject(dkId: string, data: DKProjectPayload) {
    const response = await http.post<ApiResponse<DKProject>>(
      `/daftar-kegiatan/${dkId}/projects`,
      data,
    )
    return response.data.data
  },

  async updateProject(dkId: string, id: string, data: DKProjectPayload) {
    const response = await http.put<ApiResponse<DKProject>>(
      `/daftar-kegiatan/${dkId}/projects/${id}`,
      data,
    )
    return response.data.data
  },

  async deleteProject(dkId: string, id: string) {
    await http.delete(`/daftar-kegiatan/${dkId}/projects/${id}`)
  },
}
