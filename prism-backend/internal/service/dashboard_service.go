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

func (s *DashboardService) GetGreenBookReadiness(ctx context.Context, filter model.GreenBookReadinessFilterRequest) (*model.GreenBookReadinessDashboard, error) {
	params, err := dashboardGreenBookReadinessParams(filter)
	if err != nil {
		return nil, err
	}

	summaryRow, err := s.queries.GetDashboardGreenBookReadinessSummary(ctx, queries.GetDashboardGreenBookReadinessSummaryParams(params))
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil ringkasan Green Book readiness")
	}

	disbursementRows, err := s.queries.GetDashboardGreenBookReadinessDisbursementByYear(ctx, queries.GetDashboardGreenBookReadinessDisbursementByYearParams{
		GreenBookID:     params.GreenBookID,
		PublishYear:     params.PublishYear,
		InstitutionID:   params.InstitutionID,
		LenderID:        params.LenderID,
		ReadinessStatus: params.ReadinessStatus,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil rencana disbursement Green Book")
	}

	allocationRow, err := s.queries.GetDashboardGreenBookReadinessFundingAllocation(ctx, queries.GetDashboardGreenBookReadinessFundingAllocationParams{
		GreenBookID:     params.GreenBookID,
		PublishYear:     params.PublishYear,
		InstitutionID:   params.InstitutionID,
		LenderID:        params.LenderID,
		ReadinessStatus: params.ReadinessStatus,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil funding allocation Green Book")
	}

	itemRows, err := s.queries.ListDashboardGreenBookReadinessItems(ctx, queries.ListDashboardGreenBookReadinessItemsParams{
		ReadinessStatus: params.ReadinessStatus,
		GreenBookID:     params.GreenBookID,
		PublishYear:     params.PublishYear,
		InstitutionID:   params.InstitutionID,
		LenderID:        params.LenderID,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil readiness item Green Book")
	}

	return &model.GreenBookReadinessDashboard{
		Summary:                greenBookReadinessSummary(summaryRow),
		DisbursementPlanByYear: greenBookDisbursementYears(disbursementRows),
		FundingAllocation:      greenBookFundingAllocation(allocationRow),
		ReadinessItems:         greenBookReadinessItems(itemRows),
	}, nil
}

func (s *DashboardService) GetLenderFinancingMix(ctx context.Context, filter model.LenderFinancingMixFilterRequest) (*model.LenderFinancingMixDashboard, error) {
	params, err := dashboardLenderFinancingMixParams(filter)
	if err != nil {
		return nil, err
	}

	summaryRow, err := s.queries.GetDashboardLenderFinancingMixSummary(ctx, queries.GetDashboardLenderFinancingMixSummaryParams(params))
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil ringkasan lender financing mix")
	}
	certaintyRows, err := s.queries.GetDashboardLenderCertaintyLadder(ctx, queries.GetDashboardLenderCertaintyLadderParams(params))
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil certainty ladder lender")
	}
	conversionRows, err := s.queries.GetDashboardLenderConversion(ctx, queries.GetDashboardLenderConversionParams(params))
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil conversion lender")
	}
	currencyRows, err := s.queries.GetDashboardCurrencyExposure(ctx, queries.GetDashboardCurrencyExposureParams{
		PublishYear: params.PublishYear,
		PeriodID:    params.PeriodID,
		LenderType:  params.LenderType,
		LenderID:    params.LenderID,
		Currency:    params.Currency,
		BudgetYear:  params.BudgetYear,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil currency exposure")
	}
	cofinancingRows, err := s.queries.ListDashboardCofinancingItems(ctx, queries.ListDashboardCofinancingItemsParams{
		LenderType:  params.LenderType,
		LenderID:    params.LenderID,
		Currency:    params.Currency,
		PublishYear: params.PublishYear,
		PeriodID:    params.PeriodID,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil cofinancing item")
	}

	return &model.LenderFinancingMixDashboard{
		Summary:          lenderFinancingMixSummary(summaryRow),
		CertaintyLadder:  lenderCertaintyPoints(certaintyRows),
		LenderConversion: lenderConversionItems(conversionRows),
		CurrencyExposure: currencyExposureItems(currencyRows),
		CofinancingItems: cofinancingItems(cofinancingRows),
	}, nil
}

func (s *DashboardService) GetKLPortfolioPerformance(ctx context.Context, filter model.KLPortfolioPerformanceFilterRequest) (*model.KLPortfolioPerformanceDashboard, error) {
	params, err := dashboardKLPortfolioPerformanceParams(filter)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.GetDashboardKLPortfolioPerformanceItems(ctx, params)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil K/L portfolio performance")
	}
	items := klPortfolioPerformanceItems(rows)
	return &model.KLPortfolioPerformanceDashboard{
		Summary: klPortfolioPerformanceSummary(items),
		Items:   items,
	}, nil
}

func (s *DashboardService) GetLADisbursement(ctx context.Context, filter model.LADisbursementFilterRequest) (*model.LADisbursementDashboard, error) {
	params, err := dashboardLADisbursementParams(filter)
	if err != nil {
		return nil, err
	}

	summaryRow, err := s.queries.GetDashboardLADisbursementSummary(ctx, queries.GetDashboardLADisbursementSummaryParams{
		LenderID:      params.LenderID,
		InstitutionID: params.InstitutionID,
		IsExtended:    params.IsExtended,
		ClosingMonths: params.ClosingMonths,
		BudgetYear:    params.BudgetYear,
		Quarter:       params.Quarter,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil ringkasan Loan Agreement disbursement")
	}

	trendRows, err := s.queries.GetDashboardLADisbursementQuarterlyTrend(ctx, queries.GetDashboardLADisbursementQuarterlyTrendParams{
		BudgetYear:    params.BudgetYear,
		Quarter:       params.Quarter,
		LenderID:      params.LenderID,
		InstitutionID: params.InstitutionID,
		IsExtended:    params.IsExtended,
		ClosingMonths: params.ClosingMonths,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil tren Loan Agreement disbursement")
	}

	closingRows, err := s.queries.ListDashboardLAClosingRisks(ctx, queries.ListDashboardLAClosingRisksParams{
		RiskLevel:     params.RiskLevel,
		LenderID:      params.LenderID,
		InstitutionID: params.InstitutionID,
		IsExtended:    params.IsExtended,
		ClosingMonths: params.ClosingMonths,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil risiko closing Loan Agreement")
	}

	underRows, err := s.queries.ListDashboardLAUnderDisbursementRisks(ctx, queries.ListDashboardLAUnderDisbursementRisksParams{
		RiskLevel:     params.RiskLevel,
		LenderID:      params.LenderID,
		InstitutionID: params.InstitutionID,
		IsExtended:    params.IsExtended,
		ClosingMonths: params.ClosingMonths,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil risiko under-disbursement")
	}

	componentRows, err := s.queries.GetDashboardLAComponentBreakdown(ctx, queries.GetDashboardLAComponentBreakdownParams{
		BudgetYear:    params.BudgetYear,
		Quarter:       params.Quarter,
		LenderID:      params.LenderID,
		InstitutionID: params.InstitutionID,
		IsExtended:    params.IsExtended,
		ClosingMonths: params.ClosingMonths,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil breakdown komponen monitoring")
	}

	return &model.LADisbursementDashboard{
		Summary:                laDisbursementSummary(summaryRow),
		QuarterlyTrend:         laDisbursementTrend(trendRows),
		ClosingRisks:           laClosingRisks(closingRows),
		UnderDisbursementRisks: laUnderDisbursementRisks(underRows),
		ComponentBreakdown:     laComponentBreakdown(componentRows),
	}, nil
}

func (s *DashboardService) GetDataQualityGovernance(ctx context.Context, filter model.DataQualityGovernanceFilterRequest, includeAudit bool) (*model.DataQualityGovernanceDashboard, error) {
	params, err := dashboardDataQualityParams(filter)
	if err != nil {
		return nil, err
	}

	issueRows, err := s.queries.ListDashboardDataQualityIssues(ctx, params)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil data quality issues")
	}
	issues := dataQualityIssues(issueRows)

	var auditEvents *int
	var auditSummary *model.DataQualityAuditSummary
	if includeAudit {
		auditDays := normalizeAuditDays(filter.AuditDays)
		eventCount, err := s.queries.CountDashboardAuditEvents(ctx, auditDays)
		if err != nil {
			return nil, apperrors.Internal("Gagal menghitung audit event")
		}
		eventCountInt := int(eventCount)
		auditEvents = &eventCountInt

		byUserRows, err := s.queries.GetDashboardAuditSummaryByUser(ctx, auditDays)
		if err != nil {
			return nil, apperrors.Internal("Gagal mengambil audit summary per user")
		}
		byTableRows, err := s.queries.GetDashboardAuditSummaryByTable(ctx, auditDays)
		if err != nil {
			return nil, apperrors.Internal("Gagal mengambil audit summary per table")
		}
		recentRows, err := s.queries.ListDashboardAuditRecentActivity(ctx, auditDays)
		if err != nil {
			return nil, apperrors.Internal("Gagal mengambil audit activity terbaru")
		}
		auditSummary = &model.DataQualityAuditSummary{
			ByUser:         auditSummaryByUser(byUserRows),
			ByTable:        auditSummaryByTable(byTableRows),
			RecentActivity: auditRecentActivity(recentRows),
		}
	}

	return &model.DataQualityGovernanceDashboard{
		Summary:      dataQualityIssueSummary(issues, auditEvents),
		Issues:       issues,
		AuditSummary: auditSummary,
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

func dashboardGreenBookReadinessParams(filter model.GreenBookReadinessFilterRequest) (queries.GetDashboardGreenBookReadinessSummaryParams, error) {
	greenBookID, err := optionalUUID(filter.GreenBookID)
	if err != nil {
		return queries.GetDashboardGreenBookReadinessSummaryParams{}, validation("green_book_id", "UUID tidak valid")
	}
	lenderID, err := optionalUUID(filter.LenderID)
	if err != nil {
		return queries.GetDashboardGreenBookReadinessSummaryParams{}, validation("lender_id", "UUID tidak valid")
	}
	institutionID, err := optionalUUID(filter.InstitutionID)
	if err != nil {
		return queries.GetDashboardGreenBookReadinessSummaryParams{}, validation("institution_id", "UUID tidak valid")
	}
	readinessStatus, err := optionalReadinessStatus(filter.ReadinessStatus)
	if err != nil {
		return queries.GetDashboardGreenBookReadinessSummaryParams{}, err
	}
	return queries.GetDashboardGreenBookReadinessSummaryParams{
		GreenBookID:     greenBookID,
		PublishYear:     optionalDashboardInt4(filter.PublishYear),
		InstitutionID:   institutionID,
		LenderID:        lenderID,
		ReadinessStatus: readinessStatus,
	}, nil
}

func dashboardLenderFinancingMixParams(filter model.LenderFinancingMixFilterRequest) (queries.GetDashboardLenderFinancingMixSummaryParams, error) {
	periodID, err := optionalUUID(filter.PeriodID)
	if err != nil {
		return queries.GetDashboardLenderFinancingMixSummaryParams{}, validation("period_id", "UUID tidak valid")
	}
	lenderID, err := optionalUUID(filter.LenderID)
	if err != nil {
		return queries.GetDashboardLenderFinancingMixSummaryParams{}, validation("lender_id", "UUID tidak valid")
	}
	lenderType, err := optionalLenderType(filter.LenderType)
	if err != nil {
		return queries.GetDashboardLenderFinancingMixSummaryParams{}, err
	}
	currency := pgtype.Text{}
	if filter.Currency != nil && strings.TrimSpace(*filter.Currency) != "" {
		currency = pgtype.Text{String: strings.ToUpper(strings.TrimSpace(*filter.Currency)), Valid: true}
	}
	return queries.GetDashboardLenderFinancingMixSummaryParams{
		PeriodID:    periodID,
		PublishYear: optionalDashboardInt4(filter.PublishYear),
		Currency:    currency,
		LenderType:  lenderType,
		LenderID:    lenderID,
		BudgetYear:  optionalDashboardInt4(filter.BudgetYear),
	}, nil
}

func dashboardKLPortfolioPerformanceParams(filter model.KLPortfolioPerformanceFilterRequest) (queries.GetDashboardKLPortfolioPerformanceItemsParams, error) {
	institutionID, err := optionalUUID(filter.InstitutionID)
	if err != nil {
		return queries.GetDashboardKLPortfolioPerformanceItemsParams{}, validation("institution_id", "UUID tidak valid")
	}
	periodID, err := optionalUUID(filter.PeriodID)
	if err != nil {
		return queries.GetDashboardKLPortfolioPerformanceItemsParams{}, validation("period_id", "UUID tidak valid")
	}
	institutionRole, err := optionalInstitutionRole(filter.InstitutionRole)
	if err != nil {
		return queries.GetDashboardKLPortfolioPerformanceItemsParams{}, err
	}
	quarter, err := optionalQuarter(filter.Quarter)
	if err != nil {
		return queries.GetDashboardKLPortfolioPerformanceItemsParams{}, err
	}
	return queries.GetDashboardKLPortfolioPerformanceItemsParams{
		SortBy:          optionalKLSortBy(filter.SortBy),
		InstitutionID:   institutionID,
		PeriodID:        periodID,
		PublishYear:     optionalDashboardInt4(filter.PublishYear),
		InstitutionRole: institutionRole,
		BudgetYear:      optionalDashboardInt4(filter.BudgetYear),
		Quarter:         quarter,
	}, nil
}

type dashboardLADisbursementQueryParams struct {
	BudgetYear    pgtype.Int4
	Quarter       pgtype.Text
	LenderID      pgtype.UUID
	InstitutionID pgtype.UUID
	IsExtended    pgtype.Bool
	ClosingMonths pgtype.Int4
	RiskLevel     pgtype.Text
}

func dashboardLADisbursementParams(filter model.LADisbursementFilterRequest) (dashboardLADisbursementQueryParams, error) {
	lenderID, err := optionalUUID(filter.LenderID)
	if err != nil {
		return dashboardLADisbursementQueryParams{}, validation("lender_id", "UUID tidak valid")
	}
	institutionID, err := optionalUUID(filter.InstitutionID)
	if err != nil {
		return dashboardLADisbursementQueryParams{}, validation("institution_id", "UUID tidak valid")
	}
	quarter, err := optionalQuarter(filter.Quarter)
	if err != nil {
		return dashboardLADisbursementQueryParams{}, err
	}
	closingMonths, err := optionalClosingMonths(filter.ClosingMonths)
	if err != nil {
		return dashboardLADisbursementQueryParams{}, err
	}
	riskLevel, err := optionalDashboardRiskLevel(filter.RiskLevel)
	if err != nil {
		return dashboardLADisbursementQueryParams{}, err
	}
	return dashboardLADisbursementQueryParams{
		BudgetYear:    optionalDashboardInt4(filter.BudgetYear),
		Quarter:       quarter,
		LenderID:      lenderID,
		InstitutionID: institutionID,
		IsExtended:    optionalDashboardBool(filter.IsExtended),
		ClosingMonths: closingMonths,
		RiskLevel:     riskLevel,
	}, nil
}

func dashboardDataQualityParams(filter model.DataQualityGovernanceFilterRequest) (queries.ListDashboardDataQualityIssuesParams, error) {
	severity, err := optionalDataQualitySeverity(filter.Severity)
	if err != nil {
		return queries.ListDashboardDataQualityIssuesParams{}, err
	}
	issueType, err := optionalDataQualityIssueType(filter.IssueType)
	if err != nil {
		return queries.ListDashboardDataQualityIssuesParams{}, err
	}
	return queries.ListDashboardDataQualityIssuesParams{
		Severity:       severity,
		Module:         optionalText(filter.Module),
		IssueType:      issueType,
		OnlyUnresolved: filter.OnlyUnresolved,
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

func optionalDashboardBool(value *bool) pgtype.Bool {
	if value == nil {
		return pgtype.Bool{}
	}
	return pgtype.Bool{Bool: *value, Valid: true}
}

func optionalClosingMonths(value *int32) (pgtype.Int4, error) {
	if value == nil {
		return pgtype.Int4{}, nil
	}
	switch *value {
	case 3, 6, 12:
		return pgtype.Int4{Int32: *value, Valid: true}, nil
	default:
		return pgtype.Int4{}, validation("closing_months", "closing_months harus 3, 6, atau 12")
	}
}

func optionalDashboardRiskLevel(value *string) (pgtype.Text, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return pgtype.Text{}, nil
	}
	riskLevel := strings.ToLower(strings.TrimSpace(*value))
	switch riskLevel {
	case "low", "medium", "high":
		return pgtype.Text{String: riskLevel, Valid: true}, nil
	default:
		return pgtype.Text{}, validation("risk_level", "risk_level harus low, medium, atau high")
	}
}

func optionalDataQualitySeverity(value *string) (pgtype.Text, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return pgtype.Text{}, nil
	}
	severity := strings.ToLower(strings.TrimSpace(*value))
	switch severity {
	case "info", "warning", "error":
		return pgtype.Text{String: severity, Valid: true}, nil
	default:
		return pgtype.Text{}, validation("severity", "severity harus info, warning, atau error")
	}
}

func optionalDataQualityIssueType(value *string) (pgtype.Text, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return pgtype.Text{}, nil
	}
	issueType := strings.ToUpper(strings.TrimSpace(*value))
	if _, ok := dataQualityIssueTypes[issueType]; !ok {
		return pgtype.Text{}, validation("issue_type", "issue_type tidak valid")
	}
	return pgtype.Text{String: issueType, Valid: true}, nil
}

func normalizeAuditDays(value int32) int32 {
	if value <= 0 {
		return 30
	}
	if value > 365 {
		return 365
	}
	return value
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

func optionalReadinessStatus(value *string) (pgtype.Text, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return pgtype.Text{}, nil
	}
	status := strings.ToUpper(strings.TrimSpace(*value))
	if _, ok := greenBookReadinessStatusLabels[status]; !ok {
		return pgtype.Text{}, validation("readiness_status", "readiness_status tidak valid")
	}
	return pgtype.Text{String: status, Valid: true}, nil
}

func optionalLenderType(value *string) (pgtype.Text, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return pgtype.Text{}, nil
	}
	switch strings.TrimSpace(*value) {
	case "Bilateral", "Multilateral", "KSA":
		return pgtype.Text{String: strings.TrimSpace(*value), Valid: true}, nil
	default:
		return pgtype.Text{}, validation("lender_type", "lender_type tidak valid")
	}
}

func optionalInstitutionRole(value *string) (pgtype.Text, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return pgtype.Text{}, nil
	}
	role := strings.TrimSpace(*value)
	if role != "Executing Agency" && role != "Implementing Agency" {
		return pgtype.Text{}, validation("institution_role", "institution_role tidak valid")
	}
	return pgtype.Text{String: role, Valid: true}, nil
}

func optionalQuarter(value *string) (pgtype.Text, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return pgtype.Text{}, nil
	}
	quarter := strings.ToUpper(strings.TrimSpace(*value))
	switch quarter {
	case "TW1", "TW2", "TW3", "TW4":
		return pgtype.Text{String: quarter, Valid: true}, nil
	default:
		return pgtype.Text{}, validation("quarter", "quarter tidak valid")
	}
}

func optionalKLSortBy(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{String: "pipeline_usd", Valid: true}
	}
	switch strings.TrimSpace(*value) {
	case "pipeline_usd", "la_commitment_usd", "absorption_pct", "risk_count":
		return pgtype.Text{String: strings.TrimSpace(*value), Valid: true}
	default:
		return pgtype.Text{String: "pipeline_usd", Valid: true}
	}
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

func greenBookReadinessSummary(row queries.GetDashboardGreenBookReadinessSummaryRow) model.GreenBookReadinessSummary {
	return model.GreenBookReadinessSummary{
		TotalProjects:           int(row.TotalProjects),
		TotalLoanUSD:            floatFromNumeric(row.TotalLoanUsd),
		TotalGrantUSD:           floatFromNumeric(row.TotalGrantUsd),
		TotalLocalUSD:           floatFromNumeric(row.TotalLocalUsd),
		ProjectsWithCofinancing: int(row.ProjectsWithCofinancing),
		ProjectsIncomplete:      int(row.ProjectsIncomplete),
		ProjectsReady:           int(row.ProjectsReady),
		ProjectsPartial:         int(row.ProjectsPartial),
	}
}

func greenBookDisbursementYears(rows []queries.GetDashboardGreenBookReadinessDisbursementByYearRow) []model.GreenBookDisbursementYear {
	items := make([]model.GreenBookDisbursementYear, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.GreenBookDisbursementYear{
			Year:      row.Year,
			AmountUSD: floatFromNumeric(row.AmountUsd),
		})
	}
	return items
}

func greenBookFundingAllocation(row queries.GetDashboardGreenBookReadinessFundingAllocationRow) model.GreenBookFundingAllocation {
	return model.GreenBookFundingAllocation{
		Services:      floatFromNumeric(row.Services),
		Constructions: floatFromNumeric(row.Constructions),
		Goods:         floatFromNumeric(row.Goods),
		Trainings:     floatFromNumeric(row.Trainings),
		Other:         floatFromNumeric(row.Other),
	}
}

func greenBookReadinessItems(rows []queries.ListDashboardGreenBookReadinessItemsRow) []model.GreenBookReadinessItem {
	items := make([]model.GreenBookReadinessItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.GreenBookReadinessItem{
			ProjectID:       model.UUIDToString(row.ProjectID),
			GreenBookID:     model.UUIDToString(row.GreenBookID),
			GBCode:          row.GbCode,
			ProjectName:     row.ProjectName,
			PublishYear:     row.PublishYear,
			ReadinessScore:  int(row.ReadinessScore),
			ReadinessStatus: row.ReadinessStatus,
			IsCofinancing:   row.IsCofinancing,
			MissingFields:   safeStringSlice(row.MissingFields),
			TotalFundingUSD: floatFromNumeric(row.TotalFundingUsd),
			InstitutionName: row.InstitutionName,
			LenderNames:     safeStringSlice(row.LenderNames),
		})
	}
	return items
}

func lenderFinancingMixSummary(row queries.GetDashboardLenderFinancingMixSummaryRow) model.LenderFinancingMixSummary {
	return model.LenderFinancingMixSummary{
		TotalLenders:        int(row.TotalLenders),
		BilateralUSD:        floatFromNumeric(row.BilateralUsd),
		MultilateralUSD:     floatFromNumeric(row.MultilateralUsd),
		KSAUSD:              floatFromNumeric(row.KsaUsd),
		CofinancingProjects: int(row.CofinancingProjects),
	}
}

func lenderCertaintyPoints(rows []queries.GetDashboardLenderCertaintyLadderRow) []model.LenderCertaintyPoint {
	items := make([]model.LenderCertaintyPoint, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.LenderCertaintyPoint{
			Stage:        row.Stage,
			LenderID:     model.UUIDToString(row.LenderID),
			LenderName:   row.LenderName,
			LenderType:   row.LenderType,
			ProjectCount: int(row.ProjectCount),
			AmountUSD:    floatFromNumeric(row.AmountUsd),
		})
	}
	return items
}

func lenderConversionItems(rows []queries.GetDashboardLenderConversionRow) []model.LenderConversionItem {
	items := make([]model.LenderConversionItem, 0, len(rows))
	for _, row := range rows {
		indicationCount := int(row.IndicationCount)
		laCount := int(row.LaCount)
		conversionPct := 0.0
		if indicationCount > 0 {
			conversionPct = float64(laCount) / float64(indicationCount) * 100
		}
		items = append(items, model.LenderConversionItem{
			LenderID:        model.UUIDToString(row.LenderID),
			LenderName:      row.LenderName,
			LenderType:      row.LenderType,
			IndicationCount: indicationCount,
			LoICount:        int(row.LoiCount),
			GBCount:         int(row.GbCount),
			DKCount:         int(row.DkCount),
			LACount:         laCount,
			IndicationUSD:   floatFromNumeric(row.IndicationUsd),
			LAUSD:           floatFromNumeric(row.LaUsd),
			LAConversionPct: conversionPct,
		})
	}
	return items
}

func currencyExposureItems(rows []queries.GetDashboardCurrencyExposureRow) []model.CurrencyExposureItem {
	items := make([]model.CurrencyExposureItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.CurrencyExposureItem{
			Currency:       row.Currency,
			Stage:          row.Stage,
			ProjectCount:   int(row.ProjectCount),
			AmountOriginal: floatFromNumeric(row.AmountOriginal),
			AmountUSD:      floatFromNumeric(row.AmountUsd),
		})
	}
	return items
}

func cofinancingItems(rows []queries.ListDashboardCofinancingItemsRow) []model.CofinancingItem {
	items := make([]model.CofinancingItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.CofinancingItem{
			ProjectID:     model.UUIDToString(row.ProjectID),
			ReferenceType: row.ReferenceType,
			ProjectCode:   row.ProjectCode,
			ProjectName:   row.ProjectName,
			LenderCount:   int(row.LenderCount),
			LenderNames:   safeStringSlice(row.LenderNames),
			AmountUSD:     floatFromNumeric(row.AmountUsd),
		})
	}
	return items
}

func klPortfolioPerformanceItems(rows []queries.GetDashboardKLPortfolioPerformanceItemsRow) []model.KLPortfolioPerformanceItem {
	items := make([]model.KLPortfolioPerformanceItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.KLPortfolioPerformanceItem{
			InstitutionID:       model.UUIDToString(row.InstitutionID),
			InstitutionName:     row.InstitutionName,
			BBProjectCount:      int(row.BbProjectCount),
			GBProjectCount:      int(row.GbProjectCount),
			DKProjectCount:      int(row.DkProjectCount),
			LACount:             int(row.LaCount),
			PipelineUSD:         floatFromNumeric(row.PipelineUsd),
			LACommitmentUSD:     floatFromNumeric(row.LaCommitmentUsd),
			PlannedUSD:          floatFromNumeric(row.PlannedUsd),
			RealizedUSD:         floatFromNumeric(row.RealizedUsd),
			AbsorptionPct:       floatFromNumeric(row.AbsorptionPct),
			RiskCount:           int(row.RiskCount),
			PerformanceScore:    floatFromNumeric(row.PerformanceScore),
			PerformanceCategory: row.PerformanceCategory,
		})
	}
	return items
}

func klPortfolioPerformanceSummary(items []model.KLPortfolioPerformanceItem) model.KLPortfolioPerformanceSummary {
	summary := model.KLPortfolioPerformanceSummary{TotalInstitutions: len(items)}
	if len(items) == 0 {
		return summary
	}

	topExposure := items[0]
	lowestAbsorption := items[0]
	highestRisk := items[0]
	absorptionTotal := 0.0
	for _, item := range items {
		exposure := item.PipelineUSD + item.LACommitmentUSD
		if exposure > topExposure.PipelineUSD+topExposure.LACommitmentUSD {
			topExposure = item
		}
		if item.AbsorptionPct < lowestAbsorption.AbsorptionPct {
			lowestAbsorption = item
		}
		if item.RiskCount > highestRisk.RiskCount {
			highestRisk = item
		}
		summary.TotalInstitutionExposureUSD += exposure
		summary.TotalInstitutionRiskCount += item.RiskCount
		absorptionTotal += item.AbsorptionPct
	}
	summary.TopExposureInstitution = topExposure.InstitutionName
	summary.TopExposureUSD = topExposure.PipelineUSD + topExposure.LACommitmentUSD
	summary.LowestAbsorptionInstitution = lowestAbsorption.InstitutionName
	summary.LowestAbsorptionPct = lowestAbsorption.AbsorptionPct
	summary.HighestRiskInstitution = highestRisk.InstitutionName
	summary.HighestRiskCount = highestRisk.RiskCount
	summary.AverageAbsorptionPct = absorptionTotal / float64(len(items))
	return summary
}

func laDisbursementSummary(row queries.GetDashboardLADisbursementSummaryRow) model.LADisbursementSummary {
	return model.LADisbursementSummary{
		LACount:           int(row.LaCount),
		EffectiveCount:    int(row.EffectiveCount),
		NotEffectiveCount: int(row.NotEffectiveCount),
		ExtendedCount:     int(row.ExtendedCount),
		CommitmentUSD:     floatFromNumeric(row.CommitmentUsd),
		PlannedUSD:        floatFromNumeric(row.PlannedUsd),
		RealizedUSD:       floatFromNumeric(row.RealizedUsd),
		AbsorptionPct:     floatFromNumeric(row.AbsorptionPct),
		UndisbursedUSD:    floatFromNumeric(row.UndisbursedUsd),
	}
}

func laDisbursementTrend(rows []queries.GetDashboardLADisbursementQuarterlyTrendRow) []model.LADisbursementTrendPoint {
	items := make([]model.LADisbursementTrendPoint, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.LADisbursementTrendPoint{
			Period:        fmt.Sprintf("%d-%s", row.BudgetYear, row.Quarter),
			BudgetYear:    row.BudgetYear,
			Quarter:       row.Quarter,
			PlannedUSD:    floatFromNumeric(row.PlannedUsd),
			RealizedUSD:   floatFromNumeric(row.RealizedUsd),
			AbsorptionPct: floatFromNumeric(row.AbsorptionPct),
		})
	}
	return items
}

func laClosingRisks(rows []queries.ListDashboardLAClosingRisksRow) []model.LAClosingRiskItem {
	items := make([]model.LAClosingRiskItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.LAClosingRiskItem{
			LoanAgreementID:       model.UUIDToString(row.LoanAgreementID),
			LoanCode:              row.LoanCode,
			ProjectName:           row.ProjectName,
			LenderName:            row.LenderName,
			EffectiveDate:         dateString(row.EffectiveDate),
			ClosingDate:           dateString(row.ClosingDate),
			DaysUntilClosing:      int(row.DaysUntilClosing),
			CommitmentUSD:         floatFromNumeric(row.CommitmentUsd),
			CumulativeRealizedUSD: floatFromNumeric(row.CumulativeRealizedUsd),
			UndisbursedUSD:        floatFromNumeric(row.UndisbursedUsd),
			LAAbsorptionPct:       floatFromNumeric(row.LaAbsorptionPct),
			RiskType:              row.RiskType,
			RiskLevel:             row.RiskLevel,
		})
	}
	return items
}

func laUnderDisbursementRisks(rows []queries.ListDashboardLAUnderDisbursementRisksRow) []model.LAUnderDisbursementRiskItem {
	items := make([]model.LAUnderDisbursementRiskItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.LAUnderDisbursementRiskItem{
			LoanAgreementID:                model.UUIDToString(row.LoanAgreementID),
			LoanCode:                       row.LoanCode,
			ProjectName:                    row.ProjectName,
			LenderName:                     row.LenderName,
			EffectiveDate:                  dateString(row.EffectiveDate),
			ClosingDate:                    dateString(row.ClosingDate),
			CommitmentUSD:                  floatFromNumeric(row.CommitmentUsd),
			CumulativeRealizedUSD:          floatFromNumeric(row.CumulativeRealizedUsd),
			UndisbursedUSD:                 floatFromNumeric(row.UndisbursedUsd),
			LAAbsorptionPct:                floatFromNumeric(row.LaAbsorptionPct),
			TimeElapsedPct:                 floatFromNumeric(row.TimeElapsedPct),
			AbsorptionGapPct:               floatFromNumeric(row.AbsorptionGapPct),
			RemainingMonths:                floatFromNumeric(row.RemainingMonths),
			RequiredMonthlyDisbursementUSD: floatFromNumeric(row.RequiredMonthlyDisbursementUsd),
			MonitoringCount:                int(row.MonitoringCount),
			IsExtended:                     row.IsExtended,
			RiskType:                       row.RiskType,
			RiskLevel:                      row.RiskLevel,
		})
	}
	return items
}

func laComponentBreakdown(rows []queries.GetDashboardLAComponentBreakdownRow) []model.LAComponentBreakdownItem {
	items := make([]model.LAComponentBreakdownItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.LAComponentBreakdownItem{
			ComponentName: row.ComponentName,
			LACount:       int(row.LaCount),
			PlannedUSD:    floatFromNumeric(row.PlannedUsd),
			RealizedUSD:   floatFromNumeric(row.RealizedUsd),
			AbsorptionPct: floatFromNumeric(row.AbsorptionPct),
		})
	}
	return items
}

func dataQualityIssues(rows []queries.ListDashboardDataQualityIssuesRow) []model.DataQualityIssueItem {
	items := make([]model.DataQualityIssueItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.DataQualityIssueItem{
			Severity:          row.Severity,
			Module:            row.Module,
			IssueType:         row.IssueType,
			RecordID:          model.UUIDToString(row.RecordID),
			RecordLabel:       row.RecordLabel,
			Message:           row.Message,
			RecommendedAction: row.RecommendedAction,
			IsResolved:        row.IsResolved,
		})
	}
	return items
}

func dataQualityIssueSummary(items []model.DataQualityIssueItem, auditEvents *int) model.DataQualityIssueSummary {
	summary := model.DataQualityIssueSummary{
		TotalIssues: len(items),
		AuditEvents: auditEvents,
	}
	for _, item := range items {
		switch item.Severity {
		case "error":
			summary.ErrorCount++
		case "warning":
			summary.WarningCount++
		case "info":
			summary.InfoCount++
		}
	}
	return summary
}

func auditSummaryByUser(rows []queries.GetDashboardAuditSummaryByUserRow) []model.AuditSummaryItem {
	items := make([]model.AuditSummaryItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.AuditSummaryItem{
			Label:         row.Label,
			EventCount:    int(row.EventCount),
			LastChangedAt: formatMasterTime(row.LastChangedAt),
		})
	}
	return items
}

func auditSummaryByTable(rows []queries.GetDashboardAuditSummaryByTableRow) []model.AuditSummaryItem {
	items := make([]model.AuditSummaryItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.AuditSummaryItem{
			Label:         row.Label,
			EventCount:    int(row.EventCount),
			LastChangedAt: formatMasterTime(row.LastChangedAt),
		})
	}
	return items
}

func auditRecentActivity(rows []queries.ListDashboardAuditRecentActivityRow) []model.AuditRecentActivityItem {
	items := make([]model.AuditRecentActivityItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.AuditRecentActivityItem{
			ID:        model.UUIDToString(row.ID),
			Username:  row.Username,
			Action:    row.Action,
			TableName: row.TableName,
			RecordID:  model.UUIDToString(row.RecordID),
			ChangedAt: formatMasterTime(row.ChangedAt),
		})
	}
	return items
}

var dataQualityIssueTypes = map[string]struct{}{
	"BB_WITHOUT_BAPPENAS_PARTNER":               {},
	"BB_INDICATION_WITHOUT_LOI":                 {},
	"LOI_WITHOUT_GB":                            {},
	"GB_WITHOUT_BB_REFERENCE":                   {},
	"GB_WITHOUT_FUNDING_SOURCE":                 {},
	"GB_WITHOUT_DISBURSEMENT_PLAN":              {},
	"GB_WITHOUT_ACTIVITY":                       {},
	"DK_WITHOUT_FINANCING_DETAIL":               {},
	"DK_WITHOUT_ACTIVITY_DETAIL":                {},
	"DK_WITHOUT_LA":                             {},
	"LA_NOT_EFFECTIVE":                          {},
	"EFFECTIVE_LA_WITHOUT_MONITORING":           {},
	"MONITORING_PLANNED_ZERO_REALIZED_POSITIVE": {},
	"MONITORING_COMPONENT_NAME_EMPTY":           {},
	"CURRENCY_USD_MISMATCH":                     {},
	"CLOSING_DATE_SOON_NO_RECENT_MONITORING":    {},
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

var greenBookReadinessStatusLabels = map[string]string{
	"READY":       "Ready",
	"PARTIAL":     "Partial",
	"INCOMPLETE":  "Incomplete",
	"COFINANCING": "Cofinancing",
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
