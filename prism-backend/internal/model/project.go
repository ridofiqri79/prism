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
	Search              *string
}

type ProjectMasterResponse struct {
	ID                string   `json:"id"`
	BlueBookID        string   `json:"blue_book_id"`
	BBCode            string   `json:"bb_code"`
	ProjectName       string   `json:"project_name"`
	LoanTypes         []string `json:"loan_types"`
	IndicationLenders []string `json:"indication_lenders"`
	ExecutingAgencies []string `json:"executing_agencies"`
	FixedLenders      []string `json:"fixed_lenders"`
	ProjectStatus     string   `json:"project_status"`
	PipelineStatus    string   `json:"pipeline_status"`
	ProgramTitle      string   `json:"program_title"`
	Locations         []string `json:"locations"`
	ForeignLoanUSD    float64  `json:"foreign_loan_usd"`
	DKDates           []string `json:"dk_dates"`
}
