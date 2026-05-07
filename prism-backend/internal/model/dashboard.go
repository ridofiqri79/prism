package model

type DashboardDistributionItem struct {
	ID             string  `json:"id,omitempty"`
	Label          string  `json:"label"`
	Level          string  `json:"level"`
	ProjectCount   int     `json:"project_count"`
	ForeignLoanUSD float64 `json:"foreign_loan_usd"`
}

type DashboardBlueBookDistributionResponse struct {
	AgencyGroups []DashboardDistributionItem `json:"agency_groups"`
	TopAgencies  []DashboardDistributionItem `json:"top_agencies"`
	Programs     []DashboardDistributionItem `json:"programs"`
}

type DashboardStageDistributionResponse struct {
	LenderTypes []DashboardDistributionItem `json:"lender_types"`
	TopLenders  []DashboardDistributionItem `json:"top_lenders"`
	TopAgencies []DashboardDistributionItem `json:"top_agencies"`
	Programs    []DashboardDistributionItem `json:"programs,omitempty"`
}

type DashboardGreenBookDistributionResponse = DashboardStageDistributionResponse
type DashboardDaftarKegiatanDistributionResponse = DashboardStageDistributionResponse
type DashboardLoanAgreementDistributionResponse = DashboardStageDistributionResponse
