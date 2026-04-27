import type { ZodError } from 'zod'

export type FormErrors<T extends string> = Partial<Record<T, string>>

export function toFormErrors<T extends string>(error: ZodError, fields: readonly T[]) {
  const formErrors: FormErrors<T> = {}

  for (const issue of error.issues) {
    const field = String(issue.path[0] ?? '') as T
    if (fields.includes(field) && !formErrors[field]) {
      formErrors[field] = issue.message
    }
  }

  return formErrors
}

export function formatRevision(revisionNumber: number, revisionYear?: number | null) {
  if (revisionNumber === 0) return 'Original'
  return `Revisi ke-${revisionNumber}${revisionYear ? ` (${revisionYear})` : ''}`
}

export function joinNames(items: { name?: string; title?: string }[]) {
  return items.map((item) => item.name ?? item.title).filter(Boolean).join(', ') || '-'
}

