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

func (s *GreenBookService) ListGreenBooks(ctx context.Context, filter model.GreenBookListFilter, params model.PaginationParams) (*model.ListResponse[model.GreenBookResponse], error) {
	page, limit, offset := normalizeList(params)
	queryParams, err := buildGreenBookListParams(filter, params, limit, offset)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListGreenBooks(ctx, queryParams)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar Green Book")
	}
	total, err := s.queries.CountGreenBooks(ctx, queries.CountGreenBooksParams{
		Search:       queryParams.Search,
		PublishYears: queryParams.PublishYears,
		Statuses:     queryParams.Statuses,
	})
	if err != nil {
		return nil, apperrors.Internal("Gagal menghitung Green Book")
	}
	data := make([]model.GreenBookResponse, 0, len(rows))
	for _, row := range rows {
		data = append(data, greenBookResponse(row))
	}
	return listResponse(data, page, limit, total), nil
}

func buildGreenBookListParams(filter model.GreenBookListFilter, params model.PaginationParams, limit, offset int) (queries.ListGreenBooksParams, error) {
	publishYears, err := int32Array(filter.PublishYears, "publish_year")
	if err != nil {
		return queries.ListGreenBooksParams{}, err
	}
	statuses, err := allowedValues(filter.Statuses, map[string]struct{}{"active": {}, "superseded": {}}, "status")
	if err != nil {
		return queries.ListGreenBooksParams{}, err
	}
	return queries.ListGreenBooksParams{
		Search:       nullableText(params.Search),
		PublishYears: publishYears,
		Statuses:     statuses,
		Limit:        int32(limit),
		Offset:       int32(offset),
	}, nil
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
	replacesID, err := parseOptionalUUID(req.ReplacesGreenBookID, "replaces_green_book_id")
	if err != nil {
		return nil, err
	}
	var created queries.GreenBook
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		sourceID := replacesID
		if !sourceID.Valid && req.RevisionNumber > 0 {
			active, err := qtx.GetActiveGreenBookByPublishYear(ctx, req.PublishYear)
			if err != nil && err != pgx.ErrNoRows {
				return err
			}
			if err == nil {
				sourceID = active.ID
			}
		}
		if sourceID.Valid {
			if _, err := qtx.GetGreenBook(ctx, sourceID); err != nil {
				return mapNotFound(err, "Green Book sumber revisi tidak ditemukan")
			}
		}
		if err := s.ensureGreenBookVersionAvailable(ctx, qtx, req.PublishYear, req.RevisionNumber, pgtype.UUID{}); err != nil {
			return err
		}
		if err := qtx.SupersedeGreenBooksByPublishYear(ctx, req.PublishYear); err != nil {
			return err
		}
		row, err := qtx.CreateGreenBook(ctx, queries.CreateGreenBookParams{
			PublishYear:         req.PublishYear,
			ReplacesGreenBookID: sourceID,
			RevisionNumber:      req.RevisionNumber,
		})
		if err != nil {
			return err
		}
		if sourceID.Valid {
			if err := s.cloneGreenBookProjects(ctx, qtx, sourceID, row.ID); err != nil {
				return err
			}
		}
		created = row
		return nil
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
		if err := s.ensureGreenBookVersionAvailable(ctx, qtx, req.PublishYear, req.RevisionNumber, id); err != nil {
			return err
		}
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

func (s *GreenBookService) ListGBProjects(ctx context.Context, gbID pgtype.UUID, filter model.GBProjectListFilter, params model.PaginationParams) (*model.ListResponse[model.GBProjectResponse], error) {
	page, limit, offset := normalizeList(params)
	queryParams, err := buildGBProjectListParams(gbID, filter, params, limit, offset)
	if err != nil {
		return nil, err
	}
	rows, err := s.queries.ListGBProjectsByGreenBook(ctx, queryParams)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil daftar GB Project")
	}
	total, err := s.queries.CountGBProjectsByGreenBook(ctx, queries.CountGBProjectsByGreenBookParams{
		GreenBookID:        queryParams.GreenBookID,
		Search:             queryParams.Search,
		BbProjectIds:       queryParams.BbProjectIds,
		ExecutingAgencyIds: queryParams.ExecutingAgencyIds,
		LocationIds:        queryParams.LocationIds,
		Statuses:           queryParams.Statuses,
	})
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

func buildGBProjectListParams(gbID pgtype.UUID, filter model.GBProjectListFilter, params model.PaginationParams, limit, offset int) (queries.ListGBProjectsByGreenBookParams, error) {
	bbProjectIDs, err := uuidArray(filter.BBProjectIDs, "bb_project_ids")
	if err != nil {
		return queries.ListGBProjectsByGreenBookParams{}, err
	}
	executingAgencyIDs, err := uuidArray(filter.ExecutingAgencyIDs, "executing_agency_ids")
	if err != nil {
		return queries.ListGBProjectsByGreenBookParams{}, err
	}
	locationIDs, err := uuidArray(filter.LocationIDs, "location_ids")
	if err != nil {
		return queries.ListGBProjectsByGreenBookParams{}, err
	}
	statuses, err := allowedValues(filter.Statuses, map[string]struct{}{"active": {}}, "status")
	if err != nil {
		return queries.ListGBProjectsByGreenBookParams{}, err
	}
	return queries.ListGBProjectsByGreenBookParams{
		GreenBookID:        gbID,
		Search:             nullableText(params.Search),
		BbProjectIds:       bbProjectIDs,
		ExecutingAgencyIds: executingAgencyIDs,
		LocationIds:        locationIDs,
		Statuses:           statuses,
		Limit:              int32(limit),
		Offset:             int32(offset),
	}, nil
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

	var created queries.GbProject
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		if _, err := qtx.GetGreenBook(ctx, gbID); err != nil {
			return mapNotFound(err, "Green Book tidak ditemukan")
		}
		if err := s.ensureGBCodeAvailable(ctx, qtx, gbID, req.GBCode); err != nil {
			return err
		}
		identityID, err := s.resolveGBProjectIdentity(ctx, qtx, req.GBProjectIdentityID)
		if err != nil {
			return err
		}
		project, err := qtx.CreateGBProject(ctx, queries.CreateGBProjectParams{
			GreenBookID:         gbID,
			GbProjectIdentityID: identityID,
			ProgramTitleID:      uuidOrInvalid(req.ProgramTitleID),
			GbCode:              strings.TrimSpace(req.GBCode),
			ProjectName:         strings.TrimSpace(req.ProjectName),
			Duration:            int4Ptr(req.Duration),
			Objective:           nullableTextPtr(req.Objective),
			ScopeOfProject:      nullableTextPtr(req.ScopeOfProject),
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
			Duration:       int4Ptr(req.Duration),
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

func (s *GreenBookService) DeleteGBProject(ctx context.Context, gbID, id pgtype.UUID, user *model.AuthUser) error {
	var deleted queries.GbProject
	if err := s.withTx(ctx, func(qtx *queries.Queries) error {
		if _, err := qtx.GetActiveGBProjectByGreenBook(ctx, queries.GetActiveGBProjectByGreenBookParams{GreenBookID: gbID, ID: id}); err != nil {
			return mapNotFound(err, "GB Project tidak ditemukan")
		}
		dependencies, err := qtx.ListGBProjectDeletionDependencies(ctx, id)
		if err != nil {
			return err
		}
		if len(dependencies) > 0 {
			return deletionBlockedError(user, "GB Project", gbProjectDeletionDependencies(dependencies))
		}
		row, err := qtx.HardDeleteGBProject(ctx, queries.HardDeleteGBProjectParams{GreenBookID: gbID, ID: id})
		if err != nil {
			if err == pgx.ErrNoRows {
				dependencies, depErr := qtx.ListGBProjectDeletionDependencies(ctx, id)
				if depErr == nil && len(dependencies) > 0 {
					return deletionBlockedError(user, "GB Project", gbProjectDeletionDependencies(dependencies))
				}
			}
			return mapNotFound(err, "GB Project tidak ditemukan")
		}
		deleted = row
		if err := qtx.DeleteOrphanGBProjectIdentity(ctx, deleted.GbProjectIdentityID); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	if s.broker != nil {
		s.broker.Publish("gb_project.deleted", map[string]string{"id": model.UUIDToString(deleted.ID)})
	}
	return nil
}

func gbProjectDeletionDependencies(rows []queries.ListGBProjectDeletionDependenciesRow) []deletionDependency {
	dependencies := make([]deletionDependency, 0, len(rows))
	for _, row := range rows {
		dependencies = append(dependencies, deletionDependency{
			relationType:  row.RelationType,
			relationID:    model.UUIDToString(row.RelationID),
			relationLabel: row.RelationLabel,
			relationPath:  row.RelationPath,
		})
	}
	return dependencies
}

func (s *GreenBookService) GetGBProjectHistory(ctx context.Context, id pgtype.UUID) ([]model.GBProjectHistoryItem, error) {
	if _, err := s.queries.GetGBProject(ctx, id); err != nil {
		return nil, mapNotFound(err, "GB Project tidak ditemukan")
	}
	rows, err := s.queries.ListGBProjectHistoryByProject(ctx, id)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil histori GB Project")
	}
	items := make([]model.GBProjectHistoryItem, 0, len(rows))
	for _, row := range rows {
		label := fmt.Sprintf("GB %d", row.PublishYear)
		if row.RevisionNumber > 0 {
			label = fmt.Sprintf("%s Revisi ke-%d", label, row.RevisionNumber)
		}
		bbRows, err := s.queries.GetGBProjectBBProjects(ctx, row.ID)
		if err != nil {
			return nil, apperrors.Internal("Gagal mengambil relasi BB Project")
		}
		bbProjects := make([]model.BBProjectSummary, 0, len(bbRows))
		for _, bb := range bbRows {
			bbProjects = append(bbProjects, s.bbProjectSummary(ctx, bb))
		}
		items = append(items, model.GBProjectHistoryItem{
			ID:                  model.UUIDToString(row.ID),
			GBProjectIdentityID: model.UUIDToString(row.GbProjectIdentityID),
			GreenBookID:         model.UUIDToString(row.GreenBookID),
			GBCode:              row.GbCode,
			ProjectName:         row.ProjectName,
			BookLabel:           label,
			PublishYear:         row.PublishYear,
			RevisionNumber:      row.RevisionNumber,
			BookStatus:          row.BookStatus,
			IsLatest:            row.IsLatest,
			UsedByDownstream:    row.UsedByDownstream,
			BBProjects:          bbProjects,
		})
	}
	return items, nil
}

func (s *GreenBookService) cloneGreenBookProjects(ctx context.Context, qtx *queries.Queries, sourceGreenBookID, targetGreenBookID pgtype.UUID) error {
	sourceProjects, err := qtx.ListGBProjectsForClone(ctx, sourceGreenBookID)
	if err != nil {
		return err
	}
	for _, source := range sourceProjects {
		cloned, err := qtx.CreateGBProject(ctx, queries.CreateGBProjectParams{
			GreenBookID:         targetGreenBookID,
			GbProjectIdentityID: source.GbProjectIdentityID,
			ProgramTitleID:      source.ProgramTitleID,
			GbCode:              source.GbCode,
			ProjectName:         source.ProjectName,
			Duration:            source.Duration,
			Objective:           source.Objective,
			ScopeOfProject:      source.ScopeOfProject,
		})
		if err != nil {
			return err
		}
		if err := qtx.CloneGBProjectBBProjectsWithLatestBB(ctx, queries.CloneGBProjectBBProjectsWithLatestBBParams{GbProjectID: source.ID, GbProjectID_2: cloned.ID}); err != nil {
			return err
		}
		if err := qtx.CloneGBProjectBappenasPartners(ctx, queries.CloneGBProjectBappenasPartnersParams{GbProjectID: source.ID, GbProjectID_2: cloned.ID}); err != nil {
			return err
		}
		if err := qtx.CloneGBProjectInstitutions(ctx, queries.CloneGBProjectInstitutionsParams{GbProjectID: source.ID, GbProjectID_2: cloned.ID}); err != nil {
			return err
		}
		if err := qtx.CloneGBProjectLocations(ctx, queries.CloneGBProjectLocationsParams{GbProjectID: source.ID, GbProjectID_2: cloned.ID}); err != nil {
			return err
		}
		if err := qtx.CloneGBFundingSources(ctx, queries.CloneGBFundingSourcesParams{GbProjectID: source.ID, GbProjectID_2: cloned.ID}); err != nil {
			return err
		}
		if err := qtx.CloneGBDisbursementPlans(ctx, queries.CloneGBDisbursementPlansParams{GbProjectID: source.ID, GbProjectID_2: cloned.ID}); err != nil {
			return err
		}
		activities, err := qtx.ListGBActivitiesByProject(ctx, source.ID)
		if err != nil {
			return err
		}
		for _, activity := range activities {
			clonedActivity, err := qtx.CloneGBActivity(ctx, queries.CloneGBActivityParams{ID: activity.ID, GbProjectID: cloned.ID})
			if err != nil {
				return err
			}
			if err := qtx.CloneGBFundingAllocation(ctx, queries.CloneGBFundingAllocationParams{GbActivityID: activity.ID, GbActivityID_2: clonedActivity.ID}); err != nil {
				return err
			}
		}
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
	if err := qtx.DeleteGBProjectBappenasPartners(ctx, projectID); err != nil {
		return err
	}
	if err := qtx.DeleteGBProjectInstitutions(ctx, projectID); err != nil {
		return err
	}
	if err := qtx.DeleteGBProjectLocations(ctx, projectID); err != nil {
		return err
	}

	var firstBlueBookID pgtype.UUID
	for _, id := range req.BBProjectIDs {
		bbProjectID, err := model.ParseUUID(id)
		if err != nil {
			return validation("bb_project_ids", "UUID tidak valid")
		}
		latestBB, err := qtx.GetLatestBBProjectByProject(ctx, bbProjectID)
		if err != nil {
			return mapNotFound(err, "BB Project tidak ditemukan")
		}
		if firstBlueBookID.Valid && model.UUIDToString(firstBlueBookID) != model.UUIDToString(latestBB.BlueBookID) {
			return validation("bb_project_ids", "Semua Proyek Blue Book harus berasal dari header Blue Book yang sama")
		}
		if !firstBlueBookID.Valid {
			firstBlueBookID = latestBB.BlueBookID
		}
		if err := qtx.AddGBProjectBBProject(ctx, queries.AddGBProjectBBProjectParams{GbProjectID: projectID, BbProjectID: latestBB.ID}); err != nil {
			return err
		}
	}
	if err := addProjectBappenasPartners(ctx, qtx, "bappenas_partner_ids", req.BappenasPartnerIDs, func(partnerID pgtype.UUID) error {
		return qtx.AddGBProjectBappenasPartner(ctx, queries.AddGBProjectBappenasPartnerParams{GbProjectID: projectID, BappenasPartnerID: partnerID})
	}); err != nil {
		return err
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
		currency := normalizeCurrency(item.Currency)
		if err := validateActiveCurrency(ctx, qtx, "funding_sources.currency", currency); err != nil {
			return err
		}
		loanOriginal, loanUSD := normalizeCurrencyAmountPair(currency, item.LoanOriginal, item.LoanUSD)
		grantOriginal, grantUSD := normalizeCurrencyAmountPair(currency, item.GrantOriginal, item.GrantUSD)
		localOriginal, localUSD := normalizeCurrencyAmountPair(currency, item.LocalOriginal, item.LocalUSD)
		if _, err := qtx.CreateGBFundingSource(ctx, queries.CreateGBFundingSourceParams{
			GbProjectID:   projectID,
			LenderID:      lenderID,
			InstitutionID: institutionID,
			Currency:      currency,
			LoanOriginal:  numericFromFloat(loanOriginal),
			GrantOriginal: numericFromFloat(grantOriginal),
			LocalOriginal: numericFromFloat(localOriginal),
			LoanUsd:       numericFromFloat(loanUSD),
			GrantUsd:      numericFromFloat(grantUSD),
			LocalUsd:      numericFromFloat(localUSD),
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

func (s *GreenBookService) ensureGBCodeAvailable(ctx context.Context, qtx *queries.Queries, gbID pgtype.UUID, code string) error {
	if strings.TrimSpace(code) == "" {
		return validation("gb_code", "wajib diisi")
	}
	_, err := qtx.GetGBProjectByGreenBookAndCode(ctx, queries.GetGBProjectByGreenBookAndCodeParams{GreenBookID: gbID, Lower: strings.TrimSpace(code)})
	if err == nil {
		return apperrors.Conflict("GB Code sudah digunakan dalam Green Book ini")
	}
	if err == pgx.ErrNoRows {
		return nil
	}
	return apperrors.Internal("Gagal memeriksa GB Code")
}

func (s *GreenBookService) ensureGreenBookVersionAvailable(ctx context.Context, qtx *queries.Queries, publishYear, revisionNumber int32, excludeID pgtype.UUID) error {
	var (
		count int64
		err   error
	)
	if excludeID.Valid {
		count, err = qtx.CountGreenBooksByPublishYearAndRevisionNumberExcept(ctx, queries.CountGreenBooksByPublishYearAndRevisionNumberExceptParams{
			PublishYear:    publishYear,
			RevisionNumber: revisionNumber,
			ID:             excludeID,
		})
	} else {
		count, err = qtx.CountGreenBooksByPublishYearAndRevisionNumber(ctx, queries.CountGreenBooksByPublishYearAndRevisionNumberParams{
			PublishYear:    publishYear,
			RevisionNumber: revisionNumber,
		})
	}
	if err != nil {
		return apperrors.Internal("Gagal memeriksa versi Green Book")
	}
	if count > 0 {
		return apperrors.Conflict("Green Book dengan Publish Year dan Revision number yang sama sudah ada")
	}
	return nil
}

func (s *GreenBookService) resolveGBProjectIdentity(ctx context.Context, qtx *queries.Queries, identity *string) (pgtype.UUID, error) {
	if identity == nil || strings.TrimSpace(*identity) == "" {
		row, err := qtx.CreateGBProjectIdentity(ctx)
		if err != nil {
			return pgtype.UUID{}, err
		}
		return row.ID, nil
	}
	identityID, err := model.ParseUUID(*identity)
	if err != nil {
		return pgtype.UUID{}, validation("gb_project_identity_id", "UUID tidak valid")
	}
	if _, err := qtx.GetGBProjectIdentity(ctx, identityID); err != nil {
		return pgtype.UUID{}, mapNotFound(err, "GB Project identity tidak ditemukan")
	}
	return identityID, nil
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
	partners, err := s.queries.GetGBProjectBappenasPartners(ctx, row.ID)
	if err != nil {
		return nil, apperrors.Internal("Gagal mengambil mitra Bappenas GB Project")
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
		ID:                  model.UUIDToString(row.ID),
		GreenBookID:         model.UUIDToString(row.GreenBookID),
		GBProjectIdentityID: model.UUIDToString(row.GbProjectIdentityID),
		ProgramTitleID:      stringPtrFromUUID(row.ProgramTitleID),
		GBCode:              row.GbCode,
		ProjectName:         row.ProjectName,
		Duration:            int32PtrFromInt4(row.Duration),
		Objective:           stringPtrFromText(row.Objective),
		ScopeOfProject:      stringPtrFromText(row.ScopeOfProject),
		BBProjects:          make([]model.BBProjectSummary, 0, len(bbProjects)),
		BappenasPartners:    make([]model.BappenasPartnerResponse, 0, len(partners)),
		Locations:           make([]model.RegionResponse, 0, len(locations)),
		Activities:          make([]model.GBActivityResponse, 0, len(activities)),
		FundingSources:      make([]model.GBFundingSourceResponse, 0, len(fundingSources)),
		DisbursementPlan:    make([]model.GBDisbursementPlanResponse, 0, len(disbursementPlans)),
		FundingAllocations:  make([]model.GBFundingAllocationResponse, 0, len(fundingAllocations)),
		Status:              row.Status,
		CreatedAt:           formatMasterTime(row.CreatedAt),
		UpdatedAt:           formatMasterTime(row.UpdatedAt),
	}
	latest, err := s.queries.GetLatestGBProjectByIdentity(ctx, row.GbProjectIdentityID)
	if err == nil {
		res.IsLatest = model.UUIDToString(latest.ID) == res.ID
		res.HasNewerRevision = !res.IsLatest
	} else if err == pgx.ErrNoRows {
		res.IsLatest = true
	} else {
		return nil, apperrors.Internal("Gagal memeriksa revisi GB Project")
	}
	for _, item := range bbProjects {
		res.BBProjects = append(res.BBProjects, s.bbProjectSummary(ctx, item))
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
	for _, item := range partners {
		res.BappenasPartners = append(res.BappenasPartners, toBappenasPartnerResponse(item))
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
	if req.Duration != nil && *req.Duration <= 0 {
		return validation("duration", "harus lebih dari 0 bulan")
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
	return model.GreenBookResponse{ID: model.UUIDToString(row.ID), PublishYear: row.PublishYear, ReplacesGreenBookID: stringPtrFromUUID(row.ReplacesGreenBookID), RevisionNumber: row.RevisionNumber, Status: row.Status, CreatedAt: formatMasterTime(row.CreatedAt), UpdatedAt: formatMasterTime(row.UpdatedAt)}
}

func (s *GreenBookService) bbProjectSummary(ctx context.Context, row queries.BbProject) model.BBProjectSummary {
	summary := model.BBProjectSummary{
		ID:                model.UUIDToString(row.ID),
		BlueBookID:        model.UUIDToString(row.BlueBookID),
		ProjectIdentityID: model.UUIDToString(row.ProjectIdentityID),
		BBCode:            row.BbCode,
		ProjectName:       row.ProjectName,
	}
	latest, err := s.queries.GetLatestBBProjectByIdentity(ctx, row.ProjectIdentityID)
	if err == nil {
		summary.IsLatest = model.UUIDToString(latest.ID) == summary.ID
		summary.HasNewerRevision = !summary.IsLatest
	}
	return summary
}

func gbFundingSourceResponse(row queries.ListGBFundingSourcesByProjectRow) model.GBFundingSourceResponse {
	res := model.GBFundingSourceResponse{
		ID:            model.UUIDToString(row.ID),
		Lender:        model.LenderInfo{ID: model.UUIDToString(row.LenderID), Name: row.LenderName, ShortName: stringPtrFromText(row.LenderShortName), Type: row.LenderType},
		Currency:      row.Currency,
		LoanOriginal:  floatFromNumeric(row.LoanOriginal),
		GrantOriginal: floatFromNumeric(row.GrantOriginal),
		LocalOriginal: floatFromNumeric(row.LocalOriginal),
		LoanUSD:       floatFromNumeric(row.LoanUsd),
		GrantUSD:      floatFromNumeric(row.GrantUsd),
		LocalUSD:      floatFromNumeric(row.LocalUsd),
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
