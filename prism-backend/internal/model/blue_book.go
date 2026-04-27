package model

type BlueBookRequest struct {
	PeriodID       string `json:"period_id" validate:"required"`
	PublishDate    string `json:"publish_date" validate:"required"`
	RevisionNumber int32  `json:"revision_number"`
	RevisionYear   *int32 `json:"revision_year"`
}

type PeriodInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	YearStart int32  `json:"year_start"`
	YearEnd   int32  `json:"year_end"`
}

type BlueBookResponse struct {
	ID             string     `json:"id"`
	Period         PeriodInfo `json:"period"`
	PublishDate    string     `json:"publish_date"`
	RevisionNumber int32      `json:"revision_number"`
	RevisionYear   *int32     `json:"revision_year"`
	Status         string     `json:"status"`
	CreatedAt      string     `json:"created_at,omitempty"`
	UpdatedAt      string     `json:"updated_at,omitempty"`
}

type CreateBBProjectRequest struct {
	ProgramTitleID        *string                `json:"program_title_id"`
	BappenasPartnerID     *string                `json:"bappenas_partner_id"`
	BBCode                string                 `json:"bb_code" validate:"required"`
	ProjectName           string                 `json:"project_name" validate:"required"`
	Duration              *string                `json:"duration"`
	Objective             *string                `json:"objective"`
	ScopeOfWork           *string                `json:"scope_of_work"`
	Outputs               *string                `json:"outputs"`
	Outcomes              *string                `json:"outcomes"`
	ExecutingAgencyIDs    []string               `json:"executing_agency_ids" validate:"required,min=1"`
	ImplementingAgencyIDs []string               `json:"implementing_agency_ids" validate:"required,min=1"`
	LocationIDs           []string               `json:"location_ids" validate:"required,min=1"`
	NationalPriorityIDs   []string               `json:"national_priority_ids"`
	ProjectCosts          []ProjectCostItem      `json:"project_costs"`
	LenderIndications     []LenderIndicationItem `json:"lender_indications"`
}

type UpdateBBProjectRequest = CreateBBProjectRequest

type ProjectCostItem struct {
	FundingType     string  `json:"funding_type" validate:"required"`
	FundingCategory string  `json:"funding_category" validate:"required"`
	AmountUSD       float64 `json:"amount_usd"`
}

type LenderIndicationItem struct {
	LenderID string  `json:"lender_id" validate:"required"`
	Remarks  *string `json:"remarks"`
}

type BBProjectResponse struct {
	ID                   string                     `json:"id"`
	BlueBookID           string                     `json:"blue_book_id"`
	ProgramTitleID       *string                    `json:"program_title_id,omitempty"`
	BappenasPartnerID    *string                    `json:"bappenas_partner_id,omitempty"`
	BBCode               string                     `json:"bb_code"`
	ProjectName          string                     `json:"project_name"`
	Duration             *string                    `json:"duration"`
	Objective            *string                    `json:"objective"`
	ScopeOfWork          *string                    `json:"scope_of_work"`
	Outputs              *string                    `json:"outputs"`
	Outcomes             *string                    `json:"outcomes"`
	ExecutingAgencies    []InstitutionResponse      `json:"executing_agencies"`
	ImplementingAgencies []InstitutionResponse      `json:"implementing_agencies"`
	Locations            []RegionResponse           `json:"locations"`
	NationalPriorities   []NationalPriorityResponse `json:"national_priorities"`
	ProjectCosts         []ProjectCostResponse      `json:"project_costs"`
	LenderIndications    []LenderIndicationResponse `json:"lender_indications"`
	Status               string                     `json:"status"`
	CreatedAt            string                     `json:"created_at,omitempty"`
	UpdatedAt            string                     `json:"updated_at,omitempty"`
}

type ProjectCostResponse struct {
	ID              string  `json:"id"`
	FundingType     string  `json:"funding_type"`
	FundingCategory string  `json:"funding_category"`
	AmountUSD       float64 `json:"amount_usd"`
}

type LenderInfo struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	ShortName *string `json:"short_name,omitempty"`
	Type      string  `json:"type"`
}

type LenderIndicationResponse struct {
	ID      string     `json:"id"`
	Lender  LenderInfo `json:"lender"`
	Remarks *string    `json:"remarks"`
}

type LoIRequest struct {
	LenderID     string  `json:"lender_id" validate:"required"`
	Subject      string  `json:"subject" validate:"required"`
	Date         string  `json:"date" validate:"required"`
	LetterNumber *string `json:"letter_number"`
}

type LoIResponse struct {
	ID           string     `json:"id"`
	BBProjectID  string     `json:"bb_project_id"`
	Lender       LenderInfo `json:"lender"`
	Subject      string     `json:"subject"`
	Date         string     `json:"date"`
	LetterNumber *string    `json:"letter_number"`
	CreatedAt    string     `json:"created_at,omitempty"`
	UpdatedAt    string     `json:"updated_at,omitempty"`
}
