/**
 * Shared formatter utilities for PRISM.
 *
 * Import from here instead of duplicating in page-level utils.
 *
 *   import { formatDate, formatDateTime } from '@/utils/formatters'
 */

const dateFormatter = new Intl.DateTimeFormat('id-ID', { dateStyle: 'medium' })
const dateTimeFormatter = new Intl.DateTimeFormat('id-ID', {
  dateStyle: 'medium',
  timeStyle: 'short',
})

/**
 * Formats a date string (YYYY-MM-DD or ISO) as "12 Jan 2025".
 * Returns '-' for empty/null/invalid input.
 */
export function formatDate(date?: string | null): string {
  if (!date) return '-'
  const d = new Date(date)
  if (Number.isNaN(d.getTime())) return '-'
  return dateFormatter.format(d)
}

/**
 * Formats a full ISO datetime string as "12 Jan 2025, 14.30".
 * Returns '-' for empty/null/invalid input.
 */
export function formatDateTime(value?: string | null): string {
  if (!value) return '-'
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return '-'
  return dateTimeFormatter.format(d)
}

/**
 * Parses a date string (YYYY-MM-DD) into a Date object for use with date-picker components.
 * Returns null for empty/null/invalid input.
 */
export function parseDateModel(value?: string | null): Date | null {
  if (!value) return null
  const d = new Date(`${value}T00:00:00`)
  return Number.isNaN(d.getTime()) ? null : d
}

/**
 * Converts a Date object back to a YYYY-MM-DD string.
 * Returns '' for null/undefined input.
 */
export function toDateString(value: Date | null | undefined): string {
  if (!value) return ''
  const year = value.getFullYear()
  const month = String(value.getMonth() + 1).padStart(2, '0')
  const day = String(value.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

/**
 * Extracts display names from a list of objects that have either a `name` or `title` field.
 * Used by BB/GB project detail pages to build ValueChipList items.
 */
export function toNameList(items?: { name?: string; title?: string }[]): string[] {
  return (
    items
      ?.map((item) => item.name ?? item.title)
      .filter((item): item is string => Boolean(item)) ?? []
  )
}

/**
 * Converts a Blue Book or Green Book `status` DB value to a human-readable label.
 * Replaces the duplicated `formatBlueBookStatus` / `formatGreenBookStatus` in page-utils.
 *
 *   formatBookStatus('active')      // → 'Berlaku'
 *   formatBookStatus('superseded')  // → 'Tidak Berlaku'
 */
export function formatBookStatus(status: string): string {
  if (status === 'active') return 'Berlaku'
  if (status === 'superseded') return 'Tidak Berlaku'
  return status
}

/**
 * Converts a revision number (and optional year) to a human-readable label.
 * Replaces the duplicated `formatRevision` (BB) / `formatGBRevision` (GB) in page-utils.
 *
 *   formatRevision(0)         // → 'Original'
 *   formatRevision(1)         // → 'Revisi ke-1'
 *   formatRevision(1, 2025)   // → 'Revisi ke-1 (2025)'
 */
export function formatRevision(revisionNumber: number, revisionYear?: number | null): string {
  if (revisionNumber === 0) return 'Original'
  return `Revisi ke-${revisionNumber}${revisionYear ? ` (${revisionYear})` : ''}`
}
