export type DashboardStageKey = 'BB' | 'GB' | 'DK' | 'LA'

export type DashboardStageTone = 'tealDark' | 'teal' | 'green' | 'gold'

export type DashboardMetricTone = 'neutral' | 'good' | 'warning' | 'danger'

export interface DashboardStageBreakdown {
  label: string
  count: number
  amountDisplay?: string
  percent: number
}

export interface DashboardDelayedProject {
  name: string
  meta: string
  amountDisplay: string
}

export interface DashboardStageMetric {
  label: string
  value: string
  tone?: DashboardMetricTone
}

export interface DashboardStageDetail {
  breakdownTitle: string
  topProjectsTitle: string
  metricsTitle: string
  breakdownItems: DashboardStageBreakdown[]
  topProjects: DashboardDelayedProject[]
  metrics: DashboardStageMetric[]
  notes?: string[]
}

export interface DashboardFlowStage {
  key: DashboardStageKey
  stepLabel: string
  title: string
  subtitle: string
  amountTrillion: number
  count: number
  tone: DashboardStageTone
  nextLabel?: string
  targetLabel?: string
  detail: DashboardStageDetail
}

export interface DashboardPrototypeFlowProps {
  title: string
  lastSyncLabel: string
  stages: DashboardFlowStage[]
}
