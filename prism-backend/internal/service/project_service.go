package service

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

type ProjectService struct {
	queries *queries.Queries
}

func NewProjectService(queries *queries.Queries) *ProjectService {
	return &ProjectService{queries: queries}
}

func (s *ProjectService) ListProjectMaster(ctx context.Context, filter model.ProjectMasterFilter, params model.PaginationParams) (*model.ListResponse[model.ProjectMasterResponse], error) {
	page, limit, offset := normalizeList(params)
	queryParams, err := buildProjectMasterParams(filter, params, limit, offset)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListProjectMaster(ctx, queryParams)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil master table project")
	}
	total, err := s.queries.CountProjectMaster(ctx, countProjectMasterParams(queryParams))
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung master table project")
	}
	data := make([]model.ProjectMasterResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, projectMasterResponse(row))
	}
	return listResponse(data, page, limit, total), nil
}

func buildProjectMasterParams(filter model.ProjectMasterFilter, params model.PaginationParams, limit, offset int) (queries.ListProjectMasterParams, error) {
	loanTypes, err := allowedValues(filter.LoanTypes, map[string]struct{}{"Bilateral": {}, "Multilateral": {}, "KSA": {}}, "loan_types")
	if err != nil {
		return queries.ListProjectMasterParams{}, err
	}
	projectStatuses, err := allowedValues(filter.ProjectStatuses, map[string]struct{}{"Pipeline": {}, "Ongoing": {}}, "project_statuses")
	if err != nil {
		return queries.ListProjectMasterParams{}, err
	}
	pipelineStatuses, err := allowedValues(filter.PipelineStatuses, map[string]struct{}{"BB": {}, "GB": {}, "DK": {}, "LA": {}, "Monitoring": {}}, "pipeline_statuses")
	if err != nil {
		return queries.ListProjectMasterParams{}, err
	}
	indicationLenderIDs, err := uuidArray(filter.IndicationLenderIDs, "indication_lender_ids")
	if err != nil {
		return queries.ListProjectMasterParams{}, err
	}
	executingAgencyIDs, err := uuidArray(filter.ExecutingAgencyIDs, "executing_agency_ids")
	if err != nil {
		return queries.ListProjectMasterParams{}, err
	}
	fixedLenderIDs, err := uuidArray(filter.FixedLenderIDs, "fixed_lender_ids")
	if err != nil {
		return queries.ListProjectMasterParams{}, err
	}
	programTitleIDs, err := uuidArray(filter.ProgramTitleIDs, "program_title_ids")
	if err != nil {
		return queries.ListProjectMasterParams{}, err
	}
	regionIDs, err := uuidArray(filter.RegionIDs, "region_ids")
	if err != nil {
		return queries.ListProjectMasterParams{}, err
	}
	foreignLoanMin, err := optionalNumeric(filter.ForeignLoanMin, "foreign_loan_min")
	if err != nil {
		return queries.ListProjectMasterParams{}, err
	}
	foreignLoanMax, err := optionalNumeric(filter.ForeignLoanMax, "foreign_loan_max")
	if err != nil {
		return queries.ListProjectMasterParams{}, err
	}
	dkDateFrom, err := optionalDate(filter.DKDateFrom, "dk_date_from")
	if err != nil {
		return queries.ListProjectMasterParams{}, err
	}
	dkDateTo, err := optionalDate(filter.DKDateTo, "dk_date_to")
	if err != nil {
		return queries.ListProjectMasterParams{}, err
	}
	if foreignLoanMin.Valid && foreignLoanMax.Valid && floatFromNumeric(foreignLoanMin) > floatFromNumeric(foreignLoanMax) {
		return queries.ListProjectMasterParams{}, validation("foreign_loan_max", "harus lebih besar dari nilai minimum")
	}
	if dkDateFrom.Valid && dkDateTo.Valid && dkDateFrom.Time.After(dkDateTo.Time) {
		return queries.ListProjectMasterParams{}, validation("dk_date_to", "harus setelah tanggal mulai")
	}
	sortField, sortOrder, err := normalizeProjectMasterSort(params.Sort, params.Order)
	if err != nil {
		return queries.ListProjectMasterParams{}, err
	}

	return queries.ListProjectMasterParams{
		Sort:                sortField,
		Order:               sortOrder,
		Offset:              int32(offset),
		Limit:               int32(limit),
		IncludeHistory:      filter.IncludeHistory,
		LoanTypes:           loanTypes,
		IndicationLenderIds: indicationLenderIDs,
		ExecutingAgencyIds:  executingAgencyIDs,
		FixedLenderIds:      fixedLenderIDs,
		ProjectStatuses:     projectStatuses,
		PipelineStatuses:    pipelineStatuses,
		ProgramTitleIds:     programTitleIDs,
		RegionIds:           regionIDs,
		ForeignLoanMin:      foreignLoanMin,
		ForeignLoanMax:      foreignLoanMax,
		DkDateFrom:          dkDateFrom,
		DkDateTo:            dkDateTo,
		Search:              optionalText(filter.Search),
	}, nil
}

func countProjectMasterParams(params queries.ListProjectMasterParams) queries.CountProjectMasterParams {
	return queries.CountProjectMasterParams{
		LoanTypes:           params.LoanTypes,
		IndicationLenderIds: params.IndicationLenderIds,
		ExecutingAgencyIds:  params.ExecutingAgencyIds,
		FixedLenderIds:      params.FixedLenderIds,
		ProjectStatuses:     params.ProjectStatuses,
		PipelineStatuses:    params.PipelineStatuses,
		ProgramTitleIds:     params.ProgramTitleIds,
		RegionIds:           params.RegionIds,
		ForeignLoanMin:      params.ForeignLoanMin,
		ForeignLoanMax:      params.ForeignLoanMax,
		DkDateFrom:          params.DkDateFrom,
		DkDateTo:            params.DkDateTo,
		Search:              params.Search,
		IncludeHistory:      params.IncludeHistory,
	}
}

func normalizeProjectMasterSort(sortField, sortOrder string) (string, string, error) {
	allowedSorts := map[string]struct{}{
		"bb_code":            {},
		"project_name":       {},
		"loan_types":         {},
		"indication_lenders": {},
		"executing_agencies": {},
		"fixed_lenders":      {},
		"project_status":     {},
		"pipeline_status":    {},
		"program_title":      {},
		"locations":          {},
		"foreign_loan_usd":   {},
		"dk_dates":           {},
	}

	sortField = strings.TrimSpace(sortField)
	if sortField == "" {
		sortField = "project_name"
	}
	if _, ok := allowedSorts[sortField]; !ok {
		return "", "", validation("sort", "nilai tidak valid")
	}

	sortOrder = strings.ToLower(strings.TrimSpace(sortOrder))
	if sortOrder == "" {
		sortOrder = "asc"
	}
	if sortOrder != "asc" && sortOrder != "desc" {
		return "", "", validation("order", "harus asc atau desc")
	}

	return sortField, sortOrder, nil
}

func projectMasterResponse(row queries.ListProjectMasterRow) model.ProjectMasterResponse {
	return model.ProjectMasterResponse{
		ID:                    model.UUIDToString(row.ID),
		BlueBookID:            model.UUIDToString(row.BlueBookID),
		ProjectIdentityID:     model.UUIDToString(row.ProjectIdentityID),
		BBCode:                row.BbCode,
		ProjectName:           row.ProjectName,
		LoanTypes:             safeStringSlice(row.LoanTypes),
		IndicationLenders:     safeStringSlice(row.IndicationLenders),
		ExecutingAgencies:     safeStringSlice(row.ExecutingAgencies),
		FixedLenders:          safeStringSlice(row.FixedLenders),
		ProjectStatus:         row.ProjectStatus,
		PipelineStatus:        row.PipelineStatus,
		ProgramTitle:          row.ProgramTitle,
		Locations:             safeStringSlice(row.Locations),
		ForeignLoanUSD:        floatFromNumeric(row.ForeignLoanUsd),
		DKDates:               safeStringSlice(row.DkDates),
		IsLatest:              row.IsLatest,
		HasNewerRevision:      row.HasNewerRevision,
		BlueBookRevisionLabel: row.BlueBookRevisionLabel,
	}
}

func allowedValues(values []string, allowed map[string]struct{}, field string) ([]string, error) {
	normalized := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item == "" {
			continue
		}
		if _, ok := allowed[item]; !ok {
			return nil, validation(field, "nilai tidak valid")
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		normalized = append(normalized, item)
	}
	return normalized, nil
}

func uuidArray(values []string, field string) ([]pgtype.UUID, error) {
	ids := make([]pgtype.UUID, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		id, err := model.ParseUUID(item)
		if err != nil {
			return nil, validation(field, "UUID tidak valid")
		}
		seen[item] = struct{}{}
		ids = append(ids, id)
	}
	return ids, nil
}

func optionalNumeric(value *string, field string) (pgtype.Numeric, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return pgtype.Numeric{}, nil
	}
	var numeric pgtype.Numeric
	if err := numeric.Scan(strings.TrimSpace(*value)); err != nil {
		return pgtype.Numeric{}, validation(field, "harus angka")
	}
	return numeric, nil
}

func optionalDate(value *string, field string) (pgtype.Date, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return pgtype.Date{}, nil
	}
	return parseDate(*value, field)
}

func optionalText(value *string) pgtype.Text {
	if value == nil || strings.TrimSpace(*value) == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: strings.TrimSpace(*value), Valid: true}
}

func safeStringSlice(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
