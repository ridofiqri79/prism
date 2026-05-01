package model

type MonitoringRequest struct {
	BudgetYear         int32                    `json:"budget_year" validate:"required"`
	Quarter            string                   `json:"quarter" validate:"required"`
	ExchangeRateUSDIDR float64                  `json:"exchange_rate_usd_idr"`
	ExchangeRateLAIDR  float64                  `json:"exchange_rate_la_idr"`
	PlannedLA          float64                  `json:"planned_la"`
	PlannedUSD         float64                  `json:"planned_usd"`
	PlannedIDR         float64                  `json:"planned_idr"`
	RealizedLA         float64                  `json:"realized_la"`
	RealizedUSD        float64                  `json:"realized_usd"`
	RealizedIDR        float64                  `json:"realized_idr"`
	Komponen           []MonitoringKomponenItem `json:"komponen"`
	Components         []MonitoringKomponenItem `json:"components"`
}

type MonitoringKomponenItem struct {
	ComponentName string  `json:"component_name" validate:"required"`
	PlannedLA     float64 `json:"planned_la"`
	PlannedUSD    float64 `json:"planned_usd"`
	PlannedIDR    float64 `json:"planned_idr"`
	RealizedLA    float64 `json:"realized_la"`
	RealizedUSD   float64 `json:"realized_usd"`
	RealizedIDR   float64 `json:"realized_idr"`
}

type MonitoringResponse struct {
	ID                 string                       `json:"id"`
	LoanAgreementID    string                       `json:"loan_agreement_id"`
	BudgetYear         int32                        `json:"budget_year"`
	Quarter            string                       `json:"quarter"`
	ExchangeRateUSDIDR float64                      `json:"exchange_rate_usd_idr"`
	ExchangeRateLAIDR  float64                      `json:"exchange_rate_la_idr"`
	PlannedLA          float64                      `json:"planned_la"`
	PlannedUSD         float64                      `json:"planned_usd"`
	PlannedIDR         float64                      `json:"planned_idr"`
	RealizedLA         float64                      `json:"realized_la"`
	RealizedUSD        float64                      `json:"realized_usd"`
	RealizedIDR        float64                      `json:"realized_idr"`
	AbsorptionPct      float64                      `json:"absorption_pct"`
	Komponen           []MonitoringKomponenResponse `json:"komponen"`
	CreatedAt          string                       `json:"created_at,omitempty"`
	UpdatedAt          string                       `json:"updated_at,omitempty"`
}

type MonitoringLoanAgreementResponse struct {
	ID                 string     `json:"id"`
	LoanCode           string     `json:"loan_code"`
	EffectiveDate      string     `json:"effective_date"`
	IsEffective        bool       `json:"is_effective"`
	Currency           string     `json:"currency"`
	AmountUSD          float64    `json:"amount_usd"`
	Lender             LenderInfo `json:"lender"`
	DKLetterNumber     *string    `json:"dk_letter_number,omitempty"`
	DKProjectName      string     `json:"dk_project_name"`
	MonitoringCount    int64      `json:"monitoring_count"`
	LatestMonitoringAt string     `json:"latest_monitoring_at,omitempty"`
}

type MonitoringLoanAgreementListFilter struct {
	IsEffective *string
}

type MonitoringListFilter struct {
	BudgetYear *string
	Quarter    *string
}

type MonitoringKomponenResponse struct {
	ID            string  `json:"id"`
	ComponentName string  `json:"component_name"`
	PlannedLA     float64 `json:"planned_la"`
	PlannedUSD    float64 `json:"planned_usd"`
	PlannedIDR    float64 `json:"planned_idr"`
	RealizedLA    float64 `json:"realized_la"`
	RealizedUSD   float64 `json:"realized_usd"`
	RealizedIDR   float64 `json:"realized_idr"`
}
