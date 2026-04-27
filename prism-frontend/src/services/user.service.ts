import http from '@/services/http'
import type { ApiResponse, PaginatedResponse } from '@/types/api.types'
import type { UserPermission } from '@/types/auth.types'
import type { ListParams } from '@/types/master.types'
import type {
  AppUser,
  CreateUserRequest,
  UpdatePermissionsRequest,
  UpdateUserRequest,
} from '@/types/user.types'

export const UserService = {
  async getUsers(params?: ListParams) {
    const response = await http.get<PaginatedResponse<AppUser>>('/users', { params })

    return response.data
  },

  async getUser(id: string) {
    const response = await http.get<ApiResponse<AppUser>>(`/users/${id}`)

    return response.data.data
  },

  async createUser(data: CreateUserRequest) {
    const response = await http.post<ApiResponse<AppUser>>('/users', data)

    return response.data.data
  },

  async updateUser(id: string, data: UpdateUserRequest) {
    const response = await http.put<ApiResponse<AppUser>>(`/users/${id}`, data)

    return response.data.data
  },

  async deleteUser(id: string) {
    await http.delete(`/users/${id}`)
  },

  async getUserPermissions(id: string) {
    const response = await http.get<ApiResponse<UserPermission[]>>(`/users/${id}/permissions`)

    return response.data.data
  },

  async updatePermissions(id: string, permissions: UserPermission[]) {
    const response = await http.put<ApiResponse<UserPermission[]>>(
      `/users/${id}/permissions`,
      { permissions } satisfies UpdatePermissionsRequest,
    )

    return response.data.data
  },
}
