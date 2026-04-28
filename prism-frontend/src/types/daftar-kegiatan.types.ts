import type { GBProject } from '@/types/green-book.types'
import type { Institution, Lender, ProgramTitle, Region } from '@/types/master.types'

export interface DaftarKegiatan {
  id: string
  letter_number?: string | null
  subject: string
  date: string
  project_count?: number
  created_at?: string
  updated_at?: string
}

export interface DKProject {
  id: string
  dk?: DaftarKegiatan
  dk_id?: string
  program_title_id?: string | null
  institution_id?: string | null
  program_title?: ProgramTitle
  institution?: Institution
  duration?: string | null
  objectives?: string | null
  gb_projects: GBProjectSummary[]
  locations: Region[]
  financing_details: DKFinancingDetail[]
  loan_allocations: DKLoanAllocation[]
  activity_details: DKActivityDetail[]
  created_at?: string
  updated_at?: string
}

export interface GBProjectSummary {
  id: string
  gb_project_identity_id?: string
  green_book_id?: string
  gb_code: string
  project_name: string
  is_latest?: boolean
  has_newer_revision?: boolean
}

export interface DKFinancingDetail {
  id: string
  lender?: Lender | null
  currency: string
  amount_original: number
  grant_original: number
  counterpart_original: number
  amount_usd: number
  grant_usd: number
  counterpart_usd: number
  remarks?: string | null
}

export interface DKLoanAllocation {
  id: string
  institution?: Institution | null
  currency: string
  amount_original: number
  grant_original: number
  counterpart_original: number
  amount_usd: number
  grant_usd: number
  counterpart_usd: number
  remarks?: string | null
}

export interface DKActivityDetail {
  id: string
  activity_number: number
  activity_name: string
}

export interface DaftarKegiatanPayload {
  letter_number?: string | null
  subject: string
  date: string
}

export interface DKFinancingDetailPayload {
  lender_id: string
  currency: string
  amount_original: number
  grant_original: number
  counterpart_original: number
  amount_usd: number
  grant_usd: number
  counterpart_usd: number
  remarks?: string | null
}

export interface DKLoanAllocationPayload {
  institution_id: string
  currency: string
  amount_original: number
  grant_original: number
  counterpart_original: number
  amount_usd: number
  grant_usd: number
  counterpart_usd: number
  remarks?: string | null
}

export interface DKActivityDetailPayload {
  activity_number: number
  activity_name: string
}

export interface DKProjectPayload {
  program_title_id?: string | null
  institution_id: string
  duration?: string | null
  objectives?: string | null
  gb_project_ids: string[]
  location_ids: string[]
  financing_details: DKFinancingDetailPayload[]
  loan_allocations: DKLoanAllocationPayload[]
  activity_details: DKActivityDetailPayload[]
}

export interface GBProjectOption extends Pick<GBProject, 'id' | 'gb_code' | 'project_name' | 'bb_projects' | 'funding_sources'> {
  green_book_id?: string
  gb_project_identity_id?: string
  is_latest?: boolean
  has_newer_revision?: boolean
}
