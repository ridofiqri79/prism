package service

import (
	"context"

	"github.com/jackc/pgx/v5"
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
	lois, err := s.queries.ListJourneyLoIsByBBProject(ctx, bbProjectID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil LoI")
	}
	gbRows, err := s.queries.ListJourneyGBProjectsByBBProject(ctx, bbProjectID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil Green Book Project")
	}
	gbProjects := make([]model.JourneyGBProject, 0, len(gbRows))
	for _, gb := range gbRows {
		dkRows, err := s.queries.ListJourneyDKProjectsByGBProject(ctx, gb.ID)
		if err != nil {
			return nil, apperrors.Internal("Gagal mengambil DK Project")
		}
		dkProjects := make([]model.JourneyDKProject, 0, len(dkRows))
		for _, dk := range dkRows {
			loanAgreement, err := s.getJourneyLoanAgreement(ctx, dk.ID)
			if err != nil {
				return nil, err
			}
			dkProjects = append(dkProjects, model.JourneyDKProject{
				ID:          model.UUIDToString(dk.ID),
				ProjectName: dk.ProjectName,
				Objectives:  stringPtrFromText(dk.Objectives),
				DaftarKegiatan: &model.JourneyDKHeader{
					ID:      model.UUIDToString(dk.DkID),
					Subject: dk.DkSubject,
					Date:    dateString(dk.DkDate),
				},
				LoanAgreement: loanAgreement,
			})
		}
		gbProjects = append(gbProjects, model.JourneyGBProject{
			ID:                  model.UUIDToString(gb.ID),
			GreenBookID:         model.UUIDToString(gb.GreenBookID),
			GBProjectIdentityID: model.UUIDToString(gb.GbProjectIdentityID),
			GBCode:              gb.GbCode,
			ProjectName:         gb.ProjectName,
			Status:              gb.Status,
			HasNewerRevision:    gb.HasNewerRevision,
			DKProjects:          dkProjects,
		})
	}
	return &model.JourneyResponse{
		BBProject:  model.JourneyBBProject{ID: model.UUIDToString(bb.ID), BlueBookID: model.UUIDToString(bb.BlueBookID), ProjectIdentityID: model.UUIDToString(bb.ProjectIdentityID), BBCode: bb.BbCode, ProjectName: bb.ProjectName, HasNewerRevision: bb.HasNewerRevision},
		LoI:        toJourneyLoIs(lois),
		GBProjects: gbProjects,
	}, nil
}

func (s *JourneyService) getJourneyLoanAgreement(ctx context.Context, dkProjectID pgtype.UUID) (*model.JourneyLoanAgreement, error) {
	la, err := s.queries.GetJourneyLoanAgreementByDKProject(ctx, dkProjectID)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil Loan Agreement")
	}
	monitoringRows, err := s.queries.ListJourneyMonitoringByLA(ctx, la.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil monitoring")
	}
	return &model.JourneyLoanAgreement{
		ID:                  model.UUIDToString(la.ID),
		LoanCode:            la.LoanCode,
		Lender:              model.LenderInfo{ID: model.UUIDToString(la.LenderID), Name: la.LenderName, ShortName: stringPtrFromText(la.LenderShortName), Type: la.LenderType},
		EffectiveDate:       dateString(la.EffectiveDate),
		OriginalClosingDate: dateString(la.OriginalClosingDate),
		ClosingDate:         dateString(la.ClosingDate),
		IsExtended:          isExtended(la.OriginalClosingDate, la.ClosingDate),
		ExtensionDays:       extensionDays(la.OriginalClosingDate, la.ClosingDate),
		Monitoring:          toJourneyMonitoring(monitoringRows),
	}, nil
}

func toJourneyLoIs(rows []queries.ListJourneyLoIsByBBProjectRow) []model.JourneyLoI {
	data := make([]model.JourneyLoI, 0, len(rows))
	for _, row := range rows {
		data = append(data, model.JourneyLoI{
			ID:           model.UUIDToString(row.ID),
			Lender:       model.LenderInfo{ID: model.UUIDToString(row.LenderID), Name: row.LenderName, ShortName: stringPtrFromText(row.LenderShortName), Type: row.LenderType},
			Subject:      row.Subject,
			Date:         dateString(row.Date),
			LetterNumber: stringPtrFromText(row.LetterNumber),
		})
	}
	return data
}

func toJourneyMonitoring(rows []queries.ListJourneyMonitoringByLARow) []model.JourneyMonitoringResponse {
	data := make([]model.JourneyMonitoringResponse, 0, len(rows))
	for _, row := range rows {
		planned := floatFromNumeric(row.PlannedUsd)
		realized := floatFromNumeric(row.RealizedUsd)
		data = append(data, model.JourneyMonitoringResponse{
			ID:            model.UUIDToString(row.ID),
			BudgetYear:    row.BudgetYear,
			Quarter:       row.Quarter,
			PlannedUSD:    planned,
			RealizedUSD:   realized,
			AbsorptionPct: absorptionPct(planned, realized),
		})
	}
	return data
}
