import { formatDate, parseDateModel, toDateString } from '@/utils/formatters'
import { toFormErrors } from '@/utils/form-errors'

export type { FormErrors } from '@/utils/form-errors'

// Re-export shared utilities for backward compatibility
export { formatDate, parseDateModel, toDateString, toFormErrors }

export function formatDKProjectLabel(
  project?: { id?: string; project_name?: string | null; objectives?: string | null; label?: string } | null,
) {
  if (!project) return '-'
  return project.label || project.project_name || project.objectives || project.id || '-'
}
