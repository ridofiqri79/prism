package model

type DaftarKegiatanRequest struct {
	LetterNumber *string `json:"letter_number"`
	Subject      string  `json:"subject" validate:"required"`
	Date         string  `json:"date"`
	Tanggal      string  `json:"tanggal"`
}

type DaftarKegiatanResponse struct {
	ID           string  `json:"id"`
	LetterNumber *string `json:"letter_number,omitempty"`
	Subject      string  `json:"subject"`
	Date         string  `json:"date"`
	ProjectCount int64   `json:"project_count"`
	CreatedAt    string  `json:"created_at,omitempty"`
	UpdatedAt    string  `json:"updated_at,omitempty"`
}

type DaftarKegiatanListFilter struct {
	DateFrom *string
	DateTo   *string
}

type CreateDKProjectRequest struct {
	ProgramTitleID     *string                 `json:"program_title_id"`
	InstitutionID      *string                 `json:"institution_id"`
	ProjectName        string                  `json:"project_name" validate:"required"`
	Duration           *int32                  `json:"duration"`
	Objectives         *string                 `json:"objectives"`
	GBProjectIDs       []string                `json:"gb_project_ids" validate:"required,min=1"`
	BappenasPartnerIDs []string                `json:"bappenas_partner_ids"`
	LocationIDs        []string                `json:"location_ids"`
	FinancingDetails   []DKFinancingDetailItem `json:"financing_details"`
	LoanAllocations    []DKLoanAllocationItem  `json:"loan_allocations"`
	ActivityDetails    []DKActivityDetailItem  `json:"activity_details"`
}

type UpdateDKProjectRequest = CreateDKProjectRequest

type DKProjectListFilter struct {
	GBProjectIDs       []string
	ExecutingAgencyIDs []string
	LocationIDs        []string
	LenderIDs          []string
}

type DKFinancingDetailItem struct {
	LenderID            *string `json:"lender_id"`
	Currency            string  `json:"currency"`
	AmountOriginal      float64 `json:"amount_original"`
	GrantOriginal       float64 `json:"grant_original"`
	CounterpartOriginal float64 `json:"counterpart_original"`
	AmountUSD           float64 `json:"amount_usd"`
	GrantUSD            float64 `json:"grant_usd"`
	CounterpartUSD      float64 `json:"counterpart_usd"`
	Remarks             *string `json:"remarks"`
}

type DKLoanAllocationItem struct {
	InstitutionID       *string `json:"institution_id"`
	Currency            string  `json:"currency"`
	AmountOriginal      float64 `json:"amount_original"`
	GrantOriginal       float64 `json:"grant_original"`
	CounterpartOriginal float64 `json:"counterpart_original"`
	AmountUSD           float64 `json:"amount_usd"`
	GrantUSD            float64 `json:"grant_usd"`
	CounterpartUSD      float64 `json:"counterpart_usd"`
	Remarks             *string `json:"remarks"`
}

type DKActivityDetailItem struct {
	ActivityNumber int32  `json:"activity_number" validate:"required"`
	ActivityName   string `json:"activity_name" validate:"required"`
}

type DKProjectResponse struct {
	ID               string                      `json:"id"`
	DKID             string                      `json:"dk_id"`
	ProgramTitleID   *string                     `json:"program_title_id,omitempty"`
	InstitutionID    *string                     `json:"institution_id,omitempty"`
	ProjectName      string                      `json:"project_name"`
	Duration         *int32                      `json:"duration"`
	Objectives       *string                     `json:"objectives"`
	GBProjects       []GBProjectSummary          `json:"gb_projects"`
	BappenasPartners []BappenasPartnerResponse   `json:"bappenas_partners"`
	Locations        []RegionResponse            `json:"locations"`
	FinancingDetails []DKFinancingDetailResponse `json:"financing_details"`
	LoanAllocations  []DKLoanAllocationResponse  `json:"loan_allocations"`
	ActivityDetails  []DKActivityDetailResponse  `json:"activity_details"`
	LoanAgreements   []LoanAgreementSummary      `json:"loan_agreements"`
	CreatedAt        string                      `json:"created_at,omitempty"`
	UpdatedAt        string                      `json:"updated_at,omitempty"`
}

type GBProjectSummary struct {
	ID                  string `json:"id"`
	GBProjectIdentityID string `json:"gb_project_identity_id,omitempty"`
	GreenBookID         string `json:"green_book_id,omitempty"`
	GBCode              string `json:"gb_code"`
	ProjectName         string `json:"project_name"`
	IsLatest            bool   `json:"is_latest,omitempty"`
	HasNewerRevision    bool   `json:"has_newer_revision,omitempty"`
}

type DKFinancingDetailResponse struct {
	ID                  string      `json:"id"`
	Lender              *LenderInfo `json:"lender,omitempty"`
	Currency            string      `json:"currency"`
	AmountOriginal      float64     `json:"amount_original"`
	GrantOriginal       float64     `json:"grant_original"`
	CounterpartOriginal float64     `json:"counterpart_original"`
	AmountUSD           float64     `json:"amount_usd"`
	GrantUSD            float64     `json:"grant_usd"`
	CounterpartUSD      float64     `json:"counterpart_usd"`
	Remarks             *string     `json:"remarks"`
}

type DKLoanAllocationResponse struct {
	ID                  string           `json:"id"`
	Institution         *InstitutionInfo `json:"institution,omitempty"`
	Currency            string           `json:"currency"`
	AmountOriginal      float64          `json:"amount_original"`
	GrantOriginal       float64          `json:"grant_original"`
	CounterpartOriginal float64          `json:"counterpart_original"`
	AmountUSD           float64          `json:"amount_usd"`
	GrantUSD            float64          `json:"grant_usd"`
	CounterpartUSD      float64          `json:"counterpart_usd"`
	Remarks             *string          `json:"remarks"`
}

type DKActivityDetailResponse struct {
	ID             string `json:"id"`
	ActivityNumber int32  `json:"activity_number"`
	ActivityName   string `json:"activity_name"`
}
