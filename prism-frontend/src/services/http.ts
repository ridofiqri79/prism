import axios from 'axios'
import { clearStoredSession, readStoredToken } from '@/utils/auth-session'
import { emitToast, emitUnauthorized } from '@/utils/app-events'

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
  async (error) => {
    const status = error.response?.status as number | undefined

    if (status === 401) {
      clearStoredSession()
      emitUnauthorized()

      const { default: router } = await import('@/router')

      if (router.currentRoute.value.name !== 'login') {
        await router.push({ name: 'login' })
      }
    }

    if (status === 403) {
      emitToast({
        severity: 'error',
        summary: 'Akses Ditolak',
        detail: 'Anda tidak memiliki izin untuk mengakses halaman ini.',
        life: 3000,
      })
    }

    if (status === 500) {
      emitToast({
        severity: 'error',
        summary: 'Terjadi Kesalahan Server',
        detail: 'Silakan coba lagi beberapa saat lagi.',
        life: 3000,
      })
    }

    return Promise.reject(error)
  },
)

export default http
