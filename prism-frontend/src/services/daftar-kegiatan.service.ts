import http from '@/services/http'
import type { ApiResponse, PaginatedResponse } from '@/types/api.types'
import type {
  DaftarKegiatan,
  DaftarKegiatanPayload,
  DKProject,
  DKProjectPayload,
} from '@/types/daftar-kegiatan.types'
import type { ListParams } from '@/types/master.types'

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
