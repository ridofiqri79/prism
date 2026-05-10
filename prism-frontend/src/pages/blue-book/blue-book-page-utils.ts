import { toFormErrors } from '@/utils/form-errors'
import { formatBookStatus, formatRevision as _formatRevision } from '@/utils/formatters'

export type { FormErrors } from '@/utils/form-errors'

// Re-export shared utilities for backward compatibility
export { toFormErrors }

/**
 * @deprecated Gunakan `formatRevision` dari `@/utils/formatters`.
 * Tetap di-export untuk backward compatibility.
 */
export function formatRevision(revisionNumber: number, revisionYear?: number | null) {
  return _formatRevision(revisionNumber, revisionYear)
}

/**
 * @deprecated Gunakan `formatBookStatus` dari `@/utils/formatters`.
 * Tetap di-export untuk backward compatibility.
 */
export function formatBlueBookStatus(status: string) {
  return formatBookStatus(status)
}

export function joinNames(items: { name?: string; title?: string }[]) {
  return items.map((item) => item.name ?? item.title).filter(Boolean).join(', ') || '-'
}
