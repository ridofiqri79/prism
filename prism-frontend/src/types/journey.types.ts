import type { Lender } from '@/types/master.types'

export type JourneyQuarter = 'TW1' | 'TW2' | 'TW3' | 'TW4'

export interface BBProjectSummary {
  id: string
  blue_book_id?: string
  project_identity_id?: string
  bb_code: string
  project_name: string
  blue_book_revision_label?: string
  is_latest?: boolean
  has_newer_revision?: boolean
  latest_bb_project_id?: string
  latest_blue_book_revision_label?: string
  lender_indications: LenderSummaryItem[]
}

export interface LenderSummaryItem {
  id?: string
  lender?: Pick<Lender, 'id' | 'name' | 'short_name' | 'type'>
  remarks?: string | null
}

export interface JourneyLoI {
  id: string
  lender?: Pick<Lender, 'id' | 'name' | 'short_name' | 'type'>
  subject?: string
  date?: string
  tanggal?: string
  letter_number?: string | null
}

export interface JourneyResponse {
  bb_project: BBProjectSummary
  loi: JourneyLoI[]
  gb_projects: GBProjectJourney[]
}

export interface GBProjectJourney {
  id: string
  green_book_id?: string
  gb_project_identity_id?: string
  gb_code: string
  project_name: string
  status?: string
  green_book_revision_label?: string
  is_latest?: boolean
  has_newer_revision?: boolean
  latest_gb_project_id?: string
  latest_green_book_revision_label?: string
  funding_sources: JourneyFundingSource[]
  dk_projects: DKProjectJourney[]
}

export interface JourneyFundingSource {
  id?: string
  lender?: Pick<Lender, 'id' | 'name' | 'short_name' | 'type'>
  institution?: {
    id: string
    name: string
    short_name?: string | null
  } | null
  currency?: string
  loan_original?: number
  grant_original?: number
  local_original?: number
  loan_usd?: number
  grant_usd?: number
  local_usd?: number
}

export interface DKProjectJourney {
  id: string
  project_name: string
  objectives?: string | null
  daftar_kegiatan?: {
    id?: string
    subject?: string
    date?: string
    tanggal?: string
    letter_number?: string | null
  } | null
  loan_agreements: LAJourney[]
}

export interface LAJourney {
  id: string
  loan_code: string
  lender?: Pick<Lender, 'id' | 'name' | 'short_name' | 'type'>
  agreement_date?: string
  effective_date: string
  original_closing_date?: string
  closing_date?: string
  is_extended: boolean
  extension_days?: number
  currency?: string
  amount_original?: number
  amount_usd?: number
  monitoring: MonitoringSummaryItem[]
}

export interface MonitoringSummaryItem {
  id?: string
  budget_year: number
  quarter: JourneyQuarter
  planned_usd: number
  realized_usd: number
  absorption_pct: number
}

export type JourneyStageState = 'completed' | 'pending' | 'extended' | 'warning'

export type JourneyFlowStage =
  | 'blue-book'
  | 'green-book'
  | 'daftar-kegiatan'
  | 'loan-agreement'
  | 'monitoring'

export interface JourneySummaryMetric {
  label: string
  value: string
  hint: string
  icon: string
  state: JourneyStageState
}

export interface JourneySnapshotStep {
  key: string
  label: string
  value: string
  state: JourneyStageState
  hint?: string
}

export interface JourneyFundingGroup {
  key: string
  label: string
  currency: string
  loan_usd: number
  grant_usd: number
  local_usd: number
  total_usd: number
}

export interface JourneyMatrixStage {
  key: string
  label: string
  value: string
  state: JourneyStageState
}

export interface JourneyMatrixRow {
  key: string
  project_label: string
  project_name: string
  funding_usd: number
  stages: JourneyMatrixStage[]
}

export interface JourneyFlowNode {
  name: string
  itemStyle?: {
    color: string
    borderColor?: string
    borderWidth?: number
  }
  label?: {
    color: string
    fontWeight?: number
  }
}

export interface JourneyFlowLink {
  source: string
  target: string
  value: number
  rawValue: number
  label: string
  lineStyle?: {
    color?: string
    opacity?: number
  }
}
