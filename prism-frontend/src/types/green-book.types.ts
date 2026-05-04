import type { ProjectAuditEntry } from '@/types/audit.types'
import type { BBProject } from '@/types/blue-book.types'
import type {
  BappenasPartner,
  Institution,
  ListParams,
  Lender,
  ProgramTitle,
  Region,
} from '@/types/master.types'

export type GreenBookStatus = 'active' | 'superseded'
export type GBProjectStatus = 'active'

export interface GreenBook {
  id: string
  publish_year: number
  replaces_green_book_id?: string | null
  revision_number: number
  status: GreenBookStatus
  project_count: number
  created_at?: string
  updated_at?: string
}

export interface GreenBookPayload {
  publish_year: number
  replaces_green_book_id?: string | null
  revision_number: number
  status: GreenBookStatus
}

export interface GreenBookListParams extends ListParams {
  search?: string
  publish_year?: number[]
  status?: GreenBookStatus[]
}

export interface GBProjectListParams extends ListParams {
  search?: string
  bb_project_ids?: string[]
  executing_agency_ids?: string[]
  location_ids?: string[]
  status?: GBProjectStatus[]
}

export interface BBProjectSummary {
  id: string
  blue_book_id?: string
  project_identity_id?: string
  bb_code: string
  project_name: string
  is_latest?: boolean
  has_newer_revision?: boolean
}

export interface GBProject {
  id: string
  green_book_id?: string
  gb_project_identity_id: string
  program_title_id?: string
  gb_code: string
  project_name: string
  duration?: number | null
  objective?: string | null
  scope_of_project?: string | null
  program_title?: ProgramTitle
  bb_projects: BBProjectSummary[]
  bappenas_partners: BappenasPartner[]
  executing_agencies: Institution[]
  implementing_agencies: Institution[]
  locations: Region[]
  activities: GBActivity[]
  funding_sources: GBFundingSource[]
  disbursement_plan: GBDisbursementPlan[]
  funding_allocations: GBFundingAllocation[]
  status: GBProjectStatus
  is_latest: boolean
  has_newer_revision: boolean
  created_at?: string
  updated_at?: string
}

export interface GBActivity {
  id: string
  activity_name: string
  implementation_location?: string | null
  piu?: string | null
  sort_order: number
}

export interface GBActivityPayload {
  activity_name: string
  implementation_location?: string | null
  piu?: string | null
  sort_order: number
}

export interface GBFundingSource {
  id: string
  lender: Lender
  institution?: Institution
  currency: string
  loan_original: number
  grant_original: number
  local_original: number
  loan_usd: number
  grant_usd: number
  local_usd: number
}

export interface GBFundingSourcePayload {
  lender_id: string
  institution_id?: string | null
  currency: string
  loan_original: number
  grant_original: number
  local_original: number
  loan_usd: number
  grant_usd: number
  local_usd: number
}

export interface GBDisbursementPlan {
  id: string
  year: number
  amount_usd: number
}

export interface GBDisbursementPlanPayload {
  year: number
  amount_usd: number
}

export interface GBFundingAllocation {
  id: string
  gb_activity_id: string
  activity_name?: string
  sort_order?: number
  services: number
  constructions: number
  goods: number
  trainings: number
  other: number
}

export interface GBAllocationValues {
  services: number
  constructions: number
  goods: number
  trainings: number
  other: number
}

export interface GBFundingAllocationPayload extends GBAllocationValues {
  activity_index: number
}

export interface GBProjectPayload {
  gb_project_identity_id?: string | null
  program_title_id: string
  gb_code: string
  project_name: string
  duration?: number | null
  objective?: string | null
  scope_of_project?: string | null
  bb_project_ids: string[]
  bappenas_partner_ids: string[]
  executing_agency_ids: string[]
  implementing_agency_ids: string[]
  location_ids: string[]
  activities: GBActivityPayload[]
  funding_sources: GBFundingSourcePayload[]
  disbursement_plan: GBDisbursementPlanPayload[]
  funding_allocations: GBFundingAllocationPayload[]
}

export interface BBProjectOption extends Pick<BBProject, 'id' | 'bb_code' | 'project_name'> {
  blue_book_id?: string
  project_identity_id?: string
  is_latest?: boolean
  has_newer_revision?: boolean
}

export interface GBProjectHistoryItem {
  id: string
  gb_project_identity_id: string
  green_book_id: string
  gb_code: string
  project_name: string
  book_label: string
  publish_year: number
  revision_number: number
  book_status: GreenBookStatus
  is_latest: boolean
  used_by_downstream: boolean
  bb_projects?: BBProjectSummary[]
  last_changed_by?: string
  last_changed_at?: string
  last_change_summary?: string
  audit_entries?: ProjectAuditEntry[]
}
