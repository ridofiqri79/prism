package model

type DashboardSummary struct {
	TotalBBProjects      int     `json:"total_bb_projects"`
	TotalGBProjects      int     `json:"total_gb_projects"`
	TotalLoanAgreements  int     `json:"total_loan_agreements"`
	TotalAmountUSD       float64 `json:"total_amount_usd"`
	TotalRealizedUSD     float64 `json:"total_realized_usd"`
	OverallAbsorptionPct float64 `json:"overall_absorption_pct"`
	ActiveMonitoring     int     `json:"active_monitoring"`
}

type MonitoringSummaryFilter struct {
	BudgetYear *int32
	Quarter    *string
	LenderID   *string
}

type MonitoringSummary struct {
	BudgetYear       *int32                      `json:"budget_year,omitempty"`
	Quarter          *string                     `json:"quarter,omitempty"`
	TotalPlannedUSD  float64                     `json:"total_planned_usd"`
	TotalRealizedUSD float64                     `json:"total_realized_usd"`
	AbsorptionPct    float64                     `json:"absorption_pct"`
	ByLender         []MonitoringSummaryByLender `json:"by_lender"`
}

type MonitoringSummaryByLender struct {
	Lender        LenderSummary `json:"lender"`
	PlannedUSD    float64       `json:"planned_usd"`
	RealizedUSD   float64       `json:"realized_usd"`
	AbsorptionPct float64       `json:"absorption_pct"`
}

type LenderSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
