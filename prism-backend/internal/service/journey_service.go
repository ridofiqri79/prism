package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/model"
)

type JourneyService struct {
	queries *queries.Queries
}

func NewJourneyService(queries *queries.Queries) *JourneyService {
	return &JourneyService{queries: queries}
}

func (s *JourneyService) GetProjectJourney(ctx context.Context, bbProjectID pgtype.UUID) (*model.JourneyResponse, error) {
	bb, err := s.queries.GetJourneyBBProject(ctx, bbProjectID)
	if err != nil {
		return nil, mapNotFound(err, "BB Project tidak ditemukan")
	}
	lenderIndications, err := s.queries.ListJourneyLenderIndicationsByBBProject(ctx, bbProjectID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil indikasi lender")
	}
	lois, err := s.queries.ListJourneyLoIsByBBProject(ctx, bbProjectID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil LoI")
	}
	gbRows, err := s.queries.ListJourneyGBProjectsByBBProject(ctx, bbProjectID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil Green Book Project")
	}

	gbIDs := collectJourneyGBProjectIDs(gbRows)
	fundingByGB, err := s.journeyFundingSourcesByGB(ctx, gbIDs)
	if err != nil {
		return nil, err
	}
	dkRowsByGB, dkIDs, err := s.journeyDKProjectsByGB(ctx, gbIDs)
	if err != nil {
		return nil, err
	}
	loanAgreementsByDK, err := s.journeyLoanAgreementsByDK(ctx, dkIDs)
	if err != nil {
		return nil, err
	}

	gbProjects := make([]model.JourneyGBProject, 0, len(gbRows))
	for _, gb := range gbRows {
		gbKey := uuidKey(gb.ID)
		dkRows := dkRowsByGB[gbKey]
		dkProjects := make([]model.JourneyDKProject, 0, len(dkRows))
		for _, dk := range dkRows {
			dkProjects = append(dkProjects, model.JourneyDKProject{
				ID:          model.UUIDToString(dk.ID),
				ProjectName: dk.ProjectName,
				Objectives:  stringPtrFromText(dk.Objectives),
				DaftarKegiatan: &model.JourneyDKHeader{
					ID:           model.UUIDToString(dk.DkID),
					Subject:      dk.DkSubject,
					Date:         dateString(dk.DkDate),
					LetterNumber: stringPtrFromText(dk.DkLetterNumber),
				},
				LoanAgreement: loanAgreementsByDK[uuidKey(dk.ID)],
			})
		}
		fundingSources := fundingByGB[gbKey]
		if fundingSources == nil {
			fundingSources = []model.JourneyFundingSource{}
		}
		gbProjects = append(gbProjects, model.JourneyGBProject{
			ID:                           model.UUIDToString(gb.ID),
			GreenBookID:                  model.UUIDToString(gb.GreenBookID),
			GBProjectIdentityID:          model.UUIDToString(gb.GbProjectIdentityID),
			GBCode:                       gb.GbCode,
			ProjectName:                  gb.ProjectName,
			Status:                       gb.Status,
			GreenBookRevisionLabel:       gb.GreenBookRevisionLabel,
			IsLatest:                     gb.IsLatest,
			HasNewerRevision:             gb.HasNewerRevision,
			LatestGBProjectID:            model.UUIDToString(gb.LatestGbProjectID),
			LatestGreenBookRevisionLabel: gb.LatestGreenBookRevisionLabel,
			FundingSources:               fundingSources,
			DKProjects:                   dkProjects,
		})
	}
	return &model.JourneyResponse{
		BBProject: model.JourneyBBProject{
			ID:                          model.UUIDToString(bb.ID),
			BlueBookID:                  model.UUIDToString(bb.BlueBookID),
			ProjectIdentityID:           model.UUIDToString(bb.ProjectIdentityID),
			BBCode:                      bb.BbCode,
			ProjectName:                 bb.ProjectName,
			BlueBookRevisionLabel:       bb.BlueBookRevisionLabel,
			IsLatest:                    bb.IsLatest,
			HasNewerRevision:            bb.HasNewerRevision,
			LatestBBProjectID:           model.UUIDToString(bb.LatestBbProjectID),
			LatestBlueBookRevisionLabel: bb.LatestBlueBookRevisionLabel,
			LenderIndications:           toJourneyLenderIndications(lenderIndications),
		},
		LoI:        toJourneyLoIs(lois),
		GBProjects: gbProjects,
	}, nil
}

func (s *JourneyService) journeyFundingSourcesByGB(ctx context.Context, gbIDs []pgtype.UUID) (map[string][]model.JourneyFundingSource, error) {
	data := map[string][]model.JourneyFundingSource{}
	if len(gbIDs) == 0 {
		return data, nil
	}
	rows, err := s.queries.ListJourneyFundingSourcesByGBProjects(ctx, gbIDs)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil funding source Green Book")
	}
	for _, row := range rows {
		key := uuidKey(row.GbProjectID)
		data[key] = append(data[key], model.JourneyFundingSource{
			ID:            model.UUIDToString(row.ID),
			Lender:        journeyLender(row.LenderID, row.LenderName, row.LenderShortName, row.LenderType),
			Institution:   journeyInstitution(row.InstitutionID, row.InstitutionName, row.InstitutionShortName),
			Currency:      row.Currency,
			LoanOriginal:  floatFromNumeric(row.LoanOriginal),
			GrantOriginal: floatFromNumeric(row.GrantOriginal),
			LocalOriginal: floatFromNumeric(row.LocalOriginal),
			LoanUSD:       floatFromNumeric(row.LoanUsd),
			GrantUSD:      floatFromNumeric(row.GrantUsd),
			LocalUSD:      floatFromNumeric(row.LocalUsd),
		})
	}
	return data, nil
}

func (s *JourneyService) journeyDKProjectsByGB(ctx context.Context, gbIDs []pgtype.UUID) (map[string][]queries.ListJourneyDKProjectsByGBProjectsRow, []pgtype.UUID, error) {
	data := map[string][]queries.ListJourneyDKProjectsByGBProjectsRow{}
	dkIDs := []pgtype.UUID{}
	seen := map[string]struct{}{}
	if len(gbIDs) == 0 {
		return data, dkIDs, nil
	}
	rows, err := s.queries.ListJourneyDKProjectsByGBProjects(ctx, gbIDs)
	if err != nil {
		return nil, nil, apperrors.Internal("Gagal mengambil DK Project")
	}
	for _, row := range rows {
		data[uuidKey(row.GbProjectID)] = append(data[uuidKey(row.GbProjectID)], row)
		dkIDs = appendUniqueJourneyUUID(dkIDs, seen, row.ID)
	}
	return data, dkIDs, nil
}

func (s *JourneyService) journeyLoanAgreementsByDK(ctx context.Context, dkIDs []pgtype.UUID) (map[string]*model.JourneyLoanAgreement, error) {
	data := map[string]*model.JourneyLoanAgreement{}
	if len(dkIDs) == 0 {
		return data, nil
	}
	rows, err := s.queries.ListJourneyLoanAgreementsByDKProjects(ctx, dkIDs)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil Loan Agreement")
	}
	loanIDs := []pgtype.UUID{}
	seen := map[string]struct{}{}
	for _, row := range rows {
		loanIDs = appendUniqueJourneyUUID(loanIDs, seen, row.ID)
	}
	monitoringByLA, err := s.journeyMonitoringByLA(ctx, loanIDs)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		data[uuidKey(row.DkProjectID)] = toJourneyLoanAgreement(row, monitoringByLA[uuidKey(row.ID)])
	}
	return data, nil
}

func (s *JourneyService) journeyMonitoringByLA(ctx context.Context, loanAgreementIDs []pgtype.UUID) (map[string][]model.JourneyMonitoringResponse, error) {
	data := map[string][]model.JourneyMonitoringResponse{}
	if len(loanAgreementIDs) == 0 {
		return data, nil
	}
	rows, err := s.queries.ListJourneyMonitoringByLAs(ctx, loanAgreementIDs)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil monitoring")
	}
	for _, row := range rows {
		planned := floatFromNumeric(row.PlannedUsd)
		realized := floatFromNumeric(row.RealizedUsd)
		key := uuidKey(row.LoanAgreementID)
		data[key] = append(data[key], model.JourneyMonitoringResponse{
			ID:            model.UUIDToString(row.ID),
			BudgetYear:    row.BudgetYear,
			Quarter:       row.Quarter,
			PlannedUSD:    planned,
			RealizedUSD:   realized,
			AbsorptionPct: absorptionPct(planned, realized),
		})
	}
	return data, nil
}

func toJourneyLenderIndications(rows []queries.ListJourneyLenderIndicationsByBBProjectRow) []model.JourneyLenderIndication {
	data := make([]model.JourneyLenderIndication, 0, len(rows))
	for _, row := range rows {
		data = append(data, model.JourneyLenderIndication{
			ID:      model.UUIDToString(row.ID),
			Lender:  journeyLender(row.LenderID, row.LenderName, row.LenderShortName, row.LenderType),
			Remarks: stringPtrFromText(row.Remarks),
		})
	}
	return data
}

func toJourneyLoIs(rows []queries.ListJourneyLoIsByBBProjectRow) []model.JourneyLoI {
	data := make([]model.JourneyLoI, 0, len(rows))
	for _, row := range rows {
		data = append(data, model.JourneyLoI{
			ID:           model.UUIDToString(row.ID),
			Lender:       journeyLender(row.LenderID, row.LenderName, row.LenderShortName, row.LenderType),
			Subject:      row.Subject,
			Date:         dateString(row.Date),
			LetterNumber: stringPtrFromText(row.LetterNumber),
		})
	}
	return data
}

func toJourneyLoanAgreement(row queries.ListJourneyLoanAgreementsByDKProjectsRow, monitoring []model.JourneyMonitoringResponse) *model.JourneyLoanAgreement {
	if monitoring == nil {
		monitoring = []model.JourneyMonitoringResponse{}
	}
	return &model.JourneyLoanAgreement{
		ID:                  model.UUIDToString(row.ID),
		LoanCode:            row.LoanCode,
		Lender:              journeyLender(row.LenderID, row.LenderName, row.LenderShortName, row.LenderType),
		AgreementDate:       dateString(row.AgreementDate),
		EffectiveDate:       dateString(row.EffectiveDate),
		OriginalClosingDate: dateString(row.OriginalClosingDate),
		ClosingDate:         dateString(row.ClosingDate),
		IsExtended:          isExtended(row.OriginalClosingDate, row.ClosingDate),
		ExtensionDays:       extensionDays(row.OriginalClosingDate, row.ClosingDate),
		Currency:            row.Currency,
		AmountOriginal:      floatFromNumeric(row.AmountOriginal),
		AmountUSD:           floatFromNumeric(row.AmountUsd),
		Monitoring:          monitoring,
	}
}

func journeyLender(id pgtype.UUID, name string, shortName pgtype.Text, lenderType string) model.LenderInfo {
	return model.LenderInfo{
		ID:        model.UUIDToString(id),
		Name:      name,
		ShortName: stringPtrFromText(shortName),
		Type:      lenderType,
	}
}

func journeyInstitution(id pgtype.UUID, name pgtype.Text, shortName pgtype.Text) *model.JourneyInstitutionInfo {
	if !id.Valid {
		return nil
	}
	return &model.JourneyInstitutionInfo{
		ID:        model.UUIDToString(id),
		Name:      name.String,
		ShortName: stringPtrFromText(shortName),
	}
}

func collectJourneyGBProjectIDs(rows []queries.ListJourneyGBProjectsByBBProjectRow) []pgtype.UUID {
	ids := make([]pgtype.UUID, 0, len(rows))
	seen := map[string]struct{}{}
	for _, row := range rows {
		ids = appendUniqueJourneyUUID(ids, seen, row.ID)
	}
	return ids
}

func appendUniqueJourneyUUID(ids []pgtype.UUID, seen map[string]struct{}, id pgtype.UUID) []pgtype.UUID {
	key := uuidKey(id)
	if _, ok := seen[key]; ok {
		return ids
	}
	seen[key] = struct{}{}
	return append(ids, id)
}

func uuidKey(id pgtype.UUID) string {
	return model.UUIDToString(id)
}
