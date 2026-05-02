package model

type DashboardAnalyticsFilter struct {
	BudgetYear       *int32
	Quarter          *string
	LenderIDs        []string
	LenderTypes      []string
	InstitutionIDs   []string
	PipelineStatuses []string
	ProjectStatuses  []string
	RegionIDs        []string
	ProgramTitleIDs  []string
	ForeignLoanMin   *float64
	ForeignLoanMax   *float64
	IncludeHistory   bool
}

type DashboardDrilldownQuery struct {
	Target string              `json:"target"`
	Query  map[string][]string `json:"query"`
}

type DashboardAnalyticsPortfolioSummary struct {
	ProjectCount         int     `json:"project_count"`
	PipelineProjectCount int     `json:"pipeline_project_count"`
	OngoingProjectCount  int     `json:"ongoing_project_count"`
	TotalForeignLoanUSD  float64 `json:"total_foreign_loan_usd"`
	TotalGrantUSD        float64 `json:"total_grant_usd"`
	TotalCounterpartUSD  float64 `json:"total_counterpart_usd"`
}

type DashboardAnalyticsMonitoringSummary struct {
	LoanAgreementCount int     `json:"loan_agreement_count"`
	MonitoringCount    int     `json:"monitoring_count"`
	PlannedUSD         float64 `json:"planned_usd"`
	RealizedUSD        float64 `json:"realized_usd"`
	AgreementAmountUSD float64 `json:"agreement_amount_usd"`
	AbsorptionPct      float64 `json:"absorption_pct"`
}

type DashboardAnalyticsOverviewResponse struct {
	Portfolio  DashboardAnalyticsPortfolioSummary  `json:"portfolio"`
	Monitoring DashboardAnalyticsMonitoringSummary `json:"monitoring"`
	Drilldown  DashboardDrilldownQuery             `json:"drilldown"`
}

type DashboardAnalyticsSectionSummary struct {
	ProjectCount      int     `json:"project_count"`
	MonitoringCount   int     `json:"monitoring_count"`
	PlannedUSD        float64 `json:"planned_usd"`
	RealizedUSD       float64 `json:"realized_usd"`
	AbsorptionPct     float64 `json:"absorption_pct"`
	ForeignLoanUSD    float64 `json:"foreign_loan_usd"`
	AgreementValueUSD float64 `json:"agreement_value_usd"`
}

type DashboardAnalyticsInstitutionItem struct {
	InstitutionID    string                              `json:"institution_id"`
	InstitutionName  string                              `json:"institution_name"`
	InstitutionLevel string                              `json:"institution_level"`
	ProjectCount     int                                 `json:"project_count"`
	AssignmentCount  int                                 `json:"assignment_count"`
	Monitoring       DashboardAnalyticsMonitoringSummary `json:"monitoring"`
	Drilldown        DashboardDrilldownQuery             `json:"drilldown"`
}

type DashboardAnalyticsInstitutionsResponse struct {
	Summary   DashboardAnalyticsSectionSummary    `json:"summary"`
	Items     []DashboardAnalyticsInstitutionItem `json:"items"`
	Drilldown DashboardDrilldownQuery             `json:"drilldown"`
}

type DashboardAnalyticsLenderStageItem struct {
	LenderID   string                           `json:"lender_id"`
	LenderName string                           `json:"lender_name"`
	LenderType string                           `json:"lender_type"`
	Stage      string                           `json:"stage"`
	Summary    DashboardAnalyticsSectionSummary `json:"summary"`
	Drilldown  DashboardDrilldownQuery          `json:"drilldown"`
}

type DashboardAnalyticsLendersResponse struct {
	Summary   DashboardAnalyticsSectionSummary    `json:"summary"`
	Items     []DashboardAnalyticsLenderStageItem `json:"items"`
	Drilldown DashboardDrilldownQuery             `json:"drilldown"`
}

type DashboardAnalyticsAbsorptionItem struct {
	ID            string                  `json:"id"`
	Label         string                  `json:"label"`
	Dimension     string                  `json:"dimension"`
	PlannedUSD    float64                 `json:"planned_usd"`
	RealizedUSD   float64                 `json:"realized_usd"`
	AbsorptionPct float64                 `json:"absorption_pct"`
	Drilldown     DashboardDrilldownQuery `json:"drilldown"`
}

type DashboardAnalyticsAbsorptionResponse struct {
	Summary       DashboardAnalyticsMonitoringSummary `json:"summary"`
	ByInstitution []DashboardAnalyticsAbsorptionItem  `json:"by_institution"`
	ByProject     []DashboardAnalyticsAbsorptionItem  `json:"by_project"`
	ByLender      []DashboardAnalyticsAbsorptionItem  `json:"by_lender"`
	Drilldown     DashboardDrilldownQuery             `json:"drilldown"`
}

type DashboardAnalyticsYearlyItem struct {
	BudgetYear    int32                   `json:"budget_year"`
	Quarter       string                  `json:"quarter,omitempty"`
	PlannedUSD    float64                 `json:"planned_usd"`
	RealizedUSD   float64                 `json:"realized_usd"`
	AbsorptionPct float64                 `json:"absorption_pct"`
	Drilldown     DashboardDrilldownQuery `json:"drilldown"`
}

type DashboardAnalyticsYearlyResponse struct {
	Summary   DashboardAnalyticsMonitoringSummary `json:"summary"`
	Items     []DashboardAnalyticsYearlyItem      `json:"items"`
	Drilldown DashboardDrilldownQuery             `json:"drilldown"`
}

type DashboardAnalyticsLenderProportionItem struct {
	LenderType string                  `json:"lender_type"`
	Stage      string                  `json:"stage"`
	Count      int                     `json:"count"`
	AmountUSD  float64                 `json:"amount_usd"`
	SharePct   float64                 `json:"share_pct"`
	Drilldown  DashboardDrilldownQuery `json:"drilldown"`
}

type DashboardAnalyticsLenderProportionResponse struct {
	Items     []DashboardAnalyticsLenderProportionItem `json:"items"`
	Drilldown DashboardDrilldownQuery                  `json:"drilldown"`
}

type DashboardAnalyticsRiskItem struct {
	ID        string                  `json:"id"`
	Label     string                  `json:"label"`
	Severity  string                  `json:"severity"`
	Count     int                     `json:"count"`
	Drilldown DashboardDrilldownQuery `json:"drilldown"`
}

type DashboardAnalyticsDataQualityItem struct {
	Key       string                  `json:"key"`
	Label     string                  `json:"label"`
	Count     int                     `json:"count"`
	Drilldown DashboardDrilldownQuery `json:"drilldown"`
}

type DashboardAnalyticsRisksResponse struct {
	Watchlists  []DashboardAnalyticsRiskItem        `json:"watchlists"`
	DataQuality []DashboardAnalyticsDataQualityItem `json:"data_quality"`
	Drilldown   DashboardDrilldownQuery             `json:"drilldown"`
}
