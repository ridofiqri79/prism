package model

type LoanAgreementRequest struct {
	DKProjectID            string  `json:"dk_project_id" validate:"required"`
	LenderID               string  `json:"lender_id" validate:"required"`
	LoanCode               string  `json:"loan_code" validate:"required"`
	AgreementDate          string  `json:"agreement_date" validate:"required"`
	EffectiveDate          string  `json:"effective_date" validate:"required"`
	OriginalClosingDate    string  `json:"original_closing_date"`
	ClosingDate            string  `json:"closing_date" validate:"required"`
	Currency               string  `json:"currency" validate:"required"`
	AmountOriginal         float64 `json:"amount_original"`
	AmountUSD              float64 `json:"amount_usd"`
	CumulativeDisbursement float64 `json:"cumulative_disbursement"`
}

type LoanAgreementResponse struct {
	ID                     string     `json:"id"`
	DKProjectID            string     `json:"dk_project_id"`
	Lender                 LenderInfo `json:"lender"`
	LoanCode               string     `json:"loan_code"`
	AgreementDate          string     `json:"agreement_date"`
	EffectiveDate          string     `json:"effective_date"`
	OriginalClosingDate    string     `json:"original_closing_date"`
	ClosingDate            string     `json:"closing_date"`
	IsExtended             bool       `json:"is_extended"`
	ExtensionDays          int        `json:"extension_days"`
	Currency               string     `json:"currency"`
	AmountOriginal         float64    `json:"amount_original"`
	AmountUSD              float64    `json:"amount_usd"`
	CumulativeDisbursement float64    `json:"cumulative_disbursement"`
	CreatedAt              string     `json:"created_at,omitempty"`
	UpdatedAt              string     `json:"updated_at,omitempty"`
}

type LoanAgreementSummary struct {
	ID                     string  `json:"id"`
	LoanCode               string  `json:"loan_code"`
	Currency               string  `json:"currency"`
	CumulativeDisbursement float64 `json:"cumulative_disbursement"`
}

type LoanAgreementListFilter struct {
	LenderID          *string
	IsExtended        *string
	ClosingDateBefore *string
}
