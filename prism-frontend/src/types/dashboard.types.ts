export interface DashboardDistributionItem {
  id?: string
  label: string
  level: string
  project_count: number
  foreign_loan_usd: number
}

export interface DashboardStageOverviewItem {
  stage: 'BB' | 'GB' | 'DK' | 'LA'
  project_count: number
  total_loan_usd: number
  regions: DashboardDistributionItem[]
}

export interface DashboardStageOverview {
  stages: DashboardStageOverviewItem[]
}

export interface DashboardBlueBookDistribution {
  agency_groups: DashboardDistributionItem[]
  top_agencies: DashboardDistributionItem[]
  programs: DashboardDistributionItem[]
}

export interface DashboardGreenBookDistribution {
  lender_types: DashboardDistributionItem[]
  top_lenders: DashboardDistributionItem[]
  top_agencies: DashboardDistributionItem[]
  programs?: DashboardDistributionItem[]
}

export type DashboardDaftarKegiatanDistribution = DashboardGreenBookDistribution
export type DashboardLoanAgreementDistribution = DashboardGreenBookDistribution
