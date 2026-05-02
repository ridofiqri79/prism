package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

type DashboardService struct {
	queries *queries.Queries
}

func NewDashboardService(queries *queries.Queries) *DashboardService {
	return &DashboardService{queries: queries}
}

func (s *DashboardService) GetSummary(ctx context.Context, filter model.DashboardFilterRequest) (*model.DashboardSummary, error) {
	params, err := dashboardSummaryParams(filter)
	if err != nil {
		return nil, err
	}
	row, err := s.queries.GetDashboardSummary(ctx, params)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil ringkasan dashboard")
	}
	summary := &model.DashboardSummary{
		TotalBBProjects:         int(row.TotalBbProjects),
		TotalGBProjects:         int(row.TotalGbProjects),
		TotalLoanAgreements:     int(row.TotalLoanAgreements),
		BBPipelineUSD:           floatFromNumeric(row.BbPipelineUsd),
		GBPipelineUSD:           floatFromNumeric(row.GbPipelineUsd),
		GBLocalUSD:              floatFromNumeric(row.GbLocalUsd),
		DKFinancingUSD:          floatFromNumeric(row.DkFinancingUsd),
		DKCounterpartUSD:        floatFromNumeric(row.DkCounterpartUsd),
		LACommitmentUSD:         floatFromNumeric(row.LaCommitmentUsd),
		PlannedDisbursementUSD:  floatFromNumeric(row.PlannedDisbursementUsd),
		RealizedDisbursementUSD: floatFromNumeric(row.RealizedDisbursementUsd),
		AbsorptionPct:           floatFromNumeric(row.AbsorptionPct),
		LAAbsorptionPct:         floatFromNumeric(row.LaAbsorptionPct),
		UndisbursedUSD:          floatFromNumeric(row.UndisbursedUsd),
	}
	summary.Metrics = dashboardMetricCards(summary)
	return summary, nil
}

func (s *DashboardService) GetStageFunnel(ctx context.Context, filter model.DashboardFilterRequest) ([]model.StageMetric, error) {
	periodID, err := optionalUUID(filter.PeriodID)
	if err != nil {
		return nil, validation("period_id", "UUID tidak valid")
	}
	countRows, err := s.queries.GetDashboardStageCounts(ctx, queries.GetDashboardStageCountsParams{
		PeriodID:       periodID,
		IncludeHistory: filter.IncludeHistory,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil jumlah tahap dashboard")
	}
	amountRows, err := s.queries.GetDashboardStageAmounts(ctx, queries.GetDashboardStageAmountsParams{
		PeriodID:       periodID,
		IncludeHistory: filter.IncludeHistory,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil nilai tahap dashboard")
	}

	amounts := make(map[string]float64, len(amountRows))
	for _, row := range amountRows {
		amounts[row.Stage] = floatFromNumeric(row.AmountUsd)
	}

	stages := make([]model.StageMetric, 0, len(countRows))
	for _, row := range countRows {
		stages = append(stages, model.StageMetric{
			Stage:        row.Stage,
			Label:        stageLabel(row.Stage),
			ProjectCount: int(row.ProjectCount),
			AmountUSD:    amounts[row.Stage],
		})
	}
	return stages, nil
}

func (s *DashboardService) GetMonitoringRollup(ctx context.Context, filter model.DashboardFilterRequest) ([]model.TimeSeriesPoint, error) {
	params, err := dashboardRollupParams(filter)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.GetDashboardMonitoringRollup(ctx, params)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil rollup monitoring")
	}
	points := make([]model.TimeSeriesPoint, 0, len(rows))
	for _, row := range rows {
		points = append(points, model.TimeSeriesPoint{
			Period:        fmt.Sprintf("%d-%s", row.BudgetYear, row.Quarter),
			BudgetYear:    row.BudgetYear,
			Quarter:       row.Quarter,
			PlannedUSD:    floatFromNumeric(row.PlannedUsd),
			RealizedUSD:   floatFromNumeric(row.RealizedUsd),
			AbsorptionPct: floatFromNumeric(row.AbsorptionPct),
		})
	}
	return points, nil
}

func (s *DashboardService) GetLAExposureRollup(ctx context.Context, filter model.DashboardFilterRequest) (*model.DashboardLAExposure, error) {
	params, err := dashboardLAExposureParams(filter)
	if err != nil {
		return nil, err
	}
	row, err := s.queries.GetDashboardLAExposureRollup(ctx, params)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil exposure loan agreement")
	}
	return &model.DashboardLAExposure{
		LACommitmentUSD:         floatFromNumeric(row.LaCommitmentUsd),
		RealizedDisbursementUSD: floatFromNumeric(row.RealizedDisbursementUsd),
		UndisbursedUSD:          floatFromNumeric(row.UndisbursedUsd),
		LAAbsorptionPct:         floatFromNumeric(row.LaAbsorptionPct),
	}, nil
}

func (s *DashboardService) GetLenderRollup(ctx context.Context, filter model.DashboardFilterRequest) ([]model.BreakdownItem, error) {
	params, err := dashboardLenderRollupParams(filter)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.GetDashboardLenderRollup(ctx, params)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil rollup lender")
	}
	items := make([]model.BreakdownItem, 0, len(rows))
	for _, row := range rows {
		id := model.UUIDToString(row.ID)
		items = append(items, model.BreakdownItem{
			ID:          &id,
			Label:       row.Label,
			ItemCount:   int(row.ItemCount),
			AmountUSD:   floatFromNumeric(row.AmountUsd),
			RealizedUSD: floatFromNumeric(row.RealizedUsd),
		})
	}
	return items, nil
}

func (s *DashboardService) GetInstitutionRollup(ctx context.Context, filter model.DashboardFilterRequest) ([]model.BreakdownItem, error) {
	params, err := dashboardInstitutionRollupParams(filter)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.GetDashboardInstitutionRollup(ctx, params)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil rollup institution")
	}
	items := make([]model.BreakdownItem, 0, len(rows))
	for _, row := range rows {
		id := model.UUIDToString(row.ID)
		items = append(items, model.BreakdownItem{
			ID:          &id,
			Label:       row.Label,
			ItemCount:   int(row.ItemCount),
			AmountUSD:   floatFromNumeric(row.AmountUsd),
			RealizedUSD: floatFromNumeric(row.RealizedUsd),
		})
	}
	return items, nil
}

func (s *DashboardService) GetFilterOptions(ctx context.Context) (model.DashboardFilterOptions, error) {
	rows, err := s.queries.ListDashboardFilterOptions(ctx)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil opsi filter dashboard")
	}
	options := model.DashboardFilterOptions{}
	for _, row := range rows {
		options[row.OptionType] = append(options[row.OptionType], model.BreakdownItem{
			Key:   row.Value,
			Label: row.Label,
		})
	}
	return options, nil
}

func (s *DashboardService) GetExecutivePortfolio(ctx context.Context, filter model.DashboardFilterRequest) (*model.ExecutivePortfolioDashboard, error) {
	summary, err := s.GetSummary(ctx, filter)
	if err != nil {
		return nil, err
	}

	funnelParams, err := dashboardExecutiveFunnelParams(filter)
	if err != nil {
		return nil, err
	}
	funnelRows, err := s.queries.GetDashboardExecutiveFunnel(ctx, funnelParams)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil funnel executive portfolio")
	}

	rollupParams, err := dashboardExecutiveRollupParams(filter)
	if err != nil {
		return nil, err
	}
	institutionRows, err := s.queries.GetDashboardExecutiveTopInstitutions(ctx, queries.GetDashboardExecutiveTopInstitutionsParams(rollupParams))
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil top institution executive portfolio")
	}
	lenderRows, err := s.queries.GetDashboardExecutiveTopLenders(ctx, queries.GetDashboardExecutiveTopLendersParams(rollupParams))
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil top lender executive portfolio")
	}

	riskRows, err := s.queries.ListDashboardExecutiveRiskItems(ctx, queries.ListDashboardExecutiveRiskItemsParams(funnelParams))
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil risk item executive portfolio")
	}

	funnel := executiveFunnelMetrics(funnelRows)
	risks := executiveRiskItems(riskRows)
	return &model.ExecutivePortfolioDashboard{
		Cards:           executiveMetricCards(summary),
		Funnel:          funnel,
		TopInstitutions: executiveInstitutionBreakdown(institutionRows),
		TopLenders:      executiveLenderBreakdown(lenderRows),
		RiskItems:       risks,
		Insights:        executiveInsights(summary, funnel, risks),
	}, nil
}

func (s *DashboardService) GetPipelineBottleneck(ctx context.Context, filter model.PipelineBottleneckFilterRequest, pagination model.PaginationParams) (*model.DataMetaResponse[model.PipelineBottleneckDashboard], error) {
	page, limit, offset := normalizeList(pagination)
	params, err := dashboardPipelineListParams(filter, pagination, limit, offset)
	if err != nil {
		return nil, err
	}
	summaryParams := queries.GetDashboardPipelineBottleneckStageSummaryParams{
		PeriodID:      params.PeriodID,
		PublishYear:   params.PublishYear,
		LenderID:      params.LenderID,
		InstitutionID: params.InstitutionID,
		Stage:         params.Stage,
		MinAgeDays:    params.MinAgeDays,
		Search:        params.Search,
	}
	stageRows, err := s.queries.GetDashboardPipelineBottleneckStageSummary(ctx, summaryParams)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil ringkasan pipeline bottleneck")
	}
	countParams := queries.CountDashboardPipelineBottleneckItemsParams(summaryParams)
	total, err := s.queries.CountDashboardPipelineBottleneckItems(ctx, countParams)
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung pipeline bottleneck")
	}
	itemRows, err := s.queries.ListDashboardPipelineBottleneckItems(ctx, params)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar pipeline bottleneck")
	}

	return &model.DataMetaResponse[model.PipelineBottleneckDashboard]{
		Data: model.PipelineBottleneckDashboard{
			StageSummary: pipelineStageSummary(stageRows),
			Items:        pipelineBottleneckItems(itemRows),
		},
		Meta: model.PaginationMeta{
			Page:       page,
			Limit:      limit,
			Total:      int(total),
			TotalPages: totalPages(total, limit),
		},
	}, nil
}

func dashboardSummaryParams(filter model.DashboardFilterRequest) (queries.GetDashboardSummaryParams, error) {
	periodID, err := optionalUUID(filter.PeriodID)
	if err != nil {
		return queries.GetDashboardSummaryParams{}, validation("period_id", "UUID tidak valid")
	}
	lenderID, err := optionalUUID(filter.LenderID)
	if err != nil {
		return queries.GetDashboardSummaryParams{}, validation("lender_id", "UUID tidak valid")
	}
	institutionID, err := optionalUUID(filter.InstitutionID)
	if err != nil {
		return queries.GetDashboardSummaryParams{}, validation("institution_id", "UUID tidak valid")
	}
	return queries.GetDashboardSummaryParams{
		PeriodID:       periodID,
		InstitutionID:  institutionID,
		LenderID:       lenderID,
		IncludeHistory: filter.IncludeHistory,
		PublishYear:    optionalDashboardInt4(filter.PublishYear),
		BudgetYear:     optionalDashboardInt4(filter.BudgetYear),
		Quarter:        optionalText(filter.Quarter),
	}, nil
}

func dashboardRollupParams(filter model.DashboardFilterRequest) (queries.GetDashboardMonitoringRollupParams, error) {
	lenderID, err := optionalUUID(filter.LenderID)
	if err != nil {
		return queries.GetDashboardMonitoringRollupParams{}, validation("lender_id", "UUID tidak valid")
	}
	institutionID, err := optionalUUID(filter.InstitutionID)
	if err != nil {
		return queries.GetDashboardMonitoringRollupParams{}, validation("institution_id", "UUID tidak valid")
	}
	return queries.GetDashboardMonitoringRollupParams{
		BudgetYear:    optionalDashboardInt4(filter.BudgetYear),
		Quarter:       optionalText(filter.Quarter),
		LenderID:      lenderID,
		InstitutionID: institutionID,
	}, nil
}

func dashboardLAExposureParams(filter model.DashboardFilterRequest) (queries.GetDashboardLAExposureRollupParams, error) {
	lenderID, err := optionalUUID(filter.LenderID)
	if err != nil {
		return queries.GetDashboardLAExposureRollupParams{}, validation("lender_id", "UUID tidak valid")
	}
	institutionID, err := optionalUUID(filter.InstitutionID)
	if err != nil {
		return queries.GetDashboardLAExposureRollupParams{}, validation("institution_id", "UUID tidak valid")
	}
	return queries.GetDashboardLAExposureRollupParams{
		BudgetYear:    optionalDashboardInt4(filter.BudgetYear),
		Quarter:       optionalText(filter.Quarter),
		LenderID:      lenderID,
		InstitutionID: institutionID,
	}, nil
}

func dashboardLenderRollupParams(filter model.DashboardFilterRequest) (queries.GetDashboardLenderRollupParams, error) {
	lenderID, err := optionalUUID(filter.LenderID)
	if err != nil {
		return queries.GetDashboardLenderRollupParams{}, validation("lender_id", "UUID tidak valid")
	}
	institutionID, err := optionalUUID(filter.InstitutionID)
	if err != nil {
		return queries.GetDashboardLenderRollupParams{}, validation("institution_id", "UUID tidak valid")
	}
	return queries.GetDashboardLenderRollupParams{
		BudgetYear:    optionalDashboardInt4(filter.BudgetYear),
		Quarter:       optionalText(filter.Quarter),
		LenderID:      lenderID,
		InstitutionID: institutionID,
	}, nil
}

func dashboardInstitutionRollupParams(filter model.DashboardFilterRequest) (queries.GetDashboardInstitutionRollupParams, error) {
	lenderID, err := optionalUUID(filter.LenderID)
	if err != nil {
		return queries.GetDashboardInstitutionRollupParams{}, validation("lender_id", "UUID tidak valid")
	}
	institutionID, err := optionalUUID(filter.InstitutionID)
	if err != nil {
		return queries.GetDashboardInstitutionRollupParams{}, validation("institution_id", "UUID tidak valid")
	}
	return queries.GetDashboardInstitutionRollupParams{
		BudgetYear:    optionalDashboardInt4(filter.BudgetYear),
		Quarter:       optionalText(filter.Quarter),
		LenderID:      lenderID,
		InstitutionID: institutionID,
	}, nil
}

func dashboardExecutiveFunnelParams(filter model.DashboardFilterRequest) (queries.GetDashboardExecutiveFunnelParams, error) {
	periodID, err := optionalUUID(filter.PeriodID)
	if err != nil {
		return queries.GetDashboardExecutiveFunnelParams{}, validation("period_id", "UUID tidak valid")
	}
	return queries.GetDashboardExecutiveFunnelParams{
		PeriodID:       periodID,
		IncludeHistory: filter.IncludeHistory,
		PublishYear:    optionalDashboardInt4(filter.PublishYear),
		BudgetYear:     optionalDashboardInt4(filter.BudgetYear),
		Quarter:        optionalText(filter.Quarter),
	}, nil
}

func dashboardExecutiveRollupParams(filter model.DashboardFilterRequest) (queries.GetDashboardExecutiveTopInstitutionsParams, error) {
	periodID, err := optionalUUID(filter.PeriodID)
	if err != nil {
		return queries.GetDashboardExecutiveTopInstitutionsParams{}, validation("period_id", "UUID tidak valid")
	}
	return queries.GetDashboardExecutiveTopInstitutionsParams{
		PublishYear: optionalDashboardInt4(filter.PublishYear),
		PeriodID:    periodID,
		BudgetYear:  optionalDashboardInt4(filter.BudgetYear),
		Quarter:     optionalText(filter.Quarter),
	}, nil
}

func dashboardPipelineListParams(filter model.PipelineBottleneckFilterRequest, pagination model.PaginationParams, limit, offset int) (queries.ListDashboardPipelineBottleneckItemsParams, error) {
	periodID, err := optionalUUID(filter.PeriodID)
	if err != nil {
		return queries.ListDashboardPipelineBottleneckItemsParams{}, validation("period_id", "UUID tidak valid")
	}
	lenderID, err := optionalUUID(filter.LenderID)
	if err != nil {
		return queries.ListDashboardPipelineBottleneckItemsParams{}, validation("lender_id", "UUID tidak valid")
	}
	institutionID, err := optionalUUID(filter.InstitutionID)
	if err != nil {
		return queries.ListDashboardPipelineBottleneckItemsParams{}, validation("institution_id", "UUID tidak valid")
	}
	stage, err := optionalPipelineStage(filter.Stage)
	if err != nil {
		return queries.ListDashboardPipelineBottleneckItemsParams{}, err
	}
	sort := allowedDashboardSort(pagination.Sort)
	order := strings.ToLower(strings.TrimSpace(pagination.Order))
	if order != "asc" && order != "desc" {
		order = "desc"
	}
	return queries.ListDashboardPipelineBottleneckItemsParams{
		Sort:          sort,
		Order:         order,
		Offset:        int32(offset),
		Limit:         int32(limit),
		PeriodID:      periodID,
		PublishYear:   optionalDashboardInt4(filter.PublishYear),
		LenderID:      lenderID,
		InstitutionID: institutionID,
		Stage:         stage,
		MinAgeDays:    optionalDashboardInt4(filter.MinAgeDays),
		Search:        nullableText(pagination.Search),
	}, nil
}

func optionalUUID(value *string) (pgtype.UUID, error) {
	if value == nil || *value == "" {
		return pgtype.UUID{}, nil
	}
	return model.ParseUUID(*value)
}

func optionalDashboardInt4(value *int32) pgtype.Int4 {
	if value == nil {
		return pgtype.Int4{}
	}
	return pgtype.Int4{Int32: *value, Valid: true}
}

func optionalPipelineStage(value *string) (pgtype.Text, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return pgtype.Text{}, nil
	}
	stage := strings.ToUpper(strings.TrimSpace(*value))
	if _, ok := pipelineStageLabels[stage]; !ok {
		return pgtype.Text{}, validation("stage", "stage tidak valid")
	}
	return pgtype.Text{String: stage, Valid: true}, nil
}

func allowedDashboardSort(value string) string {
	switch strings.TrimSpace(value) {
	case "stage", "project_name", "amount_usd", "age_days":
		return strings.TrimSpace(value)
	default:
		return "age_days"
	}
}

func totalPages(total int64, limit int) int {
	if total == 0 {
		return 0
	}
	return int((total + int64(limit) - 1) / int64(limit))
}

func pipelineStageSummary(rows []queries.GetDashboardPipelineBottleneckStageSummaryRow) []model.PipelineStageSummary {
	items := make([]model.PipelineStageSummary, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.PipelineStageSummary{
			Stage:        row.Stage,
			Label:        pipelineStageLabel(row.Stage),
			ProjectCount: int(row.ProjectCount),
			AmountUSD:    floatFromNumeric(row.AmountUsd),
			AvgAgeDays:   floatFromNumeric(row.AvgAgeDays),
		})
	}
	return items
}

func pipelineBottleneckItems(rows []queries.ListDashboardPipelineBottleneckItemsRow) []model.PipelineBottleneckItem {
	items := make([]model.PipelineBottleneckItem, 0, len(rows))
	for _, row := range rows {
		var relevantAt *string
		if row.RelevantAt.Valid {
			formatted := row.RelevantAt.Time.Format(time.RFC3339)
			relevantAt = &formatted
		}
		items = append(items, model.PipelineBottleneckItem{
			ProjectID:          model.UUIDToString(row.ProjectID),
			ReferenceType:      row.ReferenceType,
			JourneyBBProjectID: optionalUUIDString(row.JourneyBbProjectID),
			Code:               row.Code,
			ProjectName:        row.ProjectName,
			CurrentStage:       row.CurrentStage,
			StageLabel:         pipelineStageLabel(row.CurrentStage),
			AgeDays:            int(row.AgeDays),
			AmountUSD:          floatFromNumeric(row.AmountUsd),
			InstitutionName:    row.InstitutionName,
			LenderNames:        safeStringSlice(row.LenderNames),
			RecommendedAction:  pipelineRecommendedAction(row.CurrentStage),
			RelevantAt:         relevantAt,
		})
	}
	return items
}

var pipelineStageLabels = map[string]string{
	"BB_NO_LENDER":            "Blue Book No Lender",
	"INDICATION_NO_LOI":       "Indication No Letter of Intent",
	"LOI_NO_GB":               "Letter of Intent No Green Book",
	"GB_NO_DK":                "Green Book No Daftar Kegiatan",
	"DK_NO_LA":                "Daftar Kegiatan No Loan Agreement",
	"LA_NOT_EFFECTIVE":        "Loan Agreement Not Effective",
	"EFFECTIVE_NO_MONITORING": "Effective No Monitoring",
}

func pipelineStageLabel(stage string) string {
	if label, ok := pipelineStageLabels[stage]; ok {
		return label
	}
	return stage
}

func pipelineRecommendedAction(stage string) string {
	switch stage {
	case "BB_NO_LENDER":
		return "Market sounding / cari lender indication."
	case "INDICATION_NO_LOI":
		return "Follow up lender untuk Letter of Intent."
	case "LOI_NO_GB":
		return "Cek readiness dan usulkan Green Book."
	case "GB_NO_DK":
		return "Dorong usulan/penetapan Daftar Kegiatan."
	case "DK_NO_LA":
		return "Dorong negosiasi/legal agreement."
	case "LA_NOT_EFFECTIVE":
		return "Monitor effectiveness conditions."
	case "EFFECTIVE_NO_MONITORING":
		return "Input monitoring triwulanan."
	default:
		return "Tindak lanjuti status proyek."
	}
}

func executiveMetricCards(summary *model.DashboardSummary) []model.MetricCard {
	return []model.MetricCard{
		{Key: "bb_projects", Label: "Blue Book Projects", Value: float64(summary.TotalBBProjects), Unit: "project", Category: "pipeline"},
		{Key: "gb_projects", Label: "Green Book Projects", Value: float64(summary.TotalGBProjects), Unit: "project", Category: "pipeline"},
		{Key: "dk_financing_usd", Label: "Daftar Kegiatan Financing", Value: summary.DKFinancingUSD, Unit: "USD", Category: "commitment"},
		{Key: "la_commitment_usd", Label: "Loan Agreement Commitment", Value: summary.LACommitmentUSD, Unit: "USD", Category: "commitment"},
		{Key: "realized_disbursement_usd", Label: "Realized Disbursement", Value: summary.RealizedDisbursementUSD, Unit: "USD", Category: "monitoring"},
		{Key: "absorption_pct", Label: "Absorption", Value: summary.AbsorptionPct, Unit: "percent", Category: "monitoring"},
	}
}

func executiveFunnelMetrics(rows []queries.GetDashboardExecutiveFunnelRow) []model.StageMetric {
	byStage := make(map[string]queries.GetDashboardExecutiveFunnelRow, len(rows))
	for _, row := range rows {
		byStage[row.Stage] = row
	}
	order := []string{"BB", "GB", "DK", "LA", "MONITORING"}
	metrics := make([]model.StageMetric, 0, len(order))
	for _, stage := range order {
		row := byStage[stage]
		metrics = append(metrics, model.StageMetric{
			Stage:        stage,
			Label:        stageLabel(stage),
			ProjectCount: int(row.ProjectCount),
			AmountUSD:    floatFromNumeric(row.AmountUsd),
		})
	}
	return metrics
}

func executiveInstitutionBreakdown(rows []queries.GetDashboardExecutiveTopInstitutionsRow) []model.BreakdownItem {
	items := make([]model.BreakdownItem, 0, len(rows))
	for _, row := range rows {
		id := optionalUUIDString(row.ID)
		items = append(items, model.BreakdownItem{
			ID:          id,
			Label:       row.Label,
			ItemCount:   int(row.ItemCount),
			AmountUSD:   floatFromNumeric(row.AmountUsd),
			RealizedUSD: floatFromNumeric(row.RealizedUsd),
		})
	}
	return items
}

func executiveLenderBreakdown(rows []queries.GetDashboardExecutiveTopLendersRow) []model.BreakdownItem {
	items := make([]model.BreakdownItem, 0, len(rows))
	for _, row := range rows {
		id := optionalUUIDString(row.ID)
		items = append(items, model.BreakdownItem{
			ID:          id,
			Label:       row.Label,
			ItemCount:   int(row.ItemCount),
			AmountUSD:   floatFromNumeric(row.AmountUsd),
			RealizedUSD: floatFromNumeric(row.RealizedUsd),
		})
	}
	return items
}

func executiveRiskItems(rows []queries.ListDashboardExecutiveRiskItemsRow) []model.RiskItem {
	items := make([]model.RiskItem, 0, len(rows))
	for _, row := range rows {
		referenceID := optionalUUIDString(row.ReferenceID)
		journeyID := optionalUUIDString(row.JourneyBbProjectID)
		var daysUntilClosing *int
		if row.DaysUntilClosing > 0 {
			days := int(row.DaysUntilClosing)
			daysUntilClosing = &days
		}
		id := ""
		if referenceID != nil {
			id = *referenceID
		}
		items = append(items, model.RiskItem{
			ID:                 id,
			RiskType:           row.RiskType,
			ReferenceID:        referenceID,
			ReferenceType:      row.ReferenceType,
			JourneyBBProjectID: journeyID,
			Code:               row.Code,
			Title:              row.Title,
			Description:        row.Description,
			Severity:           row.Severity,
			AmountUSD:          floatFromNumeric(row.AmountUsd),
			DaysUntilClosing:   daysUntilClosing,
			AbsorptionPct:      floatFromNumeric(row.AbsorptionPct),
			Score:              floatFromNumeric(row.Score),
		})
	}
	return items
}

func optionalUUIDString(value pgtype.UUID) *string {
	if !value.Valid {
		return nil
	}
	id := model.UUIDToString(value)
	return &id
}

func executiveInsights(summary *model.DashboardSummary, funnel []model.StageMetric, risks []model.RiskItem) []string {
	bottleneck := model.StageMetric{}
	for _, stage := range funnel {
		if stage.ProjectCount > bottleneck.ProjectCount {
			bottleneck = stage
		}
	}
	closingCount := 0
	for _, risk := range risks {
		if risk.RiskType == "LA_CLOSING_12_MONTHS" {
			closingCount++
		}
	}
	insights := make([]string, 0, 3)
	if bottleneck.Stage != "" {
		insights = append(insights, fmt.Sprintf(
			"Bottleneck terbesar berada pada tahap %s, dengan %d proyek senilai USD %.0f.",
			bottleneck.Label,
			bottleneck.ProjectCount,
			bottleneck.AmountUSD,
		))
	}
	insights = append(insights, fmt.Sprintf("Serapan kumulatif mencapai %.2f%% dari rencana monitoring.", summary.AbsorptionPct))
	insights = append(insights, fmt.Sprintf("%d Loan Agreement akan closing dalam 12 bulan.", closingCount))
	return insights
}

func dashboardMetricCards(summary *model.DashboardSummary) []model.MetricCard {
	return []model.MetricCard{
		{Key: "bb_pipeline_usd", Label: "Blue Book Pipeline", Value: summary.BBPipelineUSD, Unit: "USD", Category: "pipeline"},
		{Key: "gb_pipeline_usd", Label: "Green Book Pipeline", Value: summary.GBPipelineUSD, Unit: "USD", Category: "pipeline"},
		{Key: "gb_local_usd", Label: "Green Book Local Funding", Value: summary.GBLocalUSD, Unit: "USD", Category: "pipeline"},
		{Key: "dk_financing_usd", Label: "Daftar Kegiatan Financing", Value: summary.DKFinancingUSD, Unit: "USD", Category: "commitment"},
		{Key: "dk_counterpart_usd", Label: "Daftar Kegiatan Counterpart", Value: summary.DKCounterpartUSD, Unit: "USD", Category: "commitment"},
		{Key: "la_commitment_usd", Label: "Loan Agreement Commitment", Value: summary.LACommitmentUSD, Unit: "USD", Category: "commitment"},
		{Key: "planned_disbursement_usd", Label: "Planned Disbursement", Value: summary.PlannedDisbursementUSD, Unit: "USD", Category: "monitoring"},
		{Key: "realized_disbursement_usd", Label: "Realized Disbursement", Value: summary.RealizedDisbursementUSD, Unit: "USD", Category: "monitoring"},
		{Key: "absorption_pct", Label: "Absorption", Value: summary.AbsorptionPct, Unit: "percent", Category: "monitoring"},
		{Key: "la_absorption_pct", Label: "LA Absorption", Value: summary.LAAbsorptionPct, Unit: "percent", Category: "monitoring"},
		{Key: "undisbursed_usd", Label: "Undisbursed", Value: summary.UndisbursedUSD, Unit: "USD", Category: "monitoring"},
	}
}

func stageLabel(stage string) string {
	switch stage {
	case "BB":
		return "Blue Book"
	case "LA":
		return "Loan Agreement"
	case "MONITORING":
		return "Monitoring"
	case "BB_ONLY":
		return "Blue Book Only"
	case "BB_WITH_LENDER_INDICATION":
		return "Blue Book With Lender Indication"
	case "BB_WITH_LOI":
		return "Blue Book With Letter of Intent"
	case "GB":
		return "Green Book"
	case "DK":
		return "Daftar Kegiatan"
	case "LA_SIGNED_NOT_EFFECTIVE":
		return "Loan Agreement Signed Not Effective"
	case "LA_EFFECTIVE_NO_MONITORING":
		return "Loan Agreement Effective No Monitoring"
	case "MONITORING_ACTIVE":
		return "Monitoring Active"
	default:
		return stage
	}
}
