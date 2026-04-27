import http from '@/services/http'
import type { ApiResponse } from '@/types/api.types'
import type { LoginResponse, MeResponse } from '@/types/auth.types'

export interface LoginRequest {
  username: string
  password: string
}

export const AuthService = {
  async login(payload: LoginRequest) {
    const response = await http.post<ApiResponse<LoginResponse>>('/auth/login', payload)

    return response.data.data
  },

  async logout() {
    await http.post('/auth/logout')
  },

  async getMe() {
    const response = await http.get<ApiResponse<MeResponse>>('/auth/me')

    return response.data.data
  },
}
