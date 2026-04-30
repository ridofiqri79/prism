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
