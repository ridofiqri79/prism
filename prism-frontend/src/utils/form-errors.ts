/**
 * Shared Zod → form-error map utility for PRISM.
 *
 * Import from here instead of duplicating in page-level utils.
 *
 *   import { toFormErrors, assignFormErrors, type FormErrors } from '@/utils/form-errors'
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

/**
 * Mutates a reactive error object in-place from a ZodError.
 * Clears all existing errors first, then sets first error per field.
 *
 * Used by all form composables (useBBProjectForm, useGBProjectForm, etc.)
 * instead of duplicating the same local `assignErrors` function.
 *
 *   assignFormErrors(errors, parsed.error)
 */
export function assignFormErrors<T extends string>(
  target: Partial<Record<T, string>>,
  error: ZodError,
): void {
  // Clear existing errors
  for (const key of Object.keys(target)) {
    delete target[key as T]
  }
  // Assign first error per field
  for (const issue of error.issues) {
    const field = String(issue.path[0]) as T
    if (!target[field]) {
      target[field] = issue.message
    }
  }
}
