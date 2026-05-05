/**
 * Centralized status label and severity mappings for PRISM.
 *
 * Usage:
 *   import { getStatusLabel, getStatusSeverity } from '@/utils/status-labels'
 *
 *   getStatusLabel('active')               // → 'Aktif'
 *   getStatusLabel('berlaku', 'blue_book') // → 'Berlaku'
 *   getStatusSeverity('active')            // → 'success'
 */

export type StatusDomain =
  | 'default'
  | 'blue_book'
  | 'green_book'
  | 'loan_agreement'
  | 'user'
  | 'import'
  | 'pipeline'

export type PrimeSeverity = 'success' | 'info' | 'warn' | 'danger' | 'secondary' | 'contrast'

/** Pipeline status labels (BB → GB → DK → LA → Monitoring) */
export const PIPELINE_STATUS_LABELS: Record<string, string> = {
  BB: 'Blue Book',
  GB: 'Green Book',
  DK: 'Daftar Kegiatan',
  LA: 'Loan Agreement',
  Monitoring: 'Monitoring Disbursement',
}

/**
 * Returns the human-readable label for a pipeline status value.
 * Falls back to the raw value if not found.
 */
export function getPipelineStatusLabel(status: string): string {
  return PIPELINE_STATUS_LABELS[status] ?? status
}

/**
 * Returns a PrimeVue severity for a pipeline status.
 */
export function getPipelineStatusSeverity(status: string): PrimeSeverity {
  if (status === 'Monitoring') return 'success'
  if (status === 'LA' || status === 'DK') return 'info'
  return 'warn'
}

interface StatusMeta {
  label: string
  severity: PrimeSeverity
}

// ---------------------------------------------------------------------------
// Generic / domain-agnostic mappings (used when no domain is specified)
// ---------------------------------------------------------------------------
const DEFAULT_MAP: Record<string, StatusMeta> = {
  active: { label: 'Aktif', severity: 'success' },
  inactive: { label: 'Tidak Aktif', severity: 'secondary' },
  deleted: { label: 'Dihapus', severity: 'danger' },
  superseded: { label: 'Digantikan', severity: 'secondary' },
  pending: { label: 'Menunggu', severity: 'warn' },
  done: { label: 'Selesai', severity: 'success' },
  failed: { label: 'Gagal', severity: 'danger' },
  tw1: { label: 'TW1', severity: 'info' },
  tw2: { label: 'TW2', severity: 'info' },
  tw3: { label: 'TW3', severity: 'info' },
  tw4: { label: 'TW4', severity: 'info' },
}

// ---------------------------------------------------------------------------
// Domain-specific overrides
// ---------------------------------------------------------------------------
const DOMAIN_MAP: Partial<Record<StatusDomain, Record<string, StatusMeta>>> = {
  blue_book: {
    berlaku: { label: 'Berlaku', severity: 'success' },
    tidak_berlaku: { label: 'Tidak Berlaku', severity: 'secondary' },
    active: { label: 'Berlaku', severity: 'success' },
    inactive: { label: 'Tidak Berlaku', severity: 'secondary' },
  },
  green_book: {
    berlaku: { label: 'Berlaku', severity: 'success' },
    tidak_berlaku: { label: 'Tidak Berlaku', severity: 'secondary' },
    active: { label: 'Berlaku', severity: 'success' },
    inactive: { label: 'Tidak Berlaku', severity: 'secondary' },
  },
  loan_agreement: {
    active: { label: 'Aktif', severity: 'success' },
    extended: { label: 'Diperpanjang', severity: 'warn' },
    closed: { label: 'Ditutup', severity: 'secondary' },
  },
  user: {
    active: { label: 'Aktif', severity: 'success' },
    inactive: { label: 'Tidak Aktif', severity: 'secondary' },
  },
  import: {
    pending: { label: 'Menunggu', severity: 'warn' },
    processing: { label: 'Diproses', severity: 'info' },
    done: { label: 'Selesai', severity: 'success' },
    failed: { label: 'Gagal', severity: 'danger' },
  },
}

function lookupMeta(status: string, domain: StatusDomain): StatusMeta | undefined {
  const normalized = status.toLowerCase()
  const domainOverrides = DOMAIN_MAP[domain]
  return domainOverrides?.[normalized] ?? DEFAULT_MAP[normalized]
}

/**
 * Returns a localized, user-facing label for the given status key.
 * Falls back to the raw status string (title-cased) if no mapping is found.
 */
export function getStatusLabel(status: string, domain: StatusDomain = 'default'): string {
  const meta = lookupMeta(status, domain)
  if (meta) return meta.label
  // Fallback: capitalize first letter so it at least looks intentional
  return status.charAt(0).toUpperCase() + status.slice(1)
}

/**
 * Returns the PrimeVue severity string for the given status key.
 * Falls back to 'secondary' if no mapping is found.
 */
export function getStatusSeverity(status: string, domain: StatusDomain = 'default'): PrimeSeverity {
  const meta = lookupMeta(status, domain)
  return meta?.severity ?? 'secondary'
}
