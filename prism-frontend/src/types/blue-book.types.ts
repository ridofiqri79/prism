import type {
  BappenasPartner,
  Institution,
  Lender,
  NationalPriority,
  Period,
  ProgramTitle,
  Region,
} from '@/types/master.types'

export type BlueBookStatus = 'active' | 'superseded'
export type BBProjectStatus = 'active' | 'deleted'
export type FundingType = 'Foreign' | 'Counterpart'

export interface BlueBook {
  id: string
  period: Period
  replaces_blue_book_id?: string | null
  publish_date: string
  revision_number: number
  revision_year?: number | null
  status: BlueBookStatus
  created_at?: string
  updated_at?: string
}

export interface BlueBookPayload {
  period_id: string
  replaces_blue_book_id?: string | null
  publish_date: string
  revision_number: number
  revision_year?: number | null
}

export interface BBProject {
  id: string
  blue_book_id?: string
  project_identity_id: string
  program_title_id?: string
  bb_code: string
  project_name: string
  program_title?: ProgramTitle
  bappenas_partners: BappenasPartner[]
  executing_agencies: Institution[]
  implementing_agencies: Institution[]
  locations: Region[]
  national_priorities: NationalPriority[]
  project_costs: BBProjectCost[]
  lender_indications: LenderIndication[]
  duration?: number | null
  objective?: string | null
  scope_of_work?: string | null
  outputs?: string | null
  outcomes?: string | null
  status: BBProjectStatus
  is_latest: boolean
  has_newer_revision: boolean
  created_at?: string
  updated_at?: string
}

export interface BBProjectCost {
  id: string
  funding_type: FundingType
  funding_category: string
  amount_usd: number
}

export interface ProjectCostPayload {
  funding_type: FundingType
  funding_category: string
  amount_usd: number
}

export interface LenderIndication {
  id: string
  lender: Lender
  remarks?: string | null
}

export interface LenderIndicationPayload {
  lender_id: string
  remarks?: string | null
}

export interface BBProjectPayload {
  project_identity_id?: string | null
  program_title_id: string
  bappenas_partner_ids: string[]
  bb_code: string
  project_name: string
  duration?: number | null
  objective?: string | null
  scope_of_work?: string | null
  outputs?: string | null
  outcomes?: string | null
  executing_agency_ids: string[]
  implementing_agency_ids: string[]
  location_ids: string[]
  national_priority_ids: string[]
  project_costs: ProjectCostPayload[]
  lender_indications: LenderIndicationPayload[]
}

export interface BBProjectHistoryItem {
  id: string
  project_identity_id: string
  blue_book_id: string
  bb_code: string
  project_name: string
  book_label: string
  revision_number: number
  revision_year?: number | null
  book_status: BlueBookStatus
  is_latest: boolean
  used_by_downstream: boolean
}

export interface LoI {
  id: string
  bb_project_id?: string
  lender: Lender
  subject: string
  date: string
  letter_number?: string | null
  created_at?: string
  updated_at?: string
}

export interface LoIPayload {
  lender_id: string
  subject: string
  date: string
  letter_number?: string | null
}
