package model

type SpatialDistributionFilter struct {
	Level            string
	ProvinceCode     *string
	LoanTypes        []string
	ProjectStatuses  []string
	PipelineStatuses []string
	Search           *string
	IncludeHistory   bool
}

type SpatialDistributionProjectFilter struct {
	SpatialDistributionFilter
	RegionCode string
}

type SpatialDistributionRegionMetric struct {
	RegionID     string  `json:"region_id"`
	RegionCode   string  `json:"region_code"`
	RegionName   string  `json:"region_name"`
	RegionType   string  `json:"region_type"`
	ParentCode   *string `json:"parent_code,omitempty"`
	ProjectCount int64   `json:"project_count"`
	TotalLoanUSD float64 `json:"total_loan_usd"`
}

type SpatialDistributionSummary struct {
	TotalRegions      int     `json:"total_regions"`
	ActiveRegions     int     `json:"active_regions"`
	TotalProjectCount int64   `json:"total_project_count"`
	TotalLoanUSD      float64 `json:"total_loan_usd"`
	MaxProjectCount   int64   `json:"max_project_count"`
	MaxLoanUSD        float64 `json:"max_loan_usd"`
}

type SpatialDistributionChoroplethResponse struct {
	Level        string                            `json:"level"`
	ProvinceCode *string                           `json:"province_code,omitempty"`
	Regions      []SpatialDistributionRegionMetric `json:"regions"`
	Summary      SpatialDistributionSummary        `json:"summary"`
}

type SpatialDistributionProjectListResponse struct {
	Level      string                      `json:"level"`
	RegionID   string                      `json:"region_id"`
	RegionCode string                      `json:"region_code"`
	RegionName string                      `json:"region_name"`
	RegionType string                      `json:"region_type"`
	ParentCode *string                     `json:"parent_code,omitempty"`
	Data       []ProjectMasterResponse     `json:"data"`
	Meta       PaginationMeta              `json:"meta"`
	Summary    ProjectMasterFundingSummary `json:"summary"`
}
