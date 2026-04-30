package model

type GreenBookRequest struct {
	PublishYear         int32   `json:"publish_year" validate:"required"`
	ReplacesGreenBookID *string `json:"replaces_green_book_id"`
	RevisionNumber      int32   `json:"revision_number"`
}

type GreenBookResponse struct {
	ID                  string  `json:"id"`
	PublishYear         int32   `json:"publish_year"`
	ReplacesGreenBookID *string `json:"replaces_green_book_id,omitempty"`
	RevisionNumber      int32   `json:"revision_number"`
	Status              string  `json:"status"`
	CreatedAt           string  `json:"created_at,omitempty"`
	UpdatedAt           string  `json:"updated_at,omitempty"`
}

type GreenBookListFilter struct {
	PublishYears []string
	Statuses     []string
}

type CreateGBProjectRequest struct {
	GBProjectIdentityID   *string                   `json:"gb_project_identity_id"`
	ProgramTitleID        *string                   `json:"program_title_id"`
	GBCode                string                    `json:"gb_code" validate:"required"`
	ProjectName           string                    `json:"project_name" validate:"required"`
	Duration              *int32                    `json:"duration"`
	Objective             *string                   `json:"objective"`
	ScopeOfProject        *string                   `json:"scope_of_project"`
	BBProjectIDs          []string                  `json:"bb_project_ids" validate:"required,min=1"`
	BappenasPartnerIDs    []string                  `json:"bappenas_partner_ids"`
	ExecutingAgencyIDs    []string                  `json:"executing_agency_ids" validate:"required,min=1"`
	ImplementingAgencyIDs []string                  `json:"implementing_agency_ids" validate:"required,min=1"`
	LocationIDs           []string                  `json:"location_ids" validate:"required,min=1"`
	Activities            []GBActivityItem          `json:"activities"`
	FundingSources        []GBFundingSourceItem     `json:"funding_sources"`
	DisbursementPlan      []GBDisbursementPlanItem  `json:"disbursement_plan"`
	FundingAllocations    []GBFundingAllocationItem `json:"funding_allocations"`
}

type UpdateGBProjectRequest = CreateGBProjectRequest

type GBProjectListFilter struct {
	BBProjectIDs       []string
	ExecutingAgencyIDs []string
	LocationIDs        []string
	Statuses           []string
}

type GBActivityItem struct {
	ActivityName           string  `json:"activity_name" validate:"required"`
	ImplementationLocation *string `json:"implementation_location"`
	PIU                    *string `json:"piu"`
	SortOrder              *int32  `json:"sort_order"`
}

type GBFundingSourceItem struct {
	LenderID      string  `json:"lender_id" validate:"required"`
	InstitutionID *string `json:"institution_id"`
	Currency      string  `json:"currency"`
	LoanOriginal  float64 `json:"loan_original"`
	GrantOriginal float64 `json:"grant_original"`
	LocalOriginal float64 `json:"local_original"`
	LoanUSD       float64 `json:"loan_usd"`
	GrantUSD      float64 `json:"grant_usd"`
	LocalUSD      float64 `json:"local_usd"`
}

type GBDisbursementPlanItem struct {
	Year      int32   `json:"year" validate:"required"`
	AmountUSD float64 `json:"amount_usd"`
}

type GBFundingAllocationItem struct {
	ActivityIndex int     `json:"activity_index"`
	Services      float64 `json:"services"`
	Constructions float64 `json:"constructions"`
	Goods         float64 `json:"goods"`
	Trainings     float64 `json:"trainings"`
	Other         float64 `json:"other"`
}

type GBProjectResponse struct {
	ID                   string                        `json:"id"`
	GreenBookID          string                        `json:"green_book_id"`
	GBProjectIdentityID  string                        `json:"gb_project_identity_id"`
	ProgramTitleID       *string                       `json:"program_title_id,omitempty"`
	GBCode               string                        `json:"gb_code"`
	ProjectName          string                        `json:"project_name"`
	Duration             *int32                        `json:"duration"`
	Objective            *string                       `json:"objective"`
	ScopeOfProject       *string                       `json:"scope_of_project"`
	BBProjects           []BBProjectSummary            `json:"bb_projects"`
	BappenasPartners     []BappenasPartnerResponse     `json:"bappenas_partners"`
	ExecutingAgencies    []InstitutionResponse         `json:"executing_agencies"`
	ImplementingAgencies []InstitutionResponse         `json:"implementing_agencies"`
	Locations            []RegionResponse              `json:"locations"`
	Activities           []GBActivityResponse          `json:"activities"`
	FundingSources       []GBFundingSourceResponse     `json:"funding_sources"`
	DisbursementPlan     []GBDisbursementPlanResponse  `json:"disbursement_plan"`
	FundingAllocations   []GBFundingAllocationResponse `json:"funding_allocations"`
	Status               string                        `json:"status"`
	IsLatest             bool                          `json:"is_latest"`
	HasNewerRevision     bool                          `json:"has_newer_revision"`
	CreatedAt            string                        `json:"created_at,omitempty"`
	UpdatedAt            string                        `json:"updated_at,omitempty"`
}

type BBProjectSummary struct {
	ID                string `json:"id"`
	BlueBookID        string `json:"blue_book_id,omitempty"`
	ProjectIdentityID string `json:"project_identity_id,omitempty"`
	BBCode            string `json:"bb_code"`
	ProjectName       string `json:"project_name"`
	IsLatest          bool   `json:"is_latest,omitempty"`
	HasNewerRevision  bool   `json:"has_newer_revision,omitempty"`
}

type GBProjectHistoryItem struct {
	ID                  string             `json:"id"`
	GBProjectIdentityID string             `json:"gb_project_identity_id"`
	GreenBookID         string             `json:"green_book_id"`
	GBCode              string             `json:"gb_code"`
	ProjectName         string             `json:"project_name"`
	BookLabel           string             `json:"book_label"`
	PublishYear         int32              `json:"publish_year"`
	RevisionNumber      int32              `json:"revision_number"`
	BookStatus          string             `json:"book_status"`
	IsLatest            bool               `json:"is_latest"`
	UsedByDownstream    bool               `json:"used_by_downstream"`
	BBProjects          []BBProjectSummary `json:"bb_projects,omitempty"`
}

type GBActivityResponse struct {
	ID                     string  `json:"id"`
	ActivityName           string  `json:"activity_name"`
	ImplementationLocation *string `json:"implementation_location"`
	PIU                    *string `json:"piu"`
	SortOrder              int32   `json:"sort_order"`
}

type GBFundingSourceResponse struct {
	ID            string           `json:"id"`
	Lender        LenderInfo       `json:"lender"`
	Institution   *InstitutionInfo `json:"institution,omitempty"`
	Currency      string           `json:"currency"`
	LoanOriginal  float64          `json:"loan_original"`
	GrantOriginal float64          `json:"grant_original"`
	LocalOriginal float64          `json:"local_original"`
	LoanUSD       float64          `json:"loan_usd"`
	GrantUSD      float64          `json:"grant_usd"`
	LocalUSD      float64          `json:"local_usd"`
}

type InstitutionInfo struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	ShortName *string `json:"short_name,omitempty"`
	Level     string  `json:"level"`
}

type GBDisbursementPlanResponse struct {
	ID        string  `json:"id"`
	Year      int32   `json:"year"`
	AmountUSD float64 `json:"amount_usd"`
}

type GBFundingAllocationResponse struct {
	ID            string  `json:"id"`
	GBActivityID  string  `json:"gb_activity_id"`
	ActivityName  string  `json:"activity_name"`
	SortOrder     int32   `json:"sort_order"`
	Services      float64 `json:"services"`
	Constructions float64 `json:"constructions"`
	Goods         float64 `json:"goods"`
	Trainings     float64 `json:"trainings"`
	Other         float64 `json:"other"`
}
