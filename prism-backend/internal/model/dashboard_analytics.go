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

	LowAbsorptionThreshold  *float64
	ClosingMonthsThreshold  *int32
	StaleMonitoringQuarters *int32
}

type DashboardDrilldownQuery struct {
	Target string              `json:"target"`
	Query  map[string][]string `json:"query"`
}

type DashboardAnalyticsEntityRef struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	ShortName *string `json:"short_name,omitempty"`
	Level     string  `json:"level,omitempty"`
	Type      string  `json:"type,omitempty"`
}

type DashboardAnalyticsPortfolioOverview struct {
	ProjectCount            int     `json:"project_count"`
	AssignmentCount         int     `json:"assignment_count"`
	TotalPipelineLoanUSD    float64 `json:"total_pipeline_loan_usd"`
	TotalAgreementAmountUSD float64 `json:"total_agreement_amount_usd"`
	TotalPlannedUSD         float64 `json:"total_planned_usd"`
	TotalRealizedUSD        float64 `json:"total_realized_usd"`
	AbsorptionPct           float64 `json:"absorption_pct"`
}

type DashboardAnalyticsPipelineFunnelItem struct {
	Stage        string                  `json:"stage"`
	ProjectCount int                     `json:"project_count"`
	TotalLoanUSD float64                 `json:"total_loan_usd"`
	Drilldown    DashboardDrilldownQuery `json:"drilldown"`
}

type DashboardAnalyticsInsight struct {
	Key       string                  `json:"key"`
	Label     string                  `json:"label"`
	Value     float64                 `json:"value"`
	Severity  string                  `json:"severity,omitempty"`
	Drilldown DashboardDrilldownQuery `json:"drilldown"`
}

type DashboardAnalyticsOverviewResponse struct {
	Portfolio      DashboardAnalyticsPortfolioOverview    `json:"portfolio"`
	PipelineFunnel []DashboardAnalyticsPipelineFunnelItem `json:"pipeline_funnel"`
	TopInsights    []DashboardAnalyticsInsight            `json:"top_insights"`
	Drilldown      DashboardDrilldownQuery                `json:"drilldown"`
}

type DashboardAnalyticsPipelineBreakdown struct {
	BB         int `json:"BB"`
	GB         int `json:"GB"`
	DK         int `json:"DK"`
	LA         int `json:"LA"`
	Monitoring int `json:"Monitoring"`
}

type DashboardAnalyticsInstitutionSummary struct {
	InstitutionCount        int     `json:"institution_count"`
	ProjectCount            int     `json:"project_count"`
	AssignmentCount         int     `json:"assignment_count"`
	TotalAgreementAmountUSD float64 `json:"total_agreement_amount_usd"`
	TotalPlannedUSD         float64 `json:"total_planned_usd"`
	TotalRealizedUSD        float64 `json:"total_realized_usd"`
	AbsorptionPct           float64 `json:"absorption_pct"`
}

type DashboardAnalyticsInstitutionItem struct {
	Institution        DashboardAnalyticsEntityRef         `json:"institution"`
	ProjectCount       int                                 `json:"project_count"`
	AssignmentCount    int                                 `json:"assignment_count"`
	LoanAgreementCount int                                 `json:"loan_agreement_count"`
	MonitoringCount    int                                 `json:"monitoring_count"`
	AgreementAmountUSD float64                             `json:"agreement_amount_usd"`
	PlannedUSD         float64                             `json:"planned_usd"`
	RealizedUSD        float64                             `json:"realized_usd"`
	AbsorptionPct      float64                             `json:"absorption_pct"`
	PipelineBreakdown  DashboardAnalyticsPipelineBreakdown `json:"pipeline_breakdown"`
	Drilldown          DashboardDrilldownQuery             `json:"drilldown"`
}

type DashboardAnalyticsInstitutionsResponse struct {
	Summary   DashboardAnalyticsInstitutionSummary `json:"summary"`
	Items     []DashboardAnalyticsInstitutionItem  `json:"items"`
	Drilldown DashboardDrilldownQuery              `json:"drilldown"`
}

type DashboardAnalyticsLenderSummary struct {
	LenderCount             int     `json:"lender_count"`
	LoanAgreementCount      int     `json:"loan_agreement_count"`
	TotalAgreementAmountUSD float64 `json:"total_agreement_amount_usd"`
	TotalPlannedUSD         float64 `json:"total_planned_usd"`
	TotalRealizedUSD        float64 `json:"total_realized_usd"`
	AbsorptionPct           float64 `json:"absorption_pct"`
}

type DashboardAnalyticsLenderItem struct {
	Lender             DashboardAnalyticsEntityRef `json:"lender"`
	LoanAgreementCount int                         `json:"loan_agreement_count"`
	ProjectCount       int                         `json:"project_count"`
	InstitutionCount   int                         `json:"institution_count"`
	MonitoringCount    int                         `json:"monitoring_count"`
	AgreementAmountUSD float64                     `json:"agreement_amount_usd"`
	PlannedUSD         float64                     `json:"planned_usd"`
	RealizedUSD        float64                     `json:"realized_usd"`
	AbsorptionPct      float64                     `json:"absorption_pct"`
	Drilldown          DashboardDrilldownQuery     `json:"drilldown"`
}

type DashboardAnalyticsLenderInstitutionMatrixItem struct {
	Institution        DashboardAnalyticsEntityRef `json:"institution"`
	Lender             DashboardAnalyticsEntityRef `json:"lender"`
	ProjectCount       int                         `json:"project_count"`
	LoanAgreementCount int                         `json:"loan_agreement_count"`
	MonitoringCount    int                         `json:"monitoring_count"`
	AgreementAmountUSD float64                     `json:"agreement_amount_usd"`
	PlannedUSD         float64                     `json:"planned_usd"`
	RealizedUSD        float64                     `json:"realized_usd"`
	AbsorptionPct      float64                     `json:"absorption_pct"`
	Drilldown          DashboardDrilldownQuery     `json:"drilldown"`
}

type DashboardAnalyticsLendersResponse struct {
	Summary                 DashboardAnalyticsLenderSummary                 `json:"summary"`
	Items                   []DashboardAnalyticsLenderItem                  `json:"items"`
	LenderInstitutionMatrix []DashboardAnalyticsLenderInstitutionMatrixItem `json:"lender_institution_matrix"`
	Drilldown               DashboardDrilldownQuery                         `json:"drilldown"`
}

type DashboardAnalyticsAbsorptionSummary struct {
	PlannedUSD    float64 `json:"planned_usd"`
	RealizedUSD   float64 `json:"realized_usd"`
	AbsorptionPct float64 `json:"absorption_pct"`
}

type DashboardAnalyticsAbsorptionItem struct {
	Rank          int                     `json:"rank"`
	ID            string                  `json:"id"`
	Name          string                  `json:"name"`
	Dimension     string                  `json:"dimension"`
	PlannedUSD    float64                 `json:"planned_usd"`
	RealizedUSD   float64                 `json:"realized_usd"`
	AbsorptionPct float64                 `json:"absorption_pct"`
	VarianceUSD   float64                 `json:"variance_usd"`
	Status        string                  `json:"status"`
	Drilldown     DashboardDrilldownQuery `json:"drilldown"`
}

type DashboardAnalyticsAbsorptionResponse struct {
	Summary       DashboardAnalyticsAbsorptionSummary `json:"summary"`
	ByInstitution []DashboardAnalyticsAbsorptionItem  `json:"by_institution"`
	ByProject     []DashboardAnalyticsAbsorptionItem  `json:"by_project"`
	ByLender      []DashboardAnalyticsAbsorptionItem  `json:"by_lender"`
	Drilldown     DashboardDrilldownQuery             `json:"drilldown"`
}

type DashboardAnalyticsYearlySummary struct {
	PlannedUSD         float64 `json:"planned_usd"`
	RealizedUSD        float64 `json:"realized_usd"`
	AbsorptionPct      float64 `json:"absorption_pct"`
	LoanAgreementCount int     `json:"loan_agreement_count"`
	ProjectCount       int     `json:"project_count"`
}

type DashboardAnalyticsYearlyItem struct {
	BudgetYear         int32                   `json:"budget_year"`
	Quarter            string                  `json:"quarter"`
	PlannedUSD         float64                 `json:"planned_usd"`
	RealizedUSD        float64                 `json:"realized_usd"`
	AbsorptionPct      float64                 `json:"absorption_pct"`
	LoanAgreementCount int                     `json:"loan_agreement_count"`
	ProjectCount       int                     `json:"project_count"`
	Drilldown          DashboardDrilldownQuery `json:"drilldown"`
}

type DashboardAnalyticsYearlyResponse struct {
	Summary   DashboardAnalyticsYearlySummary `json:"summary"`
	Items     []DashboardAnalyticsYearlyItem  `json:"items"`
	Drilldown DashboardDrilldownQuery         `json:"drilldown"`
}

type DashboardAnalyticsLenderProportionItem struct {
	Type         string                  `json:"type"`
	ProjectCount int                     `json:"project_count"`
	LenderCount  int                     `json:"lender_count"`
	AmountUSD    float64                 `json:"amount_usd"`
	SharePct     float64                 `json:"share_pct"`
	Drilldown    DashboardDrilldownQuery `json:"drilldown"`
}

type DashboardAnalyticsLenderProportionStage struct {
	Stage string                                   `json:"stage"`
	Items []DashboardAnalyticsLenderProportionItem `json:"items"`
}

type DashboardAnalyticsLenderProportionResponse struct {
	ByStage   []DashboardAnalyticsLenderProportionStage `json:"by_stage"`
	Drilldown DashboardDrilldownQuery                   `json:"drilldown"`
}

type DashboardAnalyticsRiskItem struct {
	ID        string                  `json:"id"`
	Label     string                  `json:"label"`
	Severity  string                  `json:"severity"`
	Count     int                     `json:"count"`
	Drilldown DashboardDrilldownQuery `json:"drilldown"`
}

type DashboardAnalyticsDataQualityItem struct {
	Code      string                  `json:"code"`
	Label     string                  `json:"label"`
	Stage     string                  `json:"stage"`
	Severity  string                  `json:"severity"`
	Count     int                     `json:"count"`
	Drilldown DashboardDrilldownQuery `json:"drilldown"`
}

type DashboardAnalyticsRiskThresholds struct {
	LowAbsorptionThreshold     float64 `json:"low_absorption_threshold"`
	ClosingMonthsThreshold     int32   `json:"closing_months_threshold"`
	ClosingAbsorptionThreshold float64 `json:"closing_absorption_threshold"`
	StaleMonitoringQuarters    int32   `json:"stale_monitoring_quarters"`
}

type DashboardAnalyticsRiskSummary struct {
	LowAbsorptionCount              int `json:"low_absorption_count"`
	EffectiveWithoutMonitoringCount int `json:"effective_without_monitoring_count"`
	ClosingRiskCount                int `json:"closing_risk_count"`
	ExtendedLoanCount               int `json:"extended_loan_count"`
	DataQualityIssueCount           int `json:"data_quality_issue_count"`
	BottleneckProjectCount          int `json:"bottleneck_project_count"`
}

type DashboardAnalyticsRiskCard struct {
	Code      string                  `json:"code"`
	Label     string                  `json:"label"`
	Count     int                     `json:"count"`
	Severity  string                  `json:"severity"`
	AmountUSD float64                 `json:"amount_usd,omitempty"`
	Drilldown DashboardDrilldownQuery `json:"drilldown"`
}

type DashboardAnalyticsLoanAgreementRiskItem struct {
	RiskCode            string                       `json:"risk_code"`
	RiskLabel           string                       `json:"risk_label"`
	Severity            string                       `json:"severity"`
	ProjectID           string                       `json:"project_id"`
	ProjectName         string                       `json:"project_name"`
	LoanAgreementID     string                       `json:"loan_agreement_id"`
	LoanCode            string                       `json:"loan_code"`
	Lender              DashboardAnalyticsEntityRef  `json:"lender"`
	Institution         *DashboardAnalyticsEntityRef `json:"institution,omitempty"`
	EffectiveDate       string                       `json:"effective_date"`
	OriginalClosingDate string                       `json:"original_closing_date"`
	ClosingDate         string                       `json:"closing_date"`
	BudgetYear          *int32                       `json:"budget_year,omitempty"`
	Quarter             *string                      `json:"quarter,omitempty"`
	PlannedUSD          float64                      `json:"planned_usd"`
	RealizedUSD         float64                      `json:"realized_usd"`
	AbsorptionPct       float64                      `json:"absorption_pct"`
	AgreementAmountUSD  float64                      `json:"agreement_amount_usd"`
	DaysSinceEffective  int                          `json:"days_since_effective"`
	DaysToClosing       int                          `json:"days_to_closing"`
	MonthsToClosing     int                          `json:"months_to_closing"`
	ExtensionDays       int                          `json:"extension_days"`
	StaleQuarters       int                          `json:"stale_quarters"`
	MonitoringStatus    string                       `json:"monitoring_status"`
	Drilldown           DashboardDrilldownQuery      `json:"drilldown"`
}

type DashboardAnalyticsPipelineBottleneckItem struct {
	Stage        string                  `json:"stage"`
	Label        string                  `json:"label"`
	ProjectCount int                     `json:"project_count"`
	TotalLoanUSD float64                 `json:"total_loan_usd"`
	OldestDate   string                  `json:"oldest_date,omitempty"`
	Severity     string                  `json:"severity"`
	Drilldown    DashboardDrilldownQuery `json:"drilldown"`
}

type DashboardAnalyticsExtendedLoanBreakdown struct {
	Dimension            string                      `json:"dimension"`
	Entity               DashboardAnalyticsEntityRef `json:"entity"`
	LoanAgreementCount   int                         `json:"loan_agreement_count"`
	AmountUSD            float64                     `json:"amount_usd"`
	AverageExtensionDays float64                     `json:"average_extension_days"`
	Drilldown            DashboardDrilldownQuery     `json:"drilldown"`
}

type DashboardAnalyticsExtendedLoanInsight struct {
	Count                int                                       `json:"count"`
	AmountUSD            float64                                   `json:"amount_usd"`
	AverageExtensionDays float64                                   `json:"average_extension_days"`
	ByLender             []DashboardAnalyticsExtendedLoanBreakdown `json:"by_lender"`
	ByInstitution        []DashboardAnalyticsExtendedLoanBreakdown `json:"by_institution"`
}

type DashboardAnalyticsRiskWatchlists struct {
	LowAbsorptionProjects      []DashboardAnalyticsLoanAgreementRiskItem  `json:"low_absorption_projects"`
	EffectiveWithoutMonitoring []DashboardAnalyticsLoanAgreementRiskItem  `json:"effective_without_monitoring"`
	ClosingRisks               []DashboardAnalyticsLoanAgreementRiskItem  `json:"closing_risks"`
	ExtendedLoans              []DashboardAnalyticsLoanAgreementRiskItem  `json:"extended_loans"`
	PipelineBottlenecks        []DashboardAnalyticsPipelineBottleneckItem `json:"pipeline_bottlenecks"`
}

type DashboardAnalyticsRisksResponse struct {
	Summary             DashboardAnalyticsRiskSummary         `json:"summary"`
	Thresholds          DashboardAnalyticsRiskThresholds      `json:"thresholds"`
	RiskCards           []DashboardAnalyticsRiskCard          `json:"risk_cards"`
	Watchlists          DashboardAnalyticsRiskWatchlists      `json:"watchlists"`
	ExtendedLoanInsight DashboardAnalyticsExtendedLoanInsight `json:"extended_loan_insight"`
	DataQuality         []DashboardAnalyticsDataQualityItem   `json:"data_quality"`
	Drilldown           DashboardDrilldownQuery               `json:"drilldown"`
}
