import type { Lender } from '@/types/master.types'
import type { PaginationMeta } from '@/types/api.types'

export type DashboardQuarter = 'TW1' | 'TW2' | 'TW3' | 'TW4'

export interface DashboardNavigationItem {
  key: string
  title: string
  description: string
  route_name: string
  icon: string
  accent: 'portfolio' | 'pipeline' | 'readiness' | 'financing' | 'institution' | 'disbursement' | 'quality'
}

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

export interface TimeSeriesPoint {
  period: string
  budget_year?: number
  quarter?: DashboardQuarter
  planned_usd: number
  realized_usd: number
  absorption_pct: number
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

export type GreenBookReadinessStatus = 'READY' | 'PARTIAL' | 'INCOMPLETE' | 'COFINANCING'

export interface GreenBookReadinessSummary {
  total_projects: number
  total_loan_usd: number
  total_grant_usd: number
  total_local_usd: number
  projects_with_cofinancing: number
  projects_incomplete: number
  projects_ready: number
  projects_partial: number
}

export interface GreenBookDisbursementYear {
  year: number
  amount_usd: number
}

export interface GreenBookFundingAllocation {
  services: number
  constructions: number
  goods: number
  trainings: number
  other: number
}

export interface GreenBookReadinessItem {
  project_id: string
  green_book_id: string
  gb_code: string
  project_name: string
  publish_year: number
  readiness_score: number
  readiness_status: Exclude<GreenBookReadinessStatus, 'COFINANCING'>
  is_cofinancing: boolean
  missing_fields: string[]
  total_funding_usd: number
  institution_name?: string
  lender_names: string[]
}

export interface GreenBookReadinessDashboard {
  summary: GreenBookReadinessSummary
  disbursement_plan_by_year: GreenBookDisbursementYear[]
  funding_allocation: GreenBookFundingAllocation
  readiness_items: GreenBookReadinessItem[]
}

export interface GreenBookReadinessParams {
  publish_year?: number
  green_book_id?: string
  institution_id?: string
  lender_id?: string
  readiness_status?: GreenBookReadinessStatus
}

export type LenderType = 'Bilateral' | 'Multilateral' | 'KSA'

export type LenderCertaintyStage =
  | 'LENDER_INDICATION'
  | 'LOI'
  | 'GB_FUNDING_SOURCE'
  | 'DK_FINANCING'
  | 'LA'

export interface LenderFinancingMixSummary {
  total_lenders: number
  bilateral_usd: number
  multilateral_usd: number
  ksa_usd: number
  cofinancing_projects: number
}

export interface LenderCertaintyPoint {
  stage: LenderCertaintyStage
  lender_id: string
  lender_name: string
  lender_type: LenderType | string
  project_count: number
  amount_usd: number
}

export interface LenderConversionItem {
  lender_id: string
  lender_name: string
  lender_type: LenderType | string
  indication_count: number
  loi_count: number
  gb_count: number
  dk_count: number
  la_count: number
  indication_usd: number
  la_usd: number
  la_conversion_pct: number
}

export interface CurrencyExposureItem {
  currency: string
  stage: LenderCertaintyStage | string
  project_count: number
  amount_original: number
  amount_usd: number
}

export interface CofinancingItem {
  project_id: string
  reference_type: 'GB' | 'DK' | string
  project_code?: string
  project_name: string
  lender_count: number
  lender_names: string[]
  amount_usd: number
}

export interface LenderFinancingMixDashboard {
  summary: LenderFinancingMixSummary
  certainty_ladder: LenderCertaintyPoint[]
  lender_conversion: LenderConversionItem[]
  currency_exposure: CurrencyExposureItem[]
  cofinancing_items: CofinancingItem[]
}

export interface LenderFinancingMixParams {
  lender_type?: LenderType
  lender_id?: string
  currency?: string
  period_id?: string
  publish_year?: number
  budget_year?: number
}

export type InstitutionRole = 'Executing Agency' | 'Implementing Agency'

export type KLPortfolioSortBy =
  | 'pipeline_usd'
  | 'la_commitment_usd'
  | 'absorption_pct'
  | 'risk_count'

export interface KLPortfolioPerformanceSummary {
  total_institutions: number
  top_exposure_institution?: string
  top_exposure_usd?: number
  lowest_absorption_institution?: string
  lowest_absorption_pct?: number
  highest_risk_institution?: string
  highest_risk_count?: number
  average_absorption_pct?: number
  total_institution_exposure_usd?: number
  total_institution_risk_count?: number
}

export interface KLPortfolioPerformanceItem {
  institution_id: string
  institution_name: string
  bb_project_count: number
  gb_project_count: number
  dk_project_count: number
  la_count: number
  pipeline_usd: number
  la_commitment_usd: number
  planned_usd: number
  realized_usd: number
  absorption_pct: number
  risk_count: number
  performance_score: number
  performance_category: 'Good' | 'Watch' | 'High Risk' | string
}

export interface KLPortfolioPerformanceDashboard {
  summary: KLPortfolioPerformanceSummary
  items: KLPortfolioPerformanceItem[]
}

export interface KLPortfolioPerformanceParams {
  institution_id?: string
  institution_role?: InstitutionRole
  period_id?: string
  publish_year?: number
  budget_year?: number
  quarter?: DashboardQuarter
  sort_by?: KLPortfolioSortBy
}

export type LARiskLevel = 'low' | 'medium' | 'high'

export interface LADisbursementSummary {
  la_count: number
  effective_count: number
  not_effective_count: number
  extended_count: number
  commitment_usd: number
  planned_usd: number
  realized_usd: number
  absorption_pct: number
  undisbursed_usd: number
}

export interface LADisbursementTrendPoint {
  period: string
  budget_year: number
  quarter: DashboardQuarter
  planned_usd: number
  realized_usd: number
  absorption_pct: number
}

export interface LAClosingRiskItem {
  loan_agreement_id: string
  loan_code: string
  project_name: string
  lender_name: string
  effective_date: string
  closing_date: string
  days_until_closing: number
  commitment_usd: number
  cumulative_realized_usd: number
  undisbursed_usd: number
  la_absorption_pct: number
  risk_type: string
  risk_level: LARiskLevel | string
}

export interface LAUnderDisbursementRiskItem {
  loan_agreement_id: string
  loan_code: string
  project_name: string
  lender_name: string
  effective_date: string
  closing_date: string
  commitment_usd: number
  cumulative_realized_usd: number
  undisbursed_usd: number
  la_absorption_pct: number
  time_elapsed_pct: number
  absorption_gap_pct: number
  remaining_months: number
  required_monthly_disbursement_usd: number
  monitoring_count: number
  is_extended: boolean
  risk_type: string
  risk_level: LARiskLevel | string
}

export interface LAComponentBreakdownItem {
  component_name: string
  la_count: number
  planned_usd: number
  realized_usd: number
  absorption_pct: number
}

export interface LADisbursementDashboard {
  summary: LADisbursementSummary
  quarterly_trend: LADisbursementTrendPoint[]
  closing_risks: LAClosingRiskItem[]
  under_disbursement_risks: LAUnderDisbursementRiskItem[]
  component_breakdown: LAComponentBreakdownItem[]
}

export interface LADisbursementParams {
  budget_year?: number
  quarter?: DashboardQuarter
  lender_id?: string
  institution_id?: string
  is_extended?: boolean
  closing_months?: 3 | 6 | 12
  risk_level?: LARiskLevel
}

export type DataQualitySeverity = 'info' | 'warning' | 'error'

export interface DataQualityGovernanceParams {
  severity?: DataQualitySeverity
  module?: string
  issue_type?: string
  only_unresolved?: boolean
  audit_days?: number
}

export interface DataQualityIssueSummary {
  total_issues: number
  error_count: number
  warning_count: number
  info_count: number
  audit_events?: number
}

export interface DataQualityIssueItem {
  severity: DataQualitySeverity | string
  module: string
  issue_type: string
  record_id: string
  record_label: string
  message: string
  recommended_action: string
  is_resolved: boolean
}

export interface AuditSummaryItem {
  label: string
  event_count: number
  last_changed_at?: string
}

export interface AuditRecentActivityItem {
  id: string
  username: string
  action: string
  table_name: string
  record_id: string
  changed_at?: string
}

export interface DataQualityAuditSummary {
  by_user: AuditSummaryItem[]
  by_table: AuditSummaryItem[]
  recent_activity: AuditRecentActivityItem[]
}

export interface DataQualityGovernanceDashboard {
  summary: DataQualityIssueSummary
  issues: DataQualityIssueItem[]
  audit_summary?: DataQualityAuditSummary
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
