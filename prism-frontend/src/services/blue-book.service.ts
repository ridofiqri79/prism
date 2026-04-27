import http from '@/services/http'
import type { ApiResponse, PaginatedResponse } from '@/types/api.types'
import type {
  BBProject,
  BBProjectPayload,
  BlueBook,
  BlueBookPayload,
  LoI,
  LoIPayload,
} from '@/types/blue-book.types'
import type { ListParams } from '@/types/master.types'

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

  async getProjects(blueBookId: string, params?: ListParams) {
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

