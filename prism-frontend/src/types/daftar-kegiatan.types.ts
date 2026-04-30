import type { GBProject } from '@/types/green-book.types'
import type {
  BappenasPartner,
  Institution,
  ListParams,
  Lender,
  ProgramTitle,
  Region,
} from '@/types/master.types'

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
  duration?: number | null
  objectives?: string | null
  gb_projects: GBProjectSummary[]
  bappenas_partners: BappenasPartner[]
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

export interface DaftarKegiatanListParams extends ListParams {
  search?: string
  date_from?: string
  date_to?: string
}

export interface DKProjectListParams extends ListParams {
  search?: string
  gb_project_ids?: string[]
  executing_agency_ids?: string[]
  location_ids?: string[]
  lender_ids?: string[]
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
  duration?: number | null
  objectives?: string | null
  gb_project_ids: string[]
  bappenas_partner_ids: string[]
  location_ids: string[]
  financing_details: DKFinancingDetailPayload[]
  loan_allocations: DKLoanAllocationPayload[]
  activity_details: DKActivityDetailPayload[]
}

export interface GBProjectOption extends Pick<
  GBProject,
  | 'id'
  | 'gb_code'
  | 'project_name'
  | 'program_title_id'
  | 'duration'
  | 'objective'
  | 'bb_projects'
  | 'bappenas_partners'
  | 'executing_agencies'
  | 'implementing_agencies'
  | 'locations'
  | 'activities'
  | 'funding_sources'
> {
  green_book_id?: string
  gb_project_identity_id?: string
  is_latest?: boolean
  has_newer_revision?: boolean
}
