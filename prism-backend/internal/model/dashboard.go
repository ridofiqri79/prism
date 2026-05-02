package model

type DashboardFilterRequest struct {
	PeriodID       *string `json:"period_id,omitempty"`
	PublishYear    *int32  `json:"publish_year,omitempty"`
	BudgetYear     *int32  `json:"budget_year,omitempty"`
	Quarter        *string `json:"quarter,omitempty"`
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

type TimeSeriesPoint struct {
	Period        string  `json:"period"`
	BudgetYear    int32   `json:"budget_year"`
	Quarter       string  `json:"quarter"`
	PlannedUSD    float64 `json:"planned_usd"`
	RealizedUSD   float64 `json:"realized_usd"`
	AbsorptionPct float64 `json:"absorption_pct"`
}

type BreakdownItem struct {
	ID          *string  `json:"id,omitempty"`
	Key         string   `json:"key,omitempty"`
	Label       string   `json:"label"`
	ItemCount   int      `json:"item_count,omitempty"`
	AmountUSD   float64  `json:"amount_usd,omitempty"`
	RealizedUSD float64  `json:"realized_usd,omitempty"`
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
	AbsorptionPct      float64 `json:"absorption_pct,omitempty"`
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
	PlannedDisbursementUSD  float64      `json:"planned_disbursement_usd"`
	RealizedDisbursementUSD float64      `json:"realized_disbursement_usd"`
	AbsorptionPct           float64      `json:"absorption_pct"`
	LAAbsorptionPct         float64      `json:"la_absorption_pct"`
	UndisbursedUSD          float64      `json:"undisbursed_usd"`
	Metrics                 []MetricCard `json:"metrics"`
}

type DashboardLAExposure struct {
	LACommitmentUSD         float64 `json:"la_commitment_usd"`
	RealizedDisbursementUSD float64 `json:"realized_disbursement_usd"`
	UndisbursedUSD          float64 `json:"undisbursed_usd"`
	LAAbsorptionPct         float64 `json:"la_absorption_pct"`
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
