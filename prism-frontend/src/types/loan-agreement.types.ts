import type { DKProject } from '@/types/daftar-kegiatan.types'
import type { Lender, ListParams } from '@/types/master.types'

export interface DKProjectSummary {
  id: string
  dk_id?: string
  objectives?: string | null
  gb_code?: string
  project_name?: string
}

export interface LoanAgreement {
  id: string
  dk_project?: DKProjectSummary | null
  dk_project_id?: string
  lender: Lender
  loan_code: string
  agreement_date: string
  effective_date: string
  original_closing_date: string | null
  closing_date: string
  is_extended: boolean
  extension_days: number
  currency: string
  amount_original: number
  amount_usd: number
  cumulative_disbursement: number
  cumulative_disbursement_usd: number | null
  disbursement_ratio: number | null
  estimated_time_ratio: number | null
  performance_value: number | null
  performance_status: string | null
  kurs_tengah_bi: number | null
  kurs_cut_off_date: string | null
  created_at?: string
  updated_at?: string
}

export interface LoanAgreementPayload {
  dk_project_id: string
  lender_id: string
  loan_code: string
  agreement_date: string
  effective_date: string
  original_closing_date: string
  closing_date: string
  currency: string
  amount_original: number
  amount_usd: number
  cumulative_disbursement: number
}

export interface LoanAgreementListParams extends ListParams {
  period_ids?: string[]
  search?: string
  lender_id?: string
  is_extended?: boolean
  closing_date_before?: string
}

export interface DKProjectLoanOption extends DKProject {
  label: string
  daftar_kegiatan_subject?: string
}
