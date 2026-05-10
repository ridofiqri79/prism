package model

type LoanAgreementRequest struct {
	DKProjectID            string                             `json:"dk_project_id"`
	DKProjectAllocations   []LoanAgreementDKProjectAllocation `json:"dk_project_allocations"`
	LenderID               string                             `json:"lender_id" validate:"required"`
	LoanCode               string                             `json:"loan_code" validate:"required"`
	AgreementDate          string                             `json:"agreement_date" validate:"required"`
	EffectiveDate          string                             `json:"effective_date" validate:"required"`
	OriginalClosingDate    string                             `json:"original_closing_date"`
	ClosingDate            string                             `json:"closing_date" validate:"required"`
	Currency               string                             `json:"currency" validate:"required"`
	AmountOriginal         float64                            `json:"amount_original"`
	AmountUSD              float64                            `json:"amount_usd"`
	CumulativeDisbursement float64                            `json:"cumulative_disbursement"`
}

type LoanAgreementDKProjectAllocation struct {
	DKProjectID        string  `json:"dk_project_id" validate:"required"`
	AllocationOriginal float64 `json:"allocation_original" validate:"required"`
}

type LoanAgreementResponse struct {
	ID                        string                           `json:"id"`
	DKProjectID               string                           `json:"dk_project_id,omitempty"`
	DKProjects                []LoanAgreementDKProjectResponse `json:"dk_projects"`
	Lender                    LenderInfo                       `json:"lender"`
	LoanCode                  string                           `json:"loan_code"`
	AgreementDate             string                           `json:"agreement_date"`
	EffectiveDate             string                           `json:"effective_date"`
	OriginalClosingDate       string                           `json:"original_closing_date"`
	ClosingDate               string                           `json:"closing_date"`
	IsExtended                bool                             `json:"is_extended"`
	ExtensionDays             int                              `json:"extension_days"`
	Currency                  string                           `json:"currency"`
	AmountOriginal            float64                          `json:"amount_original"`
	AmountUSD                 float64                          `json:"amount_usd"`
	CumulativeDisbursement    float64                          `json:"cumulative_disbursement"`
	CumulativeDisbursementUSD *float64                         `json:"cumulative_disbursement_usd"`
	DisbursementRatio         *float64                         `json:"disbursement_ratio"`
	EstimatedTimeRatio        *float64                         `json:"estimated_time_ratio"`
	PerformanceValue          *float64                         `json:"performance_value"`
	PerformanceStatus         *string                          `json:"performance_status"`
	KursTengahBI              *float64                         `json:"kurs_tengah_bi"`
	KursCutOffDate            *string                          `json:"kurs_cut_off_date"`
	CreatedAt                 string                           `json:"created_at,omitempty"`
	UpdatedAt                 string                           `json:"updated_at,omitempty"`
}

type LoanAgreementDKProjectResponse struct {
	ID                 string                    `json:"id"`
	DKID               string                    `json:"dk_id"`
	ProjectName        string                    `json:"project_name"`
	Objectives         *string                   `json:"objectives,omitempty"`
	GBCodes            string                    `json:"gb_codes"`
	DaftarKegiatan     LoanAgreementDKHeaderInfo `json:"daftar_kegiatan"`
	AllocationOriginal float64                   `json:"allocation_original"`
	AllocationUSD      float64                   `json:"allocation_usd"`
}

type LoanAgreementDKHeaderInfo struct {
	ID           string  `json:"id"`
	Subject      string  `json:"subject"`
	Date         string  `json:"date"`
	LetterNumber *string `json:"letter_number,omitempty"`
}

type LoanAgreementSummary struct {
	ID                     string  `json:"id"`
	LoanCode               string  `json:"loan_code"`
	Currency               string  `json:"currency"`
	AmountOriginal         float64 `json:"amount_original"`
	AmountUSD              float64 `json:"amount_usd"`
	AllocationOriginal     float64 `json:"allocation_original"`
	AllocationUSD          float64 `json:"allocation_usd"`
	CumulativeDisbursement float64 `json:"cumulative_disbursement"`
}

type LoanAgreementListFilter struct {
	PeriodIDs         []string
	LenderID          *string
	IsExtended        *string
	ClosingDateBefore *string
}
