package service

import (
	"context"

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

func (s *DashboardService) GetSummary(ctx context.Context) (*model.DashboardSummary, error) {
	row, err := s.queries.GetDashboardSummary(ctx)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil ringkasan dashboard")
	}
	totalAmount := floatFromNumeric(row.TotalAmountUsd)
	totalRealized := floatFromNumeric(row.TotalRealizedUsd)
	return &model.DashboardSummary{
		TotalBBProjects:      int(row.TotalBbProjects),
		TotalGBProjects:      int(row.TotalGbProjects),
		TotalLoanAgreements:  int(row.TotalLoanAgreements),
		TotalAmountUSD:       totalAmount,
		TotalRealizedUSD:     totalRealized,
		OverallAbsorptionPct: absorptionPct(totalAmount, totalRealized),
		ActiveMonitoring:     int(row.ActiveMonitoring),
	}, nil
}

func (s *DashboardService) GetMonitoringSummary(ctx context.Context, filter model.MonitoringSummaryFilter) (*model.MonitoringSummary, error) {
	params := queries.GetMonitoringSummaryParams{}
	if filter.BudgetYear != nil {
		params.BudgetYear = pgtype.Int4{Int32: *filter.BudgetYear, Valid: true}
	}
	if filter.Quarter != nil && *filter.Quarter != "" {
		params.Quarter = pgtype.Text{String: *filter.Quarter, Valid: true}
	}
	if filter.LenderID != nil && *filter.LenderID != "" {
		lenderID, err := model.ParseUUID(*filter.LenderID)
		if err != nil {
			return nil, validation("lender_id", "UUID tidak valid")
		}
		params.LenderID = lenderID
	}
	rows, err := s.queries.GetMonitoringSummary(ctx, params)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil ringkasan monitoring")
	}
	var totalPlanned, totalRealized float64
	byLender := make([]model.MonitoringSummaryByLender, 0, len(rows))
	for _, row := range rows {
		planned := floatFromNumeric(row.TotalPlannedUsd)
		realized := floatFromNumeric(row.TotalRealizedUsd)
		totalPlanned += planned
		totalRealized += realized
		byLender = append(byLender, model.MonitoringSummaryByLender{
			Lender:        model.LenderSummary{ID: model.UUIDToString(row.LenderID), Name: row.LenderName},
			PlannedUSD:    planned,
			RealizedUSD:   realized,
			AbsorptionPct: absorptionPct(planned, realized),
		})
	}
	return &model.MonitoringSummary{
		BudgetYear:       filter.BudgetYear,
		Quarter:          filter.Quarter,
		TotalPlannedUSD:  totalPlanned,
		TotalRealizedUSD: totalRealized,
		AbsorptionPct:    absorptionPct(totalPlanned, totalRealized),
		ByLender:         byLender,
	}, nil
}
