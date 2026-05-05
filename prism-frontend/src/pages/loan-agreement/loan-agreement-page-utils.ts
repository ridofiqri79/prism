import type { ZodError } from 'zod'

export type FormErrors<T extends string> = Partial<Record<T, string>>

export function toFormErrors<T extends string>(error: ZodError, allowedFields: T[]) {
  const allowed = new Set<string>(allowedFields)
  const errors: FormErrors<T> = {}

  for (const issue of error.issues) {
    const field = String(issue.path[0]) as T
    if (allowed.has(field) && !errors[field]) {
      errors[field] = issue.message
    }
  }

  return errors
}

export function formatDate(date?: string | null) {
  if (!date) return '-'
  return new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium' }).format(new Date(date))
}

export function formatDKProjectLabel(
  project?: { id?: string; project_name?: string | null; objectives?: string | null; label?: string } | null,
) {
  if (!project) return '-'
  return project.label || project.project_name || project.objectives || project.id || '-'
}

export function parseDateModel(value?: string | null) {
  if (!value) return null
  const date = new Date(`${value}T00:00:00`)
  return Number.isNaN(date.getTime()) ? null : date
}

export function toDateString(value: Date | null | undefined) {
  if (!value) return ''
  const year = value.getFullYear()
  const month = String(value.getMonth() + 1).padStart(2, '0')
  const day = String(value.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}
