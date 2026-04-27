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

export function joinNames(items?: { name?: string; title?: string; project_name?: string; gb_code?: string }[]) {
  if (!items?.length) return '-'
  return items
    .map((item) => item.name ?? item.title ?? [item.gb_code, item.project_name].filter(Boolean).join(' - '))
    .filter(Boolean)
    .join(', ')
}
