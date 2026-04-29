package model

type CountryRequest struct {
	Name string `json:"name" validate:"required"`
	Code string `json:"code" validate:"required"`
}

type CountryResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Code      string `json:"code"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type CountryInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type CurrencyRequest struct {
	Code      string  `json:"code" validate:"required"`
	Name      string  `json:"name" validate:"required"`
	Symbol    *string `json:"symbol"`
	IsActive  bool    `json:"is_active"`
	SortOrder int32   `json:"sort_order"`
}

type CurrencyResponse struct {
	ID        string  `json:"id"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Symbol    *string `json:"symbol"`
	IsActive  bool    `json:"is_active"`
	SortOrder int32   `json:"sort_order"`
	CreatedAt string  `json:"created_at,omitempty"`
	UpdatedAt string  `json:"updated_at,omitempty"`
}

type CreateLenderRequest struct {
	CountryID *string `json:"country_id"`
	Name      string  `json:"name" validate:"required"`
	ShortName *string `json:"short_name"`
	Type      string  `json:"type" validate:"required,oneof=Bilateral Multilateral KSA"`
}

type UpdateLenderRequest = CreateLenderRequest

type LenderResponse struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	ShortName *string      `json:"short_name"`
	Type      string       `json:"type"`
	Country   *CountryInfo `json:"country,omitempty"`
	CreatedAt string       `json:"created_at,omitempty"`
	UpdatedAt string       `json:"updated_at,omitempty"`
}

type InstitutionRequest struct {
	ParentID  *string `json:"parent_id"`
	Name      string  `json:"name" validate:"required"`
	ShortName *string `json:"short_name"`
	Level     string  `json:"level" validate:"required"`
}

type InstitutionResponse struct {
	ID         string  `json:"id"`
	ParentID   *string `json:"parent_id"`
	ParentName *string `json:"parent_name,omitempty"`
	Name       string  `json:"name"`
	ShortName  *string `json:"short_name"`
	Level      string  `json:"level"`
	CreatedAt  string  `json:"created_at,omitempty"`
	UpdatedAt  string  `json:"updated_at,omitempty"`
}

type RegionRequest struct {
	Code       string  `json:"code" validate:"required"`
	Name       string  `json:"name" validate:"required"`
	Type       string  `json:"type" validate:"required,oneof=COUNTRY PROVINCE CITY"`
	ParentCode *string `json:"parent_code"`
}

type RegionResponse struct {
	ID         string  `json:"id"`
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	ParentCode *string `json:"parent_code"`
	CreatedAt  string  `json:"created_at,omitempty"`
	UpdatedAt  string  `json:"updated_at,omitempty"`
}

type ProgramTitleRequest struct {
	ParentID *string `json:"parent_id"`
	Title    string  `json:"title" validate:"required"`
}

type ProgramTitleResponse struct {
	ID        string  `json:"id"`
	ParentID  *string `json:"parent_id"`
	Title     string  `json:"title"`
	CreatedAt string  `json:"created_at,omitempty"`
	UpdatedAt string  `json:"updated_at,omitempty"`
}

type BappenasPartnerRequest struct {
	ParentID *string `json:"parent_id"`
	Name     string  `json:"name" validate:"required"`
	Level    string  `json:"level" validate:"required,oneof=Eselon I Eselon II"`
}

type BappenasPartnerResponse struct {
	ID        string  `json:"id"`
	ParentID  *string `json:"parent_id"`
	Name      string  `json:"name"`
	Level     string  `json:"level"`
	CreatedAt string  `json:"created_at,omitempty"`
	UpdatedAt string  `json:"updated_at,omitempty"`
}

type PeriodRequest struct {
	Name      string `json:"name" validate:"required"`
	YearStart int32  `json:"year_start" validate:"required"`
	YearEnd   int32  `json:"year_end" validate:"required"`
}

type PeriodResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	YearStart int32  `json:"year_start"`
	YearEnd   int32  `json:"year_end"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type NationalPriorityRequest struct {
	PeriodID string `json:"period_id" validate:"required"`
	Title    string `json:"title" validate:"required"`
}

type NationalPriorityResponse struct {
	ID         string `json:"id"`
	PeriodID   string `json:"period_id"`
	PeriodName string `json:"period_name,omitempty"`
	Title      string `json:"title"`
	CreatedAt  string `json:"created_at,omitempty"`
	UpdatedAt  string `json:"updated_at,omitempty"`
}

type MasterImportResponse struct {
	FileName      string                    `json:"file_name"`
	TotalInserted int                       `json:"total_inserted"`
	TotalSkipped  int                       `json:"total_skipped"`
	TotalFailed   int                       `json:"total_failed"`
	Sheets        []MasterImportSheetResult `json:"sheets"`
}

type MasterImportSheetResult struct {
	Sheet    string                  `json:"sheet"`
	Inserted int                     `json:"inserted"`
	Skipped  int                     `json:"skipped"`
	Failed   int                     `json:"failed"`
	Rows     []MasterImportRowResult `json:"rows,omitempty"`
	Errors   []MasterImportRowError  `json:"errors,omitempty"`
}

type MasterImportRowResult struct {
	Row     int    `json:"row"`
	Status  string `json:"status"`
	Label   string `json:"label"`
	Message string `json:"message,omitempty"`
}

type MasterImportRowError struct {
	Row     int    `json:"row"`
	Message string `json:"message"`
}
