package model

type JourneyResponse struct {
	BBProject  JourneyBBProject   `json:"bb_project"`
	LoI        []JourneyLoI       `json:"loi"`
	GBProjects []JourneyGBProject `json:"gb_projects"`
}

type JourneyBBProject struct {
	ID          string `json:"id"`
	BBCode      string `json:"bb_code"`
	ProjectName string `json:"project_name"`
}

type JourneyLoI struct {
	ID           string     `json:"id"`
	Lender       LenderInfo `json:"lender"`
	Subject      string     `json:"subject"`
	Date         string     `json:"date"`
	LetterNumber *string    `json:"letter_number,omitempty"`
}

type JourneyGBProject struct {
	ID          string             `json:"id"`
	GBCode      string             `json:"gb_code"`
	ProjectName string             `json:"project_name"`
	Status      string             `json:"status"`
	DKProjects  []JourneyDKProject `json:"dk_projects"`
}

type JourneyDKProject struct {
	ID            string                `json:"id"`
	Objectives    *string               `json:"objectives,omitempty"`
	LoanAgreement *JourneyLoanAgreement `json:"loan_agreement"`
}

type JourneyLoanAgreement struct {
	ID                  string                      `json:"id"`
	LoanCode            string                      `json:"loan_code"`
	Lender              LenderInfo                  `json:"lender"`
	EffectiveDate       string                      `json:"effective_date"`
	OriginalClosingDate string                      `json:"original_closing_date"`
	ClosingDate         string                      `json:"closing_date"`
	IsExtended          bool                        `json:"is_extended"`
	ExtensionDays       int                         `json:"extension_days"`
	Monitoring          []JourneyMonitoringResponse `json:"monitoring"`
}

type JourneyMonitoringResponse struct {
	ID            string  `json:"id"`
	BudgetYear    int32   `json:"budget_year"`
	Quarter       string  `json:"quarter"`
	PlannedUSD    float64 `json:"planned_usd"`
	RealizedUSD   float64 `json:"realized_usd"`
	AbsorptionPct float64 `json:"absorption_pct"`
}
