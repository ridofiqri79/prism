export type DashboardStageKey = 'BB' | 'GB' | 'DK' | 'LA'

export type DashboardMetricTone = 'neutral' | 'good' | 'warning' | 'danger'

export type DashboardInsightTargetName = 'project-master' | 'spatial-distribution'

export type DashboardInsightQueryValue =
  | string
  | string[]
  | number
  | number[]
  | boolean
  | boolean[]
  | null
  | undefined

export interface DashboardInsightTarget {
  name: DashboardInsightTargetName
  query: Record<string, DashboardInsightQueryValue>
  label?: string
  exact: true
}

export type DashboardPanelKind =
  | 'bars'
  | 'card'
  | 'donut'
  | 'donutbar'
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


export interface DashboardStageCounter {
  tab: string
  label: string
  value: string
  meta: string
  tone?: DashboardMetricTone
  target?: DashboardInsightTarget
}

export interface DashboardDatum {
  label: string
  value: number
  valueLabel?: string
  amountLabel?: string
  description?: string
  color?: string
  tone?: DashboardMetricTone
  target?: DashboardInsightTarget
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
  description?: string
  data?: DashboardDatum[]
  rows?: DashboardTableRow[]
  pairs?: DashboardFlowPair[]
  target?: DashboardInsightTarget
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
  progressLabel?: string
  blockedLabel?: string
  target?: DashboardInsightTarget
  details: DashboardStageDetail
}
