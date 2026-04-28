import type { LenderType } from '@/types/master.types'

export type ProjectStatus = 'Pipeline' | 'Ongoing'
export type ProjectPipelineStatus = 'BB' | 'GB' | 'DK' | 'LA' | 'Monitoring'
export type ProjectMasterSortOrder = 'asc' | 'desc'
export type ProjectMasterSortField =
  | 'bb_code'
  | 'project_name'
  | 'loan_types'
  | 'indication_lenders'
  | 'executing_agencies'
  | 'fixed_lenders'
  | 'project_status'
  | 'pipeline_status'
  | 'program_title'
  | 'locations'
  | 'foreign_loan_usd'
  | 'dk_dates'
export type ProjectMasterColumnKey =
  | 'loan_types'
  | 'indication_lenders'
  | 'executing_agencies'
  | 'fixed_lenders'
  | 'status'
  | 'program_title'
  | 'locations'
  | 'foreign_loan_usd'
  | 'dk_dates'

export interface ProjectMasterColumnConfig {
  key: ProjectMasterColumnKey
  label: string
  sortField: ProjectMasterSortField
  defaultVisible: boolean
}

export interface ProjectMasterRow {
  id: string
  blue_book_id: string
  project_identity_id: string
  bb_code: string
  project_name: string
  loan_types: LenderType[]
  indication_lenders: string[]
  executing_agencies: string[]
  fixed_lenders: string[]
  project_status: ProjectStatus
  pipeline_status: ProjectPipelineStatus
  program_title: string
  locations: string[]
  foreign_loan_usd: number
  dk_dates: string[]
  is_latest: boolean
  has_newer_revision: boolean
  blue_book_revision_label: string
}

export interface ProjectMasterListParams {
  page?: number
  limit?: number
  sort?: ProjectMasterSortField
  order?: ProjectMasterSortOrder
  loan_types?: LenderType[]
  indication_lender_ids?: string[]
  executing_agency_ids?: string[]
  fixed_lender_ids?: string[]
  project_statuses?: ProjectStatus[]
  pipeline_statuses?: ProjectPipelineStatus[]
  program_title_ids?: string[]
  region_ids?: string[]
  foreign_loan_min?: number
  foreign_loan_max?: number
  dk_date_from?: string
  dk_date_to?: string
  search?: string
  include_history?: boolean
}

export interface ProjectMasterFilterState {
  loan_types: LenderType[]
  indication_lender_ids: string[]
  executing_agency_ids: string[]
  fixed_lender_ids: string[]
  project_statuses: ProjectStatus[]
  pipeline_statuses: ProjectPipelineStatus[]
  program_title_ids: string[]
  region_ids: string[]
  foreign_loan_min: number | null
  foreign_loan_max: number | null
  dk_date_from: string
  dk_date_to: string
  search: string
  include_history: boolean
}
