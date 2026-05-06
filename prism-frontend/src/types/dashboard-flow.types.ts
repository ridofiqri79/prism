export type DashboardStageKey = 'BB' | 'GB' | 'DK' | 'LA'

export type DashboardMetricTone = 'neutral' | 'good' | 'warning' | 'danger'

export type DashboardPanelKind =
  | 'bars'
  | 'donut'
  | 'regions'
  | 'stack'
  | 'table'
  | 'flow'
  | 'status'

export type DashboardPanelSpan = 'small' | 'medium' | 'large' | 'wide'

export interface DashboardKpi {
  label: string
  value: string
  unit?: string
  delta: string
  tone?: DashboardMetricTone
}

export type DashboardHeaderActionSeverity =
  | 'secondary'
  | 'success'
  | 'info'
  | 'warn'
  | 'danger'
  | 'help'
  | 'contrast'

export interface DashboardHeaderAction {
  key: string
  label: string
  icon: string
  severity?: DashboardHeaderActionSeverity
  outlined?: boolean
}

export interface DashboardStageCounter {
  tab: string
  label: string
  value: string
  meta: string
  tone?: DashboardMetricTone
}

export interface DashboardDatum {
  label: string
  value: number
  valueLabel?: string
  amountLabel?: string
  color?: string
  tone?: DashboardMetricTone
}

export interface DashboardTableRow {
  name: string
  status: string
  value: string
  tone?: DashboardMetricTone
}

export interface DashboardFlowPair {
  source: string
  target: string
  value: string
}

export interface DashboardStagePanel {
  tab: string
  title: string
  hint: string
  kind: DashboardPanelKind
  span?: DashboardPanelSpan
  data?: DashboardDatum[]
  rows?: DashboardTableRow[]
  pairs?: DashboardFlowPair[]
}

export interface DashboardStageTab {
  label: string
}

export interface DashboardStageDetail {
  tabs: DashboardStageTab[]
  counters: DashboardStageCounter[]
  panels: DashboardStagePanel[]
}

export interface DashboardStage {
  key: DashboardStageKey
  stepLabel: string
  title: string
  subtitle: string
  count: number
  amountLabel: string
  pipelineShare: number
  color: string
  colorSoft: string
  nextLabel?: string
  finalLabel?: string
  conversionLabel?: string
  blockedLabel?: string
  details: DashboardStageDetail
}
