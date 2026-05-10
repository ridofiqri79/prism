import { formatDate, parseDateModel, toDateString } from '@/utils/formatters'
import { toFormErrors } from '@/utils/form-errors'

export type { FormErrors } from '@/utils/form-errors'

// Re-export shared utilities for backward compatibility
export { formatDate, parseDateModel, toDateString, toFormErrors }

export function formatDKProjectLabel(
  project?: {
    id?: string
    project_name?: string | null
    objectives?: string | null
    label?: string
  } | null,
) {
  if (!project) return '-'
  return project.label || project.project_name || project.objectives || project.id || '-'
}

export function formatNumber(value?: number | null, digits = 2) {
  if (value === null || value === undefined || Number.isNaN(value)) return '-'
  return new Intl.NumberFormat('en-US', {
    minimumFractionDigits: digits,
    maximumFractionDigits: digits,
  }).format(value)
}

export function formatPercent(value?: number | null) {
  if (value === null || value === undefined || Number.isNaN(value)) return '-'
  return `${formatNumber(value)}%`
}

export function formatRatio(value?: number | null) {
  return formatNumber(value, 2)
}
