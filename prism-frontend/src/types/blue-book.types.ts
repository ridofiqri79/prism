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
  publish_date: string
  revision_number: number
  revision_year?: number | null
  status: BlueBookStatus
  created_at?: string
  updated_at?: string
}

export interface BlueBookPayload {
  period_id: string
  publish_date: string
  revision_number: number
  revision_year?: number | null
}

export interface BBProject {
  id: string
  blue_book_id?: string
  program_title_id?: string
  bappenas_partner_id?: string
  bb_code: string
  project_name: string
  program_title?: ProgramTitle
  bappenas_partner?: BappenasPartner
  executing_agencies: Institution[]
  implementing_agencies: Institution[]
  locations: Region[]
  national_priorities: NationalPriority[]
  project_costs: BBProjectCost[]
  lender_indications: LenderIndication[]
  duration?: string | null
  objective?: string | null
  scope_of_work?: string | null
  outputs?: string | null
  outcomes?: string | null
  status: BBProjectStatus
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
  program_title_id: string
  bappenas_partner_id: string
  bb_code: string
  project_name: string
  duration?: string | null
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

