export interface ApiResponse<T> {
  data: T
}

export interface PaginatedResponse<T> {
  data: T[]
  meta: PaginationMeta
}

export interface PaginationMeta {
  page: number
  limit: number
  total: number
  total_pages: number
}

export interface ApiError {
  code: string
  message: string
  details?: FieldError[]
}

export interface FieldError {
  field: string
  message: string
}
