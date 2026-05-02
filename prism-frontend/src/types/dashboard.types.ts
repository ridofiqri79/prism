import type { Lender } from '@/types/master.types'
import type { PaginationMeta } from '@/types/api.types'

export type DashboardQuarter = 'TW1' | 'TW2' | 'TW3' | 'TW4'

export interface MetricCard {
  key: string
  label: string
  value: number
  unit?: 'USD' | 'percent' | 'project' | string
  category?: string
}

export interface StageMetric {
  stage: string
  label: string
  project_count: number
  amount_usd: number
}

export interface BreakdownItem {
  id?: string
  key?: string
  label: string
  item_count?: number
  amount_usd?: number
  realized_usd?: number
  percentage?: number
}

export interface RiskItem {
  id?: string
  risk_type?: string
  reference_id?: string
  reference_type?: string
  journey_bb_project_id?: string
  code?: string
  title: string
  description?: string
  severity: 'low' | 'medium' | 'high' | string
  amount_usd?: number
  days_until_closing?: number
  absorption_pct?: number
  score?: number
}

export interface ExecutivePortfolioDashboard {
  cards: MetricCard[]
  funnel: StageMetric[]
  top_institutions: BreakdownItem[]
  top_lenders: BreakdownItem[]
  risk_items: RiskItem[]
  insights: string[]
}

export type DashboardFilterOptions = Record<string, BreakdownItem[]>

export type PipelineBottleneckStage =
  | 'BB_NO_LENDER'
  | 'INDICATION_NO_LOI'
  | 'LOI_NO_GB'
  | 'GB_NO_DK'
  | 'DK_NO_LA'
  | 'LA_NOT_EFFECTIVE'
  | 'EFFECTIVE_NO_MONITORING'

export type PipelineBottleneckSort = 'stage' | 'project_name' | 'amount_usd' | 'age_days'

export type SortOrder = 'asc' | 'desc'

export interface PipelineStageSummary {
  stage: PipelineBottleneckStage
  label: string
  project_count: number
  amount_usd: number
  avg_age_days: number
}

export interface PipelineBottleneckItem {
  project_id: string
  reference_type: string
  journey_bb_project_id?: string
  code?: string
  project_name: string
  current_stage: PipelineBottleneckStage
  stage_label: string
  age_days: number
  amount_usd: number
  institution_name?: string
  lender_names: string[]
  recommended_action: string
  relevant_at?: string
}

export interface PipelineBottleneckResponse {
  stage_summary: PipelineStageSummary[]
  items: PipelineBottleneckItem[]
}

export interface PipelineBottleneckParams {
  stage?: PipelineBottleneckStage
  period_id?: string
  publish_year?: number
  institution_id?: string
  lender_id?: string
  min_age_days?: number
  page?: number
  limit?: number
  sort?: PipelineBottleneckSort
  order?: SortOrder
  search?: string
}

export interface PipelineBottleneckApiResponse {
  data: PipelineBottleneckResponse
  meta: PaginationMeta
}

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
  loan_agreement: LAJourney | null
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
  quarter: DashboardQuarter
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

export interface DashboardFilterParams {
  period_id?: string
  publish_year?: number
  budget_year?: number
  quarter?: DashboardQuarter
  lender_id?: string
  institution_id?: string
  include_history?: boolean
}

export type DashboardSummaryApiResponse = Omit<
  DashboardSummary,
  'total_realized_usd' | 'total_realisasi_usd'
> & {
  total_realized_usd?: number
  total_realisasi_usd?: number
  bb_pipeline_usd?: number
  gb_pipeline_usd?: number
  la_commitment_usd?: number
  realized_disbursement_usd?: number
  absorption_pct?: number
  metrics?: MetricCard[]
}

export type MonitoringSummaryApiResponse = Partial<Omit<MonitoringSummary, 'by_lender'>> & {
  by_lender?: Array<
    Partial<Omit<LenderSummary, 'lender'>> & {
      lender: Lender
    }
  >
}
