package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

type ProjectService struct {
	queries *queries.Queries
}

const (
	projectMasterExportBatchSize        = 500
	projectMasterExportFilterLabelLimit = 20
)

func NewProjectService(queries *queries.Queries) *ProjectService {
	return &ProjectService{queries: queries}
}

func (s *ProjectService) ListProjectMaster(ctx context.Context, filter model.ProjectMasterFilter, params model.PaginationParams) (*model.ProjectMasterListResponse, error) {
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
	summary, err := s.queries.GetProjectMasterFundingSummary(ctx, summaryProjectMasterParams(queryParams))
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung ringkasan pendanaan project")
	}
	data := make([]model.ProjectMasterResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, projectMasterResponse(row))
	}
	res := listResponse(data, page, limit, total)
	return &model.ProjectMasterListResponse{
		Data:    res.Data,
		Meta:    res.Meta,
		Summary: projectMasterFundingSummary(summary),
	}, nil
}

func (s *ProjectService) ExportProjectMaster(ctx context.Context, filter model.ProjectMasterFilter, params model.PaginationParams) (*importTemplateFile, error) {
	queryParams, err := buildProjectMasterParams(filter, params, projectMasterExportBatchSize, 0)
	if err != nil {
		return nil, err
	}

	total, err := s.queries.CountProjectMaster(ctx, countProjectMasterParams(queryParams))
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung export project")
	}

	summaryRow, err := s.queries.GetProjectMasterFundingSummary(ctx, summaryProjectMasterParams(queryParams))
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung ringkasan export project")
	}

	rows, err := s.exportProjectMasterRows(ctx, queryParams, total)
	if err != nil {
		return nil, err
	}

	exportedAt := time.Now()
	workbook := simpleXLSXWorkbook{
		Sheets: []simpleXLSXSheet{
			buildProjectMasterExportSheet(rows),
			buildProjectMasterExportSummarySheet(
				projectMasterFundingSummary(summaryRow),
				total,
				exportedAt,
				s.projectMasterExportFilters(ctx, queryParams),
			),
		},
	}
	data, err := buildSimpleXLSX(workbook)
	if err != nil {
		return nil, apperrors.Internal("Gagal membuat file export project")
	}

	return &importTemplateFile{
		FileName:    "projects_export_" + exportedAt.Format("20060102_150405") + ".xlsx",
		ContentType: importTemplateContentType,
		Data:        data,
	}, nil
}

func (s *ProjectService) exportProjectMasterRows(ctx context.Context, params queries.ListProjectMasterParams, total int64) ([]queries.ListProjectMasterRow, error) {
	rows := make([]queries.ListProjectMasterRow, 0)
	if total > 0 {
		rows = make([]queries.ListProjectMasterRow, 0, minInt64(total, projectMasterExportBatchSize))
	}

	for offset := int32(0); int64(offset) < total || total == 0; offset += projectMasterExportBatchSize {
		params.Offset = offset
		params.Limit = projectMasterExportBatchSize

		batch, err := s.queries.ListProjectMaster(ctx, params)
		if err != nil {
			return nil, apperrors.Internal("Gagal mengambil data export project")
		}
		if len(batch) == 0 {
			break
		}

		rows = append(rows, batch...)
		if total == 0 || int64(len(rows)) >= total {
			break
		}
	}

	return rows, nil
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

func summaryProjectMasterParams(params queries.ListProjectMasterParams) queries.GetProjectMasterFundingSummaryParams {
	return queries.GetProjectMasterFundingSummaryParams{
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

func projectMasterFundingSummary(row queries.GetProjectMasterFundingSummaryRow) model.ProjectMasterFundingSummary {
	return model.ProjectMasterFundingSummary{
		TotalLoanUSD:        floatFromNumeric(row.TotalLoanUsd),
		TotalGrantUSD:       floatFromNumeric(row.TotalGrantUsd),
		TotalCounterpartUSD: floatFromNumeric(row.TotalCounterpartUsd),
	}
}

func buildProjectMasterExportSheet(rows []queries.ListProjectMasterRow) simpleXLSXSheet {
	sheetRows := [][]simpleXLSXCell{
		headerRow(
			"No",
			"BB Code",
			"Nama Proyek",
			"Jenis Pinjaman",
			"Indikasi Lender",
			"Executing Agency",
			"Fixed Lender",
			"Status Project",
			"Status Pipeline",
			"Program Title",
			"Region/Location",
			"Nilai Pinjaman USD",
			"Tanggal Daftar Kegiatan",
			"Latest Snapshot",
			"Ada Revisi Lebih Baru",
			"Revisi Blue Book",
		),
	}

	for index, row := range rows {
		project := projectMasterResponse(row)
		sheetRows = append(sheetRows, []simpleXLSXCell{
			simpleIntegerCell(int64(index + 1)),
			textCell(project.BBCode),
			textCell(project.ProjectName),
			textCell(joinExportValues(project.LoanTypes)),
			textCell(joinExportValues(project.IndicationLenders)),
			textCell(joinExportValues(project.ExecutingAgencies)),
			textCell(joinExportValues(project.FixedLenders)),
			textCell(project.ProjectStatus),
			textCell(projectPipelineLabel(project.PipelineStatus)),
			textCell(project.ProgramTitle),
			textCell(joinExportValues(project.Locations)),
			floatCell(project.ForeignLoanUSD),
			textCell(joinExportValues(project.DKDates)),
			textCell(boolExportLabel(project.IsLatest)),
			textCell(boolExportLabel(project.HasNewerRevision)),
			textCell(project.BlueBookRevisionLabel),
		})
	}

	return simpleXLSXSheet{
		Name:          "Projects",
		Rows:          sheetRows,
		Columns:       columns(8, 18, 48, 18, 30, 34, 30, 18, 20, 34, 34, 20, 26, 18, 24, 30),
		AutoFilter:    "A1:P" + strconv.Itoa(len(sheetRows)),
		FreezeRows:    1,
		ShowGridLines: true,
	}
}

type projectMasterExportFilter struct {
	Label string
	Value string
}

func buildProjectMasterExportSummarySheet(summary model.ProjectMasterFundingSummary, total int64, exportedAt time.Time, filters []projectMasterExportFilter) simpleXLSXSheet {
	rows := [][]simpleXLSXCell{
		{styledTextCell("Ringkasan Export Project", xlsxStyleTitle)},
		{textCell("Tanggal Export"), textCell(exportedAt.Format("2006-01-02 15:04:05"))},
		{textCell("Total Project Sesuai Filter"), simpleIntegerCell(total)},
		{textCell("Total Pinjaman USD"), floatCell(summary.TotalLoanUSD)},
		{textCell("Total Hibah USD"), floatCell(summary.TotalGrantUSD)},
		{textCell("Total Dana Pendamping USD"), floatCell(summary.TotalCounterpartUSD)},
		{},
		{styledTextCell("Filter Aktif", xlsxStyleSection)},
	}

	if len(filters) == 0 {
		rows = append(rows, []simpleXLSXCell{textCell("Semua data"), textCell("Tidak ada filter aktif")})
	} else {
		rows = append(rows, headerRow("Filter", "Nilai"))
		for _, filter := range filters {
			rows = append(rows, []simpleXLSXCell{textCell(filter.Label), textCell(filter.Value)})
		}
	}

	return simpleXLSXSheet{
		Name:          "Ringkasan",
		Rows:          rows,
		Columns:       columns(32, 80),
		ShowGridLines: true,
	}
}

func (s *ProjectService) projectMasterExportFilters(ctx context.Context, params queries.ListProjectMasterParams) []projectMasterExportFilter {
	filters := []projectMasterExportFilter{}

	addFilter := func(label, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		filters = append(filters, projectMasterExportFilter{Label: label, Value: value})
	}

	if params.Search.Valid {
		addFilter("Pencarian", params.Search.String)
	}
	addFilter("Jenis Pinjaman", joinExportFilterValues(params.LoanTypes))
	addFilter("Indikasi Lender", joinExportValues(s.projectLenderLabels(ctx, params.IndicationLenderIds)))
	addFilter("Executing Agency", joinExportValues(s.projectInstitutionLabels(ctx, params.ExecutingAgencyIds)))
	addFilter("Fixed Lender", joinExportValues(s.projectLenderLabels(ctx, params.FixedLenderIds)))
	addFilter("Status Project", joinExportFilterValues(params.ProjectStatuses))
	addFilter("Status Pipeline", joinExportFilterValues(projectPipelineLabels(params.PipelineStatuses)))
	addFilter("Program Title", joinExportValues(s.projectProgramTitleLabels(ctx, params.ProgramTitleIds)))
	addFilter("Region/Location", joinExportValues(s.projectRegionLabels(ctx, params.RegionIds)))

	if params.ForeignLoanMin.Valid && params.ForeignLoanMax.Valid {
		addFilter("Nilai Pinjaman USD", exportNumericLabel(params.ForeignLoanMin)+" - "+exportNumericLabel(params.ForeignLoanMax))
	} else if params.ForeignLoanMin.Valid {
		addFilter("Nilai Pinjaman USD Minimum", exportNumericLabel(params.ForeignLoanMin))
	} else if params.ForeignLoanMax.Valid {
		addFilter("Nilai Pinjaman USD Maksimum", exportNumericLabel(params.ForeignLoanMax))
	}

	if params.DkDateFrom.Valid && params.DkDateTo.Valid {
		addFilter("Tanggal Daftar Kegiatan", exportDateLabel(params.DkDateFrom)+" - "+exportDateLabel(params.DkDateTo))
	} else if params.DkDateFrom.Valid {
		addFilter("Tanggal Daftar Kegiatan Dari", exportDateLabel(params.DkDateFrom))
	} else if params.DkDateTo.Valid {
		addFilter("Tanggal Daftar Kegiatan Sampai", exportDateLabel(params.DkDateTo))
	}

	if params.IncludeHistory {
		addFilter("Snapshot historis", "Ditampilkan")
	}

	return filters
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

func int32Array(values []string, field string) ([]int32, error) {
	items := make([]int32, 0, len(values))
	seen := map[int32]struct{}{}
	for _, value := range values {
		raw := strings.TrimSpace(value)
		if raw == "" {
			continue
		}
		parsed, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			return nil, validation(field, "harus angka")
		}
		item := int32(parsed)
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		items = append(items, item)
	}
	return items, nil
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

func optionalInt4(value *string, field string) (pgtype.Int4, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return pgtype.Int4{}, nil
	}
	parsed, err := strconv.ParseInt(strings.TrimSpace(*value), 10, 32)
	if err != nil {
		return pgtype.Int4{}, validation(field, "harus angka")
	}
	return pgtype.Int4{Int32: int32(parsed), Valid: true}, nil
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

func joinExportValues(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return strings.Join(values, ", ")
}

func joinExportFilterValues(values []string) string {
	if len(values) == 0 {
		return ""
	}
	if len(values) <= projectMasterExportFilterLabelLimit {
		return strings.Join(values, ", ")
	}
	visible := append([]string{}, values[:projectMasterExportFilterLabelLimit]...)
	visible = append(visible, "+"+strconv.Itoa(len(values)-projectMasterExportFilterLabelLimit)+" lainnya")
	return strings.Join(visible, ", ")
}

func boolExportLabel(value bool) string {
	if value {
		return "Ya"
	}
	return "Tidak"
}

func projectPipelineLabel(status string) string {
	switch status {
	case "BB":
		return "Blue Book"
	case "GB":
		return "Green Book"
	case "DK":
		return "Daftar Kegiatan"
	case "LA":
		return "Loan Agreement"
	case "Monitoring":
		return "Monitoring"
	default:
		return status
	}
}

func projectPipelineLabels(statuses []string) []string {
	labels := make([]string, 0, len(statuses))
	for _, status := range statuses {
		labels = append(labels, projectPipelineLabel(status))
	}
	return labels
}

func (s *ProjectService) projectLenderLabels(ctx context.Context, ids []pgtype.UUID) []string {
	labels := make([]string, 0, minInt(len(ids), projectMasterExportFilterLabelLimit))
	for _, id := range ids[:minInt(len(ids), projectMasterExportFilterLabelLimit)] {
		row, err := s.queries.GetLender(ctx, id)
		if err != nil {
			labels = append(labels, model.UUIDToString(id))
			continue
		}
		labels = append(labels, labelWithShortName(row.Name, row.ShortName))
	}
	labels = appendRemainingLabel(labels, len(ids))
	return labels
}

func (s *ProjectService) projectInstitutionLabels(ctx context.Context, ids []pgtype.UUID) []string {
	labels := make([]string, 0, minInt(len(ids), projectMasterExportFilterLabelLimit))
	for _, id := range ids[:minInt(len(ids), projectMasterExportFilterLabelLimit)] {
		row, err := s.queries.GetInstitution(ctx, id)
		if err != nil {
			labels = append(labels, model.UUIDToString(id))
			continue
		}
		labels = append(labels, labelWithShortName(row.Name, row.ShortName))
	}
	labels = appendRemainingLabel(labels, len(ids))
	return labels
}

func (s *ProjectService) projectProgramTitleLabels(ctx context.Context, ids []pgtype.UUID) []string {
	labels := make([]string, 0, minInt(len(ids), projectMasterExportFilterLabelLimit))
	for _, id := range ids[:minInt(len(ids), projectMasterExportFilterLabelLimit)] {
		row, err := s.queries.GetProgramTitle(ctx, id)
		if err != nil {
			labels = append(labels, model.UUIDToString(id))
			continue
		}
		labels = append(labels, row.Title)
	}
	labels = appendRemainingLabel(labels, len(ids))
	return labels
}

func (s *ProjectService) projectRegionLabels(ctx context.Context, ids []pgtype.UUID) []string {
	labels := make([]string, 0, minInt(len(ids), projectMasterExportFilterLabelLimit))
	for _, id := range ids[:minInt(len(ids), projectMasterExportFilterLabelLimit)] {
		row, err := s.queries.GetRegion(ctx, id)
		if err != nil {
			labels = append(labels, model.UUIDToString(id))
			continue
		}
		labels = append(labels, row.Name)
	}
	labels = appendRemainingLabel(labels, len(ids))
	return labels
}

func appendRemainingLabel(labels []string, total int) []string {
	if total > projectMasterExportFilterLabelLimit {
		return append(labels, "+"+strconv.Itoa(total-projectMasterExportFilterLabelLimit)+" lainnya")
	}
	return labels
}

func labelWithShortName(name string, shortName pgtype.Text) string {
	if shortName.Valid && strings.TrimSpace(shortName.String) != "" {
		return name + " (" + shortName.String + ")"
	}
	return name
}

func exportNumericLabel(value pgtype.Numeric) string {
	return strconv.FormatFloat(floatFromNumeric(value), 'f', -1, 64)
}

func exportDateLabel(value pgtype.Date) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format("2006-01-02")
}

func simpleIntegerCell(value int64) simpleXLSXCell {
	return simpleXLSXCell{Value: strconv.FormatInt(value, 10), Number: true}
}

func minInt64(a int64, b int32) int {
	if a < int64(b) {
		return int(a)
	}
	return int(b)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
