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
  original_closing_date: string
  closing_date: string
  is_extended: boolean
  extension_days: number
  currency: string
  amount_original: number
  amount_usd: number
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
}

export interface LoanAgreementListParams extends ListParams {
  search?: string
  lender_id?: string
  is_extended?: boolean
  closing_date_before?: string
  risk_codes?: LoanAgreementRiskCode[]
}

export type LoanAgreementRiskCode = 'EXTENDED_LOAN' | 'CLOSING_RISK'

export interface DKProjectLoanOption extends DKProject {
  label: string
  daftar_kegiatan_subject?: string
}
