package service

import (
	"context"
	"sort"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

type DashboardAnalyticsService struct {
	queries *queries.Queries
}

const (
	defaultLowAbsorptionThreshold  = 50.0
	defaultClosingMonthsThreshold  = int32(12)
	defaultStaleMonitoringQuarters = int32(1)
	closingRiskAbsorptionThreshold = 80.0
)

type dashboardAnalyticsQueryValues struct {
	BudgetYear       pgtype.Int4
	Quarter          pgtype.Text
	LenderIDs        []pgtype.UUID
	LenderTypes      []string
	InstitutionIDs   []pgtype.UUID
	PipelineStatuses []string
	ProjectStatuses  []string
	RegionIDs        []pgtype.UUID
	ProgramTitleIDs  []pgtype.UUID
	ForeignLoanMin   pgtype.Numeric
	ForeignLoanMax   pgtype.Numeric
	IncludeHistory   bool
}

func NewDashboardAnalyticsService(queries *queries.Queries) *DashboardAnalyticsService {
	return &DashboardAnalyticsService{queries: queries}
}

func (s *DashboardAnalyticsService) Overview(ctx context.Context, filter model.DashboardAnalyticsFilter) (*model.DashboardAnalyticsOverviewResponse, error) {
	values, err := newDashboardAnalyticsQueryValues(filter)
	if err != nil {
		return nil, err
	}
	portfolioRow, err := s.queries.GetDashboardAnalyticsOverviewPortfolio(ctx, values.overviewPortfolioParams())
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil ringkasan portfolio analytics")
	}
	performanceRow, err := s.queries.GetDashboardAnalyticsAgreementPerformanceSummary(ctx, values.agreementPerformanceParams())
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil ringkasan performa analytics")
	}
	funnelRows, err := s.queries.ListDashboardAnalyticsPipelineFunnel(ctx, values.pipelineFunnelParams())
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil funnel pipeline analytics")
	}

	planned := floatFromNumeric(performanceRow.TotalPlannedUsd)
	realized := floatFromNumeric(performanceRow.TotalRealizedUsd)
	return &model.DashboardAnalyticsOverviewResponse{
		Portfolio: model.DashboardAnalyticsPortfolioOverview{
			ProjectCount:            int(portfolioRow.ProjectCount),
			AssignmentCount:         int(portfolioRow.AssignmentCount),
			TotalPipelineLoanUSD:    floatFromNumeric(portfolioRow.TotalPipelineLoanUsd),
			TotalAgreementAmountUSD: floatFromNumeric(performanceRow.TotalAgreementAmountUsd),
			TotalPlannedUSD:         planned,
			TotalRealizedUSD:        realized,
			AbsorptionPct:           absorptionPct(planned, realized),
		},
		PipelineFunnel: dashboardAnalyticsPipelineFunnel(filter, funnelRows),
		TopInsights:    []model.DashboardAnalyticsInsight{},
		Drilldown:      dashboardAnalyticsDrilldown(filter, "projects", nil),
	}, nil
}

func (s *DashboardAnalyticsService) Institutions(ctx context.Context, filter model.DashboardAnalyticsFilter) (*model.DashboardAnalyticsInstitutionsResponse, error) {
	values, err := newDashboardAnalyticsQueryValues(filter)
	if err != nil {
		return nil, err
	}
	portfolioRow, err := s.queries.GetDashboardAnalyticsOverviewPortfolio(ctx, values.overviewPortfolioParams())
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil ringkasan portfolio analytics")
	}
	performanceRow, err := s.queries.GetDashboardAnalyticsAgreementPerformanceSummary(ctx, values.agreementPerformanceParams())
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil ringkasan performa Kementerian/Lembaga")
	}
	rows, err := s.queries.ListDashboardAnalyticsInstitutions(ctx, values.institutionsParams())
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil analytics Kementerian/Lembaga")
	}

	items := make([]model.DashboardAnalyticsInstitutionItem, 0, len(rows))
	for _, row := range rows {
		planned := floatFromNumeric(row.PlannedUsd)
		realized := floatFromNumeric(row.RealizedUsd)
		id := model.UUIDToString(row.InstitutionID)
		items = append(items, model.DashboardAnalyticsInstitutionItem{
			Institution: model.DashboardAnalyticsEntityRef{
				ID:        id,
				Name:      row.InstitutionName,
				ShortName: stringPtrFromText(row.InstitutionShortName),
				Level:     row.InstitutionLevel,
			},
			ProjectCount:       int(row.PortfolioProjectCount),
			AssignmentCount:    int(row.PortfolioAssignmentCount),
			LoanAgreementCount: int(row.LoanAgreementCount),
			MonitoringCount:    int(row.MonitoringCount),
			AgreementAmountUSD: floatFromNumeric(row.AgreementAmountUsd),
			PlannedUSD:         planned,
			RealizedUSD:        realized,
			AbsorptionPct:      absorptionPct(planned, realized),
			PipelineBreakdown: model.DashboardAnalyticsPipelineBreakdown{
				BB:         int(row.BbCount),
				GB:         int(row.GbCount),
				DK:         int(row.DkCount),
				LA:         int(row.LaCount),
				Monitoring: int(row.MonitoringPipelineCount),
			},
			Drilldown: dashboardAnalyticsDrilldown(filter, "projects", map[string][]string{
				"executing_agency_ids": {id},
			}),
		})
	}

	planned := floatFromNumeric(performanceRow.TotalPlannedUsd)
	realized := floatFromNumeric(performanceRow.TotalRealizedUsd)
	return &model.DashboardAnalyticsInstitutionsResponse{
		Summary: model.DashboardAnalyticsInstitutionSummary{
			InstitutionCount:        len(items),
			ProjectCount:            int(portfolioRow.ProjectCount),
			AssignmentCount:         int(portfolioRow.AssignmentCount),
			TotalAgreementAmountUSD: floatFromNumeric(performanceRow.TotalAgreementAmountUsd),
			TotalPlannedUSD:         planned,
			TotalRealizedUSD:        realized,
			AbsorptionPct:           absorptionPct(planned, realized),
		},
		Items:     items,
		Drilldown: dashboardAnalyticsDrilldown(filter, "projects", nil),
	}, nil
}

func (s *DashboardAnalyticsService) Lenders(ctx context.Context, filter model.DashboardAnalyticsFilter) (*model.DashboardAnalyticsLendersResponse, error) {
	values, err := newDashboardAnalyticsQueryValues(filter)
	if err != nil {
		return nil, err
	}
	performanceRow, err := s.queries.GetDashboardAnalyticsAgreementPerformanceSummary(ctx, values.agreementPerformanceParams())
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil ringkasan performa lender")
	}
	lenderRows, err := s.queries.ListDashboardAnalyticsLenders(ctx, values.lendersParams())
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil analytics lender")
	}
	matrixRows, err := s.queries.ListDashboardAnalyticsLenderInstitutionMatrix(ctx, values.lenderInstitutionMatrixParams())
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil matrix lender Kementerian/Lembaga")
	}

	items := make([]model.DashboardAnalyticsLenderItem, 0, len(lenderRows))
	for _, row := range lenderRows {
		planned := floatFromNumeric(row.PlannedUsd)
		realized := floatFromNumeric(row.RealizedUsd)
		id := model.UUIDToString(row.LenderID)
		items = append(items, model.DashboardAnalyticsLenderItem{
			Lender: model.DashboardAnalyticsEntityRef{
				ID:        id,
				Name:      row.LenderName,
				ShortName: stringPtrFromText(row.LenderShortName),
				Type:      row.LenderType,
			},
			LoanAgreementCount: int(row.LoanAgreementCount),
			ProjectCount:       int(row.ProjectCount),
			InstitutionCount:   int(row.InstitutionCount),
			MonitoringCount:    int(row.MonitoringCount),
			AgreementAmountUSD: floatFromNumeric(row.AgreementAmountUsd),
			PlannedUSD:         planned,
			RealizedUSD:        realized,
			AbsorptionPct:      absorptionPct(planned, realized),
			Drilldown: dashboardAnalyticsDrilldown(filter, "projects", map[string][]string{
				"fixed_lender_ids": {id},
			}),
		})
	}

	matrix := make([]model.DashboardAnalyticsLenderInstitutionMatrixItem, 0, len(matrixRows))
	for _, row := range matrixRows {
		planned := floatFromNumeric(row.PlannedUsd)
		realized := floatFromNumeric(row.RealizedUsd)
		institutionID := model.UUIDToString(row.InstitutionID)
		lenderID := model.UUIDToString(row.LenderID)
		matrix = append(matrix, model.DashboardAnalyticsLenderInstitutionMatrixItem{
			Institution: model.DashboardAnalyticsEntityRef{
				ID:        institutionID,
				Name:      textValue(row.InstitutionName),
				ShortName: stringPtrFromText(row.InstitutionShortName),
				Level:     textValue(row.InstitutionLevel),
			},
			Lender: model.DashboardAnalyticsEntityRef{
				ID:        lenderID,
				Name:      row.LenderName,
				ShortName: stringPtrFromText(row.LenderShortName),
				Type:      row.LenderType,
			},
			ProjectCount:       int(row.ProjectCount),
			LoanAgreementCount: int(row.LoanAgreementCount),
			MonitoringCount:    int(row.MonitoringCount),
			AgreementAmountUSD: floatFromNumeric(row.AgreementAmountUsd),
			PlannedUSD:         planned,
			RealizedUSD:        realized,
			AbsorptionPct:      absorptionPct(planned, realized),
			Drilldown: dashboardAnalyticsDrilldown(filter, "projects", map[string][]string{
				"executing_agency_ids": {institutionID},
				"fixed_lender_ids":     {lenderID},
			}),
		})
	}

	planned := floatFromNumeric(performanceRow.TotalPlannedUsd)
	realized := floatFromNumeric(performanceRow.TotalRealizedUsd)
	return &model.DashboardAnalyticsLendersResponse{
		Summary: model.DashboardAnalyticsLenderSummary{
			LenderCount:             len(items),
			LoanAgreementCount:      int(performanceRow.LoanAgreementCount),
			TotalAgreementAmountUSD: floatFromNumeric(performanceRow.TotalAgreementAmountUsd),
			TotalPlannedUSD:         planned,
			TotalRealizedUSD:        realized,
			AbsorptionPct:           absorptionPct(planned, realized),
		},
		Items:                   items,
		LenderInstitutionMatrix: matrix,
		Drilldown:               dashboardAnalyticsDrilldown(filter, "projects", nil),
	}, nil
}

func (s *DashboardAnalyticsService) Absorption(ctx context.Context, filter model.DashboardAnalyticsFilter) (*model.DashboardAnalyticsAbsorptionResponse, error) {
	values, err := newDashboardAnalyticsQueryValues(filter)
	if err != nil {
		return nil, err
	}
	summaryRow, err := s.queries.GetDashboardAnalyticsAgreementPerformanceSummary(ctx, values.agreementPerformanceParams())
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil ringkasan penyerapan analytics")
	}
	institutionRows, err := s.queries.ListDashboardAnalyticsAbsorptionByInstitution(ctx, values.absorptionByInstitutionParams())
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil penyerapan by Kementerian/Lembaga")
	}
	projectRows, err := s.queries.ListDashboardAnalyticsAbsorptionByProject(ctx, values.absorptionByProjectParams())
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil penyerapan by project")
	}
	lenderRows, err := s.queries.ListDashboardAnalyticsAbsorptionByLender(ctx, values.absorptionByLenderParams())
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil penyerapan by lender")
	}

	planned := floatFromNumeric(summaryRow.TotalPlannedUsd)
	realized := floatFromNumeric(summaryRow.TotalRealizedUsd)
	return &model.DashboardAnalyticsAbsorptionResponse{
		Summary: model.DashboardAnalyticsAbsorptionSummary{
			PlannedUSD:    planned,
			RealizedUSD:   realized,
			AbsorptionPct: absorptionPct(planned, realized),
		},
		ByInstitution: absorptionInstitutionItems(filter, institutionRows),
		ByProject:     absorptionProjectItems(filter, projectRows),
		ByLender:      absorptionLenderItems(filter, lenderRows),
		Drilldown:     dashboardAnalyticsDrilldown(filter, "monitoring", nil),
	}, nil
}

func (s *DashboardAnalyticsService) Yearly(ctx context.Context, filter model.DashboardAnalyticsFilter) (*model.DashboardAnalyticsYearlyResponse, error) {
	values, err := newDashboardAnalyticsQueryValues(filter)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListDashboardAnalyticsYearlyPerformance(ctx, values.yearlyPerformanceParams())
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil performa tahunan analytics")
	}

	items := make([]model.DashboardAnalyticsYearlyItem, 0, len(rows))
	var plannedTotal, realizedTotal float64
	loanAgreementKeys := 0
	projectKeys := 0
	for _, row := range rows {
		planned := floatFromNumeric(row.PlannedUsd)
		realized := floatFromNumeric(row.RealizedUsd)
		plannedTotal += planned
		realizedTotal += realized
		loanAgreementKeys += int(row.LoanAgreementCount)
		projectKeys += int(row.ProjectCount)
		items = append(items, model.DashboardAnalyticsYearlyItem{
			BudgetYear:         row.BudgetYear,
			Quarter:            row.Quarter,
			PlannedUSD:         planned,
			RealizedUSD:        realized,
			AbsorptionPct:      absorptionPct(planned, realized),
			LoanAgreementCount: int(row.LoanAgreementCount),
			ProjectCount:       int(row.ProjectCount),
			Drilldown: dashboardAnalyticsDrilldown(filter, "monitoring", map[string][]string{
				"budget_year": {strconv.FormatInt(int64(row.BudgetYear), 10)},
				"quarter":     {row.Quarter},
			}),
		})
	}

	return &model.DashboardAnalyticsYearlyResponse{
		Summary: model.DashboardAnalyticsYearlySummary{
			PlannedUSD:         plannedTotal,
			RealizedUSD:        realizedTotal,
			AbsorptionPct:      absorptionPct(plannedTotal, realizedTotal),
			LoanAgreementCount: loanAgreementKeys,
			ProjectCount:       projectKeys,
		},
		Items:     items,
		Drilldown: dashboardAnalyticsDrilldown(filter, "monitoring", nil),
	}, nil
}

func (s *DashboardAnalyticsService) LenderProportion(ctx context.Context, filter model.DashboardAnalyticsFilter) (*model.DashboardAnalyticsLenderProportionResponse, error) {
	values, err := newDashboardAnalyticsQueryValues(filter)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListDashboardAnalyticsLenderProportion(ctx, values.lenderProportionParams())
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil proporsi lender analytics")
	}
	return &model.DashboardAnalyticsLenderProportionResponse{
		ByStage:   lenderProportionByStage(filter, rows),
		Drilldown: dashboardAnalyticsDrilldown(filter, "projects", nil),
	}, nil
}

func (s *DashboardAnalyticsService) Risks(ctx context.Context, filter model.DashboardAnalyticsFilter) (*model.DashboardAnalyticsRisksResponse, error) {
	values, err := newDashboardAnalyticsQueryValues(filter)
	if err != nil {
		return nil, err
	}
	thresholds := dashboardAnalyticsRiskThresholds(filter)
	riskRows, err := s.queries.ListDashboardAnalyticsRiskWatchlist(ctx, values.riskWatchlistParams(thresholds))
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil risk watchlist analytics")
	}
	bottleneckRows, err := s.queries.ListDashboardAnalyticsPipelineBottlenecks(ctx, values.pipelineBottleneckParams())
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil pipeline bottleneck analytics")
	}
	dataQualityRows, err := s.queries.ListDashboardAnalyticsDataQualityIssues(ctx, values.dataQualityParams())
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil data quality analytics")
	}

	watchlists := model.DashboardAnalyticsRiskWatchlists{
		LowAbsorptionProjects:      []model.DashboardAnalyticsLoanAgreementRiskItem{},
		EffectiveWithoutMonitoring: []model.DashboardAnalyticsLoanAgreementRiskItem{},
		ClosingRisks:               []model.DashboardAnalyticsLoanAgreementRiskItem{},
		ExtendedLoans:              []model.DashboardAnalyticsLoanAgreementRiskItem{},
		PipelineBottlenecks:        dashboardAnalyticsPipelineBottlenecks(filter, bottleneckRows),
	}

	for _, row := range riskRows {
		item := dashboardAnalyticsLoanAgreementRiskItem(filter, row)
		switch row.RiskCode {
		case "LOW_ABSORPTION":
			watchlists.LowAbsorptionProjects = append(watchlists.LowAbsorptionProjects, item)
		case "EFFECTIVE_WITHOUT_MONITORING":
			watchlists.EffectiveWithoutMonitoring = append(watchlists.EffectiveWithoutMonitoring, item)
		case "CLOSING_RISK":
			watchlists.ClosingRisks = append(watchlists.ClosingRisks, item)
		case "EXTENDED_LOAN":
			watchlists.ExtendedLoans = append(watchlists.ExtendedLoans, item)
		}
	}

	dataQuality := dashboardAnalyticsDataQualityItems(filter, dataQualityRows)
	extendedInsight := dashboardAnalyticsExtendedLoanInsight(filter, watchlists.ExtendedLoans)
	summary := model.DashboardAnalyticsRiskSummary{
		LowAbsorptionCount:              len(watchlists.LowAbsorptionProjects),
		EffectiveWithoutMonitoringCount: len(watchlists.EffectiveWithoutMonitoring),
		ClosingRiskCount:                len(watchlists.ClosingRisks),
		ExtendedLoanCount:               len(watchlists.ExtendedLoans),
		DataQualityIssueCount:           dashboardAnalyticsDataQualityAffectedCount(dataQuality),
		BottleneckProjectCount:          dashboardAnalyticsBottleneckProjectCount(watchlists.PipelineBottlenecks),
	}

	return &model.DashboardAnalyticsRisksResponse{
		Summary:             summary,
		Thresholds:          thresholds,
		RiskCards:           dashboardAnalyticsRiskCards(filter, summary, extendedInsight),
		Watchlists:          watchlists,
		ExtendedLoanInsight: extendedInsight,
		DataQuality:         dataQuality,
		Drilldown:           dashboardAnalyticsDrilldown(filter, "projects", nil),
	}, nil
}

func newDashboardAnalyticsQueryValues(filter model.DashboardAnalyticsFilter) (dashboardAnalyticsQueryValues, error) {
	lenderIDs, err := uuidArray(filter.LenderIDs, "lender_ids")
	if err != nil {
		return dashboardAnalyticsQueryValues{}, err
	}
	institutionIDs, err := uuidArray(filter.InstitutionIDs, "institution_ids")
	if err != nil {
		return dashboardAnalyticsQueryValues{}, err
	}
	regionIDs, err := uuidArray(filter.RegionIDs, "region_ids")
	if err != nil {
		return dashboardAnalyticsQueryValues{}, err
	}
	programTitleIDs, err := uuidArray(filter.ProgramTitleIDs, "program_title_ids")
	if err != nil {
		return dashboardAnalyticsQueryValues{}, err
	}

	values := dashboardAnalyticsQueryValues{
		LenderIDs:        lenderIDs,
		LenderTypes:      filter.LenderTypes,
		InstitutionIDs:   institutionIDs,
		PipelineStatuses: filter.PipelineStatuses,
		ProjectStatuses:  filter.ProjectStatuses,
		RegionIDs:        regionIDs,
		ProgramTitleIDs:  programTitleIDs,
		ForeignLoanMin:   optionalFloatNumeric(filter.ForeignLoanMin),
		ForeignLoanMax:   optionalFloatNumeric(filter.ForeignLoanMax),
		IncludeHistory:   filter.IncludeHistory,
	}
	if filter.BudgetYear != nil {
		values.BudgetYear = pgtype.Int4{Int32: *filter.BudgetYear, Valid: true}
	}
	if filter.Quarter != nil && *filter.Quarter != "" {
		values.Quarter = pgtype.Text{String: *filter.Quarter, Valid: true}
	}
	return values, nil
}

func (v dashboardAnalyticsQueryValues) overviewPortfolioParams() queries.GetDashboardAnalyticsOverviewPortfolioParams {
	return queries.GetDashboardAnalyticsOverviewPortfolioParams{
		IncludeHistory:   v.IncludeHistory,
		LenderTypes:      v.LenderTypes,
		LenderIds:        v.LenderIDs,
		InstitutionIds:   v.InstitutionIDs,
		PipelineStatuses: v.PipelineStatuses,
		ProjectStatuses:  v.ProjectStatuses,
		RegionIds:        v.RegionIDs,
		ProgramTitleIds:  v.ProgramTitleIDs,
		ForeignLoanMin:   v.ForeignLoanMin,
		ForeignLoanMax:   v.ForeignLoanMax,
	}
}

func (v dashboardAnalyticsQueryValues) pipelineFunnelParams() queries.ListDashboardAnalyticsPipelineFunnelParams {
	return queries.ListDashboardAnalyticsPipelineFunnelParams(v.overviewPortfolioParams())
}

func (v dashboardAnalyticsQueryValues) agreementPerformanceParams() queries.GetDashboardAnalyticsAgreementPerformanceSummaryParams {
	return queries.GetDashboardAnalyticsAgreementPerformanceSummaryParams{
		BudgetYear:       v.BudgetYear,
		Quarter:          v.Quarter,
		LenderIds:        v.LenderIDs,
		LenderTypes:      v.LenderTypes,
		InstitutionIds:   v.InstitutionIDs,
		PipelineStatuses: v.PipelineStatuses,
		ProjectStatuses:  v.ProjectStatuses,
		ForeignLoanMin:   v.ForeignLoanMin,
		ForeignLoanMax:   v.ForeignLoanMax,
		ProgramTitleIds:  v.ProgramTitleIDs,
		RegionIds:        v.RegionIDs,
	}
}

func (v dashboardAnalyticsQueryValues) institutionsParams() queries.ListDashboardAnalyticsInstitutionsParams {
	return queries.ListDashboardAnalyticsInstitutionsParams{
		IncludeHistory:   v.IncludeHistory,
		LenderTypes:      v.LenderTypes,
		LenderIds:        v.LenderIDs,
		InstitutionIds:   v.InstitutionIDs,
		PipelineStatuses: v.PipelineStatuses,
		ProjectStatuses:  v.ProjectStatuses,
		RegionIds:        v.RegionIDs,
		ProgramTitleIds:  v.ProgramTitleIDs,
		ForeignLoanMin:   v.ForeignLoanMin,
		ForeignLoanMax:   v.ForeignLoanMax,
		BudgetYear:       v.BudgetYear,
		Quarter:          v.Quarter,
	}
}

func (v dashboardAnalyticsQueryValues) lendersParams() queries.ListDashboardAnalyticsLendersParams {
	return queries.ListDashboardAnalyticsLendersParams(v.agreementPerformanceParams())
}

func (v dashboardAnalyticsQueryValues) lenderInstitutionMatrixParams() queries.ListDashboardAnalyticsLenderInstitutionMatrixParams {
	return queries.ListDashboardAnalyticsLenderInstitutionMatrixParams(v.agreementPerformanceParams())
}

func (v dashboardAnalyticsQueryValues) absorptionByInstitutionParams() queries.ListDashboardAnalyticsAbsorptionByInstitutionParams {
	return queries.ListDashboardAnalyticsAbsorptionByInstitutionParams(v.agreementPerformanceParams())
}

func (v dashboardAnalyticsQueryValues) absorptionByProjectParams() queries.ListDashboardAnalyticsAbsorptionByProjectParams {
	return queries.ListDashboardAnalyticsAbsorptionByProjectParams(v.agreementPerformanceParams())
}

func (v dashboardAnalyticsQueryValues) absorptionByLenderParams() queries.ListDashboardAnalyticsAbsorptionByLenderParams {
	return queries.ListDashboardAnalyticsAbsorptionByLenderParams(v.agreementPerformanceParams())
}

func (v dashboardAnalyticsQueryValues) yearlyPerformanceParams() queries.ListDashboardAnalyticsYearlyPerformanceParams {
	return queries.ListDashboardAnalyticsYearlyPerformanceParams(v.agreementPerformanceParams())
}

func (v dashboardAnalyticsQueryValues) lenderProportionParams() queries.ListDashboardAnalyticsLenderProportionParams {
	return queries.ListDashboardAnalyticsLenderProportionParams{
		IncludeHistory:   v.IncludeHistory,
		LenderIds:        v.LenderIDs,
		LenderTypes:      v.LenderTypes,
		InstitutionIds:   v.InstitutionIDs,
		PipelineStatuses: v.PipelineStatuses,
		ProjectStatuses:  v.ProjectStatuses,
		RegionIds:        v.RegionIDs,
		ProgramTitleIds:  v.ProgramTitleIDs,
		ForeignLoanMin:   v.ForeignLoanMin,
		ForeignLoanMax:   v.ForeignLoanMax,
		BudgetYear:       v.BudgetYear,
		Quarter:          v.Quarter,
	}
}

func (v dashboardAnalyticsQueryValues) riskWatchlistParams(thresholds model.DashboardAnalyticsRiskThresholds) queries.ListDashboardAnalyticsRiskWatchlistParams {
	return queries.ListDashboardAnalyticsRiskWatchlistParams{
		BudgetYear:              v.BudgetYear,
		Quarter:                 v.Quarter,
		LenderIds:               v.LenderIDs,
		LenderTypes:             v.LenderTypes,
		InstitutionIds:          v.InstitutionIDs,
		PipelineStatuses:        v.PipelineStatuses,
		ProjectStatuses:         v.ProjectStatuses,
		ForeignLoanMin:          v.ForeignLoanMin,
		ForeignLoanMax:          v.ForeignLoanMax,
		ProgramTitleIds:         v.ProgramTitleIDs,
		RegionIds:               v.RegionIDs,
		LowAbsorptionThreshold:  thresholds.LowAbsorptionThreshold,
		StaleMonitoringQuarters: thresholds.StaleMonitoringQuarters,
		ClosingMonthsThreshold:  thresholds.ClosingMonthsThreshold,
	}
}

func (v dashboardAnalyticsQueryValues) pipelineBottleneckParams() queries.ListDashboardAnalyticsPipelineBottlenecksParams {
	return queries.ListDashboardAnalyticsPipelineBottlenecksParams{
		PipelineStatuses: v.PipelineStatuses,
		ProjectStatuses:  v.ProjectStatuses,
		LenderIds:        v.LenderIDs,
		LenderTypes:      v.LenderTypes,
		InstitutionIds:   v.InstitutionIDs,
		ProgramTitleIds:  v.ProgramTitleIDs,
		RegionIds:        v.RegionIDs,
		ForeignLoanMin:   v.ForeignLoanMin,
		ForeignLoanMax:   v.ForeignLoanMax,
		IncludeHistory:   v.IncludeHistory,
	}
}

func (v dashboardAnalyticsQueryValues) dataQualityParams() queries.ListDashboardAnalyticsDataQualityIssuesParams {
	return queries.ListDashboardAnalyticsDataQualityIssuesParams{
		IncludeHistory:  v.IncludeHistory,
		LenderIds:       v.LenderIDs,
		LenderTypes:     v.LenderTypes,
		InstitutionIds:  v.InstitutionIDs,
		ProgramTitleIds: v.ProgramTitleIDs,
		RegionIds:       v.RegionIDs,
		ForeignLoanMin:  v.ForeignLoanMin,
		ForeignLoanMax:  v.ForeignLoanMax,
		BudgetYear:      v.BudgetYear,
		Quarter:         v.Quarter,
	}
}

func dashboardAnalyticsPipelineFunnel(filter model.DashboardAnalyticsFilter, rows []queries.ListDashboardAnalyticsPipelineFunnelRow) []model.DashboardAnalyticsPipelineFunnelItem {
	items := make([]model.DashboardAnalyticsPipelineFunnelItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.DashboardAnalyticsPipelineFunnelItem{
			Stage:        row.Stage,
			ProjectCount: int(row.ProjectCount),
			TotalLoanUSD: floatFromNumeric(row.TotalLoanUsd),
			Drilldown: dashboardAnalyticsDrilldown(filter, "projects", map[string][]string{
				"pipeline_statuses": {row.Stage},
			}),
		})
	}
	return items
}

func absorptionInstitutionItems(filter model.DashboardAnalyticsFilter, rows []queries.ListDashboardAnalyticsAbsorptionByInstitutionRow) []model.DashboardAnalyticsAbsorptionItem {
	items := make([]model.DashboardAnalyticsAbsorptionItem, 0, len(rows))
	for index, row := range rows {
		id := model.UUIDToString(row.ID)
		items = append(items, dashboardAnalyticsAbsorptionItem(filter, index, id, textValue(row.Name), row.Dimension, row.PlannedUsd, row.RealizedUsd, map[string][]string{
			"executing_agency_ids": {id},
		}))
	}
	return items
}

func absorptionProjectItems(filter model.DashboardAnalyticsFilter, rows []queries.ListDashboardAnalyticsAbsorptionByProjectRow) []model.DashboardAnalyticsAbsorptionItem {
	items := make([]model.DashboardAnalyticsAbsorptionItem, 0, len(rows))
	for index, row := range rows {
		id := model.UUIDToString(row.ID)
		items = append(items, dashboardAnalyticsAbsorptionItem(filter, index, id, row.Name, row.Dimension, row.PlannedUsd, row.RealizedUsd, map[string][]string{
			"dk_project_ids": {id},
		}))
	}
	return items
}

func absorptionLenderItems(filter model.DashboardAnalyticsFilter, rows []queries.ListDashboardAnalyticsAbsorptionByLenderRow) []model.DashboardAnalyticsAbsorptionItem {
	items := make([]model.DashboardAnalyticsAbsorptionItem, 0, len(rows))
	for index, row := range rows {
		id := model.UUIDToString(row.ID)
		items = append(items, dashboardAnalyticsAbsorptionItem(filter, index, id, row.Name, row.Dimension, row.PlannedUsd, row.RealizedUsd, map[string][]string{
			"lender_ids": {id},
		}))
	}
	return items
}

func dashboardAnalyticsAbsorptionItem(filter model.DashboardAnalyticsFilter, index int, id, name, dimension string, plannedNumeric, realizedNumeric pgtype.Numeric, drilldown map[string][]string) model.DashboardAnalyticsAbsorptionItem {
	planned := floatFromNumeric(plannedNumeric)
	realized := floatFromNumeric(realizedNumeric)
	pct := absorptionPct(planned, realized)
	return model.DashboardAnalyticsAbsorptionItem{
		Rank:          index + 1,
		ID:            id,
		Name:          name,
		Dimension:     dimension,
		PlannedUSD:    planned,
		RealizedUSD:   realized,
		AbsorptionPct: pct,
		VarianceUSD:   planned - realized,
		Status:        dashboardAnalyticsAbsorptionStatus(pct),
		Drilldown:     dashboardAnalyticsDrilldown(filter, "monitoring", drilldown),
	}
}

func lenderProportionByStage(filter model.DashboardAnalyticsFilter, rows []queries.ListDashboardAnalyticsLenderProportionRow) []model.DashboardAnalyticsLenderProportionStage {
	stageTotals := map[string]float64{}
	for _, row := range rows {
		stageTotals[row.Stage] += floatFromNumeric(row.AmountUsd)
	}

	stageOrder := []string{}
	stageItems := map[string][]model.DashboardAnalyticsLenderProportionItem{}
	seenStage := map[string]struct{}{}
	for _, row := range rows {
		if _, ok := seenStage[row.Stage]; !ok {
			seenStage[row.Stage] = struct{}{}
			stageOrder = append(stageOrder, row.Stage)
		}
		amount := floatFromNumeric(row.AmountUsd)
		stageItems[row.Stage] = append(stageItems[row.Stage], model.DashboardAnalyticsLenderProportionItem{
			Type:         row.LenderType,
			ProjectCount: int(row.ProjectCount),
			LenderCount:  int(row.LenderCount),
			AmountUSD:    amount,
			SharePct:     absorptionPct(stageTotals[row.Stage], amount),
			Drilldown: dashboardAnalyticsDrilldown(filter, "projects", map[string][]string{
				"lender_types": {row.LenderType},
			}),
		})
	}

	result := make([]model.DashboardAnalyticsLenderProportionStage, 0, len(stageOrder))
	for _, stage := range stageOrder {
		result = append(result, model.DashboardAnalyticsLenderProportionStage{
			Stage: stage,
			Items: stageItems[stage],
		})
	}
	return result
}

func dashboardAnalyticsRiskThresholds(filter model.DashboardAnalyticsFilter) model.DashboardAnalyticsRiskThresholds {
	thresholds := model.DashboardAnalyticsRiskThresholds{
		LowAbsorptionThreshold:     defaultLowAbsorptionThreshold,
		ClosingMonthsThreshold:     defaultClosingMonthsThreshold,
		ClosingAbsorptionThreshold: closingRiskAbsorptionThreshold,
		StaleMonitoringQuarters:    defaultStaleMonitoringQuarters,
	}
	if filter.LowAbsorptionThreshold != nil {
		thresholds.LowAbsorptionThreshold = *filter.LowAbsorptionThreshold
	}
	if filter.ClosingMonthsThreshold != nil {
		thresholds.ClosingMonthsThreshold = *filter.ClosingMonthsThreshold
	}
	if filter.StaleMonitoringQuarters != nil {
		thresholds.StaleMonitoringQuarters = *filter.StaleMonitoringQuarters
	}
	return thresholds
}

func dashboardAnalyticsLoanAgreementRiskItem(filter model.DashboardAnalyticsFilter, row queries.ListDashboardAnalyticsRiskWatchlistRow) model.DashboardAnalyticsLoanAgreementRiskItem {
	extra := map[string][]string{
		"risk_codes":        {row.RiskCode},
		"loan_agreement_id": {model.UUIDToString(row.LoanAgreementID)},
	}
	if row.BudgetYear.Valid {
		extra["budget_year"] = []string{strconv.FormatInt(int64(row.BudgetYear.Int32), 10)}
	}
	if row.Quarter.Valid {
		extra["quarter"] = []string{row.Quarter.String}
	}

	target := "monitoring"
	if row.RiskCode == "CLOSING_RISK" || row.RiskCode == "EXTENDED_LOAN" {
		target = "loan_agreements"
	}

	item := model.DashboardAnalyticsLoanAgreementRiskItem{
		RiskCode:            row.RiskCode,
		RiskLabel:           row.RiskLabel,
		Severity:            row.Severity,
		ProjectID:           model.UUIDToString(row.DkProjectID),
		ProjectName:         row.ProjectName,
		LoanAgreementID:     model.UUIDToString(row.LoanAgreementID),
		LoanCode:            row.LoanCode,
		Lender:              dashboardAnalyticsLenderRef(row.LenderID, row.LenderName, row.LenderShortName, row.LenderType),
		EffectiveDate:       dateString(row.EffectiveDate),
		OriginalClosingDate: dateString(row.OriginalClosingDate),
		ClosingDate:         dateString(row.ClosingDate),
		BudgetYear:          int32PtrFromInt4(row.BudgetYear),
		Quarter:             stringPtrFromText(row.Quarter),
		PlannedUSD:          floatFromNumeric(row.PlannedUsd),
		RealizedUSD:         floatFromNumeric(row.RealizedUsd),
		AbsorptionPct:       row.AbsorptionPct,
		AgreementAmountUSD:  floatFromNumeric(row.AgreementAmountUsd),
		DaysSinceEffective:  int(row.DaysSinceEffective),
		DaysToClosing:       int(row.DaysToClosing),
		MonthsToClosing:     int(row.MonthsToClosing),
		ExtensionDays:       int(row.ExtensionDays),
		StaleQuarters:       int(row.StaleQuarters),
		MonitoringStatus:    row.MonitoringStatus,
		Drilldown:           dashboardAnalyticsDrilldown(filter, target, extra),
	}
	if row.InstitutionID.Valid {
		item.Institution = &model.DashboardAnalyticsEntityRef{
			ID:        model.UUIDToString(row.InstitutionID),
			Name:      textValue(row.InstitutionName),
			ShortName: stringPtrFromText(row.InstitutionShortName),
			Level:     textValue(row.InstitutionLevel),
		}
	}
	return item
}

func dashboardAnalyticsLenderRef(id pgtype.UUID, name string, shortName pgtype.Text, lenderType string) model.DashboardAnalyticsEntityRef {
	return model.DashboardAnalyticsEntityRef{
		ID:        model.UUIDToString(id),
		Name:      name,
		ShortName: stringPtrFromText(shortName),
		Type:      lenderType,
	}
}

func dashboardAnalyticsPipelineBottlenecks(filter model.DashboardAnalyticsFilter, rows []queries.ListDashboardAnalyticsPipelineBottlenecksRow) []model.DashboardAnalyticsPipelineBottleneckItem {
	items := make([]model.DashboardAnalyticsPipelineBottleneckItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.DashboardAnalyticsPipelineBottleneckItem{
			Stage:        row.Stage,
			Label:        row.Label,
			ProjectCount: int(row.ProjectCount),
			TotalLoanUSD: floatFromNumeric(row.TotalLoanUsd),
			OldestDate:   dateString(row.OldestDate),
			Severity:     row.Severity,
			Drilldown: dashboardAnalyticsDrilldown(filter, "projects", map[string][]string{
				"pipeline_statuses": {row.Stage},
			}),
		})
	}
	return items
}

func dashboardAnalyticsDataQualityItems(filter model.DashboardAnalyticsFilter, rows []queries.ListDashboardAnalyticsDataQualityIssuesRow) []model.DashboardAnalyticsDataQualityItem {
	items := make([]model.DashboardAnalyticsDataQualityItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, model.DashboardAnalyticsDataQualityItem{
			Code:     row.Code,
			Label:    row.Label,
			Stage:    row.Stage,
			Severity: row.Severity,
			Count:    int(row.AffectedCount),
			Drilldown: dashboardAnalyticsDrilldown(filter, row.Target, map[string][]string{
				"data_quality_codes":  {row.Code},
				"data_quality_stages": {row.Stage},
			}),
		})
	}
	return items
}

func dashboardAnalyticsRiskCards(filter model.DashboardAnalyticsFilter, summary model.DashboardAnalyticsRiskSummary, extended model.DashboardAnalyticsExtendedLoanInsight) []model.DashboardAnalyticsRiskCard {
	return []model.DashboardAnalyticsRiskCard{
		{
			Code:      "LOW_ABSORPTION",
			Label:     "Penyerapan rendah",
			Count:     summary.LowAbsorptionCount,
			Severity:  "warning",
			Drilldown: dashboardAnalyticsDrilldown(filter, "monitoring", map[string][]string{"risk_codes": {"LOW_ABSORPTION"}}),
		},
		{
			Code:      "EFFECTIVE_WITHOUT_MONITORING",
			Label:     "Loan Agreement efektif tanpa monitoring terkini",
			Count:     summary.EffectiveWithoutMonitoringCount,
			Severity:  "danger",
			Drilldown: dashboardAnalyticsDrilldown(filter, "monitoring", map[string][]string{"risk_codes": {"EFFECTIVE_WITHOUT_MONITORING"}}),
		},
		{
			Code:      "CLOSING_RISK",
			Label:     "Closing risk",
			Count:     summary.ClosingRiskCount,
			Severity:  "danger",
			Drilldown: dashboardAnalyticsDrilldown(filter, "loan_agreements", map[string][]string{"risk_codes": {"CLOSING_RISK"}}),
		},
		{
			Code:      "EXTENDED_LOAN",
			Label:     "Loan Agreement diperpanjang",
			Count:     summary.ExtendedLoanCount,
			Severity:  "info",
			AmountUSD: extended.AmountUSD,
			Drilldown: dashboardAnalyticsDrilldown(filter, "loan_agreements", map[string][]string{
				"risk_codes":  {"EXTENDED_LOAN"},
				"is_extended": {"true"},
			}),
		},
		{
			Code:      "PIPELINE_BOTTLENECK",
			Label:     "Project belum berlanjut",
			Count:     summary.BottleneckProjectCount,
			Severity:  "warning",
			Drilldown: dashboardAnalyticsDrilldown(filter, "projects", map[string][]string{"project_statuses": {"Pipeline"}}),
		},
		{
			Code:      "DATA_QUALITY",
			Label:     "Data quality issues",
			Count:     summary.DataQualityIssueCount,
			Severity:  "warning",
			Drilldown: dashboardAnalyticsDrilldown(filter, "projects", map[string][]string{"data_quality": {"true"}}),
		},
	}
}

func dashboardAnalyticsDataQualityAffectedCount(items []model.DashboardAnalyticsDataQualityItem) int {
	total := 0
	for _, item := range items {
		total += item.Count
	}
	return total
}

func dashboardAnalyticsBottleneckProjectCount(items []model.DashboardAnalyticsPipelineBottleneckItem) int {
	total := 0
	for _, item := range items {
		total += item.ProjectCount
	}
	return total
}

type dashboardAnalyticsExtendedAccumulator struct {
	entity    model.DashboardAnalyticsEntityRef
	count     int
	amount    float64
	totalDays int
	extra     map[string][]string
}

func dashboardAnalyticsExtendedLoanInsight(filter model.DashboardAnalyticsFilter, loans []model.DashboardAnalyticsLoanAgreementRiskItem) model.DashboardAnalyticsExtendedLoanInsight {
	insight := model.DashboardAnalyticsExtendedLoanInsight{
		ByLender:      []model.DashboardAnalyticsExtendedLoanBreakdown{},
		ByInstitution: []model.DashboardAnalyticsExtendedLoanBreakdown{},
	}
	lenderBreakdown := map[string]*dashboardAnalyticsExtendedAccumulator{}
	institutionBreakdown := map[string]*dashboardAnalyticsExtendedAccumulator{}

	totalExtensionDays := 0
	for _, loan := range loans {
		insight.Count++
		insight.AmountUSD += loan.AgreementAmountUSD
		totalExtensionDays += loan.ExtensionDays

		lenderID := loan.Lender.ID
		if lenderID != "" {
			acc := lenderBreakdown[lenderID]
			if acc == nil {
				acc = &dashboardAnalyticsExtendedAccumulator{
					entity: loan.Lender,
					extra:  map[string][]string{"fixed_lender_ids": {lenderID}, "is_extended": {"true"}},
				}
				lenderBreakdown[lenderID] = acc
			}
			acc.count++
			acc.amount += loan.AgreementAmountUSD
			acc.totalDays += loan.ExtensionDays
		}

		if loan.Institution != nil && loan.Institution.ID != "" {
			institutionID := loan.Institution.ID
			acc := institutionBreakdown[institutionID]
			if acc == nil {
				acc = &dashboardAnalyticsExtendedAccumulator{
					entity: *loan.Institution,
					extra:  map[string][]string{"executing_agency_ids": {institutionID}, "is_extended": {"true"}},
				}
				institutionBreakdown[institutionID] = acc
			}
			acc.count++
			acc.amount += loan.AgreementAmountUSD
			acc.totalDays += loan.ExtensionDays
		}
	}

	if insight.Count > 0 {
		insight.AverageExtensionDays = float64(totalExtensionDays) / float64(insight.Count)
	}
	insight.ByLender = dashboardAnalyticsExtendedBreakdownItems(filter, "lender", lenderBreakdown)
	insight.ByInstitution = dashboardAnalyticsExtendedBreakdownItems(filter, "institution", institutionBreakdown)
	return insight
}

func dashboardAnalyticsExtendedBreakdownItems(filter model.DashboardAnalyticsFilter, dimension string, values map[string]*dashboardAnalyticsExtendedAccumulator) []model.DashboardAnalyticsExtendedLoanBreakdown {
	items := make([]model.DashboardAnalyticsExtendedLoanBreakdown, 0, len(values))
	for _, value := range values {
		avg := 0.0
		if value.count > 0 {
			avg = float64(value.totalDays) / float64(value.count)
		}
		items = append(items, model.DashboardAnalyticsExtendedLoanBreakdown{
			Dimension:            dimension,
			Entity:               value.entity,
			LoanAgreementCount:   value.count,
			AmountUSD:            value.amount,
			AverageExtensionDays: avg,
			Drilldown:            dashboardAnalyticsDrilldown(filter, "loan_agreements", value.extra),
		})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].AmountUSD == items[j].AmountUSD {
			return items[i].Entity.Name < items[j].Entity.Name
		}
		return items[i].AmountUSD > items[j].AmountUSD
	})
	return items
}

func optionalFloatNumeric(value *float64) pgtype.Numeric {
	if value == nil {
		return pgtype.Numeric{}
	}
	var numeric pgtype.Numeric
	if err := numeric.Scan(strconv.FormatFloat(*value, 'f', -1, 64)); err != nil {
		return pgtype.Numeric{}
	}
	return numeric
}

func dashboardAnalyticsAbsorptionStatus(pct float64) string {
	if pct < 50 {
		return "low"
	}
	if pct < 90 {
		return "normal"
	}
	return "high"
}

func dashboardAnalyticsDrilldown(filter model.DashboardAnalyticsFilter, target string, extra map[string][]string) model.DashboardDrilldownQuery {
	query := map[string][]string{}
	add := func(key string, values ...string) {
		for _, value := range values {
			if value == "" {
				continue
			}
			query[key] = append(query[key], value)
		}
	}

	if filter.BudgetYear != nil {
		add("budget_year", strconv.FormatInt(int64(*filter.BudgetYear), 10))
	}
	if filter.Quarter != nil {
		add("quarter", *filter.Quarter)
	}
	add("lender_ids", filter.LenderIDs...)
	add("lender_types", filter.LenderTypes...)
	add("institution_ids", filter.InstitutionIDs...)
	add("pipeline_statuses", filter.PipelineStatuses...)
	add("project_statuses", filter.ProjectStatuses...)
	add("region_ids", filter.RegionIDs...)
	add("program_title_ids", filter.ProgramTitleIDs...)
	if filter.ForeignLoanMin != nil {
		add("foreign_loan_min", strconv.FormatFloat(*filter.ForeignLoanMin, 'f', -1, 64))
	}
	if filter.ForeignLoanMax != nil {
		add("foreign_loan_max", strconv.FormatFloat(*filter.ForeignLoanMax, 'f', -1, 64))
	}
	add("include_history", strconv.FormatBool(filter.IncludeHistory))
	for key, values := range extra {
		add(key, values...)
	}

	return model.DashboardDrilldownQuery{
		Target: target,
		Query:  query,
	}
}

func textValue(value pgtype.Text) string {
	if !value.Valid {
		return ""
	}
	return value.String
}
