import { toFormErrors } from '@/utils/form-errors'

export type { FormErrors } from '@/utils/form-errors'

// Re-export shared utility for backward compatibility
export { toFormErrors }

export function formatRevision(revisionNumber: number, revisionYear?: number | null) {
  if (revisionNumber === 0) return 'Original'
  return `Revisi ke-${revisionNumber}${revisionYear ? ` (${revisionYear})` : ''}`
}

export function formatBlueBookStatus(status: string) {
  if (status === 'active') return 'Berlaku'
  if (status === 'superseded') return 'Tidak Berlaku'
  return status
}

export function joinNames(items: { name?: string; title?: string }[]) {
  return items.map((item) => item.name ?? item.title).filter(Boolean).join(', ') || '-'
}
