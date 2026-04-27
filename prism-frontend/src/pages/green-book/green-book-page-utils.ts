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

export function formatGBRevision(revisionNumber: number) {
  return revisionNumber === 0 ? 'Original' : `Revisi ke-${revisionNumber}`
}

export function joinNames(items: { name?: string; project_name?: string; bb_code?: string }[]) {
  return (
    items
      .map((item) => item.name ?? (item.bb_code ? `${item.bb_code} - ${item.project_name}` : item.project_name))
      .filter(Boolean)
      .join(', ') || '-'
  )
}

