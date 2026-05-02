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

func NewDashboardAnalyticsService(queries *queries.Queries) *DashboardAnalyticsService {
	return &DashboardAnalyticsService{queries: queries}
}

func (s *DashboardAnalyticsService) Overview(ctx context.Context, filter model.DashboardAnalyticsFilter) (*model.DashboardAnalyticsOverviewResponse, error) {
	portfolioParams, monitoringParams, err := dashboardAnalyticsQueryParams(filter)
	if err != nil {
		return nil, err
	}

	portfolioRow, err := s.queries.GetDashboardAnalyticsPortfolioFoundation(ctx, portfolioParams)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil ringkasan portfolio analytics")
	}
	monitoringRow, err := s.queries.GetDashboardAnalyticsMonitoringFoundation(ctx, monitoringParams)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil ringkasan monitoring analytics")
	}

	portfolio := dashboardAnalyticsPortfolioSummary(portfolioRow)
	monitoring := dashboardAnalyticsMonitoringSummary(monitoringRow)
	return &model.DashboardAnalyticsOverviewResponse{
		Portfolio:  portfolio,
		Monitoring: monitoring,
		Drilldown:  dashboardAnalyticsDrilldown(filter, "projects"),
	}, nil
}

func (s *DashboardAnalyticsService) Institutions(ctx context.Context, filter model.DashboardAnalyticsFilter) (*model.DashboardAnalyticsInstitutionsResponse, error) {
	overview, err := s.Overview(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &model.DashboardAnalyticsInstitutionsResponse{
		Summary:   dashboardAnalyticsSectionSummary(overview.Portfolio, overview.Monitoring),
		Items:     []model.DashboardAnalyticsInstitutionItem{},
		Drilldown: dashboardAnalyticsDrilldown(filter, "projects"),
	}, nil
}

func (s *DashboardAnalyticsService) Lenders(ctx context.Context, filter model.DashboardAnalyticsFilter) (*model.DashboardAnalyticsLendersResponse, error) {
	overview, err := s.Overview(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &model.DashboardAnalyticsLendersResponse{
		Summary:   dashboardAnalyticsSectionSummary(overview.Portfolio, overview.Monitoring),
		Items:     []model.DashboardAnalyticsLenderStageItem{},
		Drilldown: dashboardAnalyticsDrilldown(filter, "projects"),
	}, nil
}

func (s *DashboardAnalyticsService) Absorption(ctx context.Context, filter model.DashboardAnalyticsFilter) (*model.DashboardAnalyticsAbsorptionResponse, error) {
	_, monitoringParams, err := dashboardAnalyticsQueryParams(filter)
	if err != nil {
		return nil, err
	}
	monitoringRow, err := s.queries.GetDashboardAnalyticsMonitoringFoundation(ctx, monitoringParams)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil ringkasan penyerapan analytics")
	}
	return &model.DashboardAnalyticsAbsorptionResponse{
		Summary:       dashboardAnalyticsMonitoringSummary(monitoringRow),
		ByInstitution: []model.DashboardAnalyticsAbsorptionItem{},
		ByProject:     []model.DashboardAnalyticsAbsorptionItem{},
		ByLender:      []model.DashboardAnalyticsAbsorptionItem{},
		Drilldown:     dashboardAnalyticsDrilldown(filter, "monitoring"),
	}, nil
}

func (s *DashboardAnalyticsService) Yearly(ctx context.Context, filter model.DashboardAnalyticsFilter) (*model.DashboardAnalyticsYearlyResponse, error) {
	_, monitoringParams, err := dashboardAnalyticsQueryParams(filter)
	if err != nil {
		return nil, err
	}
	monitoringRow, err := s.queries.GetDashboardAnalyticsMonitoringFoundation(ctx, monitoringParams)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil ringkasan tahunan analytics")
	}
	return &model.DashboardAnalyticsYearlyResponse{
		Summary:   dashboardAnalyticsMonitoringSummary(monitoringRow),
		Items:     []model.DashboardAnalyticsYearlyItem{},
		Drilldown: dashboardAnalyticsDrilldown(filter, "monitoring"),
	}, nil
}

func (s *DashboardAnalyticsService) LenderProportion(ctx context.Context, filter model.DashboardAnalyticsFilter) (*model.DashboardAnalyticsLenderProportionResponse, error) {
	if _, _, err := dashboardAnalyticsQueryParams(filter); err != nil {
		return nil, err
	}
	return &model.DashboardAnalyticsLenderProportionResponse{
		Items:     []model.DashboardAnalyticsLenderProportionItem{},
		Drilldown: dashboardAnalyticsDrilldown(filter, "projects"),
	}, nil
}

func (s *DashboardAnalyticsService) Risks(ctx context.Context, filter model.DashboardAnalyticsFilter) (*model.DashboardAnalyticsRisksResponse, error) {
	if _, _, err := dashboardAnalyticsQueryParams(filter); err != nil {
		return nil, err
	}
	return &model.DashboardAnalyticsRisksResponse{
		Watchlists:  []model.DashboardAnalyticsRiskItem{},
		DataQuality: []model.DashboardAnalyticsDataQualityItem{},
		Drilldown:   dashboardAnalyticsDrilldown(filter, "projects"),
	}, nil
}

func dashboardAnalyticsQueryParams(filter model.DashboardAnalyticsFilter) (queries.GetDashboardAnalyticsPortfolioFoundationParams, queries.GetDashboardAnalyticsMonitoringFoundationParams, error) {
	lenderIDs, err := uuidArray(filter.LenderIDs, "lender_ids")
	if err != nil {
		return queries.GetDashboardAnalyticsPortfolioFoundationParams{}, queries.GetDashboardAnalyticsMonitoringFoundationParams{}, err
	}
	institutionIDs, err := uuidArray(filter.InstitutionIDs, "institution_ids")
	if err != nil {
		return queries.GetDashboardAnalyticsPortfolioFoundationParams{}, queries.GetDashboardAnalyticsMonitoringFoundationParams{}, err
	}
	regionIDs, err := uuidArray(filter.RegionIDs, "region_ids")
	if err != nil {
		return queries.GetDashboardAnalyticsPortfolioFoundationParams{}, queries.GetDashboardAnalyticsMonitoringFoundationParams{}, err
	}
	programTitleIDs, err := uuidArray(filter.ProgramTitleIDs, "program_title_ids")
	if err != nil {
		return queries.GetDashboardAnalyticsPortfolioFoundationParams{}, queries.GetDashboardAnalyticsMonitoringFoundationParams{}, err
	}

	foreignLoanMin := optionalFloatNumeric(filter.ForeignLoanMin)
	foreignLoanMax := optionalFloatNumeric(filter.ForeignLoanMax)

	portfolioParams := queries.GetDashboardAnalyticsPortfolioFoundationParams{
		IncludeHistory:   filter.IncludeHistory,
		LenderTypes:      filter.LenderTypes,
		LenderIds:        lenderIDs,
		InstitutionIds:   institutionIDs,
		PipelineStatuses: filter.PipelineStatuses,
		ProjectStatuses:  filter.ProjectStatuses,
		RegionIds:        regionIDs,
		ProgramTitleIds:  programTitleIDs,
		ForeignLoanMin:   foreignLoanMin,
		ForeignLoanMax:   foreignLoanMax,
	}

	monitoringParams := queries.GetDashboardAnalyticsMonitoringFoundationParams{
		LenderIds:        lenderIDs,
		LenderTypes:      filter.LenderTypes,
		InstitutionIds:   institutionIDs,
		PipelineStatuses: filter.PipelineStatuses,
		ProjectStatuses:  filter.ProjectStatuses,
		ForeignLoanMin:   foreignLoanMin,
		ForeignLoanMax:   foreignLoanMax,
		ProgramTitleIds:  programTitleIDs,
		RegionIds:        regionIDs,
	}
	if filter.BudgetYear != nil {
		monitoringParams.BudgetYear = pgtype.Int4{Int32: *filter.BudgetYear, Valid: true}
	}
	if filter.Quarter != nil && *filter.Quarter != "" {
		monitoringParams.Quarter = pgtype.Text{String: *filter.Quarter, Valid: true}
	}

	return portfolioParams, monitoringParams, nil
}

func dashboardAnalyticsPortfolioSummary(row queries.GetDashboardAnalyticsPortfolioFoundationRow) model.DashboardAnalyticsPortfolioSummary {
	return model.DashboardAnalyticsPortfolioSummary{
		ProjectCount:         int(row.ProjectCount),
		PipelineProjectCount: int(row.PipelineProjectCount),
		OngoingProjectCount:  int(row.OngoingProjectCount),
		TotalForeignLoanUSD:  floatFromNumeric(row.TotalForeignLoanUsd),
		TotalGrantUSD:        floatFromNumeric(row.TotalGrantUsd),
		TotalCounterpartUSD:  floatFromNumeric(row.TotalCounterpartUsd),
	}
}

func dashboardAnalyticsMonitoringSummary(row queries.GetDashboardAnalyticsMonitoringFoundationRow) model.DashboardAnalyticsMonitoringSummary {
	planned := floatFromNumeric(row.PlannedUsd)
	realized := floatFromNumeric(row.RealizedUsd)
	return model.DashboardAnalyticsMonitoringSummary{
		LoanAgreementCount: int(row.LoanAgreementCount),
		MonitoringCount:    int(row.MonitoringCount),
		PlannedUSD:         planned,
		RealizedUSD:        realized,
		AgreementAmountUSD: floatFromNumeric(row.AgreementAmountUsd),
		AbsorptionPct:      absorptionPct(planned, realized),
	}
}

func dashboardAnalyticsSectionSummary(portfolio model.DashboardAnalyticsPortfolioSummary, monitoring model.DashboardAnalyticsMonitoringSummary) model.DashboardAnalyticsSectionSummary {
	return model.DashboardAnalyticsSectionSummary{
		ProjectCount:      portfolio.ProjectCount,
		MonitoringCount:   monitoring.MonitoringCount,
		PlannedUSD:        monitoring.PlannedUSD,
		RealizedUSD:       monitoring.RealizedUSD,
		AbsorptionPct:     monitoring.AbsorptionPct,
		ForeignLoanUSD:    portfolio.TotalForeignLoanUSD,
		AgreementValueUSD: monitoring.AgreementAmountUSD,
	}
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

func dashboardAnalyticsDrilldown(filter model.DashboardAnalyticsFilter, target string) model.DashboardDrilldownQuery {
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

	return model.DashboardDrilldownQuery{
		Target: target,
		Query:  query,
	}
}
