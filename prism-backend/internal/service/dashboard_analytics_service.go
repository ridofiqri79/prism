package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

type DashboardAnalyticsService struct {
	queries *queries.Queries
}

func NewDashboardAnalyticsService(q *queries.Queries) *DashboardAnalyticsService {
	return &DashboardAnalyticsService{queries: q}
}

// ------ Overview ------

func (s *DashboardAnalyticsService) GetOverview(ctx context.Context, f model.DashboardAnalyticsFilter) (*model.DashboardAnalyticsOverview, error) {
	row, err := s.queries.GetDashboardAnalyticsOverview(ctx, overviewParams(f))
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil overview analytics")
	}
	totalAmount := floatFromNumeric(row.AgreementAmountUsd)
	totalPlanned := floatFromNumeric(row.TotalPlannedUsd)
	totalRealized := floatFromNumeric(row.TotalRealizedUsd)

	funnel, err := s.GetPipelineFunnel(ctx, f)
	if err != nil {
		return nil, err
	}
	topInst, err := s.GetTopInstitutions(ctx, f)
	if err != nil {
		return nil, err
	}
	topLenders, err := s.GetTopLenders(ctx, f)
	if err != nil {
		return nil, err
	}

	return &model.DashboardAnalyticsOverview{
		TotalProjects:        row.TotalProjects,
		TotalLoanAgreements:  row.TotalLoanAgreements,
		AgreementAmountUSD:   totalAmount,
		TotalPlannedUSD:      totalPlanned,
		TotalRealizedUSD:     totalRealized,
		OverallAbsorptionPct: absorptionPct(totalPlanned, totalRealized),
		ActiveMonitoring:     row.ActiveMonitoring,
		PipelineFunnel:       funnel,
		TopInstitutions:      topInst,
		TopLenders:           topLenders,
	}, nil
}

func (s *DashboardAnalyticsService) GetPipelineFunnel(ctx context.Context, f model.DashboardAnalyticsFilter) ([]model.PipelineStageSummary, error) {
	rows, err := s.queries.GetDashboardAnalyticsPipelineFunnel(ctx, queries.GetDashboardAnalyticsPipelineFunnelParams{
		IncludeHistory: f.IncludeHistory,
		BudgetYear:     nullableInt32(f.BudgetYear),
		Quarter:        nullableTextPtr(f.Quarter),
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil pipeline funnel")
	}
	out := make([]model.PipelineStageSummary, 0, len(rows))
	for _, r := range rows {
		out = append(out, model.PipelineStageSummary{
			Stage:        r.Stage,
			ProjectCount: r.ProjectCount,
			TotalLoanUSD: floatFromNumeric(r.TotalLoanUsd),
		})
	}
	return out, nil
}

func (s *DashboardAnalyticsService) GetTopInstitutions(ctx context.Context, f model.DashboardAnalyticsFilter) ([]model.InstitutionSummary, error) {
	rows, err := s.queries.GetDashboardAnalyticsTopInstitutions(ctx, queries.GetDashboardAnalyticsTopInstitutionsParams{
		InstitutionIds: parseUUIDs(f.InstitutionIDs),
		Limit:          int32(10),
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil top institutions")
	}
	out := make([]model.InstitutionSummary, 0, len(rows))
	for _, r := range rows {
		planned := floatFromNumeric(r.PlannedUsd)
		realized := floatFromNumeric(r.RealizedUsd)
		out = append(out, model.InstitutionSummary{
			Institution: model.InstitutionRef{
				ID:        model.UUIDToString(r.InstitutionID),
				Name:      r.InstitutionName,
				ShortName: textToString(r.InstitutionShortName),
				Level:     r.InstitutionLevel,
			},
			ProjectCount:       r.ProjectCount,
			LoanAgreementCount: r.LoanAgreementCount,
			MonitoringCount:    r.MonitoringCount,
			AgreementAmountUSD: floatFromNumeric(r.AgreementAmountUsd),
			PlannedUSD:         planned,
			RealizedUSD:        realized,
			AbsorptionPct:      absorptionPct(planned, realized),
		})
	}
	return out, nil
}

func (s *DashboardAnalyticsService) GetTopLenders(ctx context.Context, f model.DashboardAnalyticsFilter) ([]model.TopLenderSummary, error) {
	rows, err := s.queries.GetDashboardAnalyticsTopLenders(ctx, queries.GetDashboardAnalyticsTopLendersParams{
		LenderTypes: f.LenderTypes,
		LenderIds:   parseUUIDs(f.LenderIDs),
		Limit:       int32(10),
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil top lenders")
	}
	out := make([]model.TopLenderSummary, 0, len(rows))
	for _, r := range rows {
		planned := floatFromNumeric(r.PlannedUsd)
		realized := floatFromNumeric(r.RealizedUsd)
		out = append(out, model.TopLenderSummary{
			Lender: model.LenderRef{
				ID:        model.UUIDToString(r.LenderID),
				Name:      r.LenderName,
				ShortName: textToString(r.LenderShortName),
				Type:      r.LenderType,
			},
			LoanAgreementCount: r.LoanAgreementCount,
			ProjectCount:       r.ProjectCount,
			InstitutionCount:   r.InstitutionCount,
			AgreementAmountUSD: floatFromNumeric(r.AgreementAmountUsd),
			PlannedUSD:         planned,
			RealizedUSD:        realized,
			AbsorptionPct:      absorptionPct(planned, realized),
		})
	}
	return out, nil
}

// ------ Institutions ------

func (s *DashboardAnalyticsService) GetInstitutions(ctx context.Context, f model.DashboardAnalyticsFilter, page, limit int) (*model.DashboardAnalyticsInstitutionsResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	summary, _ := s.queries.GetDashboardAnalyticsInstitutionsSummary(ctx, parseUUIDs(f.InstitutionIDs))

	rows, err := s.queries.GetDashboardAnalyticsInstitutions(ctx, queries.GetDashboardAnalyticsInstitutionsParams{
		InstitutionIds: parseUUIDs(f.InstitutionIDs),
		BudgetYear:     nullableInt32(f.BudgetYear),
		Quarter:        nullableTextPtr(f.Quarter),
		LenderTypes:    f.LenderTypes,
		LenderIds:      parseUUIDs(f.LenderIDs),
		Limit:          int32(limit),
		Offset:         int32(offset),
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil data institution analytics")
	}

	sPlanned := floatFromNumeric(summary.PlannedUsd)
	sRealized := floatFromNumeric(summary.RealizedUsd)

	items := make([]model.InstitutionItem, 0, len(rows))
	for _, r := range rows {
		planned := floatFromNumeric(r.PlannedUsd)
		realized := floatFromNumeric(r.RealizedUsd)
		items = append(items, model.InstitutionItem{
			Institution: model.InstitutionRef{
				ID:        model.UUIDToString(r.InstitutionID),
				Name:      r.InstitutionName,
				ShortName: textToString(r.InstitutionShortName),
				Level:     r.InstitutionLevel,
			},
			ProjectCount:       r.ProjectCount,
			AssignmentCount:    r.AssignmentCount,
			LoanAgreementCount: r.LoanAgreementCount,
			MonitoringCount:    r.MonitoringCount,
			AgreementAmountUSD: floatFromNumeric(r.AgreementAmountUsd),
			PlannedUSD:         planned,
			RealizedUSD:        realized,
			AbsorptionPct:      absorptionPct(planned, realized),
			LoanTypes:          r.LoanTypes,
			Drilldown: &model.DashboardDrilldownQuery{
				Target: "projects",
				Query:  map[string][]string{"executing_agency_ids": {model.UUIDToString(r.InstitutionID)}},
			},
		})
	}

	return &model.DashboardAnalyticsInstitutionsResponse{
		Summary: model.InstitutionsSummary{
			InstitutionCount:    summary.InstitutionCount,
			ProjectCount:        summary.ProjectCount,
			AssignmentCount:     summary.AssignmentCount,
			AgreementAmountUSD:  floatFromNumeric(summary.AgreementAmountUsd),
			PlannedUSD:          sPlanned,
			RealizedUSD:         sRealized,
			AbsorptionPct:       absorptionPct(sPlanned, sRealized),
		},
		Items: items,
	}, nil
}

// ------ Lenders ------

func (s *DashboardAnalyticsService) GetLenders(ctx context.Context, f model.DashboardAnalyticsFilter, page, limit int) (*model.DashboardAnalyticsLendersResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	summary, _ := s.queries.GetDashboardAnalyticsLendersSummary(ctx, queries.GetDashboardAnalyticsLendersSummaryParams{
		LenderTypes: f.LenderTypes,
		LenderIds:   parseUUIDs(f.LenderIDs),
	})

	rows, err := s.queries.GetDashboardAnalyticsLenders(ctx, queries.GetDashboardAnalyticsLendersParams{
		BudgetYear:  nullableInt32(f.BudgetYear),
		Quarter:     nullableTextPtr(f.Quarter),
		LenderTypes: f.LenderTypes,
		LenderIds:   parseUUIDs(f.LenderIDs),
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil data lender analytics")
	}

	sPlanned := floatFromNumeric(summary.PlannedUsd)
	sRealized := floatFromNumeric(summary.RealizedUsd)

	items := make([]model.LenderItem, 0, len(rows))
	for _, r := range rows {
		planned := floatFromNumeric(r.PlannedUsd)
		realized := floatFromNumeric(r.RealizedUsd)
		items = append(items, model.LenderItem{
			Lender: model.LenderRef{
				ID:        model.UUIDToString(r.LenderID),
				Name:      r.LenderName,
				ShortName: textToString(r.LenderShortName),
				Type:      r.LenderType,
			},
			LoanAgreementCount: r.LoanAgreementCount,
			ProjectCount:       r.ProjectCount,
			InstitutionCount:   r.InstitutionCount,
			AgreementAmountUSD: floatFromNumeric(r.AgreementAmountUsd),
			PlannedUSD:         planned,
			RealizedUSD:        realized,
			AbsorptionPct:      absorptionPct(planned, realized),
			Drilldown: &model.DashboardDrilldownQuery{
				Target: "projects",
				Query:  map[string][]string{"fixed_lender_ids": {model.UUIDToString(r.LenderID)}},
			},
		})
	}

	// Lender per KL matrix
	matrix, _ := s.queries.GetDashboardAnalyticsLenderInstitutionMatrix(ctx, queries.GetDashboardAnalyticsLenderInstitutionMatrixParams{
		LenderTypes:    f.LenderTypes,
		LenderIds:      parseUUIDs(f.LenderIDs),
		InstitutionIds: parseUUIDs(f.InstitutionIDs),
		Limit:          int32(50),
		Offset:         int32(0),
	})
	matrixItems := make([]model.LenderInstitutionMatrix, 0, len(matrix))
	for _, r := range matrix {
		planned := floatFromNumeric(r.PlannedUsd)
		realized := floatFromNumeric(r.RealizedUsd)
		matrixItems = append(matrixItems, model.LenderInstitutionMatrix{
			Institution: model.InstitutionRef{
				ID:        model.UUIDToString(r.InstitutionID),
				Name:      r.InstitutionName,
				ShortName: textToString(r.InstitutionShortName),
			},
			Lender: model.LenderRef{
				ID:        model.UUIDToString(r.LenderID),
				Name:      r.LenderName,
				ShortName: textToString(r.LenderShortName),
				Type:      r.LenderType,
			},
			ProjectCount:       r.ProjectCount,
			LoanAgreementCount: r.LoanAgreementCount,
			AgreementAmountUSD: floatFromNumeric(r.AgreementAmountUsd),
			PlannedUSD:         planned,
			RealizedUSD:        realized,
			AbsorptionPct:      absorptionPct(planned, realized),
		})
	}


	return &model.DashboardAnalyticsLendersResponse{
		Summary: model.LendersSummary{
			LenderCount:        summary.LenderCount,
			LoanAgreementCount: summary.LoanAgreementCount,
			AgreementAmountUSD: floatFromNumeric(summary.AgreementAmountUsd),
			PlannedUSD:         sPlanned,
			RealizedUSD:        sRealized,
			AbsorptionPct:      absorptionPct(sPlanned, sRealized),
		},
		Items:                   items,
		LenderInstitutionMatrix: matrixItems,
	}, nil
}

// ------ Absorption ------

func (s *DashboardAnalyticsService) GetAbsorption(ctx context.Context, f model.DashboardAnalyticsFilter, page, limit int) (*model.DashboardAnalyticsAbsorptionResponse, error) {
	if limit <= 0 { limit = 20 }
	if page <= 0 { page = 1 }
	offset := (page - 1) * limit

	summary, _ := s.queries.GetDashboardAnalyticsAbsorptionSummary(ctx, queries.GetDashboardAnalyticsAbsorptionSummaryParams{
		BudgetYear:  nullableInt32(f.BudgetYear),
		Quarter:     nullableTextPtr(f.Quarter),
		LenderIds:   parseUUIDs(f.LenderIDs),
		LenderTypes: f.LenderTypes,
	})
	sPlanned := floatFromNumeric(summary.PlannedUsd)
	sRealized := floatFromNumeric(summary.RealizedUsd)

	byLender := s.absorptionByLender(ctx, f, limit, offset)
	byProject := s.absorptionByProject(ctx, f, limit, offset)
	byInstitution := s.absorptionByInstitution(ctx, f, limit, offset)

	return &model.DashboardAnalyticsAbsorptionResponse{
		Summary: model.AbsorptionSummary{
			PlannedUSD:    sPlanned,
			RealizedUSD:   sRealized,
			AbsorptionPct: absorptionPct(sPlanned, sRealized),
		},
		ByInstitution: byInstitution,
		ByProject:     byProject,
		ByLender:      byLender,
	}, nil
}

func (s *DashboardAnalyticsService) absorptionByLender(ctx context.Context, f model.DashboardAnalyticsFilter, limit, offset int) []model.AbsorptionRankItem {
	rows, _ := s.queries.GetDashboardAnalyticsAbsorptionByLender(ctx, queries.GetDashboardAnalyticsAbsorptionByLenderParams{
		BudgetYear:  nullableInt32(f.BudgetYear),
		Quarter:     nullableTextPtr(f.Quarter),
		LenderIds:   parseUUIDs(f.LenderIDs),
		LenderTypes: f.LenderTypes,
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	out := make([]model.AbsorptionRankItem, 0, len(rows))
	for i, r := range rows {
		planned := floatFromNumeric(r.PlannedUsd)
		realized := floatFromNumeric(r.RealizedUsd)
		out = append(out, model.AbsorptionRankItem{
			Rank:          int64(i + 1 + offset),
			ID:            model.UUIDToString(r.LenderID),
			Name:          r.LenderName,
			ShortName:     textToString(r.LenderShortName),
			Type:          r.LenderType,
			PlannedUSD:    planned,
			RealizedUSD:   realized,
			AbsorptionPct: absorptionPct(planned, realized),
			VarianceUSD:   planned - realized,
			Status:        absorptionStatus(absorptionPct(planned, realized)),
			Drilldown: &model.DashboardDrilldownQuery{
				Target: "monitoring",
				Query:  map[string][]string{"lender_ids": {model.UUIDToString(r.LenderID)}},
			},
		})
	}
	return out
}

func (s *DashboardAnalyticsService) absorptionByInstitution(ctx context.Context, f model.DashboardAnalyticsFilter, limit, offset int) []model.AbsorptionRankItem {
	rows, _ := s.queries.GetDashboardAnalyticsAbsorptionByInstitution(ctx, queries.GetDashboardAnalyticsAbsorptionByInstitutionParams{
		BudgetYear:     nullableInt32(f.BudgetYear),
		Quarter:        nullableTextPtr(f.Quarter),
		LenderIds:      parseUUIDs(f.LenderIDs),
		LenderTypes:    f.LenderTypes,
		InstitutionIds: parseUUIDs(f.InstitutionIDs),
		Limit:          int32(limit),
		Offset:         int32(offset),
	})
	out := make([]model.AbsorptionRankItem, 0, len(rows))
	for i, r := range rows {
		planned := floatFromNumeric(r.PlannedUsd)
		realized := floatFromNumeric(r.RealizedUsd)
		out = append(out, model.AbsorptionRankItem{
			Rank:          int64(i + 1 + offset),
			ID:            model.UUIDToString(r.InstitutionID),
			Name:          r.InstitutionName,
			ShortName:     textToString(r.InstitutionShortName),
			PlannedUSD:    planned,
			RealizedUSD:   realized,
			AbsorptionPct: absorptionPct(planned, realized),
			VarianceUSD:   planned - realized,
			Status:        absorptionStatus(absorptionPct(planned, realized)),
		})
	}
	return out
}

func (s *DashboardAnalyticsService) absorptionByProject(ctx context.Context, f model.DashboardAnalyticsFilter, limit, offset int) []model.AbsorptionRankItem {
	rows, _ := s.queries.GetDashboardAnalyticsAbsorptionByProject(ctx, queries.GetDashboardAnalyticsAbsorptionByProjectParams{
		BudgetYear:  nullableInt32(f.BudgetYear),
		Quarter:     nullableTextPtr(f.Quarter),
		LenderIds:   parseUUIDs(f.LenderIDs),
		LenderTypes: f.LenderTypes,
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	out := make([]model.AbsorptionRankItem, 0, len(rows))
	for i, r := range rows {
		planned := floatFromNumeric(r.PlannedUsd)
		realized := floatFromNumeric(r.RealizedUsd)
		out = append(out, model.AbsorptionRankItem{
			Rank:          int64(i + 1 + offset),
			ID:            model.UUIDToString(r.DkProjectID),
			Name:          r.ProjectName,
			PlannedUSD:    planned,
			RealizedUSD:   realized,
			AbsorptionPct: absorptionPct(planned, realized),
			VarianceUSD:   planned - realized,
			Status:        absorptionStatus(absorptionPct(planned, realized)),
			Drilldown: &model.DashboardDrilldownQuery{
				Target: "monitoring",
				Query:  map[string][]string{"dk_project_id": {model.UUIDToString(r.DkProjectID)}},
			},
		})
	}
	return out
}

// ------ Yearly ------

func (s *DashboardAnalyticsService) GetYearly(ctx context.Context, f model.DashboardAnalyticsFilter, page, limit int) (*model.DashboardAnalyticsYearlyResponse, error) {
	if limit <= 0 { limit = 20 }
	if page <= 0 { page = 1 }
	offset := (page - 1) * limit

	rows, err := s.queries.GetDashboardAnalyticsYearly(ctx, queries.GetDashboardAnalyticsYearlyParams{
		BudgetYear:  nullableInt32(f.BudgetYear),
		Quarter:     nullableTextPtr(f.Quarter),
		LenderIds:   parseUUIDs(f.LenderIDs),
		LenderTypes: f.LenderTypes,
		Limit:       int32(limit),
		Offset:      int32(offset),
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil data yearly")
	}

	items := make([]model.YearlyItem, 0, len(rows))
	for _, r := range rows {
		planned := floatFromNumeric(r.PlannedUsd)
		realized := floatFromNumeric(r.RealizedUsd)
		items = append(items, model.YearlyItem{
			BudgetYear:         r.BudgetYear,
			Quarter:            r.Quarter,
			PlannedUSD:         planned,
			RealizedUSD:        realized,
			AbsorptionPct:      absorptionPct(planned, realized),
			LoanAgreementCount: r.LoanAgreementCount,
			ProjectCount:       r.ProjectCount,
		})
	}

	return &model.DashboardAnalyticsYearlyResponse{Items: items}, nil
}

// ------ Lender Proportion ------

func (s *DashboardAnalyticsService) GetLenderProportion(ctx context.Context, f model.DashboardAnalyticsFilter) (*model.DashboardAnalyticsLenderProportionResponse, error) {
	bbRows, _ := s.queries.GetDashboardAnalyticsLenderProportionBB(ctx, queries.GetDashboardAnalyticsLenderProportionBBParams{
		LenderTypes: f.LenderTypes,
		LenderIds:   parseUUIDs(f.LenderIDs),
	})
	gbRows, _ := s.queries.GetDashboardAnalyticsLenderProportionGB(ctx, queries.GetDashboardAnalyticsLenderProportionGBParams{
		LenderTypes: f.LenderTypes,
		LenderIds:   parseUUIDs(f.LenderIDs),
	})
	laRows, _ := s.queries.GetDashboardAnalyticsLenderProportionLA(ctx, queries.GetDashboardAnalyticsLenderProportionLAParams{
		LenderTypes: f.LenderTypes,
		LenderIds:   parseUUIDs(f.LenderIDs),
	})
	monRows, _ := s.queries.GetDashboardAnalyticsLenderProportionMonitoring(ctx, queries.GetDashboardAnalyticsLenderProportionMonitoringParams{
		BudgetYear:  nullableInt32(f.BudgetYear),
		Quarter:     nullableTextPtr(f.Quarter),
		LenderTypes: f.LenderTypes,
		LenderIds:   parseUUIDs(f.LenderIDs),
	})

	stages := []model.LenderProportionStage{
		{Stage: "Lender Indication", Items: buildProportionItems(bbRows, "amount")},
		{Stage: "Green Book Funding Source", Items: buildProportionItems(gbRows, "amount")},
		{Stage: "Loan Agreement", Items: buildProportionItems(laRows, "amount")},
		{Stage: "Monitoring Realization", Items: buildMonitoringProportionItems(monRows)},
	}
	return &model.DashboardAnalyticsLenderProportionResponse{ByStage: stages}, nil
}

// ------ Risks ------

func (s *DashboardAnalyticsService) GetRisks(ctx context.Context, f model.DashboardAnalyticsFilter, page, limit int) (*model.DashboardRisksResponse, error) {
	if limit <= 0 { limit = 20 }
	if page <= 0 { page = 1 }
	offset := (page - 1) * limit

	closingRows, _ := s.queries.GetDashboardAnalyticsClosingRisks(ctx, queries.GetDashboardAnalyticsClosingRisksParams{Limit: int32(limit), Offset: int32(offset)})
	closingCount, _ := s.queries.CountDashboardAnalyticsClosingRisks(ctx)
	closingItems := make([]model.RiskItem, 0, len(closingRows))
	for _, r := range closingRows {
		days := int32(r.DaysToClosing)
		closingItems = append(closingItems, model.RiskItem{
			LoanAgreement: map[string]interface{}{"id": model.UUIDToString(r.LoanAgreementID), "loan_code": r.LoanCode},
			ClosingDate:   fmtDate(r.ClosingDate),
			DaysToClosing: &days,
			AbsorptionPct: numericFromInterface(r.AbsorptionPct),
			Drilldown: &model.DashboardDrilldownQuery{
				Target: "monitoring",
				Query:  map[string][]string{"loan_agreement_id": {model.UUIDToString(r.LoanAgreementID)}},
			},
		})
	}

	effRows, _ := s.queries.GetDashboardAnalyticsEffectiveWithoutMonitoring(ctx, queries.GetDashboardAnalyticsEffectiveWithoutMonitoringParams{Limit: int32(limit), Offset: int32(offset)})
	effCount, _ := s.queries.CountDashboardAnalyticsEffectiveWithoutMonitoring(ctx)
	effItems := make([]model.RiskItem, 0, len(effRows))
	effAmount := float64(0)
	for _, r := range effRows {
		amt := floatFromNumeric(r.AmountUsd)
		effAmount += amt
		effItems = append(effItems, model.RiskItem{
			LoanAgreement: map[string]interface{}{"id": model.UUIDToString(r.LoanAgreementID), "loan_code": r.LoanCode},
			EffectiveDate: fmtDate(r.EffectiveDate),
			AmountUSD:     amt,
			Drilldown: &model.DashboardDrilldownQuery{
				Target: "monitoring",
				Query:  map[string][]string{"loan_agreement_id": {model.UUIDToString(r.LoanAgreementID)}},
			},
		})
	}

	extRows, _ := s.queries.GetDashboardAnalyticsExtendedLoans(ctx, queries.GetDashboardAnalyticsExtendedLoansParams{Limit: int32(limit), Offset: int32(offset)})
	extCount, _ := s.queries.CountDashboardAnalyticsExtendedLoans(ctx)
	extItems := make([]model.RiskItem, 0, len(extRows))
	for _, r := range extRows {
		days := int32(r.ExtensionDays)
		extItems = append(extItems, model.RiskItem{
			LoanAgreement: map[string]interface{}{"id": model.UUIDToString(r.LoanAgreementID), "loan_code": r.LoanCode},
			ExtensionDays: &days,
			Drilldown: &model.DashboardDrilldownQuery{
				Target: "la",
				Query:  map[string][]string{"id": {model.UUIDToString(r.LoanAgreementID)}},
			},
		})
	}

	dq, _ := s.queries.GetDashboardAnalyticsDataQualityCounts(ctx)

	return &model.DashboardRisksResponse{
		ClosingRisk:                model.RiskSection{Count: closingCount, Items: closingItems},
		EffectiveWithoutMonitoring: model.RiskSection{Count: effCount, AgreementAmountUSD: effAmount, Items: effItems},
		ExtendedLoans:              model.RiskSection{Count: extCount, Items: extItems},
		DataQuality:                model.DataQualitySection{MissingExecutingAgencyCount: dq.MissingExecutingAgencyCount, MissingLenderIndicationCount: dq.MissingLenderIndicationCount, ProjectWithoutGBCount: dq.ProjectWithoutGbCount},
	}, nil
}

// ------ helpers ------

func overviewParams(f model.DashboardAnalyticsFilter) queries.GetDashboardAnalyticsOverviewParams {
	return queries.GetDashboardAnalyticsOverviewParams{
		BudgetYear: nullableInt32(f.BudgetYear),
		Quarter:    nullableTextPtr(f.Quarter),
	}
}

func nullableInt32(v *int32) pgtype.Int4 {
	if v == nil { return pgtype.Int4{} }
	return pgtype.Int4{Int32: *v, Valid: true}
}

func parseUUIDs(values []string) []pgtype.UUID {
	if len(values) == 0 { return nil }
	out := make([]pgtype.UUID, 0, len(values))
	for _, v := range values {
		uid, err := model.ParseUUID(v)
		if err == nil { out = append(out, uid) }
	}
	return out
}

func textToString(v pgtype.Text) string {
	if !v.Valid { return "" }
	return v.String
}

func fmtDate(d pgtype.Date) string {
	if !d.Valid { return "" }
	return d.Time.Format("2006-01-02")
}

func numericFromInterface(v interface{}) float64 {
	switch val := v.(type) {
	case float64: return val
	case int64: return float64(val)
	case pgtype.Numeric: return floatFromNumeric(val)
	default: return 0
	}
}

func absorptionStatus(pct float64) string {
	if pct < 50 { return "low" }
	if pct < 90 { return "normal" }
	return "high"
}

func buildProportionItems[T any](rows []T, amountKind string) []model.LenderProportionItem {
	var totalAmount float64
	items := make([]model.LenderProportionItem, 0)
	type proportionRow interface {
		GetLenderType() string
		GetProjectCount() int64
		GetLenderCount() int64
		GetAmountUsd() pgtype.Numeric
	}
	_ = amountKind // used conceptually to distinguish stages
	for _, r := range rows {
		switch row := any(r).(type) {
		case queries.GetDashboardAnalyticsLenderProportionBBRow:
			amt := floatFromNumeric(row.AmountUsd)
			items = append(items, model.LenderProportionItem{
				Type: row.LenderType, ProjectCount: row.ProjectCount,
				LenderCount: row.LenderCount, AmountUSD: amt,
			})
			totalAmount += amt
		case queries.GetDashboardAnalyticsLenderProportionGBRow:
			amt := floatFromNumeric(row.AmountUsd)
			items = append(items, model.LenderProportionItem{
				Type: row.LenderType, ProjectCount: row.ProjectCount,
				LenderCount: row.LenderCount, AmountUSD: amt,
			})
			totalAmount += amt
		case queries.GetDashboardAnalyticsLenderProportionLARow:
			amt := floatFromNumeric(row.AmountUsd)
			items = append(items, model.LenderProportionItem{
				Type: row.LenderType, ProjectCount: row.ProjectCount,
				LenderCount: row.LenderCount, AmountUSD: amt,
			})
			totalAmount += amt
		}
	}
	for i := range items {
		if totalAmount > 0 { items[i].SharePct = items[i].AmountUSD / totalAmount * 100 }
	}
	return items
}

func buildMonitoringProportionItems(rows []queries.GetDashboardAnalyticsLenderProportionMonitoringRow) []model.LenderProportionItem {
	items := make([]model.LenderProportionItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, model.LenderProportionItem{
			Type:         r.LenderType,
			ProjectCount: r.ProjectCount,
			LenderCount:  r.LenderCount,
			PlannedUSD:   floatFromNumeric(r.PlannedUsd),
			RealizedUSD:  floatFromNumeric(r.RealizedUsd),
		})
	}
	return items
}
