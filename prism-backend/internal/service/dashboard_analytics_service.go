package service

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

type DashboardAnalyticsService struct {
	queries *queries.Queries
}

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
	if _, err := newDashboardAnalyticsQueryValues(filter); err != nil {
		return nil, err
	}
	return &model.DashboardAnalyticsRisksResponse{
		Watchlists:  []model.DashboardAnalyticsRiskItem{},
		DataQuality: []model.DashboardAnalyticsDataQualityItem{},
		Drilldown:   dashboardAnalyticsDrilldown(filter, "projects", nil),
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
