import type { ListParams } from '@/types/master.types'

export type Quarter = 'TW1' | 'TW2' | 'TW3' | 'TW4'

export interface MonitoringKomponen {
  id?: string
  component_name: string
  planned_la: number
  planned_usd: number
  planned_idr: number
  realized_la: number
  realized_usd: number
  realized_idr: number
}

export interface MonitoringDisbursement {
  id: string
  loan_agreement_id: string
  budget_year: number
  quarter: Quarter
  exchange_rate_usd_idr: number
  exchange_rate_la_idr: number
  planned_la: number
  planned_usd: number
  planned_idr: number
  realized_la: number
  realized_usd: number
  realized_idr: number
  absorption_pct: number
  komponen: MonitoringKomponen[]
  created_at?: string
  updated_at?: string
}

export interface MonitoringPayload {
  budget_year: number
  quarter: Quarter
  exchange_rate_usd_idr: number
  exchange_rate_la_idr: number
  planned_la: number
  planned_usd: number
  planned_idr: number
  realized_la: number
  realized_usd: number
  realized_idr: number
  komponen?: MonitoringKomponen[]
}

export interface MonitoringListParams extends ListParams {
  search?: string
  budget_year?: number
  quarter?: Quarter
}

export type MonitoringApiResponse = Omit<MonitoringDisbursement, 'absorption_pct'> & {
  absorption_pct?: number
  penyerapan_pct?: number
}
