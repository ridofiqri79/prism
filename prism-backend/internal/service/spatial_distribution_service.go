package service

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

const (
	spatialLevelProvince = "province"
	spatialLevelCity     = "city"
)

type SpatialDistributionService struct {
	queries        *queries.Queries
	projectService *ProjectService
}

func NewSpatialDistributionService(queries *queries.Queries, projectService *ProjectService) *SpatialDistributionService {
	return &SpatialDistributionService{
		queries:        queries,
		projectService: projectService,
	}
}

func (s *SpatialDistributionService) Choropleth(ctx context.Context, filter model.SpatialDistributionFilter) (*model.SpatialDistributionChoroplethResponse, error) {
	level, provinceCode, err := s.normalizeLevel(ctx, filter.Level, filter.ProvinceCode)
	if err != nil {
		return nil, err
	}

	loanTypes, projectStatuses, pipelineStatuses, err := normalizeSpatialProjectFilters(filter)
	if err != nil {
		return nil, err
	}

	params := queries.ListSpatialRegionMetricsParams{
		Level:            level,
		ProvinceCode:     provinceCode,
		IncludeHistory:   filter.IncludeHistory,
		LoanTypes:        loanTypes,
		ProjectStatuses:  projectStatuses,
		PipelineStatuses: pipelineStatuses,
		Search:           optionalText(filter.Search),
	}

	rows, err := s.queries.ListSpatialRegionMetrics(ctx, params)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil data peta wilayah")
	}

	summaryRow, err := s.queries.GetSpatialRegionSummary(ctx, queries.GetSpatialRegionSummaryParams(params))
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung ringkasan peta wilayah")
	}

	regions := make([]model.SpatialDistributionRegionMetric, 0, len(rows))
	summary := model.SpatialDistributionSummary{
		TotalRegions:      len(rows),
		TotalProjectCount: summaryRow.ProjectCount,
		TotalLoanUSD:      floatFromNumeric(summaryRow.TotalLoanUsd),
	}

	for _, row := range rows {
		item := spatialRegionMetric(row)
		regions = append(regions, item)
		if item.ProjectCount > 0 {
			summary.ActiveRegions++
		}
		if item.ProjectCount > summary.MaxProjectCount {
			summary.MaxProjectCount = item.ProjectCount
		}
		if item.TotalLoanUSD > summary.MaxLoanUSD {
			summary.MaxLoanUSD = item.TotalLoanUSD
		}
	}

	return &model.SpatialDistributionChoroplethResponse{
		Level:        level,
		ProvinceCode: stringPtrFromText(provinceCode),
		Regions:      regions,
		Summary:      summary,
	}, nil
}

func (s *SpatialDistributionService) RegionProjects(ctx context.Context, filter model.SpatialDistributionProjectFilter, params model.PaginationParams) (*model.SpatialDistributionProjectListResponse, error) {
	level, normalizedProvinceCode, err := s.normalizeLevel(ctx, filter.Level, filter.ProvinceCode)
	if err != nil {
		return nil, err
	}

	region, err := s.queries.GetSpatialRegionByCode(ctx, strings.TrimSpace(filter.RegionCode))
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, validation("region_code", "wilayah tidak ditemukan")
		}
		return nil, apperrors.Internal("Gagal membaca wilayah")
	}
	if level == spatialLevelProvince && region.Type != "PROVINCE" && region.Type != "COUNTRY" {
		return nil, validation("region_code", "harus berisi kode provinsi atau nasional")
	}
	if level == spatialLevelCity && region.Type != "CITY" {
		return nil, validation("region_code", "harus berisi kode kabupaten/kota")
	}
	if level == spatialLevelCity && (!region.ParentCode.Valid || region.ParentCode.String != normalizedProvinceCode.String) {
		return nil, validation("region_code", "harus berada di province_code")
	}

	filterIDs, err := s.queries.ListSpatialRegionFilterIDs(ctx, queries.ListSpatialRegionFilterIDsParams{
		Level:      level,
		RegionCode: region.Code,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal menyiapkan filter wilayah")
	}
	if len(filterIDs) == 0 {
		return nil, validation("region_code", "tidak memiliki cakupan wilayah")
	}

	regionIDFilters := make([]string, 0, len(filterIDs))
	for _, id := range filterIDs {
		regionIDFilters = append(regionIDFilters, model.UUIDToString(id))
	}

	loanTypes, projectStatuses, pipelineStatuses, err := normalizeSpatialProjectFilters(filter.SpatialDistributionFilter)
	if err != nil {
		return nil, err
	}

	projectList, err := s.projectService.ListProjectMaster(ctx, model.ProjectMasterFilter{
		LoanTypes:        loanTypes,
		ProjectStatuses:  projectStatuses,
		PipelineStatuses: pipelineStatuses,
		RegionIDs:        regionIDFilters,
		Search:           filter.Search,
		IncludeHistory:   filter.IncludeHistory,
	}, params)
	if err != nil {
		return nil, err
	}

	return &model.SpatialDistributionProjectListResponse{
		Level:      level,
		RegionID:   model.UUIDToString(region.ID),
		RegionCode: region.Code,
		RegionName: region.Name,
		RegionType: region.Type,
		ParentCode: stringPtrFromText(region.ParentCode),
		Data:       projectList.Data,
		Meta:       projectList.Meta,
		Summary:    projectList.Summary,
	}, nil
}

func (s *SpatialDistributionService) normalizeLevel(ctx context.Context, rawLevel string, rawProvinceCode *string) (string, pgtype.Text, error) {
	level := strings.ToLower(strings.TrimSpace(rawLevel))
	if level == "" {
		level = spatialLevelProvince
	}
	if level != spatialLevelProvince && level != spatialLevelCity {
		return "", pgtype.Text{}, validation("level", "harus province atau city")
	}

	if level == spatialLevelProvince {
		return level, pgtype.Text{}, nil
	}

	if rawProvinceCode == nil || strings.TrimSpace(*rawProvinceCode) == "" {
		return "", pgtype.Text{}, validation("province_code", "wajib diisi untuk level city")
	}

	provinceCode := strings.TrimSpace(*rawProvinceCode)
	region, err := s.queries.GetSpatialRegionByCode(ctx, provinceCode)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", pgtype.Text{}, validation("province_code", "provinsi tidak ditemukan")
		}
		return "", pgtype.Text{}, apperrors.Internal("Gagal membaca provinsi")
	}
	if region.Type != "PROVINCE" {
		return "", pgtype.Text{}, validation("province_code", "harus berisi kode provinsi")
	}

	return level, pgtype.Text{String: region.Code, Valid: true}, nil
}

func normalizeSpatialProjectFilters(filter model.SpatialDistributionFilter) ([]string, []string, []string, error) {
	loanTypes, err := allowedValues(filter.LoanTypes, map[string]struct{}{"Bilateral": {}, "Multilateral": {}, "KSA": {}}, "loan_types")
	if err != nil {
		return nil, nil, nil, err
	}
	projectStatuses, err := allowedValues(filter.ProjectStatuses, map[string]struct{}{"Pipeline": {}, "Ongoing": {}}, "project_statuses")
	if err != nil {
		return nil, nil, nil, err
	}
	pipelineStatuses, err := allowedValues(filter.PipelineStatuses, map[string]struct{}{"BB": {}, "GB": {}, "DK": {}, "LA": {}, "Monitoring": {}}, "pipeline_statuses")
	if err != nil {
		return nil, nil, nil, err
	}

	return loanTypes, projectStatuses, pipelineStatuses, nil
}

func spatialRegionMetric(row queries.ListSpatialRegionMetricsRow) model.SpatialDistributionRegionMetric {
	return model.SpatialDistributionRegionMetric{
		RegionID:     model.UUIDToString(row.ID),
		RegionCode:   row.Code,
		RegionName:   row.Name,
		RegionType:   row.Type,
		ParentCode:   stringPtrFromText(row.ParentCode),
		ProjectCount: row.ProjectCount,
		TotalLoanUSD: floatFromNumeric(row.TotalLoanUsd),
	}
}
