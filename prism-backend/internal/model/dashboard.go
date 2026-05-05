package model

type DashboardFilterRequest struct {
	PeriodID       *string `json:"period_id,omitempty"`
	PublishYear    *int32  `json:"publish_year,omitempty"`
	LenderID       *string `json:"lender_id,omitempty"`
	InstitutionID  *string `json:"institution_id,omitempty"`
	IncludeHistory bool    `json:"include_history"`
}

type MetricCard struct {
	Key      string  `json:"key"`
	Label    string  `json:"label"`
	Value    float64 `json:"value"`
	Unit     string  `json:"unit,omitempty"`
	Category string  `json:"category,omitempty"`
}

type StageMetric struct {
	Stage        string  `json:"stage"`
	Label        string  `json:"label"`
	ProjectCount int     `json:"project_count"`
	AmountUSD    float64 `json:"amount_usd"`
}

type BreakdownItem struct {
	ID          *string  `json:"id,omitempty"`
	Key         string   `json:"key,omitempty"`
	Label       string   `json:"label"`
	ItemCount   int      `json:"item_count,omitempty"`
	AmountUSD   float64  `json:"amount_usd,omitempty"`
	Percentage  *float64 `json:"percentage,omitempty"`
}

type RiskItem struct {
	ID                 string  `json:"id,omitempty"`
	RiskType           string  `json:"risk_type,omitempty"`
	ReferenceID        *string `json:"reference_id,omitempty"`
	ReferenceType      string  `json:"reference_type,omitempty"`
	JourneyBBProjectID *string `json:"journey_bb_project_id,omitempty"`
	Code               string  `json:"code,omitempty"`
	Title              string  `json:"title"`
	Description        string  `json:"description,omitempty"`
	Severity           string  `json:"severity"`
	AmountUSD          float64 `json:"amount_usd,omitempty"`
	DaysUntilClosing   *int    `json:"days_until_closing,omitempty"`
	Score              float64 `json:"score,omitempty"`
}

type DashboardSummary struct {
	TotalBBProjects         int          `json:"total_bb_projects"`
	TotalGBProjects         int          `json:"total_gb_projects"`
	TotalLoanAgreements     int          `json:"total_loan_agreements"`
	BBPipelineUSD           float64      `json:"bb_pipeline_usd"`
	GBPipelineUSD           float64      `json:"gb_pipeline_usd"`
	GBLocalUSD              float64      `json:"gb_local_usd"`
	DKFinancingUSD          float64      `json:"dk_financing_usd"`
	DKCounterpartUSD        float64      `json:"dk_counterpart_usd"`
	LACommitmentUSD         float64      `json:"la_commitment_usd"`
	Metrics                 []MetricCard `json:"metrics"`
}

type DashboardFilterOptions map[string][]BreakdownItem

type ExecutivePortfolioDashboard struct {
	Cards           []MetricCard    `json:"cards"`
	Funnel          []StageMetric   `json:"funnel"`
	TopInstitutions []BreakdownItem `json:"top_institutions"`
	TopLenders      []BreakdownItem `json:"top_lenders"`
	RiskItems       []RiskItem      `json:"risk_items"`
	Insights        []string        `json:"insights"`
}

type PipelineBottleneckFilterRequest struct {
	Stage         *string `json:"stage,omitempty"`
	PeriodID      *string `json:"period_id,omitempty"`
	PublishYear   *int32  `json:"publish_year,omitempty"`
	InstitutionID *string `json:"institution_id,omitempty"`
	LenderID      *string `json:"lender_id,omitempty"`
	MinAgeDays    *int32  `json:"min_age_days,omitempty"`
}

type PipelineStageSummary struct {
	Stage        string  `json:"stage"`
	Label        string  `json:"label"`
	ProjectCount int     `json:"project_count"`
	AmountUSD    float64 `json:"amount_usd"`
	AvgAgeDays   float64 `json:"avg_age_days"`
}

type PipelineBottleneckItem struct {
	ProjectID          string   `json:"project_id"`
	ReferenceType      string   `json:"reference_type"`
	JourneyBBProjectID *string  `json:"journey_bb_project_id,omitempty"`
	Code               string   `json:"code,omitempty"`
	ProjectName        string   `json:"project_name"`
	CurrentStage       string   `json:"current_stage"`
	StageLabel         string   `json:"stage_label"`
	AgeDays            int      `json:"age_days"`
	AmountUSD          float64  `json:"amount_usd"`
	InstitutionName    string   `json:"institution_name,omitempty"`
	LenderNames        []string `json:"lender_names"`
	RecommendedAction  string   `json:"recommended_action"`
	RelevantAt         *string  `json:"relevant_at,omitempty"`
}

type PipelineBottleneckDashboard struct {
	StageSummary []PipelineStageSummary   `json:"stage_summary"`
	Items        []PipelineBottleneckItem `json:"items"`
}

type GreenBookReadinessFilterRequest struct {
	PublishYear     *int32  `json:"publish_year,omitempty"`
	GreenBookID     *string `json:"green_book_id,omitempty"`
	InstitutionID   *string `json:"institution_id,omitempty"`
	LenderID        *string `json:"lender_id,omitempty"`
	ReadinessStatus *string `json:"readiness_status,omitempty"`
}

type GreenBookReadinessSummary struct {
	TotalProjects           int     `json:"total_projects"`
	TotalLoanUSD            float64 `json:"total_loan_usd"`
	TotalGrantUSD           float64 `json:"total_grant_usd"`
	TotalLocalUSD           float64 `json:"total_local_usd"`
	ProjectsWithCofinancing int     `json:"projects_with_cofinancing"`
	ProjectsIncomplete      int     `json:"projects_incomplete"`
	ProjectsReady           int     `json:"projects_ready"`
	ProjectsPartial         int     `json:"projects_partial"`
}

type GreenBookDisbursementYear struct {
	Year      int32   `json:"year"`
	AmountUSD float64 `json:"amount_usd"`
}

type GreenBookFundingAllocation struct {
	Services      float64 `json:"services"`
	Constructions float64 `json:"constructions"`
	Goods         float64 `json:"goods"`
	Trainings     float64 `json:"trainings"`
	Other         float64 `json:"other"`
}

type GreenBookReadinessItem struct {
	ProjectID       string   `json:"project_id"`
	GreenBookID     string   `json:"green_book_id"`
	GBCode          string   `json:"gb_code"`
	ProjectName     string   `json:"project_name"`
	PublishYear     int32    `json:"publish_year"`
	ReadinessScore  int      `json:"readiness_score"`
	ReadinessStatus string   `json:"readiness_status"`
	IsCofinancing   bool     `json:"is_cofinancing"`
	MissingFields   []string `json:"missing_fields"`
	TotalFundingUSD float64  `json:"total_funding_usd"`
	InstitutionName string   `json:"institution_name,omitempty"`
	LenderNames     []string `json:"lender_names"`
}

type GreenBookReadinessDashboard struct {
	Summary                GreenBookReadinessSummary   `json:"summary"`
	DisbursementPlanByYear []GreenBookDisbursementYear `json:"disbursement_plan_by_year"`
	FundingAllocation      GreenBookFundingAllocation  `json:"funding_allocation"`
	ReadinessItems         []GreenBookReadinessItem    `json:"readiness_items"`
}

type LenderFinancingMixFilterRequest struct {
	LenderType  *string `json:"lender_type,omitempty"`
	LenderID    *string `json:"lender_id,omitempty"`
	Currency    *string `json:"currency,omitempty"`
	PeriodID    *string `json:"period_id,omitempty"`
	PublishYear *int32  `json:"publish_year,omitempty"`
}

type LenderFinancingMixSummary struct {
	TotalLenders        int     `json:"total_lenders"`
	BilateralUSD        float64 `json:"bilateral_usd"`
	MultilateralUSD     float64 `json:"multilateral_usd"`
	KSAUSD              float64 `json:"ksa_usd"`
	CofinancingProjects int     `json:"cofinancing_projects"`
}

type LenderCertaintyPoint struct {
	Stage        string  `json:"stage"`
	LenderID     string  `json:"lender_id"`
	LenderName   string  `json:"lender_name"`
	LenderType   string  `json:"lender_type"`
	ProjectCount int     `json:"project_count"`
	AmountUSD    float64 `json:"amount_usd"`
}

type LenderConversionItem struct {
	LenderID        string  `json:"lender_id"`
	LenderName      string  `json:"lender_name"`
	LenderType      string  `json:"lender_type"`
	IndicationCount int     `json:"indication_count"`
	LoICount        int     `json:"loi_count"`
	GBCount         int     `json:"gb_count"`
	DKCount         int     `json:"dk_count"`
	LACount         int     `json:"la_count"`
	IndicationUSD   float64 `json:"indication_usd"`
	LAUSD           float64 `json:"la_usd"`
	LAConversionPct float64 `json:"la_conversion_pct"`
}

type CurrencyExposureItem struct {
	Currency       string  `json:"currency"`
	Stage          string  `json:"stage"`
	ProjectCount   int     `json:"project_count"`
	AmountOriginal float64 `json:"amount_original"`
	AmountUSD      float64 `json:"amount_usd"`
}

type CofinancingItem struct {
	ProjectID     string   `json:"project_id"`
	ReferenceType string   `json:"reference_type"`
	ProjectCode   string   `json:"project_code,omitempty"`
	ProjectName   string   `json:"project_name"`
	LenderCount   int      `json:"lender_count"`
	LenderNames   []string `json:"lender_names"`
	AmountUSD     float64  `json:"amount_usd"`
}

type LenderFinancingMixDashboard struct {
	Summary          LenderFinancingMixSummary `json:"summary"`
	CertaintyLadder  []LenderCertaintyPoint    `json:"certainty_ladder"`
	LenderConversion []LenderConversionItem    `json:"lender_conversion"`
	CurrencyExposure []CurrencyExposureItem    `json:"currency_exposure"`
	CofinancingItems []CofinancingItem         `json:"cofinancing_items"`
}

type KLPortfolioPerformanceFilterRequest struct {
	InstitutionID   *string `json:"institution_id,omitempty"`
	InstitutionRole *string `json:"institution_role,omitempty"`
	PeriodID        *string `json:"period_id,omitempty"`
	PublishYear     *int32  `json:"publish_year,omitempty"`
	SortBy          *string `json:"sort_by,omitempty"`
}

type KLPortfolioPerformanceSummary struct {
	TotalInstitutions           int     `json:"total_institutions"`
	TopExposureInstitution      string  `json:"top_exposure_institution,omitempty"`
	TopExposureUSD              float64 `json:"top_exposure_usd,omitempty"`
	HighestRiskInstitution      string  `json:"highest_risk_institution,omitempty"`
	HighestRiskCount            int     `json:"highest_risk_count,omitempty"`
	TotalInstitutionExposureUSD float64 `json:"total_institution_exposure_usd,omitempty"`
	TotalInstitutionRiskCount   int     `json:"total_institution_risk_count,omitempty"`
}

type KLPortfolioPerformanceItem struct {
	InstitutionID       string  `json:"institution_id"`
	InstitutionName     string  `json:"institution_name"`
	BBProjectCount      int     `json:"bb_project_count"`
	GBProjectCount      int     `json:"gb_project_count"`
	DKProjectCount      int     `json:"dk_project_count"`
	LACount             int     `json:"la_count"`
	PipelineUSD         float64 `json:"pipeline_usd"`
	LACommitmentUSD     float64 `json:"la_commitment_usd"`
	RiskCount           int     `json:"risk_count"`
	PerformanceScore    float64 `json:"performance_score"`
	PerformanceCategory string  `json:"performance_category"`
}

type KLPortfolioPerformanceDashboard struct {
	Summary KLPortfolioPerformanceSummary `json:"summary"`
	Items   []KLPortfolioPerformanceItem  `json:"items"`
}

type DataQualityGovernanceFilterRequest struct {
	Severity       *string `json:"severity,omitempty"`
	Module         *string `json:"module,omitempty"`
	IssueType      *string `json:"issue_type,omitempty"`
	OnlyUnresolved bool    `json:"only_unresolved"`
	AuditDays      int32   `json:"audit_days"`
}

type DataQualityIssueSummary struct {
	TotalIssues  int  `json:"total_issues"`
	ErrorCount   int  `json:"error_count"`
	WarningCount int  `json:"warning_count"`
	InfoCount    int  `json:"info_count"`
	AuditEvents  *int `json:"audit_events,omitempty"`
}

type DataQualityIssueItem struct {
	Severity          string `json:"severity"`
	Module            string `json:"module"`
	IssueType         string `json:"issue_type"`
	RecordID          string `json:"record_id"`
	RecordLabel       string `json:"record_label"`
	Message           string `json:"message"`
	RecommendedAction string `json:"recommended_action"`
	IsResolved        bool   `json:"is_resolved"`
}

type AuditSummaryItem struct {
	Label         string `json:"label"`
	EventCount    int    `json:"event_count"`
	LastChangedAt string `json:"last_changed_at,omitempty"`
}

type AuditRecentActivityItem struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Action    string `json:"action"`
	TableName string `json:"table_name"`
	RecordID  string `json:"record_id"`
	ChangedAt string `json:"changed_at"`
}

type DataQualityAuditSummary struct {
	ByUser         []AuditSummaryItem        `json:"by_user"`
	ByTable        []AuditSummaryItem        `json:"by_table"`
	RecentActivity []AuditRecentActivityItem `json:"recent_activity"`
}

type DataQualityGovernanceDashboard struct {
	Summary      DataQualityIssueSummary  `json:"summary"`
	Issues       []DataQualityIssueItem   `json:"issues"`
	AuditSummary *DataQualityAuditSummary `json:"audit_summary,omitempty"`
}
