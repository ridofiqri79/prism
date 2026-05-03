package model

type JourneyResponse struct {
	BBProject  JourneyBBProject   `json:"bb_project"`
	LoI        []JourneyLoI       `json:"loi"`
	GBProjects []JourneyGBProject `json:"gb_projects"`
}

type JourneyBBProject struct {
	ID                          string                    `json:"id"`
	BlueBookID                  string                    `json:"blue_book_id"`
	ProjectIdentityID           string                    `json:"project_identity_id"`
	BBCode                      string                    `json:"bb_code"`
	ProjectName                 string                    `json:"project_name"`
	BlueBookRevisionLabel       string                    `json:"blue_book_revision_label"`
	IsLatest                    bool                      `json:"is_latest"`
	HasNewerRevision            bool                      `json:"has_newer_revision"`
	LatestBBProjectID           string                    `json:"latest_bb_project_id"`
	LatestBlueBookRevisionLabel string                    `json:"latest_blue_book_revision_label"`
	LenderIndications           []JourneyLenderIndication `json:"lender_indications"`
}

type JourneyLenderIndication struct {
	ID      string     `json:"id"`
	Lender  LenderInfo `json:"lender"`
	Remarks *string    `json:"remarks,omitempty"`
}

type JourneyLoI struct {
	ID           string     `json:"id"`
	Lender       LenderInfo `json:"lender"`
	Subject      string     `json:"subject"`
	Date         string     `json:"date"`
	LetterNumber *string    `json:"letter_number,omitempty"`
}

type JourneyGBProject struct {
	ID                           string                 `json:"id"`
	GreenBookID                  string                 `json:"green_book_id"`
	GBProjectIdentityID          string                 `json:"gb_project_identity_id"`
	GBCode                       string                 `json:"gb_code"`
	ProjectName                  string                 `json:"project_name"`
	Status                       string                 `json:"status"`
	GreenBookRevisionLabel       string                 `json:"green_book_revision_label"`
	IsLatest                     bool                   `json:"is_latest"`
	HasNewerRevision             bool                   `json:"has_newer_revision"`
	LatestGBProjectID            string                 `json:"latest_gb_project_id"`
	LatestGreenBookRevisionLabel string                 `json:"latest_green_book_revision_label"`
	FundingSources               []JourneyFundingSource `json:"funding_sources"`
	DKProjects                   []JourneyDKProject     `json:"dk_projects"`
}

type JourneyDKProject struct {
	ID             string                `json:"id"`
	ProjectName    string                `json:"project_name"`
	Objectives     *string               `json:"objectives,omitempty"`
	DaftarKegiatan *JourneyDKHeader      `json:"daftar_kegiatan,omitempty"`
	LoanAgreement  *JourneyLoanAgreement `json:"loan_agreement"`
}

type JourneyDKHeader struct {
	ID           string  `json:"id"`
	Subject      string  `json:"subject"`
	Date         string  `json:"date"`
	LetterNumber *string `json:"letter_number,omitempty"`
}

type JourneyLoanAgreement struct {
	ID                  string                      `json:"id"`
	LoanCode            string                      `json:"loan_code"`
	Lender              LenderInfo                  `json:"lender"`
	AgreementDate       string                      `json:"agreement_date"`
	EffectiveDate       string                      `json:"effective_date"`
	OriginalClosingDate string                      `json:"original_closing_date"`
	ClosingDate         string                      `json:"closing_date"`
	IsExtended          bool                        `json:"is_extended"`
	ExtensionDays       int                         `json:"extension_days"`
	Currency            string                      `json:"currency"`
	AmountOriginal      float64                     `json:"amount_original"`
	AmountUSD           float64                     `json:"amount_usd"`
	Monitoring          []JourneyMonitoringResponse `json:"monitoring"`
}

type JourneyFundingSource struct {
	ID            string                  `json:"id"`
	Lender        LenderInfo              `json:"lender"`
	Institution   *JourneyInstitutionInfo `json:"institution,omitempty"`
	Currency      string                  `json:"currency"`
	LoanOriginal  float64                 `json:"loan_original"`
	GrantOriginal float64                 `json:"grant_original"`
	LocalOriginal float64                 `json:"local_original"`
	LoanUSD       float64                 `json:"loan_usd"`
	GrantUSD      float64                 `json:"grant_usd"`
	LocalUSD      float64                 `json:"local_usd"`
}

type JourneyInstitutionInfo struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	ShortName *string `json:"short_name,omitempty"`
}

type JourneyMonitoringResponse struct {
	ID            string  `json:"id"`
	BudgetYear    int32   `json:"budget_year"`
	Quarter       string  `json:"quarter"`
	PlannedUSD    float64 `json:"planned_usd"`
	RealizedUSD   float64 `json:"realized_usd"`
	AbsorptionPct float64 `json:"absorption_pct"`
}
