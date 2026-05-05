import { toFormErrors } from '@/utils/form-errors'

export type { FormErrors } from '@/utils/form-errors'

// Re-export shared utility for backward compatibility
export { toFormErrors }

export function formatGBRevision(revisionNumber: number) {
  return revisionNumber === 0 ? 'Original' : `Revisi ke-${revisionNumber}`
}

export function formatGreenBookStatus(status: string) {
  if (status === 'active') return 'Berlaku'
  if (status === 'superseded') return 'Tidak Berlaku'
  return status
}

export function joinNames(items: { name?: string; project_name?: string; bb_code?: string }[]) {
  return (
    items
      .map((item) => item.name ?? (item.bb_code ? `${item.bb_code} - ${item.project_name}` : item.project_name))
      .filter(Boolean)
      .join(', ') || '-'
  )
}
