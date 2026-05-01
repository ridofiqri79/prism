import { isAxiosError } from 'axios'
import type { ApiErrorResponse } from '@/types/api.types'

export function formatApiError(error: unknown, fallback: string) {
  if (!isAxiosError<ApiErrorResponse>(error)) return fallback

  const apiError = error.response?.data?.error
  if (!apiError) return fallback

  const details = apiError.details?.map((item) => item.message).filter(Boolean) ?? []
  if (details.length === 0) return apiError.message || fallback

  return [apiError.message, ...details].join('\n')
}
