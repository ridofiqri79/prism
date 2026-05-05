/**
 * Shared Zod → form-error map utility for PRISM.
 *
 * Import from here instead of duplicating in page-level utils.
 *
 *   import { toFormErrors, type FormErrors } from '@/utils/form-errors'
 */

import type { ZodError } from 'zod'

export type FormErrors<T extends string> = Partial<Record<T, string>>

/**
 * Converts a ZodError into a flat field → message map.
 * Only includes fields that appear in the `fields` allowlist.
 * First error per field wins (consistent with all existing page-utils implementations).
 */
export function toFormErrors<T extends string>(
  error: ZodError,
  fields: readonly T[],
): FormErrors<T> {
  const allowed = new Set<string>(fields)
  const errors: FormErrors<T> = {}

  for (const issue of error.issues) {
    const field = String(issue.path[0] ?? '') as T
    if (allowed.has(field) && !errors[field]) {
      errors[field] = issue.message
    }
  }

  return errors
}
