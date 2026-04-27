import http from '@/services/http'
import type { ApiResponse, PaginatedResponse } from '@/types/api.types'
import type {
  GBProject,
  GBProjectPayload,
  GreenBook,
  GreenBookPayload,
} from '@/types/green-book.types'
import type { ListParams } from '@/types/master.types'

export const GreenBookService = {
  async getGreenBooks(params?: ListParams) {
    const response = await http.get<PaginatedResponse<GreenBook>>('/green-books', { params })
    return response.data
  },

  async getGreenBook(id: string) {
    const response = await http.get<ApiResponse<GreenBook>>(`/green-books/${id}`)
    return response.data.data
  },

  async createGreenBook(data: GreenBookPayload) {
    const response = await http.post<ApiResponse<GreenBook>>('/green-books', data)
    return response.data.data
  },

  async updateGreenBook(id: string, data: GreenBookPayload) {
    const response = await http.put<ApiResponse<GreenBook>>(`/green-books/${id}`, data)
    return response.data.data
  },

  async deleteGreenBook(id: string) {
    await http.delete(`/green-books/${id}`)
  },

  async getProjects(greenBookId: string, params?: ListParams) {
    const response = await http.get<PaginatedResponse<GBProject>>(
      `/green-books/${greenBookId}/projects`,
      { params },
    )
    return response.data
  },

  async getProject(greenBookId: string, id: string) {
    const response = await http.get<ApiResponse<GBProject>>(
      `/green-books/${greenBookId}/projects/${id}`,
    )
    return response.data.data
  },

  async createProject(greenBookId: string, data: GBProjectPayload) {
    const response = await http.post<ApiResponse<GBProject>>(
      `/green-books/${greenBookId}/projects`,
      data,
    )
    return response.data.data
  },

  async updateProject(greenBookId: string, id: string, data: GBProjectPayload) {
    const response = await http.put<ApiResponse<GBProject>>(
      `/green-books/${greenBookId}/projects/${id}`,
      data,
    )
    return response.data.data
  },

  async deleteProject(greenBookId: string, id: string) {
    await http.delete(`/green-books/${greenBookId}/projects/${id}`)
  },
}

