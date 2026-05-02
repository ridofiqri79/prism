package model

type ProjectMasterFilter struct {
	LoanTypes           []string
	IndicationLenderIDs []string
	ExecutingAgencyIDs  []string
	FixedLenderIDs      []string
	ProjectStatuses     []string
	PipelineStatuses    []string
	ProgramTitleIDs     []string
	RegionIDs           []string
	ForeignLoanMin      *string
	ForeignLoanMax      *string
	DKDateFrom          *string
	DKDateTo            *string
	DataQualityCodes    []string
	DataQualityStages   []string
	Search              *string
	IncludeHistory      bool
}

type ProjectMasterResponse struct {
	ID                    string   `json:"id"`
	BlueBookID            string   `json:"blue_book_id"`
	ProjectIdentityID     string   `json:"project_identity_id"`
	BBCode                string   `json:"bb_code"`
	ProjectName           string   `json:"project_name"`
	LoanTypes             []string `json:"loan_types"`
	IndicationLenders     []string `json:"indication_lenders"`
	ExecutingAgencies     []string `json:"executing_agencies"`
	FixedLenders          []string `json:"fixed_lenders"`
	ProjectStatus         string   `json:"project_status"`
	PipelineStatus        string   `json:"pipeline_status"`
	ProgramTitle          string   `json:"program_title"`
	Locations             []string `json:"locations"`
	ForeignLoanUSD        float64  `json:"foreign_loan_usd"`
	DKDates               []string `json:"dk_dates"`
	IsLatest              bool     `json:"is_latest"`
	HasNewerRevision      bool     `json:"has_newer_revision"`
	BlueBookRevisionLabel string   `json:"blue_book_revision_label"`
}

type ProjectMasterFundingSummary struct {
	TotalLoanUSD        float64 `json:"total_loan_usd"`
	TotalGrantUSD       float64 `json:"total_grant_usd"`
	TotalCounterpartUSD float64 `json:"total_counterpart_usd"`
}

type ProjectMasterListResponse struct {
	Data    []ProjectMasterResponse     `json:"data"`
	Meta    PaginationMeta              `json:"meta"`
	Summary ProjectMasterFundingSummary `json:"summary"`
}
