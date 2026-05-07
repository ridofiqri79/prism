package service

import (
	"context"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

type DashboardService struct {
	queries *queries.Queries
}

func NewDashboardService(q *queries.Queries) *DashboardService {
	return &DashboardService{queries: q}
}

func (s *DashboardService) GetBlueBookDistribution(ctx context.Context) (*model.DashboardBlueBookDistributionResponse, error) {
	groupRows, err := s.queries.ListBlueBookTopLevelExecutingAgencyGroups(ctx)
	if err != nil {
		return nil, err
	}

	agencyRows, err := s.queries.ListBlueBookTopLevelExecutingAgencies(ctx)
	if err != nil {
		return nil, err
	}

	programRows, err := s.queries.ListBlueBookPrograms(ctx)
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

func (s *DashboardService) GetGreenBookDistribution(ctx context.Context) (*model.DashboardGreenBookDistributionResponse, error) {
	lenderTypeRows, err := s.queries.ListGreenBookLenderTypes(ctx)
	if err != nil {
		return nil, err
	}

	lenderRows, err := s.queries.ListGreenBookTopLenders(ctx)
	if err != nil {
		return nil, err
	}

	agencyRows, err := s.queries.ListGreenBookTopLevelExecutingAgencies(ctx)
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

func (s *DashboardService) GetDaftarKegiatanDistribution(ctx context.Context) (*model.DashboardDaftarKegiatanDistributionResponse, error) {
	lenderTypeRows, err := s.queries.ListDaftarKegiatanLenderTypes(ctx)
	if err != nil {
		return nil, err
	}

	lenderRows, err := s.queries.ListDaftarKegiatanTopLenders(ctx)
	if err != nil {
		return nil, err
	}

	agencyRows, err := s.queries.ListDaftarKegiatanTopLevelExecutingAgencies(ctx)
	if err != nil {
		return nil, err
	}

	programRows, err := s.queries.ListDaftarKegiatanPrograms(ctx)
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

func (s *DashboardService) GetLoanAgreementDistribution(ctx context.Context) (*model.DashboardLoanAgreementDistributionResponse, error) {
	lenderTypeRows, err := s.queries.ListLoanAgreementLenderTypes(ctx)
	if err != nil {
		return nil, err
	}

	lenderRows, err := s.queries.ListLoanAgreementTopLenders(ctx)
	if err != nil {
		return nil, err
	}

	agencyRows, err := s.queries.ListLoanAgreementTopLevelExecutingAgencies(ctx)
	if err != nil {
		return nil, err
	}

	programRows, err := s.queries.ListLoanAgreementPrograms(ctx)
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
