import axios, { type AxiosError } from 'axios'
import type { ApiErrorResponse } from '@/types/api.types'
import { clearStoredSession, readStoredToken } from '@/utils/auth-session'
import { emitToast, emitUnauthorized } from '@/utils/app-events'

declare module 'axios' {
  export interface AxiosRequestConfig {
    skipAuthRedirect?: boolean
  }
}

const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
  timeout: 30000,
})

http.interceptors.request.use((config) => {
  const token = readStoredToken()

  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }

  return config
})

http.interceptors.response.use(
  (response) => response,
  async (error: AxiosError<ApiErrorResponse>) => {
    const status = error.response?.status
    const skipAuthRedirect = error.config?.skipAuthRedirect === true

    if (status === 401 && !skipAuthRedirect) {
      clearStoredSession()
      emitUnauthorized()

      const { default: router } = await import('@/router')

      if (router.currentRoute.value.name !== 'login') {
        await router.push({ name: 'login' })
      }
    }

    if (status === 403 && !error.response?.data?.error?.details?.length) {
      emitToast({
        severity: 'error',
        summary: 'Akses Ditolak',
        detail: 'Anda tidak memiliki izin untuk melakukan tindakan ini',
        life: 5000,
      })
    }

    if (status === 500) {
      emitToast({
        severity: 'error',
        summary: 'Terjadi Kesalahan',
        detail: 'Server mengalami masalah, silakan coba lagi',
        life: 5000,
      })
    }

    if (!error.response) {
      emitToast({
        severity: 'error',
        summary: 'Koneksi Bermasalah',
        detail: 'Periksa koneksi internet Anda',
        life: 5000,
      })
    }

    return Promise.reject(error)
  },
)

export default http
