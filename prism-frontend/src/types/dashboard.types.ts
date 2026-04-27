import type { Lender } from '@/types/master.types'

export type DashboardQuarter = 'TW1' | 'TW2' | 'TW3' | 'TW4'

export interface DashboardSummary {
  total_bb_projects: number
  total_gb_projects: number
  total_loan_agreements: number
  total_amount_usd: number
  total_realized_usd: number
  total_realisasi_usd: number
  overall_absorption_pct: number
  active_monitoring: number
}

export interface MonitoringSummary {
  budget_year?: number
  tahun_anggaran?: number
  quarter?: DashboardQuarter
  triwulan?: DashboardQuarter
  total_planned_usd: number
  total_rencana_usd: number
  total_realized_usd: number
  total_realisasi_usd: number
  absorption_pct: number
  by_lender: LenderSummary[]
}

export interface LenderSummary {
  lender: Lender
  planned_usd: number
  rencana_usd: number
  realized_usd: number
  realisasi_usd: number
  absorption_pct: number
}

export interface BBProjectSummary {
  id: string
  blue_book_id?: string
  bb_code: string
  project_name: string
  lender_indications?: LenderSummaryItem[]
}

export interface LenderSummaryItem {
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
  gb_code: string
  project_name: string
  funding_sources?: JourneyFundingSource[]
  dk_projects: DKProjectJourney[]
}

export interface JourneyFundingSource {
  lender?: Pick<Lender, 'id' | 'name' | 'short_name' | 'type'>
  currency?: string
  amount_usd?: number
  amount_original?: number
}

export interface DKProjectJourney {
  id: string
  objectives?: string | null
  daftar_kegiatan?: {
    id?: string
    subject?: string
    date?: string
    tanggal?: string
    letter_number?: string | null
  } | null
  loan_agreement: LAJourney | null
}

export interface LAJourney {
  id: string
  loan_code: string
  effective_date: string
  closing_date?: string
  is_extended: boolean
  monitoring: MonitoringSummaryItem[]
}

export interface MonitoringSummaryItem {
  id?: string
  budget_year: number
  quarter: DashboardQuarter
  planned_usd: number
  realized_usd: number
  absorption_pct: number
}

export interface DashboardFilterParams {
  budget_year?: number
  quarter?: DashboardQuarter
  lender_id?: string
}

export type DashboardSummaryApiResponse = Omit<
  DashboardSummary,
  'total_realized_usd' | 'total_realisasi_usd'
> & {
  total_realized_usd?: number
  total_realisasi_usd?: number
}

export type MonitoringSummaryApiResponse = Partial<
  Omit<MonitoringSummary, 'by_lender'>
> & {
  by_lender?: Array<
    Partial<Omit<LenderSummary, 'lender'>> & {
      lender: Lender
    }
  >
}
