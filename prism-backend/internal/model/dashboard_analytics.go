package model

// DashboardAnalyticsFilter holds parsed query params common to all analytics endpoints.
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

// DashboardDrilldownQuery enables frontend navigation from analytics to
// the relevant workspace (Project Master, Monitoring, LA, etc.).
type DashboardDrilldownQuery struct {
	Target string              `json:"target"`
	Query  map[string][]string `json:"query"`
}

// ------ Overview ------

type DashboardAnalyticsOverview struct {
	TotalProjects        int64                  `json:"total_projects"`
	TotalLoanAgreements  int64                  `json:"total_loan_agreements"`
	AgreementAmountUSD   float64                `json:"agreement_amount_usd"`
	TotalPlannedUSD      float64                `json:"total_planned_usd"`
	TotalRealizedUSD     float64                `json:"total_realized_usd"`
	OverallAbsorptionPct float64                `json:"overall_absorption_pct"`
	ActiveMonitoring     int64                  `json:"active_monitoring"`
	PipelineFunnel       []PipelineStageSummary `json:"pipeline_funnel"`
	TopInstitutions      []InstitutionSummary   `json:"top_institutions"`
	TopLenders           []TopLenderSummary     `json:"top_lenders"`
}

type PipelineStageSummary struct {
	Stage        string  `json:"stage"`
	ProjectCount int64   `json:"project_count"`
	TotalLoanUSD float64 `json:"total_loan_usd"`
}

// ------ Institutions ------

type DashboardAnalyticsInstitutionsResponse struct {
	Summary InstitutionsSummary `json:"summary"`
	Items   []InstitutionItem   `json:"items"`
}

type InstitutionsSummary struct {
	InstitutionCount    int64   `json:"institution_count"`
	ProjectCount        int64   `json:"project_count"`
	AssignmentCount     int64   `json:"assignment_count"`
	AgreementAmountUSD  float64 `json:"agreement_amount_usd"`
	PlannedUSD          float64 `json:"planned_usd"`
	RealizedUSD         float64 `json:"realized_usd"`
	AbsorptionPct       float64 `json:"absorption_pct"`
}

type InstitutionItem struct {
	Institution         InstitutionRef              `json:"institution"`
	ProjectCount        int64                       `json:"project_count"`
	AssignmentCount     int64                       `json:"assignment_count"`
	LoanAgreementCount  int64                       `json:"loan_agreement_count"`
	MonitoringCount     int64                       `json:"monitoring_count"`
	AgreementAmountUSD  float64                     `json:"agreement_amount_usd"`
	PlannedUSD          float64                      `json:"planned_usd"`
	RealizedUSD         float64                      `json:"realized_usd"`
	AbsorptionPct       float64                      `json:"absorption_pct"`
	LoanTypes           []string                    `json:"loan_types"`
	Drilldown           *DashboardDrilldownQuery    `json:"drilldown"`
}

type InstitutionRef struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name,omitempty"`
	Level     string `json:"level"`
}

type InstitutionSummary struct {
	Institution        InstitutionRef `json:"institution"`
	ProjectCount       int64          `json:"project_count"`
	LoanAgreementCount int64          `json:"loan_agreement_count"`
	MonitoringCount    int64          `json:"monitoring_count"`
	AgreementAmountUSD float64        `json:"agreement_amount_usd"`
	PlannedUSD         float64        `json:"planned_usd"`
	RealizedUSD        float64        `json:"realized_usd"`
	AbsorptionPct      float64        `json:"absorption_pct"`
}

// ------ Lenders ------

type DashboardAnalyticsLendersResponse struct {
	Summary                  LendersSummary            `json:"summary"`
	Items                    []LenderItem              `json:"items"`
	LenderInstitutionMatrix  []LenderInstitutionMatrix `json:"lender_institution_matrix"`
}

type LendersSummary struct {
	LenderCount           int64   `json:"lender_count"`
	LoanAgreementCount    int64   `json:"loan_agreement_count"`
	AgreementAmountUSD    float64 `json:"agreement_amount_usd"`
	PlannedUSD            float64 `json:"planned_usd"`
	RealizedUSD           float64 `json:"realized_usd"`
	AbsorptionPct         float64 `json:"absorption_pct"`
}

type LenderItem struct {
	Lender               LenderRef                 `json:"lender"`
	LoanAgreementCount   int64                     `json:"loan_agreement_count"`
	ProjectCount         int64                     `json:"project_count"`
	InstitutionCount     int64                     `json:"institution_count"`
	AgreementAmountUSD   float64                   `json:"agreement_amount_usd"`
	PlannedUSD           float64                   `json:"planned_usd"`
	RealizedUSD          float64                   `json:"realized_usd"`
	AbsorptionPct        float64                   `json:"absorption_pct"`
	Drilldown            *DashboardDrilldownQuery  `json:"drilldown"`
}

type LenderRef struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name,omitempty"`
	Type      string `json:"type"`
}

type LenderInstitutionMatrix struct {
	Institution        InstitutionRef `json:"institution"`
	Lender             LenderRef      `json:"lender"`
	ProjectCount       int64          `json:"project_count"`
	LoanAgreementCount int64          `json:"loan_agreement_count"`
	AgreementAmountUSD float64        `json:"agreement_amount_usd"`
	PlannedUSD         float64        `json:"planned_usd"`
	RealizedUSD        float64        `json:"realized_usd"`
	AbsorptionPct      float64        `json:"absorption_pct"`
}

type TopLenderSummary struct {
	Lender             LenderRef `json:"lender"`
	LoanAgreementCount int64     `json:"loan_agreement_count"`
	ProjectCount       int64     `json:"project_count"`
	InstitutionCount   int64     `json:"institution_count"`
	AgreementAmountUSD float64   `json:"agreement_amount_usd"`
	PlannedUSD         float64   `json:"planned_usd"`
	RealizedUSD        float64   `json:"realized_usd"`
	AbsorptionPct      float64   `json:"absorption_pct"`
}

// ------ Absorption ------

type DashboardAnalyticsAbsorptionResponse struct {
	Summary       AbsorptionSummary     `json:"summary"`
	ByInstitution []AbsorptionRankItem  `json:"by_institution"`
	ByProject     []AbsorptionRankItem  `json:"by_project"`
	ByLender      []AbsorptionRankItem  `json:"by_lender"`
}

type AbsorptionSummary struct {
	PlannedUSD    float64 `json:"planned_usd"`
	RealizedUSD   float64 `json:"realized_usd"`
	AbsorptionPct float64 `json:"absorption_pct"`
}

type AbsorptionRankItem struct {
	Rank           int64                     `json:"rank"`
	ID             string                    `json:"id"`
	Name           string                    `json:"name"`
	ShortName      string                    `json:"short_name,omitempty"`
	Type           string                    `json:"type,omitempty"`
	PlannedUSD     float64                   `json:"planned_usd"`
	RealizedUSD    float64                   `json:"realized_usd"`
	AbsorptionPct  float64                   `json:"absorption_pct"`
	VarianceUSD    float64                   `json:"variance_usd"`
	Status         string                    `json:"status"`
	Drilldown      *DashboardDrilldownQuery  `json:"drilldown,omitempty"`
}

// ------ Yearly ------

type DashboardAnalyticsYearlyResponse struct {
	Items []YearlyItem `json:"items"`
}

type YearlyItem struct {
	BudgetYear          int32   `json:"budget_year"`
	Quarter             string  `json:"quarter"`
	PlannedUSD          float64 `json:"planned_usd"`
	RealizedUSD         float64 `json:"realized_usd"`
	AbsorptionPct       float64 `json:"absorption_pct"`
	LoanAgreementCount  int64   `json:"loan_agreement_count"`
	ProjectCount        int64   `json:"project_count"`
}

// ------ Lender Proportion ------

type DashboardAnalyticsLenderProportionResponse struct {
	ByStage []LenderProportionStage `json:"by_stage"`
}

type LenderProportionStage struct {
	Stage string                    `json:"stage"`
	Items []LenderProportionItem    `json:"items"`
}

type LenderProportionItem struct {
	Type         string  `json:"type"`
	ProjectCount int64   `json:"project_count"`
	LenderCount  int64   `json:"lender_count"`
	AmountUSD    float64 `json:"amount_usd"`
	SharePct     float64 `json:"share_pct"`
	PlannedUSD   float64 `json:"planned_usd,omitempty"`
	RealizedUSD  float64 `json:"realized_usd,omitempty"`
}

// ------ Risk ------

type DashboardRisksResponse struct {
	ClosingRisk                RiskSection       `json:"closing_risk"`
	EffectiveWithoutMonitoring RiskSection       `json:"effective_without_monitoring"`
	ExtendedLoans              RiskSection       `json:"extended_loans"`
	DataQuality                DataQualitySection `json:"data_quality"`
}

type RiskSection struct {
	Count              int64       `json:"count"`
	AgreementAmountUSD float64     `json:"agreement_amount_usd,omitempty"`
	Items              []RiskItem  `json:"items,omitempty"`
}

type RiskItem struct {
	LoanAgreement     map[string]interface{}    `json:"loan_agreement"`
	ClosingDate       string                    `json:"closing_date,omitempty"`
	DaysToClosing     *int32                    `json:"days_to_closing,omitempty"`
	AbsorptionPct     float64                   `json:"absorption_pct,omitempty"`
	EffectiveDate     string                    `json:"effective_date,omitempty"`
	AmountUSD         float64                   `json:"amount_usd,omitempty"`
	ExtensionDays     *int32                    `json:"extension_days,omitempty"`
	Drilldown         *DashboardDrilldownQuery  `json:"drilldown,omitempty"`
}

type DataQualitySection struct {
	MissingExecutingAgencyCount  int64 `json:"missing_executing_agency_count"`
	MissingLenderIndicationCount int64 `json:"missing_lender_indication_count"`
	ProjectWithoutGBCount        int64 `json:"project_without_gb_count"`
}
