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
