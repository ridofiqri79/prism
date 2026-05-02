import type { Institution, Lender, LenderType, ProgramTitle, Region } from '@/types/master.types'

export type DashboardQuarter = 'TW1' | 'TW2' | 'TW3' | 'TW4'
export type DashboardAnalyticsTarget =
  | 'projects'
  | 'monitoring'
  | 'loan_agreements'
  | 'spatial_distribution'
export type DashboardAnalyticsPipelineStage = 'BB' | 'GB' | 'DK' | 'LA' | 'Monitoring'
export type DashboardAnalyticsProjectStatus = 'Pipeline' | 'Ongoing'
export type DashboardAnalyticsSeverity = 'info' | 'success' | 'warning' | 'danger' | 'secondary'
export type DashboardAnalyticsAbsorptionStatus = 'low' | 'normal' | 'high'

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

export type MonitoringSummaryApiResponse = Partial<Omit<MonitoringSummary, 'by_lender'>> & {
  by_lender?: Array<
    Partial<Omit<LenderSummary, 'lender'>> & {
      lender: Lender
    }
  >
}

export interface DashboardDrilldownQuery {
  target: DashboardAnalyticsTarget | string
  query: Record<string, string[]>
}

export interface DashboardAnalyticsFilterParams {
  budget_year?: number
  quarter?: DashboardQuarter
  lender_ids?: string[]
  lender_types?: LenderType[]
  institution_ids?: string[]
  pipeline_statuses?: DashboardAnalyticsPipelineStage[]
  project_statuses?: DashboardAnalyticsProjectStatus[]
  region_ids?: string[]
  program_title_ids?: string[]
  foreign_loan_min?: number
  foreign_loan_max?: number
  include_history?: boolean
  low_absorption_threshold?: number
  closing_months_threshold?: number
  stale_monitoring_quarters?: number
}

export interface DashboardAnalyticsFilterState {
  budget_year: number | null
  quarter: DashboardQuarter | null
  lender_ids: string[]
  lender_types: LenderType[]
  institution_ids: string[]
  pipeline_statuses: DashboardAnalyticsPipelineStage[]
  project_statuses: DashboardAnalyticsProjectStatus[]
  region_ids: string[]
  program_title_ids: string[]
  foreign_loan_min: number | null
  foreign_loan_max: number | null
  include_history: boolean
}

export type DashboardAnalyticsLenderRef = Pick<Lender, 'id' | 'name' | 'short_name' | 'type'>
export type DashboardAnalyticsInstitutionRef = Pick<
  Institution,
  'id' | 'name' | 'short_name' | 'level'
>
export type DashboardAnalyticsRegionRef = Pick<Region, 'id' | 'name' | 'type'>
export type DashboardAnalyticsProgramTitleRef = Pick<ProgramTitle, 'id' | 'title'>

export interface AnalyticsMoneyMetric {
  key: string
  label: string
  value: number
  format?: 'number' | 'currency' | 'percent'
  unit?: string
  severity?: DashboardAnalyticsSeverity
  drilldown?: DashboardDrilldownQuery
}

export interface AnalyticsRankedItem {
  rank?: number
  id: string
  name: string
  dimension?: string
  planned_usd?: number
  realized_usd?: number
  absorption_pct?: number
  variance_usd?: number
  status?: DashboardAnalyticsAbsorptionStatus | string
  drilldown?: DashboardDrilldownQuery
}

export interface AnalyticsStageBreakdown {
  BB: number
  GB: number
  DK: number
  LA: number
  Monitoring: number
}

export interface AnalyticsBreakdownTableColumn {
  key: string
  label: string
  kind?: 'text' | 'number' | 'currency' | 'percent' | 'badge' | 'drilldown' | 'absorption'
  align?: 'left' | 'right' | 'center'
}

export interface AnalyticsBreakdownTableRow {
  id: string
  label?: string
  severity?: DashboardAnalyticsSeverity | string
  cells: Record<string, string | number | null | undefined>
  drilldown?: DashboardDrilldownQuery
}

export interface DashboardAnalyticsPortfolioOverview {
  project_count: number
  assignment_count: number
  total_pipeline_loan_usd: number
  total_agreement_amount_usd: number
  total_planned_usd: number
  total_realized_usd: number
  absorption_pct: number
}

export interface DashboardAnalyticsPipelineFunnelItem {
  stage: DashboardAnalyticsPipelineStage
  project_count: number
  total_loan_usd: number
  drilldown: DashboardDrilldownQuery
}

export interface DashboardAnalyticsInsight {
  key: string
  label: string
  value: number
  severity?: DashboardAnalyticsSeverity | string
  drilldown: DashboardDrilldownQuery
}

export interface DashboardAnalyticsOverview {
  portfolio: DashboardAnalyticsPortfolioOverview
  pipeline_funnel: DashboardAnalyticsPipelineFunnelItem[]
  top_insights: DashboardAnalyticsInsight[]
  drilldown: DashboardDrilldownQuery
}

export interface DashboardInstitutionAnalyticsSummary {
  institution_count: number
  project_count: number
  assignment_count: number
  total_agreement_amount_usd: number
  total_planned_usd: number
  total_realized_usd: number
  absorption_pct: number
}

export interface DashboardInstitutionAnalyticsItem {
  institution: DashboardAnalyticsInstitutionRef
  project_count: number
  assignment_count: number
  loan_agreement_count: number
  monitoring_count: number
  agreement_amount_usd: number
  planned_usd: number
  realized_usd: number
  absorption_pct: number
  pipeline_breakdown: AnalyticsStageBreakdown
  drilldown: DashboardDrilldownQuery
}

export interface DashboardInstitutionAnalytics {
  summary: DashboardInstitutionAnalyticsSummary
  items: DashboardInstitutionAnalyticsItem[]
  drilldown: DashboardDrilldownQuery
}

export interface DashboardLenderAnalyticsSummary {
  lender_count: number
  loan_agreement_count: number
  total_agreement_amount_usd: number
  total_planned_usd: number
  total_realized_usd: number
  absorption_pct: number
}

export interface DashboardLenderAnalyticsItem {
  lender: DashboardAnalyticsLenderRef
  loan_agreement_count: number
  project_count: number
  institution_count: number
  monitoring_count: number
  agreement_amount_usd: number
  planned_usd: number
  realized_usd: number
  absorption_pct: number
  drilldown: DashboardDrilldownQuery
}

export interface DashboardLenderInstitutionMatrixItem {
  institution: DashboardAnalyticsInstitutionRef
  lender: DashboardAnalyticsLenderRef
  project_count: number
  loan_agreement_count: number
  monitoring_count: number
  agreement_amount_usd: number
  planned_usd: number
  realized_usd: number
  absorption_pct: number
  drilldown: DashboardDrilldownQuery
}

export interface DashboardLenderAnalytics {
  summary: DashboardLenderAnalyticsSummary
  items: DashboardLenderAnalyticsItem[]
  lender_institution_matrix: DashboardLenderInstitutionMatrixItem[]
  drilldown: DashboardDrilldownQuery
}

export interface DashboardAbsorptionSummary {
  planned_usd: number
  realized_usd: number
  absorption_pct: number
}

export interface DashboardAbsorptionRankedItem extends AnalyticsRankedItem {
  planned_usd: number
  realized_usd: number
  absorption_pct: number
  variance_usd: number
  status: DashboardAnalyticsAbsorptionStatus | string
}

export interface DashboardAbsorptionAnalytics {
  summary: DashboardAbsorptionSummary
  by_institution: DashboardAbsorptionRankedItem[]
  by_project: DashboardAbsorptionRankedItem[]
  by_lender: DashboardAbsorptionRankedItem[]
  drilldown: DashboardDrilldownQuery
}

export interface DashboardYearlySummary extends DashboardAbsorptionSummary {
  loan_agreement_count: number
  project_count: number
}

export interface DashboardYearlyItem {
  budget_year: number
  quarter: DashboardQuarter
  planned_usd: number
  realized_usd: number
  absorption_pct: number
  loan_agreement_count: number
  project_count: number
  drilldown: DashboardDrilldownQuery
}

export interface DashboardYearlyAnalytics {
  summary: DashboardYearlySummary
  items: DashboardYearlyItem[]
  drilldown: DashboardDrilldownQuery
}

export interface DashboardLenderProportionItem {
  type: LenderType
  project_count: number
  lender_count: number
  amount_usd: number
  share_pct: number
  drilldown: DashboardDrilldownQuery
}

export interface DashboardLenderProportionStage {
  stage: string
  items: DashboardLenderProportionItem[]
}

export interface DashboardLenderProportionAnalytics {
  by_stage: DashboardLenderProportionStage[]
  drilldown: DashboardDrilldownQuery
}

export interface DashboardRiskSummary {
  low_absorption_count: number
  effective_without_monitoring_count: number
  closing_risk_count: number
  extended_loan_count: number
  data_quality_issue_count: number
  bottleneck_project_count: number
}

export interface DashboardRiskThresholds {
  low_absorption_threshold: number
  closing_months_threshold: number
  closing_absorption_threshold: number
  stale_monitoring_quarters: number
}

export interface DashboardRiskCard {
  code: string
  label: string
  count: number
  severity: DashboardAnalyticsSeverity | string
  amount_usd?: number
  drilldown: DashboardDrilldownQuery
}

export interface DashboardLoanAgreementRiskItem {
  risk_code: string
  risk_label?: string
  severity: DashboardAnalyticsSeverity | string
  project_id: string
  project_name: string
  loan_agreement_id: string
  loan_code: string
  lender: DashboardAnalyticsLenderRef
  institution?: DashboardAnalyticsInstitutionRef
  effective_date?: string
  original_closing_date?: string
  closing_date?: string
  budget_year?: number
  quarter?: DashboardQuarter
  planned_usd: number
  realized_usd: number
  absorption_pct: number
  agreement_amount_usd?: number
  days_since_effective?: number
  days_to_closing?: number
  months_to_closing?: number
  extension_days?: number
  stale_quarters?: number
  monitoring_status?: string
  drilldown: DashboardDrilldownQuery
}

export interface DashboardPipelineBottleneckItem {
  stage: DashboardAnalyticsPipelineStage
  label: string
  project_count: number
  total_loan_usd: number
  oldest_date?: string
  severity: DashboardAnalyticsSeverity | string
  drilldown: DashboardDrilldownQuery
}

export interface DashboardRiskWatchlists {
  low_absorption_projects: DashboardLoanAgreementRiskItem[]
  effective_without_monitoring: DashboardLoanAgreementRiskItem[]
  closing_risks: DashboardLoanAgreementRiskItem[]
  extended_loans: DashboardLoanAgreementRiskItem[]
  pipeline_bottlenecks: DashboardPipelineBottleneckItem[]
}

export interface DashboardExtendedLoanBreakdown {
  dimension: string
  entity: DashboardAnalyticsLenderRef | DashboardAnalyticsInstitutionRef
  loan_agreement_count: number
  amount_usd: number
  average_extension_days: number
  drilldown: DashboardDrilldownQuery
}

export interface DashboardExtendedLoanInsight {
  count: number
  amount_usd: number
  average_extension_days: number
  by_lender: DashboardExtendedLoanBreakdown[]
  by_institution: DashboardExtendedLoanBreakdown[]
}

export interface DashboardDataQualityItem {
  code: string
  label: string
  stage: string
  severity: DashboardAnalyticsSeverity | string
  count: number
  drilldown: DashboardDrilldownQuery
}

export interface DashboardRiskAnalytics {
  summary: DashboardRiskSummary
  thresholds: DashboardRiskThresholds
  risk_cards: DashboardRiskCard[]
  watchlists: DashboardRiskWatchlists
  extended_loan_insight: DashboardExtendedLoanInsight
  data_quality: DashboardDataQualityItem[]
  drilldown: DashboardDrilldownQuery
}
