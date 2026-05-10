package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

type DashboardService struct {
	queries *queries.Queries
}

func NewDashboardService(q *queries.Queries) *DashboardService {
	return &DashboardService{queries: q}
}

func (s *DashboardService) GetStageOverview(ctx context.Context, periodIDs []pgtype.UUID, regionIDs []pgtype.UUID) (*model.DashboardStageOverviewResponse, error) {
	params := queries.ListDashboardStageSummariesParams{
		PeriodIds: periodIDs,
		RegionIds: regionIDs,
	}
	summaryRows, err := s.queries.ListDashboardStageSummaries(ctx, params)
	if err != nil {
		return nil, err
	}

	regionRows, err := s.queries.ListDashboardStageRegionGroups(ctx, queries.ListDashboardStageRegionGroupsParams{
		PeriodIds: periodIDs,
		RegionIds: regionIDs,
	})
	if err != nil {
		return nil, err
	}

	regionsByStage := make(map[string][]model.DashboardDistributionItem)
	for _, row := range regionRows {
		regionsByStage[row.Stage] = append(regionsByStage[row.Stage], model.DashboardDistributionItem{
			ID:             row.ID,
			Label:          row.Label,
			Level:          row.Level,
			ProjectCount:   int(row.ProjectCount),
			ForeignLoanUSD: floatFromNumeric(row.ForeignLoanUsd),
		})
	}

	stages := make([]model.DashboardStageOverviewItem, 0, len(summaryRows))
	for _, row := range summaryRows {
		regions := regionsByStage[row.Stage]
		if regions == nil {
			regions = []model.DashboardDistributionItem{}
		}

		stages = append(stages, model.DashboardStageOverviewItem{
			Stage:        row.Stage,
			ProjectCount: int(row.ProjectCount),
			TotalLoanUSD: floatFromNumeric(row.TotalLoanUsd),
			Regions:      regions,
		})
	}

	return &model.DashboardStageOverviewResponse{Stages: stages}, nil
}

func (s *DashboardService) GetBlueBookDistribution(ctx context.Context, periodIDs []pgtype.UUID) (*model.DashboardBlueBookDistributionResponse, error) {
	groupRows, err := s.queries.ListBlueBookTopLevelExecutingAgencyGroups(ctx, periodIDs)
	if err != nil {
		return nil, err
	}

	agencyRows, err := s.queries.ListBlueBookTopLevelExecutingAgencies(ctx, periodIDs)
	if err != nil {
		return nil, err
	}

	programRows, err := s.queries.ListBlueBookPrograms(ctx, periodIDs)
	if err != nil {
		return nil, err
	}

	agencyGroups := make([]model.DashboardDistributionItem, 0, len(groupRows))
	for _, row := range groupRows {
		agencyGroups = append(agencyGroups, model.DashboardDistributionItem{
			ID:             row.ID,
			Label:          row.Label,
			Level:          row.Level,
			ProjectCount:   int(row.ProjectCount),
			ForeignLoanUSD: floatFromNumeric(row.ForeignLoanUsd),
		})
	}

	topAgencies := make([]model.DashboardDistributionItem, 0, len(agencyRows))
	for _, row := range agencyRows {
		topAgencies = append(topAgencies, model.DashboardDistributionItem{
			ID:             row.ID,
			Label:          row.Label,
			Level:          row.Level,
			ProjectCount:   int(row.ProjectCount),
			ForeignLoanUSD: floatFromNumeric(row.ForeignLoanUsd),
		})
	}

	programs := make([]model.DashboardDistributionItem, 0, len(programRows))
	for _, row := range programRows {
		programs = append(programs, model.DashboardDistributionItem{
			ID:             row.ID,
			Label:          row.Label,
			Level:          row.Level,
			ProjectCount:   int(row.ProjectCount),
			ForeignLoanUSD: floatFromNumeric(row.ForeignLoanUsd),
		})
	}

	return &model.DashboardBlueBookDistributionResponse{
		AgencyGroups: agencyGroups,
		TopAgencies:  topAgencies,
		Programs:     programs,
	}, nil
}

func (s *DashboardService) GetGreenBookDistribution(ctx context.Context, periodIDs []pgtype.UUID) (*model.DashboardGreenBookDistributionResponse, error) {
	lenderTypeRows, err := s.queries.ListGreenBookLenderTypes(ctx, periodIDs)
	if err != nil {
		return nil, err
	}

	lenderRows, err := s.queries.ListGreenBookTopLenders(ctx, periodIDs)
	if err != nil {
		return nil, err
	}

	agencyRows, err := s.queries.ListGreenBookTopLevelExecutingAgencies(ctx, periodIDs)
	if err != nil {
		return nil, err
	}

	lenderTypes := make([]model.DashboardDistributionItem, 0, len(lenderTypeRows))
	for _, row := range lenderTypeRows {
		lenderTypes = append(lenderTypes, model.DashboardDistributionItem{
			ID:             row.ID,
			Label:          row.Label,
			Level:          row.Level,
			ProjectCount:   int(row.ProjectCount),
			ForeignLoanUSD: floatFromNumeric(row.ForeignLoanUsd),
		})
	}

	topLenders := make([]model.DashboardDistributionItem, 0, len(lenderRows))
	for _, row := range lenderRows {
		topLenders = append(topLenders, model.DashboardDistributionItem{
			ID:             row.ID,
			Label:          row.Label,
			Level:          row.Level,
			ProjectCount:   int(row.ProjectCount),
			ForeignLoanUSD: floatFromNumeric(row.ForeignLoanUsd),
		})
	}

	topAgencies := make([]model.DashboardDistributionItem, 0, len(agencyRows))
	for _, row := range agencyRows {
		topAgencies = append(topAgencies, model.DashboardDistributionItem{
			ID:             row.ID,
			Label:          row.Label,
			Level:          row.Level,
			ProjectCount:   int(row.ProjectCount),
			ForeignLoanUSD: floatFromNumeric(row.ForeignLoanUsd),
		})
	}

	return &model.DashboardGreenBookDistributionResponse{
		LenderTypes: lenderTypes,
		TopLenders:  topLenders,
		TopAgencies: topAgencies,
	}, nil
}

func (s *DashboardService) GetDaftarKegiatanDistribution(ctx context.Context, periodIDs []pgtype.UUID) (*model.DashboardDaftarKegiatanDistributionResponse, error) {
	lenderTypeRows, err := s.queries.ListDaftarKegiatanLenderTypes(ctx, periodIDs)
	if err != nil {
		return nil, err
	}

	lenderRows, err := s.queries.ListDaftarKegiatanTopLenders(ctx, periodIDs)
	if err != nil {
		return nil, err
	}

	agencyRows, err := s.queries.ListDaftarKegiatanTopLevelExecutingAgencies(ctx, periodIDs)
	if err != nil {
		return nil, err
	}

	programRows, err := s.queries.ListDaftarKegiatanPrograms(ctx, periodIDs)
	if err != nil {
		return nil, err
	}

	lenderTypes := make([]model.DashboardDistributionItem, 0, len(lenderTypeRows))
	for _, row := range lenderTypeRows {
		lenderTypes = append(lenderTypes, model.DashboardDistributionItem{
			ID:             row.ID,
			Label:          row.Label,
			Level:          row.Level,
			ProjectCount:   int(row.ProjectCount),
			ForeignLoanUSD: floatFromNumeric(row.ForeignLoanUsd),
		})
	}

	topLenders := make([]model.DashboardDistributionItem, 0, len(lenderRows))
	for _, row := range lenderRows {
		topLenders = append(topLenders, model.DashboardDistributionItem{
			ID:             row.ID,
			Label:          row.Label,
			Level:          row.Level,
			ProjectCount:   int(row.ProjectCount),
			ForeignLoanUSD: floatFromNumeric(row.ForeignLoanUsd),
		})
	}

	topAgencies := make([]model.DashboardDistributionItem, 0, len(agencyRows))
	for _, row := range agencyRows {
		topAgencies = append(topAgencies, model.DashboardDistributionItem{
			ID:             row.ID,
			Label:          row.Label,
			Level:          row.Level,
			ProjectCount:   int(row.ProjectCount),
			ForeignLoanUSD: floatFromNumeric(row.ForeignLoanUsd),
		})
	}

	programs := make([]model.DashboardDistributionItem, 0, len(programRows))
	for _, row := range programRows {
		programs = append(programs, model.DashboardDistributionItem{
			ID:             row.ID,
			Label:          row.Label,
			Level:          row.Level,
			ProjectCount:   int(row.ProjectCount),
			ForeignLoanUSD: floatFromNumeric(row.ForeignLoanUsd),
		})
	}

	return &model.DashboardDaftarKegiatanDistributionResponse{
		LenderTypes: lenderTypes,
		TopLenders:  topLenders,
		TopAgencies: topAgencies,
		Programs:    programs,
	}, nil
}

func (s *DashboardService) GetLoanAgreementDistribution(ctx context.Context, periodIDs []pgtype.UUID) (*model.DashboardLoanAgreementDistributionResponse, error) {
	lenderTypeRows, err := s.queries.ListLoanAgreementLenderTypes(ctx, periodIDs)
	if err != nil {
		return nil, err
	}

	lenderRows, err := s.queries.ListLoanAgreementTopLenders(ctx, periodIDs)
	if err != nil {
		return nil, err
	}

	agencyRows, err := s.queries.ListLoanAgreementTopLevelExecutingAgencies(ctx, periodIDs)
	if err != nil {
		return nil, err
	}

	programRows, err := s.queries.ListLoanAgreementPrograms(ctx, periodIDs)
	if err != nil {
		return nil, err
	}

	lenderTypes := make([]model.DashboardDistributionItem, 0, len(lenderTypeRows))
	for _, row := range lenderTypeRows {
		lenderTypes = append(lenderTypes, model.DashboardDistributionItem{
			ID:             row.ID,
			Label:          row.Label,
			Level:          row.Level,
			ProjectCount:   int(row.ProjectCount),
			ForeignLoanUSD: floatFromNumeric(row.ForeignLoanUsd),
		})
	}

	topLenders := make([]model.DashboardDistributionItem, 0, len(lenderRows))
	for _, row := range lenderRows {
		topLenders = append(topLenders, model.DashboardDistributionItem{
			ID:             row.ID,
			Label:          row.Label,
			Level:          row.Level,
			ProjectCount:   int(row.ProjectCount),
			ForeignLoanUSD: floatFromNumeric(row.ForeignLoanUsd),
		})
	}

	topAgencies := make([]model.DashboardDistributionItem, 0, len(agencyRows))
	for _, row := range agencyRows {
		topAgencies = append(topAgencies, model.DashboardDistributionItem{
			ID:             row.ID,
			Label:          row.Label,
			Level:          row.Level,
			ProjectCount:   int(row.ProjectCount),
			ForeignLoanUSD: floatFromNumeric(row.ForeignLoanUsd),
		})
	}

	programs := make([]model.DashboardDistributionItem, 0, len(programRows))
	for _, row := range programRows {
		programs = append(programs, model.DashboardDistributionItem{
			ID:             row.ID,
			Label:          row.Label,
			Level:          row.Level,
			ProjectCount:   int(row.ProjectCount),
			ForeignLoanUSD: floatFromNumeric(row.ForeignLoanUsd),
		})
	}

	return &model.DashboardLoanAgreementDistributionResponse{
		LenderTypes: lenderTypes,
		TopLenders:  topLenders,
		TopAgencies: topAgencies,
		Programs:    programs,
	}, nil
}
