import { toFormErrors } from '@/utils/form-errors'
import { formatBookStatus, formatRevision as _formatRevision } from '@/utils/formatters'

export type { FormErrors } from '@/utils/form-errors'

// Re-export shared utilities for backward compatibility
export { toFormErrors }

/**
 * @deprecated Gunakan `formatRevision` dari `@/utils/formatters`.
 * Tetap di-export untuk backward compatibility.
 */
export function formatGBRevision(revisionNumber: number) {
  return _formatRevision(revisionNumber)
}

/**
 * @deprecated Gunakan `formatBookStatus` dari `@/utils/formatters`.
 * Tetap di-export untuk backward compatibility.
 */
export function formatGreenBookStatus(status: string) {
  return formatBookStatus(status)
}

export function joinNames(items: { name?: string; project_name?: string; bb_code?: string }[]) {
  return (
    items
      .map((item) => item.name ?? (item.bb_code ? `${item.bb_code} - ${item.project_name}` : item.project_name))
      .filter(Boolean)
      .join(', ') || '-'
  )
}
