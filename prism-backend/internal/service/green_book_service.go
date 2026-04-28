package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ridofiqri79/prism-backend/internal/database/queries"
	apperrors "github.com/ridofiqri79/prism-backend/internal/errors"
	"github.com/ridofiqri79/prism-backend/internal/middleware"
	"github.com/ridofiqri79/prism-backend/internal/model"
	"github.com/ridofiqri79/prism-backend/internal/sse"
)

type GreenBookService struct {
	db      *pgxpool.Pool
	queries *queries.Queries
	broker  *sse.Broker
}

func NewGreenBookService(db *pgxpool.Pool, queries *queries.Queries, broker *sse.Broker) *GreenBookService {
	return &GreenBookService{db: db, queries: queries, broker: broker}
}

func (s *GreenBookService) ListGreenBooks(ctx context.Context, params model.PaginationParams) (*model.ListResponse[model.GreenBookResponse], error) {
	page, limit, offset := normalizeList(params)
	rows, err := s.queries.ListGreenBooks(ctx, queries.ListGreenBooksParams{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar Green Book")
	}
	total, err := s.queries.CountGreenBooks(ctx)
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung Green Book")
	}
	data := make([]model.GreenBookResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, greenBookResponse(row))
	}
	return listResponse(data, page, limit, total), nil
}

func (s *GreenBookService) GetGreenBook(ctx context.Context, id pgtype.UUID) (*model.GreenBookResponse, error) {
	row, err := s.queries.GetGreenBook(ctx, id)
	if err != nil {
		return nil, mapNotFound(err, "Green Book tidak ditemukan")
	}
	res := greenBookResponse(row)
	return &res, nil
}

func (s *GreenBookService) CreateGreenBook(ctx context.Context, req model.GreenBookRequest) (*model.GreenBookResponse, error) {
	if req.PublishYear <= 0 {
		return nil, validation("publish_year", "wajib diisi")
	}
	var created queries.GreenBook
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		if err := qtx.SupersedeGreenBooksByPublishYear(ctx, req.PublishYear); err != nil {
			return err
		}
		row, err := qtx.CreateGreenBook(ctx, queries.CreateGreenBookParams{
			PublishYear:    req.PublishYear,
			RevisionNumber: req.RevisionNumber,
		})
		created = row
		return err
	}); err != nil {
		return nil, err
	}
	return s.GetGreenBook(ctx, created.ID)
}

func (s *GreenBookService) UpdateGreenBook(ctx context.Context, id pgtype.UUID, req model.GreenBookRequest) (*model.GreenBookResponse, error) {
	if req.PublishYear <= 0 {
		return nil, validation("publish_year", "wajib diisi")
	}
	var updated queries.GreenBook
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		row, err := qtx.UpdateGreenBook(ctx, queries.UpdateGreenBookParams{
			ID:             id,
			PublishYear:    req.PublishYear,
			RevisionNumber: req.RevisionNumber,
		})
		if err != nil {
			return mapNotFound(err, "Green Book tidak ditemukan")
		}
		updated = row
		return nil
	}); err != nil {
		return nil, err
	}
	return s.GetGreenBook(ctx, updated.ID)
}

func (s *GreenBookService) DeleteGreenBook(ctx context.Context, id pgtype.UUID) error {
	return s.withTx(ctx, func(qtx *queries.Queries) error {
		if _, err := qtx.SupersedeGreenBook(ctx, id); err != nil {
			return mapNotFound(err, "Green Book tidak ditemukan")
		}
		return nil
	})
}

func (s *GreenBookService) ListGBProjects(ctx context.Context, gbID pgtype.UUID, params model.PaginationParams) (*model.ListResponse[model.GBProjectResponse], error) {
	page, limit, offset := normalizeList(params)
	rows, err := s.queries.ListGBProjectsByGreenBook(ctx, queries.ListGBProjectsByGreenBookParams{GreenBookID: gbID, Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar GB Project")
	}
	total, err := s.queries.CountGBProjectsByGreenBook(ctx, gbID)
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung GB Project")
	}
	data := make([]model.GBProjectResponse, 0, len(rows))
	for _, row := range rows {
		res, err := s.buildGBProjectResponse(ctx, row)
		if err != nil {
			return nil, err
		}
		data = append(data, *res)
	}
	return listResponse(data, page, limit, total), nil
}

func (s *GreenBookService) GetGBProject(ctx context.Context, gbID, id pgtype.UUID) (*model.GBProjectResponse, error) {
	row, err := s.queries.GetActiveGBProjectByGreenBook(ctx, queries.GetActiveGBProjectByGreenBookParams{GreenBookID: gbID, ID: id})
	if err != nil {
		return nil, mapNotFound(err, "GB Project tidak ditemukan")
	}
	return s.buildGBProjectResponse(ctx, row)
}

func (s *GreenBookService) CreateGBProject(ctx context.Context, gbID pgtype.UUID, req model.CreateGBProjectRequest) (*model.GBProjectResponse, error) {
	if err := validateGBProjectRequest(req, true); err != nil {
		return nil, err
	}
	if err := s.ensureGBCodeAvailable(ctx, req.GBCode); err != nil {
		return nil, err
	}

	var created queries.GbProject
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		if _, err := qtx.GetGreenBook(ctx, gbID); err != nil {
			return mapNotFound(err, "Green Book tidak ditemukan")
		}
		project, err := qtx.CreateGBProject(ctx, queries.CreateGBProjectParams{
			GreenBookID:    gbID,
			ProgramTitleID: uuidOrInvalid(req.ProgramTitleID),
			GbCode:         strings.TrimSpace(req.GBCode),
			ProjectName:    strings.TrimSpace(req.ProjectName),
			Duration:       nullableTextPtr(req.Duration),
			Objective:      nullableTextPtr(req.Objective),
			ScopeOfProject: nullableTextPtr(req.ScopeOfProject),
		})
		if err != nil {
			return err
		}
		if err := s.replaceGBProjectChildren(ctx, qtx, project.ID, req); err != nil {
			return err
		}
		created = project
		return nil
	}); err != nil {
		return nil, err
	}
	if s.broker != nil {
		s.broker.Publish("gb_project.created", map[string]string{"id": model.UUIDToString(created.ID)})
	}
	return s.buildGBProjectResponse(ctx, created)
}

func (s *GreenBookService) UpdateGBProject(ctx context.Context, gbID, id pgtype.UUID, req model.UpdateGBProjectRequest) (*model.GBProjectResponse, error) {
	if err := validateGBProjectRequest(req, false); err != nil {
		return nil, err
	}

	var updated queries.GbProject
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		if _, err := qtx.GetActiveGBProjectByGreenBook(ctx, queries.GetActiveGBProjectByGreenBookParams{GreenBookID: gbID, ID: id}); err != nil {
			return mapNotFound(err, "GB Project tidak ditemukan")
		}
		project, err := qtx.UpdateGBProject(ctx, queries.UpdateGBProjectParams{
			ID:             id,
			ProgramTitleID: uuidOrInvalid(req.ProgramTitleID),
			ProjectName:    strings.TrimSpace(req.ProjectName),
			Duration:       nullableTextPtr(req.Duration),
			Objective:      nullableTextPtr(req.Objective),
			ScopeOfProject: nullableTextPtr(req.ScopeOfProject),
		})
		if err != nil {
			return mapNotFound(err, "GB Project tidak ditemukan")
		}
		if err := s.replaceGBProjectChildren(ctx, qtx, id, req); err != nil {
			return err
		}
		updated = project
		return nil
	}); err != nil {
		return nil, err
	}
	if s.broker != nil {
		s.broker.Publish("gb_project.updated", map[string]string{"id": model.UUIDToString(updated.ID)})
	}
	return s.buildGBProjectResponse(ctx, updated)
}

func (s *GreenBookService) DeleteGBProject(ctx context.Context, gbID, id pgtype.UUID) error {
	var deleted queries.GbProject
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		if _, err := qtx.GetActiveGBProjectByGreenBook(ctx, queries.GetActiveGBProjectByGreenBookParams{GreenBookID: gbID, ID: id}); err != nil {
			return mapNotFound(err, "GB Project tidak ditemukan")
		}
		row, err := qtx.SoftDeleteGBProject(ctx, id)
		if err != nil {
			return mapNotFound(err, "GB Project tidak ditemukan")
		}
		deleted = row
		return nil
	}); err != nil {
		return err
	}
	if s.broker != nil {
		s.broker.Publish("gb_project.deleted", map[string]string{"id": model.UUIDToString(deleted.ID)})
	}
	return nil
}

func (s *GreenBookService) replaceGBProjectChildren(ctx context.Context, qtx *queries.Queries, projectID pgtype.UUID, req model.CreateGBProjectRequest) error {
	if err := qtx.DeleteGBFundingAllocationsByProject(ctx, projectID); err != nil {
		return err
	}
	if err := qtx.DeleteGBActivities(ctx, projectID); err != nil {
		return err
	}
	if err := qtx.DeleteGBFundingSources(ctx, projectID); err != nil {
		return err
	}
	if err := qtx.DeleteGBDisbursementPlans(ctx, projectID); err != nil {
		return err
	}
	if err := qtx.DeleteGBProjectBBProjects(ctx, projectID); err != nil {
		return err
	}
	if err := qtx.DeleteGBProjectInstitutions(ctx, projectID); err != nil {
		return err
	}
	if err := qtx.DeleteGBProjectLocations(ctx, projectID); err != nil {
		return err
	}

	for _, id := range req.BBProjectIDs {
		bbProjectID, err := model.ParseUUID(id)
		if err != nil {
			return validation("bb_project_ids", "UUID tidak valid")
		}
		if err := qtx.AddGBProjectBBProject(ctx, queries.AddGBProjectBBProjectParams{GbProjectID: projectID, BbProjectID: bbProjectID}); err != nil {
			return err
		}
	}
	for _, id := range req.ExecutingAgencyIDs {
		institutionID, err := model.ParseUUID(id)
		if err != nil {
			return validation("executing_agency_ids", "UUID tidak valid")
		}
		if err := qtx.AddGBProjectInstitution(ctx, queries.AddGBProjectInstitutionParams{GbProjectID: projectID, InstitutionID: institutionID, Role: roleExecutingAgency}); err != nil {
			return err
		}
	}
	for _, id := range req.ImplementingAgencyIDs {
		institutionID, err := model.ParseUUID(id)
		if err != nil {
			return validation("implementing_agency_ids", "UUID tidak valid")
		}
		if err := qtx.AddGBProjectInstitution(ctx, queries.AddGBProjectInstitutionParams{GbProjectID: projectID, InstitutionID: institutionID, Role: roleImplementingAgency}); err != nil {
			return err
		}
	}
	for _, id := range req.LocationIDs {
		regionID, err := model.ParseUUID(id)
		if err != nil {
			return validation("location_ids", "UUID tidak valid")
		}
		if err := qtx.AddGBProjectLocation(ctx, queries.AddGBProjectLocationParams{GbProjectID: projectID, RegionID: regionID}); err != nil {
			return err
		}
	}

	activityIDs := make([]pgtype.UUID, 0, len(req.Activities))
	for i, item := range req.Activities {
		sortOrder := int32(i)
		if item.SortOrder != nil {
			sortOrder = *item.SortOrder
		}
		row, err := qtx.CreateGBActivity(ctx, queries.CreateGBActivityParams{
			GbProjectID:            projectID,
			ActivityName:           strings.TrimSpace(item.ActivityName),
			ImplementationLocation: nullableTextPtr(item.ImplementationLocation),
			Piu:                    nullableTextPtr(item.PIU),
			SortOrder:              sortOrder,
		})
		if err != nil {
			return err
		}
		activityIDs = append(activityIDs, row.ID)
	}
	for _, item := range req.FundingSources {
		lenderID, err := model.ParseUUID(item.LenderID)
		if err != nil {
			return validation("funding_sources.lender_id", "UUID tidak valid")
		}
		institutionID, err := parseOptionalUUID(item.InstitutionID, "funding_sources.institution_id")
		if err != nil {
			return err
		}
		if _, err := qtx.CreateGBFundingSource(ctx, queries.CreateGBFundingSourceParams{
			GbProjectID:   projectID,
			LenderID:      lenderID,
			InstitutionID: institutionID,
			LoanUsd:       numericFromFloat(item.LoanUSD),
			GrantUsd:      numericFromFloat(item.GrantUSD),
			LocalUsd:      numericFromFloat(item.LocalUSD),
		}); err != nil {
			return err
		}
	}
	for _, item := range req.DisbursementPlan {
		if _, err := qtx.UpsertGBDisbursementPlan(ctx, queries.UpsertGBDisbursementPlanParams{
			GbProjectID: projectID,
			Year:        item.Year,
			AmountUsd:   numericFromFloat(item.AmountUSD),
		}); err != nil {
			return err
		}
	}
	for _, item := range req.FundingAllocations {
		if item.ActivityIndex < 0 || item.ActivityIndex >= len(activityIDs) {
			return validation("funding_allocations.activity_index", "activity_index tidak valid")
		}
		if _, err := qtx.CreateGBFundingAllocation(ctx, queries.CreateGBFundingAllocationParams{
			GbActivityID:  activityIDs[item.ActivityIndex],
			Services:      numericFromFloat(item.Services),
			Constructions: numericFromFloat(item.Constructions),
			Goods:         numericFromFloat(item.Goods),
			Trainings:     numericFromFloat(item.Trainings),
			Other:         numericFromFloat(item.Other),
		}); err != nil {
			return err
		}
	}
	return nil
}

func (s *GreenBookService) ensureGBCodeAvailable(ctx context.Context, code string) error {
	if strings.TrimSpace(code) == "" {
		return validation("gb_code", "wajib diisi")
	}
	_, err := s.queries.GetGBProjectByCode(ctx, strings.TrimSpace(code))
	if err == nil {
		return apperrors.Conflict("GB Code sudah digunakan")
	}
	if err == pgx.ErrNoRows {
		return nil
	}
	return apperrors.Internal("Gagal memeriksa GB Code")
}

func (s *GreenBookService) buildGBProjectResponse(ctx context.Context, row queries.GbProject) (*model.GBProjectResponse, error) {
	bbProjects, err := s.queries.GetGBProjectBBProjects(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil relasi BB Project")
	}
	institutions, err := s.queries.GetGBProjectInstitutions(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil institution GB Project")
	}
	locations, err := s.queries.GetGBProjectLocations(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil lokasi GB Project")
	}
	activities, err := s.queries.ListGBActivitiesByProject(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil activities GB Project")
	}
	fundingSources, err := s.queries.ListGBFundingSourcesByProject(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil funding source GB Project")
	}
	disbursementPlans, err := s.queries.ListGBDisbursementPlansByProject(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil disbursement plan GB Project")
	}
	fundingAllocations, err := s.queries.ListGBFundingAllocationsByProject(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil funding allocation GB Project")
	}

	res := model.GBProjectResponse{
		ID:                 model.UUIDToString(row.ID),
		GreenBookID:        model.UUIDToString(row.GreenBookID),
		ProgramTitleID:     stringPtrFromUUID(row.ProgramTitleID),
		GBCode:             row.GbCode,
		ProjectName:        row.ProjectName,
		Duration:           stringPtrFromText(row.Duration),
		Objective:          stringPtrFromText(row.Objective),
		ScopeOfProject:     stringPtrFromText(row.ScopeOfProject),
		BBProjects:         make([]model.BBProjectSummary, 0, len(bbProjects)),
		Locations:          make([]model.RegionResponse, 0, len(locations)),
		Activities:         make([]model.GBActivityResponse, 0, len(activities)),
		FundingSources:     make([]model.GBFundingSourceResponse, 0, len(fundingSources)),
		DisbursementPlan:   make([]model.GBDisbursementPlanResponse, 0, len(disbursementPlans)),
		FundingAllocations: make([]model.GBFundingAllocationResponse, 0, len(fundingAllocations)),
		Status:             row.Status,
		CreatedAt:          formatMasterTime(row.CreatedAt),
		UpdatedAt:          formatMasterTime(row.UpdatedAt),
	}
	for _, item := range bbProjects {
		res.BBProjects = append(res.BBProjects, model.BBProjectSummary{ID: model.UUIDToString(item.ID), BBCode: item.BbCode, ProjectName: item.ProjectName})
	}
	for _, item := range institutions {
		institution := model.InstitutionResponse{ID: model.UUIDToString(item.ID), ParentID: stringPtrFromUUID(item.ParentID), Name: item.Name, ShortName: stringPtrFromText(item.ShortName), Level: item.Level, CreatedAt: formatMasterTime(item.CreatedAt), UpdatedAt: formatMasterTime(item.UpdatedAt)}
		if item.Role == roleExecutingAgency {
			res.ExecutingAgencies = append(res.ExecutingAgencies, institution)
		}
		if item.Role == roleImplementingAgency {
			res.ImplementingAgencies = append(res.ImplementingAgencies, institution)
		}
	}
	for _, item := range locations {
		res.Locations = append(res.Locations, toRegionResponse(item))
	}
	for _, item := range activities {
		res.Activities = append(res.Activities, model.GBActivityResponse{ID: model.UUIDToString(item.ID), ActivityName: item.ActivityName, ImplementationLocation: stringPtrFromText(item.ImplementationLocation), PIU: stringPtrFromText(item.Piu), SortOrder: item.SortOrder})
	}
	for _, item := range fundingSources {
		res.FundingSources = append(res.FundingSources, gbFundingSourceResponse(item))
	}
	for _, item := range disbursementPlans {
		res.DisbursementPlan = append(res.DisbursementPlan, model.GBDisbursementPlanResponse{ID: model.UUIDToString(item.ID), Year: item.Year, AmountUSD: floatFromNumeric(item.AmountUsd)})
	}
	for _, item := range fundingAllocations {
		res.FundingAllocations = append(res.FundingAllocations, model.GBFundingAllocationResponse{ID: model.UUIDToString(item.ID), GBActivityID: model.UUIDToString(item.GbActivityID), ActivityName: item.ActivityName, SortOrder: item.SortOrder, Services: floatFromNumeric(item.Services), Constructions: floatFromNumeric(item.Constructions), Goods: floatFromNumeric(item.Goods), Trainings: floatFromNumeric(item.Trainings), Other: floatFromNumeric(item.Other)})
	}
	return &res, nil
}

func validateGBProjectRequest(req model.CreateGBProjectRequest, validateCode bool) error {
	if validateCode && strings.TrimSpace(req.GBCode) == "" {
		return validation("gb_code", "wajib diisi")
	}
	if strings.TrimSpace(req.ProjectName) == "" {
		return validation("project_name", "wajib diisi")
	}
	if len(req.BBProjectIDs) == 0 {
		return validation("bb_project_ids", "Minimal 1 BB Project")
	}
	if len(req.ExecutingAgencyIDs) == 0 {
		return validation("executing_agency_ids", "minimal satu")
	}
	if len(req.ImplementingAgencyIDs) == 0 {
		return validation("implementing_agency_ids", "minimal satu")
	}
	if len(req.LocationIDs) == 0 {
		return validation("location_ids", "minimal satu")
	}
	years := make(map[int32]struct{}, len(req.DisbursementPlan))
	for _, item := range req.DisbursementPlan {
		if _, exists := years[item.Year]; exists {
			return apperrors.BusinessRule(fmt.Sprintf("Tahun %d duplikat di disbursement plan", item.Year))
		}
		years[item.Year] = struct{}{}
	}
	for _, item := range req.Activities {
		if strings.TrimSpace(item.ActivityName) == "" {
			return validation("activities.activity_name", "wajib diisi")
		}
	}
	return nil
}

func greenBookResponse(row queries.GreenBook) model.GreenBookResponse {
	return model.GreenBookResponse{ID: model.UUIDToString(row.ID), PublishYear: row.PublishYear, RevisionNumber: row.RevisionNumber, Status: row.Status, CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)}
}

func gbFundingSourceResponse(row queries.ListGBFundingSourcesByProjectRow) model.GBFundingSourceResponse {
	res := model.GBFundingSourceResponse{
		ID:       model.UUIDToString(row.ID),
		Lender:   model.LenderInfo{ID: model.UUIDToString(row.LenderID), Name: row.LenderName, ShortName: stringPtrFromText(row.LenderShortName), Type: row.LenderType},
		LoanUSD:  floatFromNumeric(row.LoanUsd),
		GrantUSD: floatFromNumeric(row.GrantUsd),
		LocalUSD: floatFromNumeric(row.LocalUsd),
	}
	if row.InstitutionID.Valid {
		res.Institution = &model.InstitutionInfo{ID: model.UUIDToString(row.InstitutionID), Name: row.InstitutionName.String, ShortName: stringPtrFromText(row.InstitutionShortName), Level: row.InstitutionLevel.String}
	}
	return res
}

func (s *GreenBookService) withTx(ctx context.Context, fn func(*queries.Queries) error) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return apperrors.Internal("Gagal memulai transaksi")
	}
	defer tx.Rollback(ctx)
	if err := middleware.ApplyAuditUser(ctx, tx); err != nil {
		return apperrors.Internal("Gagal menyiapkan audit user")
	}
	if err := fn(s.queries.WithTx(tx)); err != nil {
		if _, ok := err.(*apperrors.AppError); ok {
			return err
		}
		return apperrors.FromPgError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return apperrors.Internal("Gagal menyimpan data")
	}
	return nil
}

func parseOptionalUUID(value *string, field string) (pgtype.UUID, error) {
	if value == nil || strings.TrimSpace(*value) == "" {
		return pgtype.UUID{}, nil
	}
	parsed, err := model.ParseUUID(*value)
	if err != nil {
		return pgtype.UUID{}, validation(field, "UUID tidak valid")
	}
	return parsed, nil
}
