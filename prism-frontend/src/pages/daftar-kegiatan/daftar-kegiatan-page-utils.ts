import { formatDate } from '@/utils/formatters'
import { toFormErrors } from '@/utils/form-errors'

export type { FormErrors } from '@/utils/form-errors'

// Re-export shared utilities for backward compatibility
export { formatDate, toFormErrors }

export function joinNames(items?: { name?: string; title?: string; project_name?: string; gb_code?: string }[]) {
  if (!items?.length) return '-'
  return items
    .map((item) => item.name ?? item.title ?? [item.gb_code, item.project_name].filter(Boolean).join(' - '))
    .filter(Boolean)
    .join(', ')
}
